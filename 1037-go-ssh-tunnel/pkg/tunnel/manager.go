package tunnel

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"ssh-tunnel/pkg/common"
)

type Forwarder interface {
	Start() error
	Stop()
}

type Tunnel struct {
	config    *common.TunnelConfig
	state     common.TunnelState
	conn      *SSHConnection
	forwarder Forwarder
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *log.Logger
	mu        sync.RWMutex
}

func NewTunnel(cfg *common.TunnelConfig) *Tunnel {
	ctx, cancel := context.WithCancel(context.Background())
	return &Tunnel{
		config: cfg,
		state: common.TunnelState{
			ID:        cfg.ID,
			Name:      cfg.Name,
			Status:    common.StatusStopped,
			StartedAt: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
		logger: log.New(os.Stdout, fmt.Sprintf("[tunnel-%s] ", cfg.ID), log.LstdFlags),
	}
}

func (t *Tunnel) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state.Status == common.StatusRunning || t.state.Status == common.StatusConnecting {
		return fmt.Errorf("tunnel already running")
	}

	t.state.Status = common.StatusConnecting
	t.state.StartedAt = time.Now()

	t.conn = NewSSHConnection(t.config)
	if t.config.Mode == common.ModeLocal {
		t.forwarder = NewLocalForward(t.config, t.conn)
	} else {
		t.forwarder = NewRemoteForward(t.config, t.conn)
	}

	go t.run()
	return nil
}

func (t *Tunnel) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state.Status == common.StatusStopped {
		return
	}

	t.cancel()

	if t.forwarder != nil {
		t.forwarder.Stop()
		t.forwarder = nil
	}

	if t.conn != nil {
		_ = t.conn.Close()
		t.conn = nil
	}

	t.state.Status = common.StatusStopped
	t.logger.Printf("隧道已停止")
}

func (t *Tunnel) GetState() common.TunnelState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

func (t *Tunnel) run() {
	interval := 1
	maxInterval := 10
	if t.config.MaxRetryInterval > 0 {
		maxInterval = t.config.MaxRetryInterval
	}

	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		t.mu.Lock()
		t.state.Status = common.StatusConnecting
		t.mu.Unlock()

		t.logger.Printf("正在建立SSH连接...")
		if err := t.conn.Connect(); err != nil {
			t.logger.Printf("SSH连接失败: %v", err)
			t.mu.Lock()
			t.state.Status = common.StatusReconnecting
			t.state.Error = err.Error()
			t.mu.Unlock()

			if !t.waitRetry(interval) {
				return
			}
			interval = nextInterval(interval, maxInterval)
			continue
		}

		t.mu.Lock()
		t.state.Status = common.StatusRunning
		t.state.Error = ""
		t.state.LastActive = time.Now()
		t.mu.Unlock()

		t.logger.Printf("开始启动转发器...")
		if err := t.forwarder.Start(); err != nil {
			t.logger.Printf("启动转发器失败: %v", err)
			_ = t.conn.Close()
			t.mu.Lock()
			t.state.Status = common.StatusReconnecting
			t.state.Error = err.Error()
			t.mu.Unlock()

			if !t.waitRetry(interval) {
				return
			}
			interval = nextInterval(interval, maxInterval)
			continue
		}

		t.logger.Printf("隧道已建立")
		interval = 1

		for {
			select {
			case <-t.ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}

			if !t.conn.IsConnected() {
				t.logger.Printf("检测到SSH连接断开，准备重连...")
				t.forwarder.Stop()
				_ = t.conn.Close()
				t.conn.CloseActiveConnections()
				t.mu.Lock()
				t.state.Status = common.StatusReconnecting
				t.mu.Unlock()
				break
			}

			t.mu.Lock()
			t.state.LastActive = time.Now()
			t.mu.Unlock()
		}

		if !t.waitRetry(interval) {
			return
		}
		interval = nextInterval(interval, maxInterval)
	}
}

func (t *Tunnel) waitRetry(seconds int) bool {
	t.logger.Printf("等待 %d 秒后重试...", seconds)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	remaining := seconds
	for remaining > 0 {
		select {
		case <-t.ctx.Done():
			return false
		case <-ticker.C:
			remaining--
		}
	}
	return true
}

func nextInterval(current, max int) int {
	next := current * 2
	if next > max {
		return max
	}
	return next
}

type Manager struct {
	tunnels map[string]*Tunnel
	mu      sync.RWMutex
	logger  *log.Logger
}

func NewManager() *Manager {
	return &Manager{
		tunnels: make(map[string]*Tunnel),
		logger:  log.New(os.Stdout, "[manager] ", log.LstdFlags),
	}
}

func (m *Manager) CreateTunnel(cfg *common.TunnelConfig) (*Tunnel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cfg.ID == "" {
		return nil, fmt.Errorf("tunnel ID is required")
	}

	if _, exists := m.tunnels[cfg.ID]; exists {
		return nil, fmt.Errorf("tunnel with ID %s already exists", cfg.ID)
	}

	tunnel := NewTunnel(cfg)
	m.tunnels[cfg.ID] = tunnel

	if err := tunnel.Start(); err != nil {
		delete(m.tunnels, cfg.ID)
		return nil, err
	}

	m.logger.Printf("隧道已创建并启动: %s (%s)", cfg.ID, cfg.Name)
	return tunnel, nil
}

func (m *Manager) StopTunnel(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tunnel, exists := m.tunnels[id]
	if !exists {
		return fmt.Errorf("tunnel with ID %s not found", id)
	}

	tunnel.Stop()
	delete(m.tunnels, id)
	m.logger.Printf("隧道已停止并移除: %s", id)
	return nil
}

func (m *Manager) ListTunnels() []common.TunnelState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make([]common.TunnelState, 0, len(m.tunnels))
	for _, tunnel := range m.tunnels {
		states = append(states, tunnel.GetState())
	}
	return states
}

func (m *Manager) GetTunnel(id string) (*Tunnel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tunnel, exists := m.tunnels[id]
	if !exists {
		return nil, fmt.Errorf("tunnel with ID %s not found", id)
	}
	return tunnel, nil
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, tunnel := range m.tunnels {
		tunnel.Stop()
		delete(m.tunnels, id)
	}
	m.logger.Printf("所有隧道已停止")
}

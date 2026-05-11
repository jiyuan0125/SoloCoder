package tunnel

import (
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
	logger    *log.Logger
	mu        sync.RWMutex
}

func NewTunnel(cfg *common.TunnelConfig) *Tunnel {
	return &Tunnel{
		config: cfg,
		state: common.TunnelState{
			ID:        cfg.ID,
			Name:      cfg.Name,
			Status:    common.StatusStopped,
			StartedAt: time.Now(),
		},
		logger: log.New(os.Stdout, fmt.Sprintf("[tunnel-%s] ", cfg.ID), log.LstdFlags),
	}
}

func (t *Tunnel) updateStateStatus(status common.TunnelStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.state.Status = status
	if status == common.StatusRunning {
		t.state.LastActive = time.Now()
	}
}

func (t *Tunnel) updateStateError(err string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.state.Error = err
}

func (t *Tunnel) clearStateError() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.state.Error = ""
}

func (t *Tunnel) startForwarder(conn *SSHConnection) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.conn = conn
	if t.config.Mode == common.ModeLocal {
		t.forwarder = NewLocalForward(t.config, conn)
	} else {
		t.forwarder = NewRemoteForward(t.config, conn)
	}

	t.logger.Printf("正在启动转发器...")
	return t.forwarder.Start()
}

func (t *Tunnel) stopForwarder() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.forwarder != nil {
		t.forwarder.Stop()
		t.forwarder = nil
	}
	t.conn = nil
}

func (t *Tunnel) GetState() common.TunnelState {
	t.mu.RLock()
	defer t.mu.RUnlock()
	state := t.state
	return state
}

func (t *Tunnel) GetConfig() *common.TunnelConfig {
	return t.config
}

func nextInterval(current, max int) int {
	next := current * 2
	if next > max {
		return max
	}
	return next
}

type Manager struct {
	tunnels  map[string]*Tunnel
	handlers map[ConnectionKey]*SharedConnectionHandler
	pool     *ConnectionPool
	mu       sync.RWMutex
	logger   *log.Logger
}

func NewManager() *Manager {
	return &Manager{
		tunnels:  make(map[string]*Tunnel),
		handlers: make(map[ConnectionKey]*SharedConnectionHandler),
		pool:     NewConnectionPool(),
		logger:   log.New(os.Stdout, "[manager] ", log.LstdFlags),
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

	key := keyFromConfig(cfg)

	tunnel := NewTunnel(cfg)
	m.tunnels[cfg.ID] = tunnel

	handler, exists := m.handlers[key]
	if !exists {
		pooled, err := m.pool.GetOrCreate(cfg)
		if err != nil {
			delete(m.tunnels, cfg.ID)
			return nil, err
		}

		handler = NewSharedConnectionHandler(pooled, cfg.MaxRetryInterval)
		m.handlers[key] = handler
		handler.Start()
		m.logger.Printf("为 %s@%s:%d 创建新的共享连接处理器", key.User, key.Server, key.Port)
	}

	handler.AddTunnel(tunnel)
	tunnel.updateStateStatus(common.StatusConnecting)

	if handler.GetStatus() == SharedStatusRunning {
		if err := tunnel.startForwarder(handler.pooled.GetConn()); err != nil {
			tunnel.updateStateError(err.Error())
		} else {
			tunnel.updateStateStatus(common.StatusRunning)
		}
	}

	m.logger.Printf("隧道已创建并加入共享连接: %s (%s)", cfg.ID, cfg.Name)
	return tunnel, nil
}

func (m *Manager) StopTunnel(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tunnel, exists := m.tunnels[id]
	if !exists {
		return fmt.Errorf("tunnel with ID %s not found", id)
	}

	key := keyFromConfig(tunnel.config)

	handler, exists := m.handlers[key]
	if exists {
		handler.RemoveTunnel(id)
		if handler.TunnelCount() == 0 {
			m.logger.Printf("共享连接处理器没有隧道了，停止它: %s@%s:%d", key.User, key.Server, key.Port)
			handler.Stop()
			delete(m.handlers, key)
			_ = m.pool.Release(handler.pooled)
		}
	}

	tunnel.stopForwarder()
	tunnel.updateStateStatus(common.StatusStopped)
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

	for key, handler := range m.handlers {
		m.logger.Printf("停止共享连接处理器: %s@%s:%d", key.User, key.Server, key.Port)
		handler.Stop()
		delete(m.handlers, key)
	}

	for id, tunnel := range m.tunnels {
		tunnel.stopForwarder()
		tunnel.updateStateStatus(common.StatusStopped)
		delete(m.tunnels, id)
	}

	m.pool.CloseAll()
	m.logger.Printf("所有隧道已停止")
}

func (m *Manager) ConnectionCount() int {
	return m.pool.Size()
}

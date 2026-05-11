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

type ConnectionKey struct {
	Server     string
	Port       int
	User       string
	AuthMethod common.AuthMethod
	Password   string
	KeyFile    string
}

func keyFromConfig(cfg *common.TunnelConfig) ConnectionKey {
	return ConnectionKey{
		Server:     cfg.SSHServer,
		Port:       cfg.SSHPort,
		User:       cfg.SSHUser,
		AuthMethod: cfg.AuthMethod,
		Password:   cfg.Password,
		KeyFile:    cfg.KeyFilePath,
	}
}

type PooledSSHConnection struct {
	conn       *SSHConnection
	refCount   int
	key        ConnectionKey
	config     *common.TunnelConfig
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	onDisconnect func()
	active     bool
}

func NewPooledSSHConnection(key ConnectionKey, cfg *common.TunnelConfig) *PooledSSHConnection {
	ctx, cancel := context.WithCancel(context.Background())
	connCfg := *cfg
	return &PooledSSHConnection{
		conn:     NewSSHConnection(&connCfg),
		refCount: 0,
		key:      key,
		config:   &connCfg,
		ctx:      ctx,
		cancel:   cancel,
		active:   true,
	}
}

func (p *PooledSSHConnection) AddRef() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.refCount++
}

func (p *PooledSSHConnection) Release() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.refCount--
	return p.refCount
}

func (p *PooledSSHConnection) RefCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.refCount
}

func (p *PooledSSHConnection) Connect() error {
	return p.conn.Connect()
}

func (p *PooledSSHConnection) Close() error {
	p.cancel()
	p.active = false
	return p.conn.Close()
}

func (p *PooledSSHConnection) IsConnected() bool {
	return p.conn.IsConnected()
}

func (p *PooledSSHConnection) GetConn() *SSHConnection {
	return p.conn
}

func (p *PooledSSHConnection) SetOnDisconnect(fn func()) {
	p.onDisconnect = fn
}

type ConnectionPool struct {
	connections map[ConnectionKey]*PooledSSHConnection
	mu          sync.RWMutex
	logger      *log.Logger
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[ConnectionKey]*PooledSSHConnection),
		logger:      log.New(os.Stdout, "[conn-pool] ", log.LstdFlags),
	}
}

func (p *ConnectionPool) GetOrCreate(cfg *common.TunnelConfig) (*PooledSSHConnection, error) {
	key := keyFromConfig(cfg)

	p.mu.Lock()
	defer p.mu.Unlock()

	if pooled, exists := p.connections[key]; exists && pooled.active {
		pooled.AddRef()
		p.logger.Printf("复用现有连接: %s@%s:%d (引用计数: %d)", key.User, key.Server, key.Port, pooled.RefCount())
		return pooled, nil
	}

	pooled := NewPooledSSHConnection(key, cfg)
	pooled.AddRef()
	p.connections[key] = pooled

	p.logger.Printf("创建新连接: %s@%s:%d", key.User, key.Server, key.Port)
	return pooled, nil
}

func (p *ConnectionPool) Release(pooled *PooledSSHConnection) error {
	if pooled == nil {
		return nil
	}

	remaining := pooled.Release()
	p.logger.Printf("释放连接: %s@%s:%d (剩余引用: %d)", pooled.key.User, pooled.key.Server, pooled.key.Port, remaining)

	if remaining <= 0 {
		p.mu.Lock()
		defer p.mu.Unlock()

		if remaining <= 0 {
			key := pooled.key
			if existing, exists := p.connections[key]; exists && existing == pooled {
				delete(p.connections, key)
				p.logger.Printf("关闭空闲连接: %s@%s:%d", key.User, key.Server, key.Port)
				return pooled.Close()
			}
		}
	}
	return nil
}

func (p *ConnectionPool) ForceClose(key ConnectionKey) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if pooled, exists := p.connections[key]; exists {
		delete(p.connections, key)
		p.logger.Printf("强制关闭连接: %s@%s:%d", key.User, key.Server, key.Port)
		return pooled.Close()
	}
	return nil
}

func (p *ConnectionPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for key, pooled := range p.connections {
		p.logger.Printf("关闭连接: %s@%s:%d", key.User, key.Server, key.Port)
		_ = pooled.Close()
		delete(p.connections, key)
	}
	p.logger.Printf("所有连接已关闭")
}

func (p *ConnectionPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.connections)
}

type SharedTunnelStatus string

const (
	SharedStatusConnecting   SharedTunnelStatus = "connecting"
	SharedStatusRunning      SharedTunnelStatus = "running"
	SharedStatusReconnecting SharedTunnelStatus = "reconnecting"
)

type SharedConnectionHandler struct {
	pooled    *PooledSSHConnection
	status    SharedTunnelStatus
	mu        sync.RWMutex
	tunnels   map[string]*Tunnel
	logger    *log.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	interval  int
	maxRetry  int
}

func NewSharedConnectionHandler(pooled *PooledSSHConnection, maxRetryInterval int) *SharedConnectionHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &SharedConnectionHandler{
		pooled:   pooled,
		status:   SharedStatusConnecting,
		tunnels:  make(map[string]*Tunnel),
		logger:   log.New(os.Stdout, fmt.Sprintf("[shared-%s:%d] ", pooled.key.Server, pooled.key.Port), log.LstdFlags),
		ctx:      ctx,
		cancel:   cancel,
		interval: 1,
		maxRetry: maxRetryInterval,
	}
}

func (h *SharedConnectionHandler) AddTunnel(t *Tunnel) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.tunnels[t.config.ID] = t
}

func (h *SharedConnectionHandler) RemoveTunnel(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.tunnels, id)
}

func (h *SharedConnectionHandler) TunnelCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.tunnels)
}

func (h *SharedConnectionHandler) GetTunnels() []*Tunnel {
	h.mu.RLock()
	defer h.mu.RUnlock()
	tunnels := make([]*Tunnel, 0, len(h.tunnels))
	for _, t := range h.tunnels {
		tunnels = append(tunnels, t)
	}
	return tunnels
}

func (h *SharedConnectionHandler) GetStatus() SharedTunnelStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

func (h *SharedConnectionHandler) Start() {
	h.wg.Add(1)
	go h.run()
}

func (h *SharedConnectionHandler) Stop() {
	h.cancel()
	h.wg.Wait()
}

func (h *SharedConnectionHandler) run() {
	defer h.wg.Done()

	maxInterval := 60
	if h.maxRetry > 0 {
		maxInterval = h.maxRetry
	}
	interval := 1

	for {
		select {
		case <-h.ctx.Done():
			return
		default:
		}

		h.mu.Lock()
		h.status = SharedStatusConnecting
		h.mu.Unlock()
		h.notifyStatus(common.StatusConnecting, "")

		h.logger.Printf("正在建立共享SSH连接...")
		if err := h.pooled.Connect(); err != nil {
			h.logger.Printf("共享SSH连接失败: %v", err)
			h.mu.Lock()
			h.status = SharedStatusReconnecting
			h.mu.Unlock()
			h.notifyStatus(common.StatusReconnecting, err.Error())

			if !h.waitRetry(interval) {
				return
			}
			interval = nextInterval(interval, maxInterval)
			continue
		}

		h.mu.Lock()
		h.status = SharedStatusRunning
		h.mu.Unlock()
		h.notifyStatus(common.StatusRunning, "")
		interval = 1

		h.logger.Printf("共享SSH连接已建立，正在启动所有隧道转发器...")
		h.startAllForwarders()

		for {
			select {
			case <-h.ctx.Done():
				h.stopAllForwarders()
				return
			case <-time.After(1 * time.Second):
			}

			if !h.pooled.IsConnected() {
				h.logger.Printf("检测到共享SSH连接断开，准备重连...")
				h.mu.Lock()
				h.status = SharedStatusReconnecting
				h.mu.Unlock()
				h.notifyStatus(common.StatusReconnecting, "connection lost, reconnecting...")
				h.stopAllForwarders()
				h.pooled.GetConn().CloseActiveConnections()
				break
			}
		}

		if !h.waitRetry(interval) {
			return
		}
		interval = nextInterval(interval, maxInterval)
	}
}

func (h *SharedConnectionHandler) startAllForwarders() {
	h.mu.RLock()
	tunnels := make([]*Tunnel, 0, len(h.tunnels))
	for _, t := range h.tunnels {
		tunnels = append(tunnels, t)
	}
	h.mu.RUnlock()

	for _, t := range tunnels {
		if err := t.startForwarder(h.pooled.GetConn()); err != nil {
			h.logger.Printf("隧道 %s 启动转发器失败: %v", t.config.ID, err)
			t.updateStateError(err.Error())
		} else {
			t.updateStateStatus(common.StatusRunning)
		}
	}
}

func (h *SharedConnectionHandler) stopAllForwarders() {
	h.mu.RLock()
	tunnels := make([]*Tunnel, 0, len(h.tunnels))
	for _, t := range h.tunnels {
		tunnels = append(tunnels, t)
	}
	h.mu.RUnlock()

	for _, t := range tunnels {
		t.stopForwarder()
		t.updateStateStatus(common.StatusReconnecting)
	}
}

func (h *SharedConnectionHandler) notifyStatus(status common.TunnelStatus, errMsg string) {
	h.mu.RLock()
	tunnels := make([]*Tunnel, 0, len(h.tunnels))
	for _, t := range h.tunnels {
		tunnels = append(tunnels, t)
	}
	h.mu.RUnlock()

	for _, t := range tunnels {
		t.updateStateStatus(status)
		if errMsg != "" {
			t.updateStateError(errMsg)
		} else {
			t.clearStateError()
		}
	}
}

func (h *SharedConnectionHandler) waitRetry(seconds int) bool {
	h.logger.Printf("等待 %d 秒后重试...", seconds)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	remaining := seconds
	for remaining > 0 {
		select {
		case <-h.ctx.Done():
			return false
		case <-ticker.C:
			remaining--
		}
	}
	return true
}

package pool

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"connpool/internal/models"
)

var (
	ErrPoolNotFound = errors.New("pool not found")
	ErrPoolInUse    = errors.New("pool is in use")
	ErrWaitTimeout  = errors.New("wait timeout")
	ErrInvalidMax   = errors.New("max connections must be positive")
	ErrConnNotFound = errors.New("connection not found")
)

type poolConn struct {
	id        string
	netConn   net.Conn
	createdAt time.Time
	lastUsed  time.Time
	healthy   bool
}

type ConnectionPool struct {
	config     *models.PoolConfig
	mu         sync.Mutex
	active     map[string]*poolConn
	idle       []*poolConn
	waiting    []chan struct{}
	connCount  int
	stopChan   chan struct{}
	closed     bool
}

type Manager struct {
	pools   map[string]*ConnectionPool
	mu      sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		pools: make(map[string]*ConnectionPool),
	}
}

func (m *Manager) CreatePool(config *models.PoolConfig) error {
	if config.MaxConnections <= 0 {
		return ErrInvalidMax
	}

	if config.WaitTimeout == 0 {
		config.WaitTimeout = 30 * time.Second
	}

	if config.IdleTimeout == 0 {
		config.IdleTimeout = 5 * time.Minute
	}

	if config.MaxLifetime == 0 {
		config.MaxLifetime = 1 * time.Hour
	}

	pool := &ConnectionPool{
		config:   config,
		active:   make(map[string]*poolConn),
		idle:     make([]*poolConn, 0),
		waiting:  make([]chan struct{}, 0),
		stopChan: make(chan struct{}),
	}

	m.mu.Lock()
	m.pools[config.ID] = pool
	m.mu.Unlock()

	go pool.startCleanup()

	return nil
}

func (m *Manager) GetPool(id string) (*ConnectionPool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pool, exists := m.pools[id]
	if !exists {
		return nil, ErrPoolNotFound
	}
	return pool, nil
}

func (m *Manager) DeletePool(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pool, exists := m.pools[id]
	if !exists {
		return ErrPoolNotFound
	}

	pool.mu.Lock()
	if len(pool.active) > 0 || len(pool.waiting) > 0 {
		pool.mu.Unlock()
		return ErrPoolInUse
	}
	pool.mu.Unlock()

	close(pool.stopChan)
	pool.closed = true

	for _, pc := range pool.idle {
		if pc.netConn != nil {
			pc.netConn.Close()
		}
	}

	delete(m.pools, id)
	return nil
}

func (m *Manager) ListPoolIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.pools))
	for id := range m.pools {
		ids = append(ids, id)
	}
	return ids
}

func (p *ConnectionPool) GetStatus() *models.PoolStatus {
	p.mu.Lock()
	defer p.mu.Unlock()

	return &models.PoolStatus{
		PoolID:       p.config.ID,
		ActiveCount:  len(p.active),
		IdleCount:    len(p.idle),
		WaitingCount: len(p.waiting),
		MaxCount:     p.config.MaxConnections,
	}
}

func (p *ConnectionPool) Acquire(ctx context.Context) (string, error) {
	for {
		p.mu.Lock()

		for i := len(p.idle) - 1; i >= 0; i-- {
			pc := p.idle[i]
			if p.isValid(pc) {
				p.idle = append(p.idle[:i], p.idle[i+1:]...)
				pc.lastUsed = time.Now()
				p.active[pc.id] = pc
				p.mu.Unlock()
				return pc.id, nil
			} else {
				p.idle = append(p.idle[:i], p.idle[i+1:]...)
				if pc.netConn != nil {
					pc.netConn.Close()
				}
				p.connCount--
			}
		}

		if p.connCount < p.config.MaxConnections {
			pc, err := p.createConn()
			if err != nil {
				p.mu.Unlock()
				return "", err
			}
			p.active[pc.id] = pc
			p.connCount++
			p.mu.Unlock()
			return pc.id, nil
		}

		waitChan := make(chan struct{}, 1)
		p.waiting = append(p.waiting, waitChan)
		p.mu.Unlock()

		timeout := p.config.WaitTimeout
		if deadline, ok := ctx.Deadline(); ok {
			remaining := time.Until(deadline)
			if remaining < timeout {
				timeout = remaining
			}
		}

		select {
		case <-waitChan:
			continue
		case <-time.After(timeout):
			p.mu.Lock()
			for i, ch := range p.waiting {
				if ch == waitChan {
					p.waiting = append(p.waiting[:i], p.waiting[i+1:]...)
					break
				}
			}
			p.mu.Unlock()
			return "", ErrWaitTimeout
		case <-ctx.Done():
			p.mu.Lock()
			for i, ch := range p.waiting {
				if ch == waitChan {
					p.waiting = append(p.waiting[:i], p.waiting[i+1:]...)
					break
				}
			}
			p.mu.Unlock()
			return "", ctx.Err()
		}
	}
}

func (p *ConnectionPool) Release(connID string, healthy bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	pc, exists := p.active[connID]
	if !exists {
		return ErrConnNotFound
	}

	delete(p.active, connID)

	if !healthy {
		if pc.netConn != nil {
			pc.netConn.Close()
		}
		p.connCount--
		p.notifyWaiter()
		return nil
	}

	if !p.isValid(pc) {
		if pc.netConn != nil {
			pc.netConn.Close()
		}
		p.connCount--
		p.notifyWaiter()
		return nil
	}

	pc.healthy = true
	pc.lastUsed = time.Now()
	p.idle = append(p.idle, pc)
	p.notifyWaiter()
	return nil
}

func (p *ConnectionPool) notifyWaiter() {
	if len(p.waiting) > 0 {
		waiter := p.waiting[0]
		p.waiting = p.waiting[1:]
		select {
		case waiter <- struct{}{}:
		default:
		}
	}
}

func (p *ConnectionPool) isValid(pc *poolConn) bool {
	if !pc.healthy {
		return false
	}

	now := time.Now()
	if p.config.MaxLifetime > 0 && now.Sub(pc.createdAt) > p.config.MaxLifetime {
		return false
	}

	if p.config.IdleTimeout > 0 && now.Sub(pc.lastUsed) > p.config.IdleTimeout {
		return false
	}

	return true
}

func (p *ConnectionPool) createConn() (*poolConn, error) {
	connID := generateID()
	conn, err := net.DialTimeout("tcp", p.config.BackendAddress, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("create connection: %w", err)
	}

	return &poolConn{
		id:        connID,
		netConn:   conn,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		healthy:   true,
	}, nil
}

func (p *ConnectionPool) startCleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.cleanup()
		case <-p.stopChan:
			return
		}
	}
}

func (p *ConnectionPool) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	newIdle := make([]*poolConn, 0)

	for _, pc := range p.idle {
		expired := (p.config.MaxLifetime > 0 && now.Sub(pc.createdAt) > p.config.MaxLifetime) ||
			(p.config.IdleTimeout > 0 && now.Sub(pc.lastUsed) > p.config.IdleTimeout)

		if expired {
			if pc.netConn != nil {
				pc.netConn.Close()
			}
			p.connCount--
		} else {
			newIdle = append(newIdle, pc)
		}
	}

	p.idle = newIdle
}

func (p *ConnectionPool) CheckHealth() {
	p.mu.Lock()
	defer p.mu.Unlock()

	healthyIdle := make([]*poolConn, 0)
	removedCount := 0

	for _, pc := range p.idle {
		if p.checkConnHealth(pc) {
			healthyIdle = append(healthyIdle, pc)
		} else {
			if pc.netConn != nil {
				pc.netConn.Close()
			}
			p.connCount--
			removedCount++
		}
	}

	p.idle = healthyIdle

	for i := 0; i < removedCount && p.connCount < p.config.MaxConnections; i++ {
		if pc, err := p.createConn(); err == nil {
			p.idle = append(p.idle, pc)
			p.connCount++
		}
	}
}

func (p *ConnectionPool) checkConnHealth(pc *poolConn) bool {
	if p.config.HealthCheckURL == "" {
		return true
	}

	if pc.netConn != nil {
		pc.netConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		buf := make([]byte, 1)
		_, err := pc.netConn.Read(buf)
		if err != nil {
			return false
		}
	}
	return true
}

func (p *ConnectionPool) UpdateConfig(newMax int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config.MaxConnections = newMax
}

func (p *ConnectionPool) Config() *models.PoolConfig {
	return p.config
}

func (m *Manager) GetStatistics() *models.Statistics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &models.Statistics{}
	for _, pool := range m.pools {
		status := pool.GetStatus()
		stats.TotalActive += status.ActiveCount
		stats.TotalIdle += status.IdleCount
		stats.TotalWaiting += status.WaitingCount
		stats.TotalMax += status.MaxCount
	}
	stats.UpdatedAt = time.Now()
	return stats
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

package pool

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

var (
	ErrPoolNotInitialized = errors.New("pool not initialized")
	ErrAcquireTimeout     = errors.New("acquire connection timeout")
	ErrConnectionNotFound = errors.New("connection not found")
	ErrPoolClosed         = errors.New("pool is closed")
)

type Pool struct {
	config      Config
	idleConns   []*pooledConn
	activeConns map[string]*pooledConn
	totalConns  int
	mu          sync.Mutex
	createSem   chan struct{}
	closed      bool
	connFactory func() (net.Conn, error)
	notify      chan struct{}
}

func NewPool(cfg Config) (*Pool, error) {
	if cfg.TargetAddress == "" {
		return nil, errors.New("target address is required")
	}
	if cfg.MinIdle < 0 {
		return nil, errors.New("min idle cannot be negative")
	}
	if cfg.MaxActive <= 0 {
		return nil, errors.New("max active must be positive")
	}
	if cfg.MinIdle > cfg.MaxActive {
		return nil, errors.New("min idle cannot be greater than max active")
	}

	connFactory := func() (net.Conn, error) {
		conn, err := net.Dial("tcp", cfg.TargetAddress)
		if err != nil {
			return nil, err
		}
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(30 * time.Second)
		}
		return conn, nil
	}

	semCapacity := cfg.MaxActive - cfg.MinIdle
	if semCapacity <= 0 {
		semCapacity = 1
	}

	p := &Pool{
		config:      cfg,
		activeConns: make(map[string]*pooledConn),
		idleConns:   make([]*pooledConn, 0, cfg.MaxActive),
		createSem:   make(chan struct{}, semCapacity),
		connFactory: connFactory,
		notify:      make(chan struct{}, cfg.MaxActive),
	}

	for i := 0; i < cfg.MinIdle; i++ {
		conn, err := p.createNewConnLocked()
		if err != nil {
			for _, pc := range p.idleConns {
				pc.close()
			}
			return nil, err
		}
		p.idleConns = append(p.idleConns, conn)
		p.totalConns++
	}

	return p, nil
}

func (p *Pool) createNewConnLocked() (*pooledConn, error) {
	conn, err := p.connFactory()
	if err != nil {
		return nil, err
	}
	return &pooledConn{
		conn:      conn,
		id:        generateConnID(),
		createdAt: time.Now(),
		lastUsed:  time.Now(),
		inUse:     false,
	}, nil
}

func (p *Pool) tryAcquireFromIdleLocked() (string, net.Conn, bool) {
	for len(p.idleConns) > 0 {
		pc := p.idleConns[0]
		p.idleConns = p.idleConns[1:]

		if time.Since(pc.createdAt) > p.config.MaxLifetime {
			pc.close()
			p.totalConns--
			continue
		}

		if !pc.isAlive() {
			pc.close()
			p.totalConns--
			continue
		}

		pc.inUse = true
		pc.lastUsed = time.Now()
		p.activeConns[pc.id] = pc
		return pc.id, pc.conn, true
	}
	return "", nil, false
}

func (p *Pool) Acquire() (string, net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.config.AcquireTimeout)
	defer cancel()

	for {
		p.mu.Lock()

		if p.closed {
			p.mu.Unlock()
			return "", nil, ErrPoolClosed
		}

		if connID, conn, ok := p.tryAcquireFromIdleLocked(); ok {
			p.mu.Unlock()
			return connID, conn, nil
		}

		if p.totalConns >= p.config.MaxActive {
			p.mu.Unlock()
			select {
			case <-p.notify:
			case <-ctx.Done():
				return "", nil, ErrAcquireTimeout
			}
			continue
		}

		canCreate := false
		select {
		case p.createSem <- struct{}{}:
			canCreate = true
		default:
		}

		if !canCreate {
			p.mu.Unlock()
			select {
			case <-p.notify:
			case <-ctx.Done():
				return "", nil, ErrAcquireTimeout
			}
			continue
		}

		p.mu.Unlock()

		pc, err := p.connFactory()
		<-p.createSem

		if err != nil {
			select {
			case <-ctx.Done():
				return "", nil, ErrAcquireTimeout
			default:
			}
			continue
		}

		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			pc.Close()
			return "", nil, ErrPoolClosed
		}

		newPC := &pooledConn{
			conn:      pc,
			id:        generateConnID(),
			createdAt: time.Now(),
			lastUsed:  time.Now(),
			inUse:     true,
		}
		p.activeConns[newPC.id] = newPC
		p.totalConns++
		p.mu.Unlock()

		return newPC.id, newPC.conn, nil
	}
}

func (p *Pool) Release(connID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return ErrPoolClosed
	}

	pc, ok := p.activeConns[connID]
	if !ok {
		return ErrConnectionNotFound
	}

	delete(p.activeConns, connID)
	pc.inUse = false

	if time.Since(pc.createdAt) > p.config.MaxLifetime {
		pc.close()
		p.totalConns--
		p.notifyWaiters()
		return nil
	}

	if p.totalConns > p.config.MinIdle && time.Since(pc.lastUsed) > p.config.IdleTimeout {
		pc.close()
		p.totalConns--
		p.notifyWaiters()
		return nil
	}

	pc.lastUsed = time.Now()
	p.idleConns = append(p.idleConns, pc)
	p.notifyWaiters()
	return nil
}

func (p *Pool) notifyWaiters() {
	select {
	case p.notify <- struct{}{}:
	default:
	}
}

func (p *Pool) Stats() (int, int, int, int, int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.config.MinIdle, p.config.MaxActive, p.totalConns, len(p.idleConns), len(p.activeConns)
}

func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.closed = true

	for _, pc := range p.idleConns {
		pc.close()
	}
	for _, pc := range p.activeConns {
		pc.close()
	}

	p.idleConns = nil
	p.activeConns = nil
	p.totalConns = 0

	close(p.notify)
}

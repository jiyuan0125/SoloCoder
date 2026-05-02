package connpool

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

var (
	ErrPoolClosed   = errors.New("connpool: pool is closed")
	ErrPoolFull     = errors.New("connpool: pool is full")
	ErrConnInvalid  = errors.New("connpool: connection is invalid")
)

type Pool struct {
	config      config
	metrics     poolMetrics
	idleConns   []*ConnWrapper
	mu          sync.Mutex
	cond        *sync.Cond
	isClosing   bool
	closeOnce   sync.Once
	done        chan struct{}
}

type PoolOption func(*config)

func WithMaxOpen(n int32) PoolOption {
	return func(c *config) {
		c.maxOpen = n
	}
}

func WithMaxIdle(n int32) PoolOption {
	return func(c *config) {
		c.maxIdle = n
	}
}

func WithMaxLifetime(d time.Duration) PoolOption {
	return func(c *config) {
		c.maxLifetime = d
	}
}

func WithMaxIdleTime(d time.Duration) PoolOption {
	return func(c *config) {
		c.maxIdleTime = d
	}
}

func WithHealthCheck(fn HealthCheckFunc) PoolOption {
	return func(c *config) {
		c.healthCheck = fn
	}
}

func New(newFunc NewFunc, opts ...PoolOption) *Pool {
	cfg := config{
		maxOpen:       10,
		maxIdle:       5,
		maxLifetime:   0,
		maxIdleTime:   0,
		newFunc:       newFunc,
		healthCheck:   nil,
		checkInterval: 30 * time.Second,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.maxIdle > cfg.maxOpen && cfg.maxOpen > 0 {
		cfg.maxIdle = cfg.maxOpen
	}

	p := &Pool{
		config:    cfg,
		idleConns: make([]*ConnWrapper, 0, cfg.maxIdle),
		done:      make(chan struct{}),
	}
	p.cond = sync.NewCond(&p.mu)

	if cfg.healthCheck != nil && cfg.checkInterval > 0 {
		go p.healthCheckLoop()
	}

	return p
}

func (p *Pool) Get(ctx context.Context) (*ConnWrapper, error) {
	for {
		conn, err := p.acquire(ctx)
		if err != nil {
			return nil, err
		}

		if p.isExpired(conn) {
			p.mu.Lock()
			p.metrics.decActive()
			p.mu.Unlock()
			_ = conn.conn.Close()
			continue
		}

		return conn, nil
	}
}

func (p *Pool) acquire(ctx context.Context) (*ConnWrapper, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isClosing {
		return nil, ErrPoolClosed
	}

	for {
		if n := len(p.idleConns); n > 0 {
			conn := p.idleConns[n-1]
			p.idleConns = p.idleConns[:n-1]
			p.metrics.decIdle()
			p.metrics.incActive()
			return conn, nil
		}

		currentTotal := p.metrics.total()
		if p.config.maxOpen <= 0 || currentTotal < p.config.maxOpen {
			p.metrics.incCreating()
			p.mu.Unlock()
			netConn, err := p.config.newFunc()
			p.mu.Lock()
			p.metrics.decCreating()

			if p.isClosing {
				if err == nil {
					_ = netConn.Close()
				}
				return nil, ErrPoolClosed
			}

			if err != nil {
				return nil, err
			}

			now := time.Now()
			conn := &ConnWrapper{
				conn:         netConn,
				createTime:   now,
				lastUsedTime: now,
				owner:        p,
			}
			p.metrics.incActive()
			return conn, nil
		}

		if err := p.wait(ctx); err != nil {
			return nil, err
		}
	}
}

func (p *Pool) wait(ctx context.Context) error {
	p.metrics.incWait()
	defer p.metrics.decWait()

	ctxDone := ctx.Done()
	if ctxDone == nil {
		p.cond.Wait()
		return nil
	}

	waitDone := make(chan struct{})
	go func() {
		defer close(waitDone)
		p.cond.Wait()
	}()

	select {
	case <-ctxDone:
		p.cond.Signal()
		<-waitDone
		return ctx.Err()
	case <-waitDone:
		return nil
	}
}

func (p *Pool) Put(conn *ConnWrapper) error {
	if conn == nil || conn.conn == nil {
		return ErrConnInvalid
	}

	if conn.owner != p {
		return ErrConnInvalid
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isClosing {
		p.metrics.decActive()
		_ = conn.conn.Close()
		return nil
	}

	if p.isExpired(conn) {
		p.metrics.decActive()
		_ = conn.conn.Close()
		p.cond.Signal()
		return nil
	}

	if int32(len(p.idleConns)) >= p.config.maxIdle && p.config.maxIdle > 0 {
		p.metrics.decActive()
		_ = conn.conn.Close()
		p.cond.Signal()
		return nil
	}

	conn.lastUsedTime = time.Now()
	p.idleConns = append(p.idleConns, conn)
	p.metrics.decActive()
	p.metrics.incIdle()
	p.cond.Signal()

	return nil
}

func (p *Pool) Close() {
	p.closeOnce.Do(func() {
		p.mu.Lock()
		p.isClosing = true
		p.mu.Unlock()

		close(p.done)
		p.cond.Broadcast()

		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			stats := p.metrics.stats()
			if stats.Active == 0 {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		p.mu.Lock()
		for _, conn := range p.idleConns {
			_ = conn.conn.Close()
		}
		p.idleConns = p.idleConns[:0]
		p.mu.Unlock()
	})
}

func (p *Pool) Stats() PoolStats {
	return p.metrics.stats()
}

func (p *Pool) isExpired(conn *ConnWrapper) bool {
	now := time.Now()

	if p.config.maxLifetime > 0 {
		if now.Sub(conn.createTime) > p.config.maxLifetime {
			return true
		}
	}

	if p.config.maxIdleTime > 0 {
		if now.Sub(conn.lastUsedTime) > p.config.maxIdleTime {
			return true
		}
	}

	return false
}

func (p *Pool) healthCheckLoop() {
	ticker := time.NewTicker(p.config.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.runHealthCheck()
		case <-p.done:
			return
		}
	}
}

func (p *Pool) runHealthCheck() {
	p.mu.Lock()
	if p.isClosing {
		p.mu.Unlock()
		return
	}

	var validConns []*ConnWrapper
	var discarded int

	for _, conn := range p.idleConns {
		if p.config.healthCheck(conn.conn) {
			validConns = append(validConns, conn)
		} else {
			_ = conn.conn.Close()
			discarded++
		}
	}

	p.idleConns = validConns
	if discarded > 0 {
		for i := 0; i < discarded; i++ {
			p.metrics.decIdle()
		}
		for i := 0; i < discarded; i++ {
			p.cond.Signal()
		}
	}

	p.mu.Unlock()
}

package pool

import (
	"errors"
	"sync"
	"time"
)

type GenericPool struct {
	name         string
	factory      Factory
	factoryType  string
	minIdle      int
	maxTotal     int
	maxLifetime  time.Duration
	idleTimeout  time.Duration

	mu           sync.Mutex
	conns        []*wrappedConn
	activeCount  int
	waitingCount int

	totalBorrows    int64
	totalBorrowTime time.Duration
	createdAt       time.Time

	stopChan chan struct{}
	wg       sync.WaitGroup

	borrowTimes map[Connection]time.Time
}

func NewGenericPool(cfg PoolConfig) (*GenericPool, error) {
	if cfg.Name == "" {
		return nil, errors.New("pool name is required")
	}
	if cfg.Factory == nil {
		return nil, errors.New("factory function is required")
	}
	if cfg.MaxTotal <= 0 {
		cfg.MaxTotal = 10
	}
	if cfg.MinIdle < 0 {
		cfg.MinIdle = 0
	}
	if cfg.MinIdle > cfg.MaxTotal {
		cfg.MinIdle = cfg.MaxTotal
	}

	p := &GenericPool{
		name:        cfg.Name,
		factory:     cfg.Factory,
		factoryType: cfg.FactoryType,
		minIdle:     cfg.MinIdle,
		maxTotal:    cfg.MaxTotal,
		maxLifetime: cfg.MaxLifetime,
		idleTimeout: cfg.IdleTimeout,
		conns:       make([]*wrappedConn, 0),
		stopChan:    make(chan struct{}),
		borrowTimes: make(map[Connection]time.Time),
		createdAt:   time.Now(),
	}

	if err := p.preheat(); err != nil {
		return nil, err
	}

	p.wg.Add(1)
	go p.maintainLoop()

	return p, nil
}

func (p *GenericPool) preheat() error {
	for i := 0; i < p.minIdle; i++ {
		wc, err := p.createConn()
		if err != nil {
			return err
		}
		p.conns = append(p.conns, wc)
	}
	return nil
}

func (p *GenericPool) createConn() (*wrappedConn, error) {
	conn, err := p.factory()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &wrappedConn{
		conn:      conn,
		createdAt: now,
		lastUsed:  now,
		inUse:     false,
	}, nil
}

func (p *GenericPool) Borrow() (Connection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		for i := len(p.conns) - 1; i >= 0; i-- {
			wc := p.conns[i]
			if !wc.inUse {
				if p.isExpired(wc) {
					p.closeConn(wc)
					p.conns = append(p.conns[:i], p.conns[i+1:]...)
					continue
				}

				wc.inUse = true
				wc.lastUsed = time.Now()
				p.activeCount++
				p.totalBorrows++
				p.borrowTimes[wc.conn] = time.Now()
				return wc.conn, nil
			}
		}

		totalConns := len(p.conns)
		if totalConns < p.maxTotal {
			wc, err := p.createConn()
			if err != nil {
				return nil, err
			}
			wc.inUse = true
			p.conns = append(p.conns, wc)
			p.activeCount++
			p.totalBorrows++
			p.borrowTimes[wc.conn] = time.Now()
			return wc.conn, nil
		}

		p.waitingCount++
		p.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		p.mu.Lock()
		p.waitingCount--
	}
}

func (p *GenericPool) Return(conn Connection) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if borrowTime, ok := p.borrowTimes[conn]; ok {
		elapsed := time.Since(borrowTime)
		p.totalBorrowTime += elapsed
		delete(p.borrowTimes, conn)
	}

	for _, wc := range p.conns {
		if wc.conn == conn {
			wc.inUse = false
			wc.lastUsed = time.Now()
			p.activeCount--
			return nil
		}
	}

	return errors.New("connection not found in pool")
}

func (p *GenericPool) isExpired(wc *wrappedConn) bool {
	now := time.Now()

	if p.maxLifetime > 0 && now.Sub(wc.createdAt) > p.maxLifetime {
		return true
	}

	if p.idleTimeout > 0 && !wc.inUse && now.Sub(wc.lastUsed) > p.idleTimeout {
		return true
	}

	return false
}

func (p *GenericPool) closeConn(wc *wrappedConn) {
	if wc.conn != nil {
		_ = wc.conn.Close()
	}
}

func (p *GenericPool) maintainLoop() {
	defer p.wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.maintain()
		}
	}
}

func (p *GenericPool) maintain() {
	p.mu.Lock()
	defer p.mu.Unlock()

	idleCount := 0
	for _, wc := range p.conns {
		if !wc.inUse {
			idleCount++
		}
	}

	toRemove := make([]int, 0)
	for i, wc := range p.conns {
		if !wc.inUse && p.isExpired(wc) {
			toRemove = append(toRemove, i)
		}
	}

	for i := len(toRemove) - 1; i >= 0; i-- {
		idx := toRemove[i]
		p.closeConn(p.conns[idx])
		p.conns = append(p.conns[:idx], p.conns[idx+1:]...)
	}

	idleCount = 0
	for _, wc := range p.conns {
		if !wc.inUse {
			idleCount++
		}
	}

	for idleCount < p.minIdle && len(p.conns) < p.maxTotal {
		wc, err := p.createConn()
		if err != nil {
			break
		}
		p.conns = append(p.conns, wc)
		idleCount++
	}
}

func (p *GenericPool) Stats() PoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	idleCount := 0
	for _, wc := range p.conns {
		if !wc.inUse {
			idleCount++
		}
	}

	avgBorrowTime := time.Duration(0)
	if p.totalBorrows > 0 {
		avgBorrowTime = p.totalBorrowTime / time.Duration(p.totalBorrows)
	}

	return PoolStats{
		Name:              p.name,
		FactoryType:       p.factoryType,
		ActiveConnections: p.activeCount,
		IdleConnections:   idleCount,
		WaitingRequests:   p.waitingCount,
		TotalBorrows:      p.totalBorrows,
		AvgBorrowTime:     avgBorrowTime,
		TotalBorrowTime:   p.totalBorrowTime,
		MinIdle:           p.minIdle,
		MaxTotal:          p.maxTotal,
		MaxLifetime:       p.maxLifetime,
		IdleTimeout:       p.idleTimeout,
		CreatedAt:         p.createdAt,
	}
}

func (p *GenericPool) Close() error {
	close(p.stopChan)
	p.wg.Wait()

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, wc := range p.conns {
		p.closeConn(wc)
	}
	p.conns = nil
	return nil
}

func (p *GenericPool) Name() string {
	return p.name
}

func (p *GenericPool) FactoryType() string {
	return p.factoryType
}

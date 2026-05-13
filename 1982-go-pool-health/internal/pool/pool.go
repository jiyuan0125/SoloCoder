package pool

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrPoolClosed       = errors.New("pool is closed")
	ErrGetTimeout       = errors.New("get connection timeout")
	ErrDataSourceClosed = errors.New("data source is closed")
)

type Pool struct {
	name          string
	config        PoolConfig
	configMu      sync.RWMutex

	idleConns     *list.List
	idleMu        sync.Mutex

	activeCount   int32
	leakCount     int32
	isClosed      atomic.Bool
	isReady       atomic.Bool

	lastHealthCheck atomic.Pointer[HealthCheckResult]

	cond          *sync.Cond

	stopOnce      sync.Once
	stopChan      chan struct{}

	dialFunc      func(network, addr string) (net.Conn, error)
}

type HealthCheckResult struct {
	CheckedAt     time.Time
	HealthyCount  int
	UnhealthyCount int
	Details       []string
}

func NewPool(name string, config PoolConfig, dialFunc func(network, addr string) (net.Conn, error)) *Pool {
	p := &Pool{
		name:       name,
		config:     config,
		idleConns:  list.New(),
		stopChan:   make(chan struct{}),
		dialFunc:   dialFunc,
	}
	p.cond = sync.NewCond(&p.idleMu)
	return p
}

func (p *Pool) Name() string {
	return p.name
}

func (p *Pool) IsReady() bool {
	return p.isReady.Load()
}

func (p *Pool) WarmUp() error {
	if p.isClosed.Load() {
		return ErrPoolClosed
	}

	for i := 0; i < p.config.MinIdle; i++ {
		conn, err := p.createConnectionWithRetry()
		if err != nil {
			log.Printf("[Pool %s] WarmUp failed: %v", p.name, err)
			continue
		}
		p.idleConns.PushBack(conn)
	}

	p.isReady.Store(true)
	log.Printf("[Pool %s] WarmUp completed, idle conns: %d", p.name, p.idleConns.Len())
	return nil
}

func (p *Pool) StartHealthCheck() {
	go p.healthCheckLoop()
	go p.leakDetectorLoop()
}

func (p *Pool) Get() (*Conn, error) {
	if p.isClosed.Load() {
		return nil, ErrPoolClosed
	}

	if !p.isReady.Load() {
		return nil, errors.New("pool is not ready yet")
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.config.GetTimeout)
	defer cancel()

	for {
		p.idleMu.Lock()
		for p.idleConns.Len() == 0 && atomic.LoadInt32(&p.activeCount) >= int32(p.getMaxConns()) {
			if !p.waitIdle(ctx) {
				p.idleMu.Unlock()
				return nil, ErrGetTimeout
			}
		}

		if p.idleConns.Len() > 0 {
			elem := p.idleConns.Front()
			p.idleConns.Remove(elem)
			conn := elem.Value.(*Conn)
			p.idleMu.Unlock()

			conn.mu.Lock()
			conn.isBorrowed = true
			conn.borrowedAt = time.Now()
			conn.lastUsedTime = time.Now()
			conn.mu.Unlock()

			atomic.AddInt32(&p.activeCount, 1)
			return conn, nil
		}

		activeCount := atomic.LoadInt32(&p.activeCount)
		p.idleMu.Unlock()

		if activeCount < int32(p.getMaxConns()) {
			conn, err := p.createConnectionWithRetry()
			if err != nil {
				return nil, err
			}

			conn.mu.Lock()
			conn.isBorrowed = true
			conn.borrowedAt = time.Now()
			conn.lastUsedTime = time.Now()
			conn.mu.Unlock()

			atomic.AddInt32(&p.activeCount, 1)
			return conn, nil
		}
	}
}

func (p *Pool) waitIdle(ctx context.Context) bool {
	waitChan := make(chan struct{})
	go func() {
		p.cond.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		return true
	case <-ctx.Done():
		p.cond.Broadcast()
		return false
	}
}

func (p *Pool) putBack(conn *Conn) {
	if p.isClosed.Load() {
		conn.Close()
		return
	}

	conn.mu.Lock()
	conn.isBorrowed = false
	conn.borrowedAt = time.Time{}
	conn.lastUsedTime = time.Now()
	conn.mu.Unlock()

	atomic.AddInt32(&p.activeCount, -1)

	p.idleMu.Lock()
	p.idleConns.PushBack(conn)
	p.cond.Signal()
	p.idleMu.Unlock()
}

func (p *Pool) createConnectionWithRetry() (*Conn, error) {
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		conn, err := p.createConnection()
		if err == nil {
			return conn, nil
		}
		lastErr = err
		log.Printf("[Pool %s] Connection attempt %d failed: %v", p.name, attempt, err)
		if attempt < 3 {
			waitTime := time.Duration(attempt*100) * time.Millisecond
			time.Sleep(waitTime)
		}
	}

	log.Printf("[Pool %s] All connection attempts failed, marking data source as temporarily unhealthy", p.name)
	return nil, fmt.Errorf("connection creation failed after 3 retries: %w", lastErr)
}

func (p *Pool) createConnection() (*Conn, error) {
	network := "tcp"
	addr := p.config.HealthCheckAddr
	if addr == "" {
		return nil, errors.New("health check address is not configured")
	}

	netConn, err := p.dialFunc(network, addr)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%s-%d", p.name, time.Now().UnixNano())
	return newConn(p, id, netConn), nil
}

func (p *Pool) healthCheckLoop() {
	interval := p.getHealthCheckInterval()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.performHealthCheck()
		}
	}
}

func (p *Pool) performHealthCheck() {
	result := &HealthCheckResult{
		CheckedAt: time.Now(),
	}

	p.idleMu.Lock()
	connsToCheck := make([]*list.Element, 0, p.idleConns.Len())
	for e := p.idleConns.Front(); e != nil; e = e.Next() {
		connsToCheck = append(connsToCheck, e)
	}
	p.idleMu.Unlock()

	healthyCount := 0
	unhealthyCount := 0
	details := make([]string, 0)
	connsToReplace := make([]*Conn, 0)

	for _, e := range connsToCheck {
		conn := e.Value.(*Conn)
		if p.checkConnectionHealth(conn) {
			healthyCount++
			details = append(details, fmt.Sprintf("Conn %s: healthy", conn.id))
		} else {
			unhealthyCount++
			details = append(details, fmt.Sprintf("Conn %s: unhealthy", conn.id))
			connsToReplace = append(connsToReplace, conn)

			p.idleMu.Lock()
			p.idleConns.Remove(e)
			p.idleMu.Unlock()
			conn.Close()
		}
	}

	for _, unhealthyConn := range connsToReplace {
		newConn, err := p.createConnectionWithRetry()
		if err != nil {
			log.Printf("[Pool %s] Failed to replace unhealthy connection %s: %v", p.name, unhealthyConn.id, err)
			result.Details = append(details, fmt.Sprintf("Replace %s failed: %v", unhealthyConn.id, err))
			continue
		}

		p.idleMu.Lock()
		p.idleConns.PushBack(newConn)
		p.cond.Signal()
		p.idleMu.Unlock()
		result.Details = append(details, fmt.Sprintf("Replaced %s with %s", unhealthyConn.id, newConn.id))
	}

	result.HealthyCount = healthyCount
	result.UnhealthyCount = unhealthyCount
	result.Details = details
	p.lastHealthCheck.Store(result)

	if unhealthyCount > 0 {
		log.Printf("[Pool %s] Health check: %d healthy, %d unhealthy", p.name, healthyCount, unhealthyCount)
	}
}

func (p *Pool) checkConnectionHealth(conn *Conn) bool {
	if conn.IsClosed() {
		return false
	}

	switch p.getHealthCheckType() {
	case HealthCheckTypeTCP:
		return p.checkTCPHealth(conn)
	case HealthCheckTypeCommand:
		return p.checkCommandHealth(conn)
	default:
		return true
	}
}

func (p *Pool) checkTCPHealth(conn *Conn) bool {
	if conn.conn == nil {
		return false
	}

	conn.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	buf := make([]byte, 1)
	_, err := conn.conn.Read(buf)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return true
		}
		return false
	}
	return true
}

func (p *Pool) checkCommandHealth(conn *Conn) bool {
	if conn.conn == nil {
		return false
	}
	return true
}

func (p *Pool) leakDetectorLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.checkLeaks()
		}
	}
}

func (p *Pool) checkLeaks() {
	leakThreshold := p.getLeakThreshold()
	if leakThreshold <= 0 {
		return
	}

	now := time.Now()
	p.idleMu.Lock()

	for e := p.idleConns.Front(); e != nil; e = e.Next() {
		conn := e.Value.(*Conn)
		conn.mu.RLock()
		isBorrowed := conn.isBorrowed
		borrowedAt := conn.borrowedAt
		conn.mu.RUnlock()

		if isBorrowed && !borrowedAt.IsZero() && now.Sub(borrowedAt) > leakThreshold {
			log.Printf("[Pool %s] LEAK DETECTED: Connection %s borrowed for %s, threshold %s",
				p.name, conn.id, now.Sub(borrowedAt), leakThreshold)

			conn.mu.Lock()
			conn.isBorrowed = false
			conn.borrowedAt = time.Time{}
			conn.mu.Unlock()

			atomic.AddInt32(&p.activeCount, -1)
			atomic.AddInt32(&p.leakCount, 1)

			p.idleConns.Remove(e)
			conn.Close()
		}
	}

	p.idleMu.Unlock()
}

func (p *Pool) getMaxConns() int {
	p.configMu.RLock()
	defer p.configMu.RUnlock()
	return p.config.MaxConns
}

func (p *Pool) getHealthCheckType() HealthCheckType {
	p.configMu.RLock()
	defer p.configMu.RUnlock()
	return p.config.HealthCheckType
}

func (p *Pool) getHealthCheckInterval() time.Duration {
	p.configMu.RLock()
	defer p.configMu.RUnlock()
	if p.config.HealthCheckInterval <= 0 {
		return 30 * time.Second
	}
	return p.config.HealthCheckInterval
}

func (p *Pool) getLeakThreshold() time.Duration {
	p.configMu.RLock()
	defer p.configMu.RUnlock()
	return p.config.LeakThreshold
}

func (p *Pool) GetStats() PoolStats {
	p.idleMu.Lock()
	idleCount := p.idleConns.Len()
	p.idleMu.Unlock()

	activeCount := atomic.LoadInt32(&p.activeCount)
	leakCount := atomic.LoadInt32(&p.leakCount)
	lastCheck := p.lastHealthCheck.Load()

	return PoolStats{
		Name:           p.name,
		ActiveCount:    activeCount,
		IdleCount:      int32(idleCount),
		LeakCount:      leakCount,
		IsReady:        p.isReady.Load(),
		IsClosed:       p.isClosed.Load(),
		LastHealthCheck: lastCheck,
	}
}

func (p *Pool) UpdateConfig(newConfig PoolConfig) {
	p.configMu.Lock()
	p.config = newConfig
	p.configMu.Unlock()

	log.Printf("[Pool %s] Config updated: %+v", p.name, newConfig)
}

func (p *Pool) GetConfig() PoolConfig {
	p.configMu.RLock()
	defer p.configMu.RUnlock()
	return p.config
}

func (p *Pool) Close() error {
	p.stopOnce.Do(func() {
		p.isClosed.Store(true)
		close(p.stopChan)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	waitTicker := time.NewTicker(100 * time.Millisecond)
	defer waitTicker.Stop()

	for {
		activeCount := atomic.LoadInt32(&p.activeCount)
		if activeCount == 0 {
			break
		}

		select {
		case <-ctx.Done():
			log.Printf("[Pool %s] Close timeout (%ds), forcing remaining connections closed", p.name, 60)
			p.idleMu.Lock()
			for e := p.idleConns.Front(); e != nil; e = e.Next() {
				conn := e.Value.(*Conn)
				conn.Close()
			}
			p.idleConns.Init()
			p.idleMu.Unlock()
			return nil
		case <-waitTicker.C:
		}
	}

	p.idleMu.Lock()
	for e := p.idleConns.Front(); e != nil; e = e.Next() {
		conn := e.Value.(*Conn)
		conn.Close()
	}
	p.idleConns.Init()
	p.idleMu.Unlock()

	log.Printf("[Pool %s] Closed successfully", p.name)
	return nil
}

type PoolStats struct {
	Name            string
	ActiveCount     int32
	IdleCount       int32
	LeakCount       int32
	IsReady         bool
	IsClosed        bool
	LastHealthCheck *HealthCheckResult
}

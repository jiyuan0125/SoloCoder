package pool

import (
	"context"
	"sync"
	"time"
)

type Connection struct {
	ID        string
	CreatedAt time.Time
	LastUsed  time.Time
	IsActive  bool
}

type ShardStats struct {
	ShardID        int
	ActiveConns    int
	IdleConns      int
	WaitingCount   int64
	TotalRequests  int64
	AcquireLatency time.Duration
}

type Shard struct {
	id             int
	cfg            ShardConfig
	activeConns    map[string]*Connection
	idleConns      []*Connection
	waiters        []chan struct{}
	mu             sync.Mutex
	totalRequests  int64
	waitingCount   int64
	totalLatency   time.Duration
	connCounter    int64
}

func NewShard(id int, cfg ShardConfig) *Shard {
	return &Shard{
		id:          id,
		cfg:         cfg,
		activeConns: make(map[string]*Connection),
		idleConns:   make([]*Connection, 0),
		waiters:     make([]chan struct{}, 0),
	}
}

func (s *Shard) ID() int {
	return s.id
}

func (s *Shard) Acquire(ctx context.Context) (*Connection, error) {
	start := time.Now()

	s.mu.Lock()

	if len(s.idleConns) > 0 {
		conn := s.idleConns[0]
		s.idleConns = s.idleConns[1:]

		if s.isValidConn(conn) {
			conn.IsActive = true
			conn.LastUsed = time.Now()
			s.activeConns[conn.ID] = conn
			s.mu.Unlock()

			s.totalRequests++
			s.totalLatency += time.Since(start)
			return conn, nil
		}
	}

	if len(s.activeConns) < s.cfg.MaxConnections {
		conn := s.newConnection()
		s.activeConns[conn.ID] = conn
		s.mu.Unlock()

		s.totalRequests++
		s.totalLatency += time.Since(start)
		return conn, nil
	}

	wait := make(chan struct{})
	s.waiters = append(s.waiters, wait)
	s.waitingCount++
	s.mu.Unlock()

	select {
	case <-ctx.Done():
		s.mu.Lock()
		for i, w := range s.waiters {
			if w == wait {
				s.waiters = append(s.waiters[:i], s.waiters[i+1:]...)
				break
			}
		}
		s.waitingCount--
		s.mu.Unlock()
		return nil, ctx.Err()
	case <-wait:
		return s.Acquire(ctx)
	}
}

func (s *Shard) Release(connID string) {
	s.mu.Lock()

	conn, ok := s.activeConns[connID]
	if !ok {
		s.mu.Unlock()
		return
	}

	delete(s.activeConns, connID)

	if s.isValidConn(conn) {
		conn.IsActive = false
		conn.LastUsed = time.Now()
		s.idleConns = append(s.idleConns, conn)
	}

	if len(s.waiters) > 0 {
		waiter := s.waiters[0]
		s.waiters = s.waiters[1:]
		close(waiter)
	}

	s.mu.Unlock()
}

func (s *Shard) Stats() ShardStats {
	s.mu.Lock()
	defer s.mu.Unlock()

	var avgLatency time.Duration
	if s.totalRequests > 0 {
		avgLatency = s.totalLatency / time.Duration(s.totalRequests)
	}

	return ShardStats{
		ShardID:        s.id,
		ActiveConns:    len(s.activeConns),
		IdleConns:      len(s.idleConns),
		WaitingCount:   s.waitingCount,
		TotalRequests:  s.totalRequests,
		AcquireLatency: avgLatency,
	}
}

func (s *Shard) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	validIdle := make([]*Connection, 0)
	for _, conn := range s.idleConns {
		if s.isValidConn(conn) {
			validIdle = append(validIdle, conn)
		}
	}
	s.idleConns = validIdle
}

func (s *Shard) newConnection() *Connection {
	s.connCounter++
	return &Connection{
		ID:        s.generateConnID(),
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		IsActive:  true,
	}
}

func (s *Shard) generateConnID() string {
	return "conn-" + string('0'+s.id) + "-" + time.Now().Format("20060102150405.000000") + "-" + string('0'+s.id)
}

func (s *Shard) isValidConn(conn *Connection) bool {
	now := time.Now()

	if s.cfg.IdleTimeout > 0 && now.Sub(conn.LastUsed) > s.cfg.IdleTimeout {
		return false
	}

	if s.cfg.MaxLifetime > 0 && now.Sub(conn.CreatedAt) > s.cfg.MaxLifetime {
		return false
	}

	return true
}

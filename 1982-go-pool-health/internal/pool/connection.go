package pool

import (
	"net"
	"sync"
	"time"
)

type Conn struct {
	id           string
	conn         net.Conn
	pool         *Pool
	createdAt    time.Time
	lastUsedTime time.Time
	borrowedAt   time.Time
	isBorrowed   bool
	mu           sync.RWMutex
}

func newConn(pool *Pool, id string, conn net.Conn) *Conn {
	return &Conn{
		id:           id,
		conn:         conn,
		pool:         pool,
		createdAt:    time.Now(),
		lastUsedTime: time.Now(),
	}
}

func (c *Conn) ID() string {
	return c.id
}

func (c *Conn) Raw() net.Conn {
	return c.conn
}

func (c *Conn) Release() {
	c.pool.putBack(c)
}

func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

func (c *Conn) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn == nil
}

func (c *Conn) LastUsedTime() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastUsedTime
}

func (c *Conn) BorrowedAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.borrowedAt
}

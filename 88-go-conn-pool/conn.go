package connpool

import (
	"net"
	"sync/atomic"
	"time"
)

type ConnWrapper struct {
	conn         net.Conn
	createTime   time.Time
	lastUsedTime time.Time
	owner        *Pool
}

func (cw *ConnWrapper) NetConn() net.Conn {
	return cw.conn
}

func (cw *ConnWrapper) Read(b []byte) (n int, err error) {
	cw.lastUsedTime = time.Now()
	return cw.conn.Read(b)
}

func (cw *ConnWrapper) Write(b []byte) (n int, err error) {
	cw.lastUsedTime = time.Now()
	return cw.conn.Write(b)
}

func (cw *ConnWrapper) Close() error {
	return cw.conn.Close()
}

func (cw *ConnWrapper) LocalAddr() net.Addr {
	return cw.conn.LocalAddr()
}

func (cw *ConnWrapper) RemoteAddr() net.Addr {
	return cw.conn.RemoteAddr()
}

func (cw *ConnWrapper) SetDeadline(t time.Time) error {
	return cw.conn.SetDeadline(t)
}

func (cw *ConnWrapper) SetReadDeadline(t time.Time) error {
	return cw.conn.SetReadDeadline(t)
}

func (cw *ConnWrapper) SetWriteDeadline(t time.Time) error {
	return cw.conn.SetWriteDeadline(t)
}

type PoolStats struct {
	Active    int32
	Idle      int32
	WaitCount int32
	WaitTotal int64
}

type NewFunc func() (net.Conn, error)

type HealthCheckFunc func(net.Conn) bool

type config struct {
	maxOpen       int32
	maxIdle       int32
	maxLifetime   time.Duration
	maxIdleTime   time.Duration
	newFunc       NewFunc
	healthCheck   HealthCheckFunc
	checkInterval time.Duration
}

type poolMetrics struct {
	active    int32
	idle      int32
	creating  int32
	waitCount int32
	waitTotal int64
}

func (m *poolMetrics) incActive() {
	atomic.AddInt32(&m.active, 1)
}

func (m *poolMetrics) decActive() {
	atomic.AddInt32(&m.active, -1)
}

func (m *poolMetrics) incIdle() {
	atomic.AddInt32(&m.idle, 1)
}

func (m *poolMetrics) decIdle() {
	atomic.AddInt32(&m.idle, -1)
}

func (m *poolMetrics) incCreating() {
	atomic.AddInt32(&m.creating, 1)
}

func (m *poolMetrics) decCreating() {
	atomic.AddInt32(&m.creating, -1)
}

func (m *poolMetrics) incWait() {
	atomic.AddInt32(&m.waitCount, 1)
	atomic.AddInt64(&m.waitTotal, 1)
}

func (m *poolMetrics) decWait() {
	atomic.AddInt32(&m.waitCount, -1)
}

func (m *poolMetrics) stats() PoolStats {
	return PoolStats{
		Active:    atomic.LoadInt32(&m.active),
		Idle:      atomic.LoadInt32(&m.idle),
		WaitCount: atomic.LoadInt32(&m.waitCount),
		WaitTotal: atomic.LoadInt64(&m.waitTotal),
	}
}

func (m *poolMetrics) total() int32 {
	return atomic.LoadInt32(&m.active) + atomic.LoadInt32(&m.idle) + atomic.LoadInt32(&m.creating)
}

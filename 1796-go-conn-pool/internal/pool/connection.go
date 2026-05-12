package pool

import (
	"net"
	"sync/atomic"
)

type PooledConnection struct {
	conn   net.Conn
	uses   atomic.Int64
	closed atomic.Bool
}

func newPooledConnection(conn net.Conn) *PooledConnection {
	return &PooledConnection{
		conn: conn,
	}
}

func (pc *PooledConnection) GetConn() net.Conn {
	return pc.conn
}

func (pc *PooledConnection) GetUses() int64 {
	return pc.uses.Load()
}

func (pc *PooledConnection) IncrementUses() {
	pc.uses.Add(1)
}

func (pc *PooledConnection) Close() error {
	if !pc.closed.CompareAndSwap(false, true) {
		return pc.conn.Close()
	}
	return nil
}

func (pc *PooledConnection) IsClosed() bool {
	return pc.closed.Load()
}

package pool

import (
	"net"
	"sync"
	"time"
)

type pooledConn struct {
	conn      net.Conn
	id        string
	createdAt time.Time
	lastUsed  time.Time
	inUse     bool
}

var (
	connIDCounter uint64
	connIDMu      sync.Mutex
)

func generateConnID() string {
	connIDMu.Lock()
	defer connIDMu.Unlock()
	connIDCounter++
	return time.Now().Format("20060102150405-") + string(rune('0'+connIDCounter%26))
}

func (pc *pooledConn) isAlive() bool {
	return true
}

func (pc *pooledConn) close() error {
	return pc.conn.Close()
}

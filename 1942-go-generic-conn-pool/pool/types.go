package pool

import "time"

type Connection interface {
	Close() error
}

type Factory func() (Connection, error)

type PoolConfig struct {
	Name         string
	Factory      Factory
	FactoryType  string
	MinIdle      int
	MaxTotal     int
	MaxLifetime  time.Duration
	IdleTimeout  time.Duration
}

type PoolStats struct {
	Name              string
	FactoryType       string
	ActiveConnections int
	IdleConnections   int
	WaitingRequests   int
	TotalBorrows      int64
	AvgBorrowTime     time.Duration
	TotalBorrowTime   time.Duration
	MinIdle           int
	MaxTotal          int
	MaxLifetime       time.Duration
	IdleTimeout       time.Duration
	CreatedAt         time.Time
}

type wrappedConn struct {
	conn      Connection
	createdAt time.Time
	lastUsed  time.Time
	inUse     bool
}

type Pool interface {
	Borrow() (Connection, error)
	Return(Connection) error
	Stats() PoolStats
	Close() error
	Name() string
	FactoryType() string
}

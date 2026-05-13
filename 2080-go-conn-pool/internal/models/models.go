package models

import (
	"time"
)

type PoolConfig struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	MaxConnections  int           `json:"max_connections"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	MaxLifetime     time.Duration `json:"max_lifetime"`
	BackendAddress  string        `json:"backend_address"`
	WaitTimeout     time.Duration `json:"wait_timeout"`
	HealthCheckURL  string        `json:"health_check_url,omitempty"`
	Quota           int           `json:"quota"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

type PoolStatus struct {
	PoolID       string `json:"pool_id"`
	ActiveCount  int    `json:"active_count"`
	IdleCount    int    `json:"idle_count"`
	WaitingCount int    `json:"waiting_count"`
	MaxCount     int    `json:"max_count"`
}

type Connection struct {
	ID          string
	PoolID      string
	Conn        interface{}
	Addr        string
	CreatedAt   time.Time
	LastUsed    time.Time
	Healthy     bool
	InUse       bool
	maxLifetime time.Duration
}

func (c *Connection) MaxLifetime() time.Duration {
	return c.maxLifetime
}

func (c *Connection) SetMaxLifetime(d time.Duration) {
	c.maxLifetime = d
}

type Statistics struct {
	TotalActive  int `json:"total_active"`
	TotalIdle    int `json:"total_idle"`
	TotalWaiting int `json:"total_waiting"`
	TotalMax     int `json:"total_max"`
	UpdatedAt    time.Time `json:"updated_at"`
}

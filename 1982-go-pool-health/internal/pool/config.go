package pool

import (
	"time"
)

type HealthCheckType string

const (
	HealthCheckTypeTCP     HealthCheckType = "tcp"
	HealthCheckTypeCommand HealthCheckType = "command"
)

type PoolConfig struct {
	MaxConns          int               `json:"maxConns"`
	MinIdle           int               `json:"minIdle"`
	GetTimeout        time.Duration     `json:"getTimeout"`
	IdleTimeout       time.Duration     `json:"idleTimeout"`
	LeakThreshold     time.Duration     `json:"leakThreshold"`
	HealthCheckType   HealthCheckType   `json:"healthCheckType"`
	HealthCheckAddr   string            `json:"healthCheckAddr"`
	HealthCheckCmd    string            `json:"healthCheckCmd"`
	HealthCheckInterval time.Duration   `json:"healthCheckInterval"`
}

func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxConns:          20,
		MinIdle:           5,
		GetTimeout:        5 * time.Second,
		IdleTimeout:       10 * time.Minute,
		LeakThreshold:     5 * time.Minute,
		HealthCheckType:   HealthCheckTypeTCP,
		HealthCheckInterval: 30 * time.Second,
	}
}

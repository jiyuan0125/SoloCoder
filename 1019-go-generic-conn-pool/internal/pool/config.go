package pool

import "time"

type Config struct {
	TargetAddress   string
	MinIdle         int
	MaxActive       int
	IdleTimeout     time.Duration
	MaxLifetime     time.Duration
	AcquireTimeout  time.Duration
	HealthCheckTime time.Duration
}

func DefaultConfig() Config {
	return Config{
		MinIdle:         2,
		MaxActive:       10,
		IdleTimeout:     60 * time.Second,
		MaxLifetime:     5 * time.Minute,
		AcquireTimeout:  10 * time.Second,
		HealthCheckTime: 1 * time.Second,
	}
}

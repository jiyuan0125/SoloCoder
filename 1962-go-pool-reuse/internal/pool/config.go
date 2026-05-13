package pool

import (
	"os"
	"strconv"
	"time"
)

type ShardConfig struct {
	ShardCount      int
	VirtualNodeCount int
	MaxConnections  int
	IdleTimeout     time.Duration
	MaxLifetime     time.Duration
}

func DefaultConfig() ShardConfig {
	return ShardConfig{
		ShardCount:      16,
		VirtualNodeCount: 100,
		MaxConnections:  10,
		IdleTimeout:     5 * time.Minute,
		MaxLifetime:     30 * time.Minute,
	}
}

func LoadConfigFromEnv() ShardConfig {
	cfg := DefaultConfig()

	if v := os.Getenv("SHARD_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.ShardCount = n
		}
	}

	if v := os.Getenv("VIRTUAL_NODE_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.VirtualNodeCount = n
		}
	}

	if v := os.Getenv("MAX_CONNECTIONS_PER_SHARD"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxConnections = n
		}
	}

	if v := os.Getenv("IDLE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.IdleTimeout = d
		}
	}

	if v := os.Getenv("MAX_LIFETIME"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.MaxLifetime = d
		}
	}

	return cfg
}

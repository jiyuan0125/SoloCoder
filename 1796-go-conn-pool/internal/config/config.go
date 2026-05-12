package config

type PoolConfig struct {
	Address        string
	MinIdle        int
	MaxOpen        int
	MaxUsesPerConn int
}

func DefaultPoolConfig(address string) PoolConfig {
	return PoolConfig{
		Address:        address,
		MinIdle:        5,
		MaxOpen:        20,
		MaxUsesPerConn: 1000,
	}
}

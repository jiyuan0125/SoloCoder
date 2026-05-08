package dnscache

import "dns-cache-service/common"

type Config struct {
	NXDomainTTLMultiplier float64
	DefaultTTL            int
}

func DefaultConfig() *Config {
	return &Config{
		NXDomainTTLMultiplier: common.DefaultNXDomainTTLMultiplier,
		DefaultTTL:            common.DefaultTTL,
	}
}

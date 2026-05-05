package common

import "time"

const (
	MinRefreshRate       = 5 * time.Second
	MaxReconnectWindow   = 5 * time.Minute
	DataRetentionDays    = 7
	MaxClients           = 100
	ExpirationThreshold  = 1 * time.Hour
	TrendWindow          = 1 * time.Hour
	TrendInterval        = 5 * time.Minute
	HeartbeatInterval    = 30 * time.Second
	DefaultDelayWarning  = 1000
	DefaultDelayCritical = 3000
)

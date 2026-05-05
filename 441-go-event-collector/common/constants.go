package common

const (
	MaxEventNameLength    = 50
	MaxPropertiesSize     = 10 * 1024
	TimestampPastLimit    = 24 * 60 * 60
	TimestampFutureLimit  = 5 * 60
	DeduplicationWindow   = 1
	ReportRetentionDays   = 90
	MaxRetryAttempts      = 3
	DefaultPageSize       = 20
	MaxPageSize           = 100
	DefaultServerPort     = 8080
)

var EventNamePattern = `^[a-zA-Z0-9_]+$`

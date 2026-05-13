package config

const (
	Port               = "8801"
	MaxMemoryUsage     = 512 * 1024 * 1024
	MaxNestingDepth    = 200
	DBPath             = "./validator.db"
	MaxArrayElements   = 10000
	MaxStringLength    = 10 * 1024 * 1024
	MemoryCheckEnabled = true
)

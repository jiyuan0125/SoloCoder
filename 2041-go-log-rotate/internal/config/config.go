package config

const (
	DefaultMaxFileSizeMB   = 100
	DefaultRotateHour      = 3
	DefaultRotateMinute    = 0
	DefaultMaxArchiveFiles = 30
	MaxLogEntrySize        = 1 * 1024 * 1024
	ServerPort             = ":8100"
	DefaultLogDir          = "./logs"
	DefaultDBPath          = "./log_rotate_test.db"
	TruncateMarker         = "[已截断]"
)

var DefaultConfig = map[string]interface{}{
	"max_file_size_mb":    DefaultMaxFileSizeMB,
	"rotate_hour":         DefaultRotateHour,
	"rotate_minute":       DefaultRotateMinute,
	"max_archive_files":   DefaultMaxArchiveFiles,
	"log_dir":             DefaultLogDir,
	"db_path":             DefaultDBPath,
}

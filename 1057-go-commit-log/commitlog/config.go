package commitlog

import "time"

type Config struct {
	SegmentSize       int64
	IndexInterval     int
	RetentionDuration time.Duration
	CleanupInterval   time.Duration
}

func DefaultConfig() Config {
	return Config{
		SegmentSize:       10 * 1024 * 1024,
		IndexInterval:     4 * 1024,
		RetentionDuration: 24 * time.Hour,
		CleanupInterval:   5 * time.Minute,
	}
}

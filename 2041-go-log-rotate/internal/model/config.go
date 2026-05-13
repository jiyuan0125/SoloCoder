package model

import "time"

type LogConfig struct {
	ID              int64     `json:"id"`
	MaxFileSizeMB   int       `json:"max_file_size_mb"`
	RotateHour      int       `json:"rotate_hour"`
	RotateMinute    int       `json:"rotate_minute"`
	MaxArchiveFiles int       `json:"max_archive_files"`
	LogDir          string    `json:"log_dir"`
	DBPath          string    `json:"db_path"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type LogArchive struct {
	ID        int64     `json:"id"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	Compressed bool     `json:"compressed"`
	CreatedAt time.Time `json:"created_at"`
	Deleted   bool      `json:"deleted"`
}

type LogEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Truncated bool      `json:"truncated,omitempty"`
}

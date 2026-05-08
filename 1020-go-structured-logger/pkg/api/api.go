package api

import "time"

type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     LogLevel          `json:"level"`
	Message   string            `json:"message"`
	File      string            `json:"file"`
	Line      int               `json:"line"`
	Fields    map[string]any    `json:"fields,omitempty"`
}

type WriteLogRequest struct {
	Level   LogLevel       `json:"level"`
	Message string         `json:"message"`
	Fields  map[string]any `json:"fields,omitempty"`
}

type WriteLogResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type UpdateLevelRequest struct {
	Level LogLevel `json:"level"`
}

type UpdateLevelResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetConfigResponse struct {
	Level        LogLevel         `json:"level"`
	MaxFileSize  int64            `json:"max_file_size"`
	MaxBackups   int              `json:"max_backups"`
	FieldLimits  map[string]int   `json:"field_limits"`
	OutputToFile bool             `json:"output_to_file"`
	OutputToStd  bool             `json:"output_to_std"`
}

type GetRecentLogsRequest struct {
	Count int `json:"count"`
}

type GetRecentLogsResponse struct {
	Logs    []LogEntry `json:"logs"`
	Success bool       `json:"success"`
	Error   string     `json:"error,omitempty"`
}

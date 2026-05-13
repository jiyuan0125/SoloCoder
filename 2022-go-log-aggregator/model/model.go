package model

import (
	"encoding/json"
	"time"
)

type LogEntry struct {
	ID        int64     `json:"id,omitempty" db:"id"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Level     string    `json:"level" db:"level"`
	Service   string    `json:"service" db:"service"`
	Message   string    `json:"message" db:"message"`
}

type AggregatedLogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Level       string    `json:"level"`
	Service     string    `json:"service"`
	Message     string    `json:"message"`
	Count       int       `json:"count"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}

type LogRequest struct {
	Timestamp *time.Time `json:"timestamp"`
	Level     *string    `json:"level"`
	Service   *string    `json:"service"`
	Message   *string    `json:"message"`
}

func (l *LogRequest) Validate() []string {
	missing := []string{}
	if l.Timestamp == nil {
		missing = append(missing, "timestamp")
	}
	if l.Level == nil || *l.Level == "" {
		missing = append(missing, "level")
	}
	if l.Service == nil || *l.Service == "" {
		missing = append(missing, "service")
	}
	if l.Message == nil || *l.Message == "" {
		missing = append(missing, "message")
	}
	return missing
}

func (l *LogRequest) ToLogEntry() *LogEntry {
	return &LogEntry{
		Timestamp: *l.Timestamp,
		Level:     *l.Level,
		Service:   *l.Service,
		Message:   *l.Message,
	}
}

type LogQueryParams struct {
	StartTime time.Time
	EndTime   time.Time
	Levels    []string
	Services  []string
	SortOrder string
}

type DailyStats struct {
	Date        string  `json:"date"`
	Service     string  `json:"service"`
	TotalCount  int     `json:"total_count"`
	ErrorCount  int     `json:"error_count"`
	FatalCount  int     `json:"fatal_count"`
	ErrorRate   float64 `json:"error_rate"`
}

type RetentionConfig struct {
	RetentionDays int `json:"retention_days"`
}

func (r *RetentionConfig) Validate() error {
	if r.RetentionDays < 1 {
		return nil
	}
	return nil
}

func (r *RetentionConfig) ToJSON() string {
	b, _ := json.Marshal(r)
	return string(b)
}

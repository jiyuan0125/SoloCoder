package types

import (
	"encoding/json"
	"time"
)

type LogEntry struct {
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	Timestamp  time.Time              `json:"timestamp"`
	Source     string                 `json:"source"`
	CustomData map[string]interface{} `json:"-"`
}

func (l *LogEntry) UnmarshalJSON(data []byte) error {
	type Alias LogEntry
	aux := &struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Alias: (*Alias)(l),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Timestamp != "" {
		t, err := time.Parse(time.RFC3339, aux.Timestamp)
		if err != nil {
			l.Timestamp = time.Now()
		} else {
			l.Timestamp = t
		}
	} else {
		l.Timestamp = time.Now()
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	delete(raw, "level")
	delete(raw, "message")
	delete(raw, "timestamp")
	delete(raw, "source")
	if len(raw) > 0 {
		l.CustomData = raw
	}

	return nil
}

func (l *LogEntry) MarshalJSON() ([]byte, error) {
	type Alias LogEntry
	aux := &struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Timestamp: l.Timestamp.Format(time.RFC3339),
		Alias:     (*Alias)(l),
	}
	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	if len(l.CustomData) > 0 {
		var merged map[string]interface{}
		if err := json.Unmarshal(data, &merged); err != nil {
			return nil, err
		}
		for k, v := range l.CustomData {
			merged[k] = v
		}
		return json.Marshal(merged)
	}

	return data, nil
}

type Target interface {
	ID() string
	Type() string
	Config() interface{}
	Status() TargetStatus
	Write([]*LogEntry) error
	Start() error
	Stop() error
}

type TargetStatus struct {
	ID                string      `json:"id"`
	Type              string      `json:"type"`
	Config            interface{} `json:"config"`
	LastWriteTime     time.Time   `json:"last_write_time"`
	SuccessCount      int64       `json:"success_count"`
	FailureCount      int64       `json:"failure_count"`
	PendingRetryCount int64       `json:"pending_retry_count,omitempty"`
	WriteRate         float64     `json:"write_rate"`
}

type FileRotationPolicy struct {
	Strategy   string        `json:"strategy"`
	MaxSize    int64         `json:"max_size,omitempty"`
	MaxAge     time.Duration `json:"max_age,omitempty"`
	MaxBackups int           `json:"max_backups,omitempty"`
}

type FileTargetConfig struct {
	Path     string             `json:"path"`
	Rotation FileRotationPolicy `json:"rotation,omitempty"`
}

type HTTPTargetConfig struct {
	URL          string        `json:"url"`
	BatchSize    int           `json:"batch_size"`
	FlushTimeout time.Duration `json:"flush_timeout"`
	Retries      int           `json:"retries"`
	RetryDelay   time.Duration `json:"retry_delay"`
	Timeout      time.Duration `json:"timeout"`
}

type MemoryTargetConfig struct {
	MaxEntries int `json:"max_entries"`
}

type Stats struct {
	ReceivedPerSecond   float64                `json:"received_per_second"`
	TotalReceived       int64                  `json:"total_received"`
	TargetStats         map[string]TargetStats `json:"targets"`
}

type TargetStats struct {
	WriteRate      float64 `json:"write_rate"`
	SuccessPerSec  float64 `json:"success_per_second"`
	FailurePerSec  float64 `json:"failure_per_second"`
	TotalSuccess   int64   `json:"total_success"`
	TotalFailure   int64   `json:"total_failure"`
	PendingRetries int64   `json:"pending_retries,omitempty"`
}

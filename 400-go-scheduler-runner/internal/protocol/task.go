package protocol

import (
	"time"
)

type TaskConfig struct {
	Name        string        `json:"name"`
	Command     string        `json:"command"`
	CronExpr    string        `json:"cron_expr"`
	Timeout     time.Duration `json:"timeout"`
	MaxRetry    int           `json:"max_retry"`
	RetryInterval time.Duration `json:"retry_interval"`
	Disabled    bool          `json:"disabled"`
}

type ExecutionRecord struct {
	ID         string    `json:"id"`
	TaskName   string    `json:"task_name"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Duration   time.Duration `json:"duration"`
	Success    bool      `json:"success"`
	Output     string    `json:"output"`
	Error      string    `json:"error,omitempty"`
	RetryCount int       `json:"retry_count"`
}

type TaskStatus struct {
	TaskConfig
	NextRun       time.Time         `json:"next_run"`
	LastExecution *ExecutionRecord  `json:"last_execution,omitempty"`
	IsRunning     bool              `json:"is_running"`
}

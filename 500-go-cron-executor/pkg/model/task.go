package model

import (
	"time"
)

type Task struct {
	ID             string
	Name           string
	CronExpr       string
	Command        string
	Timeout        time.Duration
	MaxRetries     int
	RetryInterval  time.Duration
	Dependencies   []string
	NextRun        time.Time
	LastRun        time.Time
	Status         TaskStatus
	ActiveExecID   string
}

type TaskStatus string

const (
	TaskStatusIdle       TaskStatus = "idle"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusSuccess    TaskStatus = "success"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusRetrying   TaskStatus = "retrying"
)

type Execution struct {
	ID              string
	TaskID          string
	TaskName        string
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	ExitCode        int
	Stdout          string
	Stderr          string
	Status          ExecutionStatus
	RetryCount      int
	MaxRetries      int
	IsTimeout       bool
	IsManualTrigger bool
}

type ExecutionStatus string

const (
	ExecStatusRunning   ExecutionStatus = "running"
	ExecStatusSuccess   ExecutionStatus = "success"
	ExecStatusFailed    ExecutionStatus = "failed"
	ExecStatusTimeout   ExecutionStatus = "timeout"
	ExecStatusSkipped   ExecutionStatus = "skipped"
)

type TaskConfig struct {
	Name           string        `json:"name"`
	CronExpr       string        `json:"cron_expr"`
	Command        string        `json:"command"`
	Timeout        time.Duration `json:"timeout"`
	MaxRetries     int           `json:"max_retries"`
	RetryInterval  time.Duration `json:"retry_interval"`
	Dependencies   []string      `json:"dependencies"`
}

type Config struct {
	TCPPort     int          `json:"tcp_port"`
	WebAPIPort  int          `json:"web_api_port"`
	LogDir      string       `json:"log_dir"`
	DataDir     string       `json:"data_dir"`
	Tasks       []TaskConfig `json:"tasks"`
}

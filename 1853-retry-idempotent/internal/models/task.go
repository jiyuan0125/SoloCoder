package models

import (
	"time"
)

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusRunning    TaskStatus = "running"
	StatusSuccess    TaskStatus = "success"
	StatusFailed     TaskStatus = "failed"
)

type RetryStrategy string

const (
	StrategyFixed         RetryStrategy = "fixed"
	StrategyExponential   RetryStrategy = "exponential"
)

type Task struct {
	ID             string
	IdempotencyKey string
	URL            string
	Method         string
	RequestBody    []byte
	MaxRetries     int
	RetryStrategy  RetryStrategy
	BaseInterval   time.Duration
	Attempts       int
	Status         TaskStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Result         *TaskResult
	History        []*ExecutionRecord
	Callbacks      []string
}

type TaskResult struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
	Success    bool
	Error      string
}

type ExecutionRecord struct {
	Attempt    int
	StartTime  time.Time
	EndTime    time.Time
	StatusCode int
	Body       []byte
	Success    bool
	Error      string
}

type Callback struct {
	TaskID string
	URL    string
}

type CallbackResult struct {
	TaskID   string
	URL      string
	Success  bool
	Error    string
	SentAt   time.Time
}

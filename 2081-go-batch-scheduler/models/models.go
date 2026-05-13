package models

import (
	"time"
)

type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
)

func (p Priority) String() string {
	switch p {
	case PriorityHigh:
		return "high"
	case PriorityMedium:
		return "medium"
	case PriorityLow:
		return "low"
	default:
		return "unknown"
	}
}

func PriorityFromString(s string) (Priority, bool) {
	switch s {
	case "high":
		return PriorityHigh, true
	case "medium":
		return PriorityMedium, true
	case "low":
		return PriorityLow, true
	default:
		return PriorityMedium, false
	}
}

type TaskStatus string

const (
	StatusQueued    TaskStatus = "queued"
	StatusRunning   TaskStatus = "running"
	StatusSuccess TaskStatus = "success"
	StatusFailed  TaskStatus = "failed"
)

type FlowStatus string

const (
	FlowToSubmit  FlowStatus = "to_submit"
	FlowReviewing FlowStatus = "reviewing"
	FlowApproved  FlowStatus = "approved"
	FlowExecuting FlowStatus = "executing"
	FlowCompleted FlowStatus = "completed"
	FlowRejected  FlowStatus = "rejected"
)

type Task struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Type           string     `json:"type"`
	Priority       Priority   `json:"priority"`
	Timeout        time.Duration `json:"timeout"`
	MaxRetries    int        `json:"max_retries"`
	RetryCount   int        `json:"retry_count"`
	Status       TaskStatus `json:"status"`
	FlowStatus   FlowStatus `json:"flow_status"`
	ResourceID   string      `json:"resource_id"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	FailedReason  string      `json:"failed_reason"`
	IsFinalFailure bool       `json:"is_final_failure"`
}

type Dependency struct {
	TaskID       string `json:"task_id"`
	DependsOnID  string `json:"depends_on_id"`
}

type Resource struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

type ResourceRelation struct {
	FromID    string `json:"from_id"`
	ToID      string `json:"to_id"`
	Relation  string `json:"relation"`
}

type Stats struct {
	TotalTasks        int64         `json:"total_tasks"`
	QueuedCount     int64         `json:"queued_count"`
	RunningCount    int64         `json:"running_count"`
	SuccessCount    int64         `json:"success_count"`
	FailedCount     int64         `json:"failed_count"`
	AvgExecTime     time.Duration `json:"avg_exec_time"`
}

type ResourceSummary struct {
	ResourceID    string  `json:"resource_id"`
	ResourceName  string  `json:"resource_name"`
	ResourceType  string  `json:"resource_type"`
	TotalTasks    int64   `json:"total_tasks"`
	QueuedCount   int64   `json:"queued_count"`
	RunningCount  int64   `json:"running_count"`
	SuccessCount  int64   `json:"success_count"`
	FailedCount   int64   `json:"failed_count"`
}

type SubmitTaskRequest struct {
	Task
	Dependencies []string `json:"dependencies"`
}

type SubmitTaskRequestJSON struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Priority    string   `json:"priority"`
	TimeoutSec  int64    `json:"timeout_sec"`
	MaxRetries  int      `json:"max_retries"`
	ResourceID  string   `json:"resource_id"`
	Dependencies []string `json:"dependencies,omitempty"`
}

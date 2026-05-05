package common

import (
	"time"
)

type TaskStatus string

const (
	TaskStatusActive   TaskStatus = "active"
	TaskStatusPaused   TaskStatus = "paused"
	TaskStatusError    TaskStatus = "error"
	TaskStatusDeleted  TaskStatus = "deleted"
)

type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusSuccess   ExecutionStatus = "success"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusSkipped   ExecutionStatus = "skipped"
)

type TriggerType string

const (
	TriggerTypeScheduled TriggerType = "scheduled"
	TriggerTypeManual    TriggerType = "manual"
)

type Task struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	CronExpression  string            `json:"cron_expression"`
	NextRunTime     time.Time         `json:"next_run_time"`
	LastRunTime     *time.Time        `json:"last_run_time,omitempty"`
	Parameters      map[string]string `json:"parameters"`
	Status          TaskStatus        `json:"status"`
	RetryCount      int               `json:"retry_count"`
	ConsecutiveFailures int            `json:"consecutive_failures"`
	PausedAt        *time.Time        `json:"paused_at,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	DependsOn       []string          `json:"depends_on,omitempty"`
}

type Execution struct {
	ID             string           `json:"id"`
	TaskID         string           `json:"task_id"`
	TriggerType    TriggerType      `json:"trigger_type"`
	Status         ExecutionStatus  `json:"status"`
	StartTime      *time.Time       `json:"start_time,omitempty"`
	EndTime        *time.Time       `json:"end_time,omitempty"`
	DurationMs     int64            `json:"duration_ms"`
	ErrorMsg       string           `json:"error_msg,omitempty"`
	RetryNumber    int              `json:"retry_number"`
	CreatedAt      time.Time        `json:"created_at"`
}

type CreateTaskRequest struct {
	Name            string            `json:"name"`
	CronExpression  string            `json:"cron_expression"`
	Parameters      map[string]string `json:"parameters"`
	DependsOn       []string          `json:"depends_on,omitempty"`
}

type CreateTaskResponse struct {
	Task *Task  `json:"task"`
	Err  string `json:"err,omitempty"`
}

type GetTaskRequest struct {
	ID string `json:"id"`
}

type GetTaskResponse struct {
	Task *Task  `json:"task"`
	Err  string `json:"err,omitempty"`
}

type ListTasksRequest struct {
	Status *TaskStatus `json:"status,omitempty"`
}

type ListTasksResponse struct {
	Tasks []*Task `json:"tasks"`
	Err   string  `json:"err,omitempty"`
}

type UpdateTaskRequest struct {
	ID             string            `json:"id"`
	Name           *string           `json:"name,omitempty"`
	CronExpression *string           `json:"cron_expression,omitempty"`
	Parameters     map[string]string `json:"parameters,omitempty"`
}

type UpdateTaskResponse struct {
	Task *Task  `json:"task"`
	Err  string `json:"err,omitempty"`
}

type DeleteTaskRequest struct {
	ID string `json:"id"`
}

type DeleteTaskResponse struct {
	Err string `json:"err,omitempty"`
}

type PauseTaskRequest struct {
	ID string `json:"id"`
}

type PauseTaskResponse struct {
	Task *Task  `json:"task"`
	Err  string `json:"err,omitempty"`
}

type ResumeTaskRequest struct {
	ID string `json:"id"`
}

type ResumeTaskResponse struct {
	Task *Task  `json:"task"`
	Err  string `json:"err,omitempty"`
}

type ManualTriggerRequest struct {
	TaskID string `json:"task_id"`
}

type ManualTriggerResponse struct {
	ExecutionID string `json:"execution_id"`
	Err         string `json:"err,omitempty"`
}

type GetExecutionRequest struct {
	ID string `json:"id"`
}

type GetExecutionResponse struct {
	Execution *Execution `json:"execution"`
	Err       string     `json:"err,omitempty"`
}

type ListExecutionsRequest struct {
	TaskID      *string           `json:"task_id,omitempty"`
	Status      *ExecutionStatus  `json:"status,omitempty"`
	TriggerType *TriggerType      `json:"trigger_type,omitempty"`
}

type ListExecutionsResponse struct {
	Executions []*Execution `json:"executions"`
	Err        string       `json:"err,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

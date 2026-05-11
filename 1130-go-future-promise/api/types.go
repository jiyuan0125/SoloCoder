package api

import "time"

type TaskState string

const (
	StatePending   TaskState = "pending"
	StateRunning   TaskState = "running"
	StateCompleted TaskState = "completed"
	StateFailed    TaskState = "failed"
	StateCancelled TaskState = "cancelled"
	StateTimeout   TaskState = "timeout"
)

type TaskLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	TaskID    string    `json:"task_id"`
	Action    string    `json:"action"`
	Details   string    `json:"details,omitempty"`
}

type SubmitTaskRequest struct {
	Type     string            `json:"type"`
	Input    interface{}       `json:"input,omitempty"`
	Steps    []TaskStep        `json:"steps,omitempty"`
	Timeout  time.Duration     `json:"timeout,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type TaskStep struct {
	Name   string      `json:"name"`
	Input  interface{} `json:"input,omitempty"`
	Delay  time.Duration `json:"delay,omitempty"`
	Fail   bool        `json:"fail,omitempty"`
}

type SubmitTaskResponse struct {
	Success bool   `json:"success"`
	TaskID  string `json:"task_id"`
	Message string `json:"message,omitempty"`
}

type GetTaskResponse struct {
	Success   bool        `json:"success"`
	TaskID    string      `json:"task_id"`
	State     TaskState   `json:"state"`
	Input     interface{} `json:"input,omitempty"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	StartTime time.Time   `json:"start_time,omitempty"`
	EndTime   time.Time   `json:"end_time,omitempty"`
}

type CancelTaskResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GetLogsResponse struct {
	Success bool            `json:"success"`
	Logs    []TaskLogEntry  `json:"logs"`
}

package common

import "dag-scheduler/core"

type TaskDefinition struct {
	ID            string   `json:"id"`
	Dependencies  []string `json:"dependencies"`
	Duration      int      `json:"duration,omitempty"`
	Retries       *int     `json:"retries,omitempty"`
	RetryInterval *int     `json:"retry_interval,omitempty"`
	ShouldFail    bool     `json:"should_fail,omitempty"`
}

type SubmitRequest struct {
	Tasks                []TaskDefinition `json:"tasks"`
	MaxConcurrency       int              `json:"max_concurrency,omitempty"`
	DefaultRetries       int              `json:"default_retries,omitempty"`
	DefaultRetryInterval int              `json:"default_retry_interval,omitempty"`
}

type SubmitResponse struct {
	Success bool     `json:"success"`
	JobID   string   `json:"job_id,omitempty"`
	Error   string   `json:"error,omitempty"`
	Cycle   []string `json:"cycle,omitempty"`
}

type StartRequest struct {
	JobID string `json:"job_id"`
}

type StartResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type StatusRequest struct {
	JobID string `json:"job_id"`
}

type TaskInfo struct {
	ID           string          `json:"id"`
	Status       core.TaskStatus `json:"status"`
	StartTime    int64           `json:"start_time,omitempty"`
	EndTime      int64           `json:"end_time,omitempty"`
	DurationMs   int64           `json:"duration_ms,omitempty"`
	Attempts     int             `json:"attempts,omitempty"`
	LastError    string          `json:"last_error,omitempty"`
	Dependencies []string        `json:"dependencies,omitempty"`
}

type StatusResponse struct {
	Success   bool                `json:"success"`
	JobID     string              `json:"job_id"`
	IsRunning bool                `json:"is_running"`
	Tasks     map[string]TaskInfo `json:"tasks"`
	Error     string              `json:"error,omitempty"`
}

type CancelRequest struct {
	JobID string `json:"job_id"`
}

type CancelResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

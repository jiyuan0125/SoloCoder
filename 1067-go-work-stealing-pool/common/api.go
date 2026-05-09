package common

import "time"

type SubmitTaskRequest struct {
	TaskType string        `json:"task_type"`
	Payload  interface{}   `json:"payload"`
	Timeout  time.Duration `json:"timeout"`
}

type SubmitTaskResponse struct {
	TaskID  string `json:"task_id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetTaskResultRequest struct {
	TaskID string `json:"task_id"`
}

type GetTaskResultResponse struct {
	TaskID  string      `json:"task_id"`
	Status  string      `json:"status"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
	Done    bool        `json:"done"`
}

type PoolStatsResponse struct {
	WorkerCount int    `json:"worker_count"`
	Status      string `json:"status"`
}

type AdjustWorkersRequest struct {
	Add    int `json:"add"`
	Remove int `json:"remove"`
}

type ShutdownRequest struct {
	Force bool `json:"force"`
}

type ShutdownResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

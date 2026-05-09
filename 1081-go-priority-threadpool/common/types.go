package common

type Priority int

const (
	Low Priority = iota
	Medium
	High
)

type SubmitRequest struct {
	Priority Priority `json:"priority"`
	Task     string   `json:"task"`
}

type SubmitResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type StatusResponse struct {
	WorkerCount        int    `json:"worker_count"`
	IdleWorkerCount  int    `json:"idle_worker_count"`
	QueueHighPriority int    `json:"queue_high_priority"`
	QueueMediumPriority int `json:"queue_medium_priority"`
	QueueLowPriority    int `json:"queue_low_priority"`
	StarvedTasksCount    int    `json:"starved_tasks_count"`
	Success            bool   `json:"success"`
}

type ShutdownResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

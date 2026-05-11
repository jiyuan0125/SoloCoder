package api

import "time"

type AddTaskRequest struct {
	ID       string        `json:"id"`
	Callback string        `json:"callback"`
	Delay    time.Duration `json:"delay"`
}

type AddTaskResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type CancelTaskRequest struct {
	ID string `json:"id"`
}

type CancelTaskResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ResetTaskRequest struct {
	ID       string        `json:"id"`
	NewDelay time.Duration `json:"new_delay"`
}

type ResetTaskResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetTaskRequest struct {
	ID string `json:"id"`
}

type GetTaskResponse struct {
	Success bool       `json:"success"`
	Task    *TaskInfo  `json:"task,omitempty"`
	Error   string     `json:"error,omitempty"`
}

type TaskInfo struct {
	ID          string        `json:"id"`
	Callback    string        `json:"callback"`
	Delay       time.Duration `json:"delay"`
	CreatedAt   time.Time     `json:"created_at"`
	ExpireAt    time.Time     `json:"expire_at"`
	Remaining   time.Duration `json:"remaining"`
	Status      string        `json:"status"`
}

type ListTasksResponse struct {
	Success bool        `json:"success"`
	Tasks   []*TaskInfo `json:"tasks,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type StatsResponse struct {
	Success        bool          `json:"success"`
	TotalPending   int           `json:"total_pending"`
	ExecutedCount  int64         `json:"executed_count"`
	CancelledCount int64         `json:"cancelled_count"`
	HourTasks      int           `json:"hour_tasks"`
	MinuteTasks    int           `json:"minute_tasks"`
	SecondTasks    int           `json:"second_tasks"`
	HourTick       int           `json:"hour_tick"`
	MinuteTick     int           `json:"minute_tick"`
	SecondTick     int           `json:"second_tick"`
	Error          string        `json:"error,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

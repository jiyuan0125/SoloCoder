package api

import "time"

type Priority int

const (
	PriorityLow    Priority = 1
	PriorityNormal Priority = 2
	PriorityHigh   Priority = 3
	PriorityUrgent Priority = 4
)

type SubmitRequest struct {
	ID        string    `json:"id"`
	Payload   string    `json:"payload"`
	ExecuteAt time.Time `json:"execute_at"`
	Priority  Priority  `json:"priority"`
}

type SubmitResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ModifyRequest struct {
	ID        string    `json:"id"`
	ExecuteAt time.Time `json:"execute_at"`
}

type ModifyResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type PromoteRequest struct {
	ID       string   `json:"id"`
	Priority Priority `json:"priority"`
}

type PromoteResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type Task struct {
	ID        string    `json:"id"`
	Payload   string    `json:"payload"`
	ExecuteAt time.Time `json:"execute_at"`
	Priority  Priority  `json:"priority"`
}

type PollResponse struct {
	Task *Task `json:"task,omitempty"`
}

type PeekResponse struct {
	Task *Task `json:"task,omitempty"`
}

type DrainResponse struct {
	Tasks []*Task `json:"tasks"`
}

type StatusResponse struct {
	Queued int `json:"queued"`
}

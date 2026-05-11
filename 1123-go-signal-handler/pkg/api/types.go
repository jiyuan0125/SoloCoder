package api

type CallbackStatus string

const (
	StatusPending    CallbackStatus = "pending"
	StatusRunning    CallbackStatus = "running"
	StatusCompleted  CallbackStatus = "completed"
	StatusFailed     CallbackStatus = "failed"
	StatusSkipped    CallbackStatus = "skipped"
)

type CallbackInfo struct {
	Name     string         `json:"name"`
	Order    int            `json:"order"`
	Status   CallbackStatus `json:"status"`
	Duration string         `json:"duration,omitempty"`
	Error    string         `json:"error,omitempty"`
	StartTime string        `json:"start_time,omitempty"`
	EndTime   string        `json:"end_time,omitempty"`
}

type ShutdownState string

const (
	StateIdle       ShutdownState = "idle"
	StateGraceful   ShutdownState = "graceful"
	StateCompleted  ShutdownState = "completed"
	StateForceExit  ShutdownState = "force_exit"
)

type StatusResponse struct {
	State     ShutdownState  `json:"state"`
	Callbacks []CallbackInfo `json:"callbacks"`
	Message   string         `json:"message,omitempty"`
	StartTime string         `json:"start_time,omitempty"`
}

type RegisterRequest struct {
	Name   string `json:"name"`
	Order  int    `json:"order"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ShutdownResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

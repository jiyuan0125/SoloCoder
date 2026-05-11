package api

import "time"

type AcquireRequest struct {
	Name     string `json:"name"`
	HolderID string `json:"holder_id"`
	ExpireMs int64  `json:"expire_ms"`
}

type ReleaseRequest struct {
	Name     string `json:"name"`
	HolderID string `json:"holder_id"`
}

type RenewRequest struct {
	Name     string `json:"name"`
	HolderID string `json:"holder_id"`
}

type LockInfo struct {
	Name          string    `json:"name"`
	HolderID      string    `json:"holder_id"`
	ExpireMs      int64     `json:"expire_ms"`
	CreatedAt     time.Time `json:"created_at"`
	LastRenewedAt time.Time `json:"last_renewed_at"`
	RemainingMs   int64     `json:"remaining_ms"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Success: true,
		Data:    data,
	}
}

func NewErrorResponse(message string) *Response {
	return &Response{
		Success: false,
		Message: message,
	}
}

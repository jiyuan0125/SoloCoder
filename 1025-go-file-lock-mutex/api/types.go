package api

import "time"

type LockRequest struct {
	FilePath string        `json:"file_path"`
	Mode     string        `json:"mode"`
	Timeout  time.Duration `json:"timeout"`
}

type LockResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	LockID  string `json:"lock_id,omitempty"`
}

type UnlockRequest struct {
	FilePath string `json:"file_path"`
}

type UnlockResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type StatusRequest struct {
	FilePath string `json:"file_path"`
}

type StatusResponse struct {
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	IsHeld       bool   `json:"is_held"`
	Mode         string `json:"mode,omitempty"`
	LockFilePath string `json:"lock_file_path"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

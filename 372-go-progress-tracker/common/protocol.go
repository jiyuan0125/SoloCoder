package common

import (
	"time"
)

type ProgressStatus struct {
	ID           string        `json:"id"`
	Current      int64         `json:"current"`
	Total        int64         `json:"total"`
	Percentage   float64       `json:"percentage"`
	Remaining    time.Duration `json:"remaining_seconds"`
	Description  string        `json:"description"`
	IsDone       bool          `json:"is_done"`
	IsCancelled  bool          `json:"is_cancelled"`
	CreatedAt    time.Time     `json:"created_at"`
}

type CreateProgressRequest struct {
	Total       int64  `json:"total"`
	Description string `json:"description"`
}

type CreateProgressResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type UpdateProgressRequest struct {
	Delta int64 `json:"delta"`
}

type SetProgressRequest struct {
	Current int64 `json:"current"`
}

type ListProgressResponse struct {
	Progresses []ProgressStatus `json:"progresses"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type AddSubTaskRequest struct {
	ParentID string  `json:"parent_id"`
	Total    int64   `json:"total"`
	Weight   float64 `json:"weight"`
}

type AddSubTaskResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

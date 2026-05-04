package models

import "time"

type FeedbackType string

const (
	FeedbackTypeFeature FeedbackType = "feature"
	FeedbackTypeBug     FeedbackType = "bug"
	FeedbackTypeComplaint FeedbackType = "complaint"
)

var ValidFeedbackTypes = []FeedbackType{
	FeedbackTypeFeature,
	FeedbackTypeBug,
	FeedbackTypeComplaint,
}

type FeedbackStatus string

const (
	StatusPending   FeedbackStatus = "pending"
	StatusProcessing FeedbackStatus = "processing"
	StatusClosed    FeedbackStatus = "closed"
)

var ValidFeedbackStatuses = []FeedbackStatus{
	StatusPending,
	StatusProcessing,
	StatusClosed,
}

type Feedback struct {
	ID             int64          `db:"id"`
	UserID         string         `db:"user_id"`
	FeedbackType   FeedbackType   `db:"feedback_type"`
	Rating         int            `db:"rating"`
	Description    string         `db:"description"`
	Status         FeedbackStatus `db:"status"`
	InternalNote   string         `db:"internal_note"`
	ProcessingNote string         `db:"processing_note"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

type FeedbackStatistics struct {
	Type         FeedbackType `db:"feedback_type"`
	Count        int          `db:"count"`
	AverageRating float64     `db:"average_rating"`
}

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type CreateFeedbackRequest struct {
	UserID       string       `json:"user_id"`
	FeedbackType string       `json:"feedback_type"`
	Rating       interface{}  `json:"rating"`
	Description  string       `json:"description"`
}

type UpdateFeedbackStatusRequest struct {
	Status         string `json:"status"`
	ProcessingNote string `json:"processing_note"`
}

type AddInternalNoteRequest struct {
	Note string `json:"note"`
}

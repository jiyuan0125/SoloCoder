package protocol

import "time"

type FeedbackType string

const (
	FeedbackTypeBug         FeedbackType = "bug"
	FeedbackTypeFeature     FeedbackType = "feature"
	FeedbackTypeComplaint   FeedbackType = "complaint"
	FeedbackTypeConsultation FeedbackType = "consultation"
)

type FeedbackPriority string

const (
	PriorityLow    FeedbackPriority = "low"
	PriorityMedium FeedbackPriority = "medium"
	PriorityHigh   FeedbackPriority = "high"
	PriorityUrgent FeedbackPriority = "urgent"
)

type FeedbackStatus string

const (
	StatusPending    FeedbackStatus = "pending"
	StatusProcessing FeedbackStatus = "processing"
	StatusResolved   FeedbackStatus = "resolved"
	StatusClosed     FeedbackStatus = "closed"
)

type Comment struct {
	ID        string    `json:"id"`
	FeedbackID string   `json:"feedback_id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Content   string    `json:"content"`
	IsFromSupport bool  `json:"is_from_support"`
	CreatedAt time.Time `json:"created_at"`
}

type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	IsSystem  bool      `json:"is_system"`
}

type Feedback struct {
	ID          string            `json:"id"`
	UserID      string            `json:"user_id"`
	UserName    string            `json:"user_name"`
	Type        FeedbackType      `json:"type"`
	Content     string            `json:"content"`
	Priority    FeedbackPriority  `json:"priority"`
	Status      FeedbackStatus    `json:"status"`
	HandlerID   string            `json:"handler_id"`
	HandlerName string            `json:"handler_name"`
	Comments    []*Comment        `json:"comments"`
	Tags        []*Tag            `json:"tags"`
	TagIDs      []string          `json:"tag_ids"`
	IsEscalated bool              `json:"is_escalated"`
	EscalatedAt time.Time         `json:"escalated_at"`
	FirstResponseAt time.Time     `json:"first_response_at"`
	ResolvedAt  time.Time         `json:"resolved_at"`
	ClosedAt    time.Time         `json:"closed_at"`
	LastUpdatedAt time.Time       `json:"last_updated_at"`
	LastNotifiedAt time.Time      `json:"last_notified_at"`
	MergedFrom   []string         `json:"merged_from"`
	MergedInto   string           `json:"merged_into"`
	InReviewQueue bool            `json:"in_review_queue"`
	IsInvalid    bool             `json:"is_invalid"`
	CreatedAt    time.Time        `json:"created_at"`
}

type UserLimitInfo struct {
	UserID             string        `json:"user_id"`
	InvalidCloseCount  int           `json:"invalid_close_count"`
	IsUnderReview      bool          `json:"is_under_review"`
	ReviewStartedAt    time.Time     `json:"review_started_at"`
}

type CreateFeedbackRequest struct {
	UserID   string         `json:"user_id"`
	UserName string         `json:"user_name"`
	Type     FeedbackType   `json:"type"`
	Content  string         `json:"content"`
}

type CreateFeedbackResponse struct {
	Feedback    *Feedback `json:"feedback"`
	IsMerged    bool      `json:"is_merged"`
	MergedInto  string    `json:"merged_into,omitempty"`
	InReview    bool      `json:"in_review"`
}

type AddCommentRequest struct {
	FeedbackID    string `json:"feedback_id"`
	UserID        string `json:"user_id"`
	UserName      string `json:"user_name"`
	Content       string `json:"content"`
	IsFromSupport bool   `json:"is_from_support"`
}

type UpdateStatusRequest struct {
	FeedbackID string         `json:"feedback_id"`
	Status     FeedbackStatus `json:"status"`
	HandlerID  string         `json:"handler_id"`
	HandlerName string        `json:"handler_name"`
	IsInvalid  bool           `json:"is_invalid"`
}

type AddTagRequest struct {
	FeedbackID string `json:"feedback_id"`
	TagID      string `json:"tag_id"`
}

type RemoveTagRequest struct {
	FeedbackID string `json:"feedback_id"`
	TagID      string `json:"tag_id"`
}

type CreateTagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type AssignHandlerRequest struct {
	FeedbackID  string `json:"feedback_id"`
	HandlerID   string `json:"handler_id"`
	HandlerName string `json:"handler_name"`
}

type ListFeedbackRequest struct {
	Status     FeedbackStatus   `json:"status,omitempty"`
	Type       FeedbackType     `json:"type,omitempty"`
	Priority   FeedbackPriority `json:"priority,omitempty"`
	UserID     string           `json:"user_id,omitempty"`
	HandlerID  string           `json:"handler_id,omitempty"`
	TagID      string           `json:"tag_id,omitempty"`
	InReview   *bool            `json:"in_review,omitempty"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
}

type ListFeedbackResponse struct {
	Feedbacks []*Feedback `json:"feedbacks"`
	Total     int         `json:"total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
}

type MonthlyReport struct {
	Month                  string                  `json:"month"`
	TotalFeedbacks         int                     `json:"total_feedbacks"`
	ByType                 map[FeedbackType]int    `json:"by_type"`
	ByStatus               map[FeedbackStatus]int  `json:"by_status"`
	ByPriority             map[FeedbackPriority]int`json:"by_priority"`
	AvgResolutionTime      float64                 `json:"avg_resolution_time_hours"`
	AvgFirstResponseTime   float64                 `json:"avg_first_response_time_hours"`
	SLAComplianceRate      float64                 `json:"sla_compliance_rate"`
	TopTags                []TagCount              `json:"top_tags"`
	MergedCount            int                     `json:"merged_count"`
	EscalatedCount         int                     `json:"escalated_count"`
	InvalidCount           int                     `json:"invalid_count"`
	TrendAnalysis          string                  `json:"trend_analysis"`
	ImprovementSuggestions []string                `json:"improvement_suggestions"`
	CreatedAt              time.Time               `json:"created_at"`
}

type TagCount struct {
	TagID   string `json:"tag_id"`
	TagName string `json:"tag_name"`
	Count   int    `json:"count"`
}

type KPIStats struct {
	HandlerID           string  `json:"handler_id"`
	HandlerName         string  `json:"handler_name"`
	Period              string  `json:"period"`
	TotalHandled        int     `json:"total_handled"`
	AvgResolutionTime   float64 `json:"avg_resolution_time_hours"`
	AvgFirstResponseTime float64`json:"avg_first_response_time_hours"`
	SLAComplianceRate   float64 `json:"sla_compliance_rate"`
	ResolvedCount       int     `json:"resolved_count"`
	ClosedCount         int     `json:"closed_count"`
	EscalatedCount      int     `json:"escalated_count"`
}

type Notification struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	TargetID   string    `json:"target_id"`
	Message    string    `json:"message"`
	UserID     string    `json:"user_id"`
	Read       bool      `json:"read"`
	CreatedAt  time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

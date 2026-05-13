package models

import "time"

type AuditLog struct {
	ID          int64     `json:"id" db:"id"`
	Operator    string    `json:"operator" db:"operator"`
	ResourceType string   `json:"resource_type" db:"resource_type"`
	Operation   string    `json:"operation" db:"operation"`
	ResourceID  string    `json:"resource_id" db:"resource_id"`
	BeforeValue string    `json:"before_value" db:"before_value"`
	AfterValue  string    `json:"after_value" db:"after_value"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	Archived    bool      `json:"archived" db:"archived"`
	ArchiveFile string    `json:"archive_file,omitempty" db:"archive_file"`
}

type AuditLogSummary struct {
	ID          int64     `json:"id" db:"id"`
	Operator    string    `json:"operator" db:"operator"`
	ResourceType string   `json:"resource_type" db:"resource_type"`
	Operation   string    `json:"operation" db:"operation"`
	ResourceID  string    `json:"resource_id" db:"resource_id"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	Archived    bool      `json:"archived" db:"archived"`
	ArchiveFile string    `json:"archive_file,omitempty" db:"archive_file"`
}

type CreateAuditLogRequest struct {
	Operator     string `json:"operator" binding:"required"`
	ResourceType string `json:"resource_type" binding:"required"`
	Operation    string `json:"operation" binding:"required"`
	ResourceID   string `json:"resource_id" binding:"required"`
	BeforeValue  string `json:"before_value"`
	AfterValue   string `json:"after_value"`
}

type QueryAuditLogRequest struct {
	Operator     string    `form:"operator"`
	ResourceType string    `form:"resource_type"`
	StartTime    time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime      time.Time `form:"end_time" time_format:"2006-01-02T15:04:05Z07:00"`
	Page         int       `form:"page,default=1"`
	PageSize     int       `form:"page_size,default=20"`
}

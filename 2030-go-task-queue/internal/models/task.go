package models

import (
	"time"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusSuccess   TaskStatus = "success"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID              int64      `json:"id" gorm:"primaryKey"`
	Type            string     `json:"type" gorm:"not null"`
	Priority        int        `json:"priority" gorm:"not null;default:5"`
	Payload         string     `json:"payload" gorm:"type:text;not null"`
	Status          TaskStatus `json:"status" gorm:"not null;default:pending"`
	RetryCount      int        `json:"retry_count" gorm:"default:0"`
	TimeoutCount    int        `json:"timeout_count" gorm:"default:0"`
	MaxRetries      int        `json:"max_retries" gorm:"default:3"`
	TimeoutSeconds  int        `json:"timeout_seconds" gorm:"default:1800"`
	Result          string     `json:"result" gorm:"type:text"`
	Error           string     `json:"error" gorm:"type:text"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ProcessingAt    *time.Time `json:"processing_at"`
	CompletedAt     *time.Time `json:"completed_at"`
}

func (Task) TableName() string {
	return "tasks"
}

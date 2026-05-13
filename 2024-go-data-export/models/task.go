package models

import "time"

type ExportTask struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       string    `json:"user_id" gorm:"index"`
	TableName    string    `json:"table_name"`
	Fields       string    `json:"fields"`
	Filter       string    `json:"filter"`
	Format       string    `json:"format"`
	Status       string    `json:"status"`
	Progress     int       `json:"progress"`
	FilePath     string    `json:"file_path"`
	FileName     string    `json:"file_name"`
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ExportRequest struct {
	UserID    string   `json:"user_id" binding:"required"`
	TableName string   `json:"table_name" binding:"required"`
	Fields    []string `json:"fields" binding:"required"`
	Filter    Filter   `json:"filter"`
	Format    string   `json:"format" binding:"required"`
}

type Filter struct {
	Conditions string `json:"conditions"`
}

type ProgressResponse struct {
	ID       uint   `json:"id"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	FileName string `json:"file_name,omitempty"`
}

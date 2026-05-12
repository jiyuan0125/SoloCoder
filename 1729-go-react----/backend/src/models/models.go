package models

import (
	"time"
)

type Department struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DepartmentID string `json:"department_id"`
}

type ProjectStatus string

const (
	ProjectStatusPreparing  ProjectStatus = "preparing"
	ProjectStatusInProgress ProjectStatus = "in_progress"
	ProjectStatusCompleted  ProjectStatus = "completed"
)

type Project struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	LeadDepartment   string        `json:"lead_department"`
	ParticipatingDepts []string     `json:"participating_depts"`
	LeaderID         string        `json:"leader_id"`
	StartDate        time.Time     `json:"start_date"`
	EndDate          time.Time     `json:"end_date"`
	Status           ProjectStatus `json:"status"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

type TaskPriority string

const (
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityLow    TaskPriority = "low"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type Task struct {
	ID         string        `json:"id"`
	ProjectID  string        `json:"project_id"`
	Title      string        `json:"title"`
	AssigneeID string        `json:"assignee_id"`
	DueDate    time.Time     `json:"due_date"`
	Priority   TaskPriority  `json:"priority"`
	Status     TaskStatus    `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type DataFormat string

const (
	DataFormatCSV   DataFormat = "csv"
	DataFormatJSON  DataFormat = "json"
	DataFormatExcel DataFormat = "excel"
	DataFormatImage DataFormat = "image"
	DataFormatOther DataFormat = "other"
)

type SharedData struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Format      DataFormat `json:"format"`
	FileSize    int64      `json:"file_size"`
	ProjectID   string     `json:"project_id"`
	UploaderID  string     `json:"uploader_id"`
	CreatedAt   time.Time  `json:"created_at"`
}

type DownloadLog struct {
	ID         string    `json:"id"`
	DataID     string    `json:"data_id"`
	UserID     string    `json:"user_id"`
	DownloadedAt time.Time `json:"downloaded_at"`
}

type AchievementType string

const (
	AchievementTypePaper     AchievementType = "paper"
	AchievementTypePatent    AchievementType = "patent"
	AchievementTypeSoftware  AchievementType = "software"
	AchievementTypeReport    AchievementType = "report"
)

type AchievementStatus string

const (
	AchievementStatusSubmitted AchievementStatus = "submitted"
	AchievementStatusPublished AchievementStatus = "published"
	AchievementStatusAuthorized AchievementStatus = "authorized"
)

type ProjectContribution struct {
	ProjectID string `json:"project_id"`
	Ratio     int    `json:"ratio"`
}

type Achievement struct {
	ID            string                 `json:"id"`
	Type          AchievementType        `json:"type"`
	Title         string                 `json:"title"`
	OutputDate    time.Time              `json:"output_date"`
	Participants  []string               `json:"participants"`
	Status        AchievementStatus      `json:"status"`
	Contributions []ProjectContribution  `json:"contributions"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type ApprovalStatus string

const (
	ApprovalStatusSubmitted ApprovalStatus = "submitted"
	ApprovalStatusReview1   ApprovalStatus = "review_1"
	ApprovalStatusReview2   ApprovalStatus = "review_2"
	ApprovalStatusFinal     ApprovalStatus = "final"
	ApprovalStatusApproved  ApprovalStatus = "approved"
	ApprovalStatusRejected  ApprovalStatus = "rejected"
)

type Approval struct {
	ID          string         `json:"id"`
	ReferenceID string         `json:"reference_id"`
	Type        string         `json:"type"`
	Status      ApprovalStatus `json:"status"`
	Amount      float64        `json:"amount"`
	History     []ApprovalStep `json:"history"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type ApprovalStep struct {
	Status     ApprovalStatus `json:"status"`
	ActionBy   string         `json:"action_by"`
	ActionAt   time.Time      `json:"action_at"`
	Comments   string         `json:"comments"`
	OldAmount  *float64       `json:"old_amount,omitempty"`
	NewAmount  *float64       `json:"new_amount,omitempty"`
}

type TodoItem struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	RefID     string    `json:"ref_id"`
	RefType   string    `json:"ref_type"`
	CreatedAt time.Time `json:"created_at"`
}

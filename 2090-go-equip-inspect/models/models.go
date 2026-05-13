package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending       TaskStatus = "pending"
	TaskStatusInspecting    TaskStatus = "inspecting"
	TaskStatusNormal        TaskStatus = "normal"
	TaskStatusAbnormal      TaskStatus = "abnormal"
	TaskStatusPendingReview TaskStatus = "pending_review"
	TaskStatusReviewPass    TaskStatus = "review_pass"
	TaskStatusReviewFail    TaskStatus = "review_fail"
	TaskStatusClosed        TaskStatus = "closed"
)

func (ts TaskStatus) Value() (driver.Value, error) {
	return string(ts), nil
}

func (ts *TaskStatus) Scan(value interface{}) error {
	if value == nil {
		return errors.New("task status cannot be nil")
	}
	sv, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid task status type: %T", value)
	}
	*ts = TaskStatus(sv)
	return nil
}

type UserRole string

const (
	UserRoleInspector UserRole = "inspector"
	UserRoleRepair    UserRole = "repair"
	UserRoleReviewer  UserRole = "reviewer"
	UserRoleAdmin     UserRole = "admin"
)

type User struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username"`
	Password  string     `json:"-"`
	Role      UserRole   `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Device struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CheckItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxScore    int    `json:"max_score"`
}

type InspectionPoint struct {
	ID         int64     `json:"id"`
	DeviceID   int64     `json:"device_id"`
	Name       string    `json:"name"`
	OrderIndex int       `json:"order_index"`
	CreatedAt  time.Time `json:"created_at"`
}

type InspectionPlan struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Frequency      string    `json:"frequency"`
	StartDate      time.Time `json:"start_date"`
	LastGenerateAt *time.Time `json:"last_generate_at,omitempty"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PlanPoint struct {
	ID        int64 `json:"id"`
	PlanID    int64 `json:"plan_id"`
	PointID   int64 `json:"point_id"`
	OrderIndex int  `json:"order_index"`
}

type InspectionTask struct {
	ID             int64       `json:"id"`
	PlanID         int64       `json:"plan_id"`
	Code           string      `json:"code"`
	Status         TaskStatus  `json:"status"`
	InspectorID    *int64      `json:"inspector_id,omitempty"`
	AssignedAt     *time.Time  `json:"assigned_at,omitempty"`
	StartedAt      *time.Time  `json:"started_at,omitempty"`
	CompletedAt    *time.Time  `json:"completed_at,omitempty"`
	ReviewerID     *int64      `json:"reviewer_id,omitempty"`
	ReviewedAt     *time.Time  `json:"reviewed_at,omitempty"`
	ReviewComment  string      `json:"review_comment,omitempty"`
	DueDate        time.Time   `json:"due_date"`
	TotalScore     *int        `json:"total_score,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type TaskPoint struct {
	ID              int64      `json:"id"`
	TaskID          int64      `json:"task_id"`
	PointID         int64      `json:"point_id"`
	OrderIndex      int        `json:"order_index"`
	Checked         bool       `json:"checked"`
	CheckedAt       *time.Time `json:"checked_at,omitempty"`
	CheckedBy       *int64     `json:"checked_by,omitempty"`
}

type InspectionRecord struct {
	ID          int64      `json:"id"`
	TaskPointID int64      `json:"task_point_id"`
	CheckItemID int64      `json:"check_item_id"`
	Score       int        `json:"score"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
}

type RepairOrder struct {
	ID          int64     `json:"id"`
	TaskID      int64     `json:"task_id"`
	DeviceID    int64     `json:"device_id"`
	RepairerID  *int64    `json:"repairer_id,omitempty"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	AssignedAt  *time.Time `json:"assigned_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Statistics struct {
	ID                int64     `json:"id"`
	Date              time.Time `json:"date"`
	TotalTasks        int       `json:"total_tasks"`
	PendingTasks      int       `json:"pending_tasks"`
	CompletedTasks    int       `json:"completed_tasks"`
	AbnormalTasks     int       `json:"abnormal_tasks"`
	NormalTasks       int       `json:"normal_tasks"`
	ReviewPassTasks   int       `json:"review_pass_tasks"`
	ReviewFailTasks   int       `json:"review_fail_tasks"`
	ClosedTasks       int       `json:"closed_tasks"`
	OverdueTasks      int       `json:"overdue_tasks"`
	UpdatedAt         time.Time `json:"updated_at"`
}

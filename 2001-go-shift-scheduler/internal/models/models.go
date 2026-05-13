package models

import "time"

type ShiftStatus string

const (
	StatusDraft       ShiftStatus = "draft"
	StatusReviewing   ShiftStatus = "reviewing"
	StatusApproved    ShiftStatus = "approved"
	StatusInProgress  ShiftStatus = "in_progress"
	StatusCompleted   ShiftStatus = "completed"
	StatusRejected    ShiftStatus = "rejected"
)

type Department struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Employee struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	DepartmentID int64     `json:"department_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type Holiday struct {
	ID   int64     `json:"id"`
	Date time.Time `json:"date"`
	Name string    `json:"name"`
}

type ShiftMaster struct {
	ID          int64       `json:"id"`
	EmployeeID  int64       `json:"employee_id"`
	ShiftDate   time.Time   `json:"shift_date"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     time.Time   `json:"end_time"`
	Position    string      `json:"position"`
	Status      ShiftStatus `json:"status"`
	ScheduledHours float64   `json:"scheduled_hours"`
	ActualHours    float64   `json:"actual_hours"`
	IsHoliday      bool      `json:"is_holiday"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type ShiftDetail struct {
	ID         int64       `json:"id"`
	MasterID   int64       `json:"master_id"`
	EmployeeID int64       `json:"employee_id"`
	ShiftDate  time.Time   `json:"shift_date"`
	StartTime  time.Time   `json:"start_time"`
	EndTime    time.Time   `json:"end_time"`
	Position   string      `json:"position"`
	Status     ShiftStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
}

type OperationHistory struct {
	ID         int64     `json:"id"`
	MasterID   int64     `json:"master_id"`
	OperatorID int64     `json:"operator_id"`
	Action     string    `json:"action"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

type SwapRequest struct {
	ID              int64     `json:"id"`
	RequesterShiftID int64    `json:"requester_shift_id"`
	ResponderShiftID int64    `json:"responder_shift_id"`
	RequesterID     int64     `json:"requester_id"`
	ResponderID     int64     `json:"responder_id"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	ConfirmedAt     *time.Time `json:"confirmed_at"`
}

type MonthlyStats struct {
	EmployeeID     int64   `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	DepartmentID   int64   `json:"department_id"`
	DepartmentName string  `json:"department_name"`
	Month          string  `json:"month"`
	ScheduledHours float64 `json:"scheduled_hours"`
	ActualHours    float64 `json:"actual_hours"`
	Difference     float64 `json:"difference"`
}

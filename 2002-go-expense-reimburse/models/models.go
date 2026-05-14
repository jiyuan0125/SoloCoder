package models

import "time"

type ExpenseType string

const (
	Travel    ExpenseType = "travel"
	Meal      ExpenseType = "meal"
	Office    ExpenseType = "office"
	Other     ExpenseType = "other"
)

type ApprovalRole string

const (
	RoleManager    ApprovalRole = "manager"
	RoleDepartment ApprovalRole = "department_manager"
	RoleCFO        ApprovalRole = "cfo"
)

type ReimbursementStatus string

const (
	StatusPending    ReimbursementStatus = "pending"
	StatusApproved   ReimbursementStatus = "approved"
	StatusRejected   ReimbursementStatus = "rejected"
	StatusExpired    ReimbursementStatus = "expired"
	StatusCompleted  ReimbursementStatus = "completed"
)

type Reimbursement struct {
	ID                  int64               `json:"id"`
	EmployeeID          string              `json:"employee_id"`
	AmountCent          int64               `json:"amount_cent"`
	AmountYuan          float64             `json:"amount_yuan"`
	ExpenseType         ExpenseType         `json:"expense_type"`
	OccurredDate        time.Time           `json:"occurred_date"`
	Description         string              `json:"description"`
	Status              ReimbursementStatus `json:"status"`
	CurrentApprovalStep int                 `json:"current_approval_step"`
	ApprovalChain       []ApprovalRole      `json:"approval_chain"`
	ModifyCount         int                 `json:"modify_count"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	Stages              []*Stage            `json:"stages,omitempty"`
}

type Stage struct {
	ID               int64        `json:"id"`
	ReimbursementID  int64        `json:"reimbursement_id"`
	ApprovalRole     ApprovalRole `json:"approval_role"`
	AmountCent       int64        `json:"amount_cent"`
	AmountYuan       float64      `json:"amount_yuan"`
	ApprovalStep     int          `json:"approval_step"`
}

type StageDB struct {
	ID              int64
	ReimbursementID int64
	ApprovalRole    string
	AmountCent      int64
	ApprovalStep    int
}

type StatusHistory struct {
	ID             int64                  `json:"id"`
	ReimbursementID int64                 `json:"reimbursement_id"`
	FromStatus     ReimbursementStatus    `json:"from_status"`
	ToStatus       ReimbursementStatus    `json:"to_status"`
	ChangedBy      string                 `json:"changed_by"`
	ChangedAt      time.Time              `json:"changed_at"`
	Note           string                 `json:"note,omitempty"`
}

type SubmitRequest struct {
	EmployeeID   string      `json:"employee_id" binding:"required"`
	AmountYuan   float64     `json:"amount_yuan" binding:"required,gt=0"`
	ExpenseType  ExpenseType `json:"expense_type" binding:"required,oneof=travel meal office other"`
	OccurredDate string      `json:"occurred_date" binding:"required"`
	Description  string      `json:"description"`
}

type ApprovalRequest struct {
	ApproverID   string `json:"approver_id" binding:"required"`
	ApproverRole string `json:"approver_role" binding:"required,oneof=manager department_manager cfo"`
	Action       string `json:"action" binding:"required,oneof=approve reject"`
	Note         string `json:"note"`
}

type UpdateRequest struct {
	ReimbursementID int64   `json:"reimbursement_id" binding:"required"`
	EmployeeID      string  `json:"employee_id" binding:"required"`
	AmountYuan      float64 `json:"amount_yuan" binding:"required,gt=0"`
}

package core

import "time"

type ServiceWindow struct {
	StartTime string
	EndTime   string
	Weekdays  []int
}

type ContractStatus string

const (
	ContractStatusActive   ContractStatus = "active"
	ContractStatusCompleted ContractStatus = "completed"
	ContractStatusCancelled ContractStatus = "cancelled"
)

type MilestoneStatus string

const (
	MilestoneStatusPending    MilestoneStatus = "pending"
	MilestoneStatusInProgress MilestoneStatus = "in_progress"
	MilestoneStatusCompleted  MilestoneStatus = "completed"
	MilestoneStatusApproved   MilestoneStatus = "approved"
)

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusDue      PaymentStatus = "due"
	PaymentStatusPaid     PaymentStatus = "paid"
)

type Contract struct {
	ID            string
	ContractNo    string
	PartyA        string
	PartyB        string
	TotalAmount   int64
	SignDate      time.Time
	ContractType  string
	Status        ContractStatus
	ServiceWindow *ServiceWindow
	Milestones    []*Milestone
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Milestone struct {
	ID                string
	ContractID        string
	Name              string
	PlannedDate       time.Time
	PlannedPercentage int
	PlannedAmount     int64
	Owner             string
	Status            MilestoneStatus
	ActualDate        *time.Time
	IsOffHours        bool
	Payment           *Payment
}

type Payment struct {
	ID         string
	MilestoneID string
	ContractID  string
	DueAmount  int64
	PaidAmount *int64
	PaidDate   *time.Time
	Status     PaymentStatus
}

type AuditLog struct {
	ID         string
	ContractID string
	Operation  string
	Operator   string
	Timestamp  time.Time
	Field      string
	OldValue   string
	NewValue   string
}

type ContractAdjustment struct {
	ID            string
	ContractID    string
	PaymentID     string
	AdjustmentAmount int64
	Reason        string
	CreatedAt     time.Time
}

package models

import (
	"time"
)

type WorkOrderType string

const (
	WorkOrderTypeFaultReport  WorkOrderType = "故障报修"
	WorkOrderTypeServiceRequest WorkOrderType = "服务请求"
	WorkOrderTypeComplaint   WorkOrderType = "投诉"
)

type WorkOrderStatus string

const (
	WorkOrderStatusPending    WorkOrderStatus = "待处理"
	WorkOrderStatusProcessing WorkOrderStatus = "处理中"
	WorkOrderStatusConfirming WorkOrderStatus = "待确认"
	WorkOrderStatusClosed     WorkOrderStatus = "已关闭"
)

type ProcessStatus string

const (
	ProcessStatusSubmitted ProcessStatus = "待提交"
	ProcessStatusReviewing ProcessStatus = "审核中"
	ProcessStatusApproved  ProcessStatus = "已通过"
	ProcessStatusExecuting ProcessStatus = "执行中"
	ProcessStatusCompleted ProcessStatus = "已完成"
	ProcessStatusRejected  ProcessStatus = "退回修改"
)

type Team struct {
	ID        int64
	Name      string
	Type      WorkOrderType
	LeaderID  int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID        int64
	Username  string
	Email     string
	Phone     string
	TeamID    *int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DutySchedule struct {
	ID        int64
	TeamID    int64
	UserID    int64
	DutyDate  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ResourceType struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Resource struct {
	ID             int64
	ResourceTypeID int64
	Name           string
	Description    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type WorkOrderResource struct {
	ID          int64
	WorkOrderID int64
	ResourceID  int64
	CreatedAt   time.Time
}

type WorkOrder struct {
	ID               int64
	Type             WorkOrderType
	Title            string
	Description      string
	Content          string
	CurrentHandlerID *int64
	CurrentTeamID    int64
	ProcessStatus    ProcessStatus
	WorkOrderStatus  WorkOrderStatus
	SubmitterID      int64
	Escalated        bool
	Rating           *int
	EscalatedAt      *time.Time
	DeadlineAt       *time.Time
	ProcessStartAt   *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type HistoryRecord struct {
	ID                  int64
	WorkOrderID         int64
	OperationType       string
	OperatorID          int64
	OldProcessStatus    ProcessStatus
	NewProcessStatus    ProcessStatus
	OldWorkOrderStatus  WorkOrderStatus
	NewWorkOrderStatus  WorkOrderStatus
	AssignedFrom        *int64
	AssignedTo          *int64
	Reason              string
	Comment             string
	CreatedAt           time.Time
}

type CommunicationRecord struct {
	ID          int64
	WorkOrderID int64
	SenderID    int64
	Content     string
	CreatedAt   time.Time
}

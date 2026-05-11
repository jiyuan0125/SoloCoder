package common

import "time"

type ServiceWindow struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Weekdays  []int  `json:"weekdays"`
}

type MilestoneRequest struct {
	Name               string    `json:"name"`
	PlannedDate        time.Time `json:"planned_date"`
	PlannedPercentage  int       `json:"planned_percentage"`
	Owner              string    `json:"owner"`
}

type CreateContractRequest struct {
	ContractNo     string              `json:"contract_no"`
	PartyA         string              `json:"party_a"`
	PartyB         string              `json:"party_b"`
	TotalAmount    int64               `json:"total_amount"`
	SignDate       time.Time           `json:"sign_date"`
	ContractType   string              `json:"contract_type"`
	Milestones     []MilestoneRequest  `json:"milestones"`
	ServiceWindow  *ServiceWindow      `json:"service_window,omitempty"`
	Operator       string              `json:"operator"`
}

type UpdateContractAmountRequest struct {
	NewTotalAmount int64  `json:"new_total_amount"`
	Operator       string `json:"operator"`
}

type CompleteMilestoneRequest struct {
	ActualDate time.Time `json:"actual_date"`
	Approved   bool      `json:"approved"`
	Operator   string    `json:"operator"`
}

type PayMilestoneRequest struct {
	PaidAmount int64  `json:"paid_amount"`
	Operator   string `json:"operator"`
}

type ContractResponse struct {
	ID             string            `json:"id"`
	ContractNo     string            `json:"contract_no"`
	PartyA         string            `json:"party_a"`
	PartyB         string            `json:"party_b"`
	TotalAmount    int64             `json:"total_amount"`
	SignDate       time.Time         `json:"sign_date"`
	ContractType   string            `json:"contract_type"`
	Status         string            `json:"status"`
	ServiceWindow  *ServiceWindow    `json:"service_window,omitempty"`
	Milestones     []MilestoneResp   `json:"milestones"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type MilestoneResp struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	PlannedDate        time.Time `json:"planned_date"`
	PlannedPercentage  int       `json:"planned_percentage"`
	PlannedAmount      int64     `json:"planned_amount"`
	Owner              string    `json:"owner"`
	Status             string    `json:"status"`
	ActualDate         *time.Time `json:"actual_date,omitempty"`
	IsOffHours         bool      `json:"is_off_hours"`
	Payment            *PaymentResp `json:"payment,omitempty"`
}

type PaymentResp struct {
	ID          string     `json:"id"`
	DueAmount   int64      `json:"due_amount"`
	PaidAmount  *int64     `json:"paid_amount,omitempty"`
	PaidDate    *time.Time `json:"paid_date,omitempty"`
	Status      string     `json:"status"`
}

type AuditLogEntry struct {
	ID           string    `json:"id"`
	ContractID   string    `json:"contract_id"`
	Operation    string    `json:"operation"`
	Operator     string    `json:"operator"`
	Timestamp    time.Time `json:"timestamp"`
	Field        string    `json:"field"`
	OldValue     string    `json:"old_value"`
	NewValue     string    `json:"new_value"`
}

type ListContractsResponse struct {
	Contracts []ContractResponse `json:"contracts"`
}

type ListAuditLogsResponse struct {
	Logs []AuditLogEntry `json:"logs"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

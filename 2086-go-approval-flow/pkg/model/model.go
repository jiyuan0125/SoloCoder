package model

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Token     string    `json:"token,omitempty"`
	ManagerID string    `json:"manager_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ApprovalChain struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Nodes       []*ApprovalNode `json:"nodes,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ApprovalNode struct {
	ID           string   `json:"id"`
	ChainID      string   `json:"chain_id"`
	Level        int      `json:"level"`
	NodeType     string   `json:"node_type"`
	ApproverIDs  []string `json:"approver_ids"`
	Condition    string   `json:"condition,omitempty"`
	IsSignAll    bool     `json:"is_sign_all"`
	CreatedAt    time.Time `json:"created_at"`
}

type Application struct {
	ID                 string    `json:"id"`
	ChainID            string    `json:"chain_id"`
	ApplicantID        string    `json:"applicant_id"`
	Title              string    `json:"title"`
	Description        string    `json:"description,omitempty"`
	Data               map[string]interface{} `json:"data"`
	CurrentLevel       int       `json:"current_level"`
	CurrentApproverIDs []string  `json:"current_approver_ids,omitempty"`
	Status             string    `json:"status"`
	RejectReason       string    `json:"reject_reason,omitempty"`
	SubmissionCount    int       `json:"submission_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ApprovalOperation struct {
	ID            string    `json:"id"`
	ApplicationID string    `json:"application_id"`
	OperatorID    string    `json:"operator_id"`
	Level         int       `json:"level"`
	Operation     string    `json:"operation"`
	Reason        string    `json:"reason,omitempty"`
	TargetUserID  string    `json:"target_user_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type Notification struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	ApplicationID string    `json:"application_id"`
	Type          string    `json:"type"`
	Message       string    `json:"message"`
	Read          bool      `json:"read"`
	CreatedAt     time.Time `json:"created_at"`
}

type Report struct {
	ID            string                    `json:"id"`
	ReportDate    string                    `json:"report_date"`
	ChainID       string                    `json:"chain_id"`
	TotalAmount   float64                   `json:"total_amount"`
	ApprovedCount int                       `json:"approved_count"`
	RejectedCount int                       `json:"rejected_count"`
	PendingCount  int                       `json:"pending_count"`
	DetailStats   map[string]map[string]int `json:"detail_stats,omitempty"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

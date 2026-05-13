package model

import (
	"time"
)

type ReleaseStatus string

const (
	StatusDraft       ReleaseStatus = "draft"
	StatusPending     ReleaseStatus = "pending_review"
	StatusApproved    ReleaseStatus = "approved"
	StatusDeploying   ReleaseStatus = "deploying"
	StatusCompleted   ReleaseStatus = "completed"
	StatusRejected    ReleaseStatus = "rejected"
	StatusRollbacked  ReleaseStatus = "rollbacked"
)

type Release struct {
	ID          int64         `json:"id"`
	Version     string        `json:"version"`
	Description string        `json:"description"`
	Status      ReleaseStatus `json:"status"`
	SubmittedBy string        `json:"submitted_by"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type ChangeItem struct {
	ID        int64  `json:"id"`
	ReleaseID int64  `json:"release_id"`
	Content   string `json:"content"`
	Type      string `json:"type"`
}

type Approval struct {
	ID        int64     `json:"id"`
	ReleaseID int64     `json:"release_id"`
	Approver  string    `json:"approver"`
	Approved  bool      `json:"approved"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Deployment struct {
	ID            int64     `json:"id"`
	ReleaseID     int64     `json:"release_id"`
	Environment   string    `json:"environment"`
	Status        string    `json:"status"`
	SmokeTestPass bool      `json:"smoke_test_pass,omitempty"`
	GrayRelease   bool      `json:"gray_release,omitempty"`
	GrayPercent   int       `json:"gray_percent,omitempty"`
	IsRollback    bool      `json:"is_rollback,omitempty"`
	RollbackFrom  int64     `json:"rollback_from,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type OperationHistory struct {
	ID          int64     `json:"id"`
	ReleaseID   int64     `json:"release_id"`
	Operation   string    `json:"operation"`
	Operator    string    `json:"operator"`
	Details     string    `json:"details"`
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

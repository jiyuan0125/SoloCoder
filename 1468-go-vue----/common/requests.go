package common

import (
	"time"
)

type CreateProjectRequest struct {
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	Method           TenderMethod     `json:"method"`
	BidDeadline      time.Time        `json:"bid_deadline"`
	OpenTime         time.Time        `json:"open_time"`
	EvaluationMethod EvaluationMethod `json:"evaluation_method"`
	InvitedSuppliers []string         `json:"invited_suppliers,omitempty"`
}

type UpdateProjectRequest struct {
	Name             string           `json:"name,omitempty"`
	Description      string           `json:"description,omitempty"`
	Method           TenderMethod     `json:"method,omitempty"`
	BidDeadline      *time.Time       `json:"bid_deadline,omitempty"`
	OpenTime         *time.Time       `json:"open_time,omitempty"`
	EvaluationMethod EvaluationMethod `json:"evaluation_method,omitempty"`
	InvitedSuppliers []string         `json:"invited_suppliers,omitempty"`
}

type SubmitBidRequest struct {
	ProjectID     string `json:"project_id"`
	SupplierID    string `json:"supplier_id"`
	Amount        int64  `json:"amount"`
	TechnicalPlan string `json:"technical_plan"`
	DurationDays  int    `json:"duration_days"`
}

type UpdateBidRequest struct {
	Amount        int64  `json:"amount"`
	TechnicalPlan string `json:"technical_plan"`
	DurationDays  int    `json:"duration_days"`
}

type OpenProjectRequest struct {
	ProjectID string `json:"project_id"`
}

type ListProjectsRequest struct {
	UserRole     string `json:"user_role"`
	SupplierID   string `json:"supplier_id,omitempty"`
	StatusFilter string `json:"status_filter,omitempty"`
}

package common

import "time"

type Policy struct {
	PolicyNumber   string       `json:"policy_number"`
	PolicyType     PolicyType   `json:"policy_type"`
	TotalAmount    float64      `json:"total_amount"`
	RemainingAmount float64     `json:"remaining_amount"`
	EffectiveDate  time.Time    `json:"effective_date"`
	Status         PolicyStatus `json:"status"`
	CreatedAt      time.Time    `json:"created_at"`
}

type Claim struct {
	ClaimID          string          `json:"claim_id"`
	PolicyNumber     string          `json:"policy_number"`
	IncidentDate     time.Time       `json:"incident_date"`
	IncidentReason   string          `json:"incident_reason"`
	RequestedAmount  float64         `json:"requested_amount"`
	LiabilityRatio   LiabilityRatio  `json:"liability_ratio,omitempty"`
	PayoutPercentage float64         `json:"payout_percentage"`
	PayoutAmount     float64         `json:"payout_amount"`
	Status           ClaimStatus     `json:"status"`
	RejectReason     string          `json:"reject_reason,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	ReviewedAt       *time.Time      `json:"reviewed_at,omitempty"`
	PaidAt           *time.Time      `json:"paid_at,omitempty"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

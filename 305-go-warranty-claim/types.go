package main

import "time"

type ClaimStatus string

const (
	StatusPending    ClaimStatus = "pending"
	StatusApproved   ClaimStatus = "approved"
	StatusRejected   ClaimStatus = "rejected"
	StatusProcessing ClaimStatus = "processing"
)

type Appeal struct {
	ID          string    `json:"id"`
	ClaimID     string    `json:"claim_id"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
	Resolved    bool      `json:"resolved"`
	ResolvedAt  time.Time `json:"resolved_at,omitempty"`
	Resolution  string    `json:"resolution,omitempty"`
}

type WarrantyClaim struct {
	ID              string       `json:"id"`
	SerialNumber    string       `json:"serial_number"`
	PurchaseDate    time.Time    `json:"purchase_date"`
	FailureDesc     string       `json:"failure_desc"`
	Status          ClaimStatus  `json:"status"`
	RejectReason    string       `json:"reject_reason,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	ApprovedAt      time.Time    `json:"approved_at,omitempty"`
	WarrantyStartAt time.Time    `json:"warranty_start_at,omitempty"`
	Appeals         []Appeal     `json:"appeals,omitempty"`
	UserID          string       `json:"user_id"`
}

type CreateClaimRequest struct {
	SerialNumber string `json:"serial_number"`
	PurchaseDate string `json:"purchase_date"`
	FailureDesc  string `json:"failure_desc"`
	UserID       string `json:"user_id"`
}

type ReviewRequest struct {
	Approved bool   `json:"approved"`
	Reason   string `json:"reason,omitempty"`
}

type AppealRequest struct {
	Reason string `json:"reason"`
}

type Statistics struct {
	TotalClaims     int            `json:"total_claims"`
	PendingClaims   int            `json:"pending_claims"`
	ApprovedClaims  int            `json:"approved_claims"`
	RejectedClaims  int            `json:"rejected_claims"`
	ProcessingClaims int           `json:"processing_claims"`
	ClaimsByStatus  map[string]int `json:"claims_by_status"`
}

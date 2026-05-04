package common

import "time"

type CreatePolicyRequest struct {
	PolicyNumber  string       `json:"policy_number"`
	PolicyType    PolicyType   `json:"policy_type"`
	TotalAmount   float64      `json:"total_amount"`
	EffectiveDate time.Time    `json:"effective_date"`
}

type SubmitClaimRequest struct {
	PolicyNumber     string         `json:"policy_number"`
	IncidentDate     time.Time      `json:"incident_date"`
	IncidentReason   string         `json:"incident_reason"`
	RequestedAmount  float64        `json:"requested_amount"`
	LiabilityRatio   LiabilityRatio `json:"liability_ratio,omitempty"`
}

type ReviewClaimRequest struct {
	ClaimID      string      `json:"claim_id"`
	Approved     bool        `json:"approved"`
	RejectReason string      `json:"reject_reason,omitempty"`
}

type ConfirmPaymentRequest struct {
	ClaimID string `json:"claim_id"`
}

type GetPolicyClaimsRequest struct {
	PolicyNumber string `json:"policy_number"`
}

type GetClaimRequest struct {
	ClaimID string `json:"claim_id"`
}

type PolicyResponse struct {
	Policy
}

type ClaimResponse struct {
	Claim
}

type ClaimsListResponse struct {
	Claims []*Claim `json:"claims"`
}

type PoliciesListResponse struct {
	Policies []*Policy `json:"policies"`
}

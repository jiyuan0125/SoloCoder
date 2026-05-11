package common

import "time"

type SubmitClaimRequest struct {
	PolicyNo            string       `json:"policy_no"`
	AccidentTime        time.Time    `json:"accident_time"`
	AccidentLocation    string       `json:"accident_location"`
	AccidentType        AccidentType `json:"accident_type"`
	AccidentDescription string       `json:"accident_description"`
	EstimatedAmount     int64        `json:"estimated_amount"`
}

type AssignInvestigatorRequest struct {
	CaseID           string `json:"case_id"`
	InvestigatorID   string `json:"investigator_id"`
	InvestigatorName string `json:"investigator_name"`
	Operator         string `json:"operator"`
}

type SubmitInvestigationRequest struct {
	CaseID            string        `json:"case_id"`
	InvestigatorID    string        `json:"investigator_id"`
	InvestigatorName  string        `json:"investigator_name"`
	CauseAnalysis     string        `json:"cause_analysis"`
	Liability         LiabilityType `json:"liability"`
	DamagedParts      string        `json:"damaged_parts"`
	HasValidReason    bool          `json:"has_valid_reason"`
	ReasonDescription string        `json:"reason_description"`
}

type AssignAssessorRequest struct {
	CaseID         string `json:"case_id"`
	AssessorID     string `json:"assessor_id"`
	AssessorName   string `json:"assessor_name"`
	Operator       string `json:"operator"`
}

type DamageItemRequest struct {
	ItemName   string     `json:"item_name"`
	RepairType RepairType `json:"repair_type"`
	Amount     int64      `json:"amount"`
}

type SubmitAssessmentRequest struct {
	CaseID       string              `json:"case_id"`
	AssessorID   string              `json:"assessor_id"`
	AssessorName string              `json:"assessor_name"`
	Items        []DamageItemRequest `json:"items"`
}

type FlagForReviewRequest struct {
	CaseID   string `json:"case_id"`
	Operator string `json:"operator"`
	Reason   string `json:"reason"`
}

type ResolveReviewRequest struct {
	CaseID       string `json:"case_id"`
	Operator     string `json:"operator"`
	ShouldReject bool   `json:"should_reject"`
	Reason       string `json:"reason"`
}

type ListCasesRequest struct {
	Status CaseStatus `json:"status,omitempty"`
	PolicyNo string `json:"policy_no,omitempty"`
}

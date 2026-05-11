package common

import "time"

type AccidentType string

const (
	AccidentTypeTraffic     AccidentType = "traffic"
	AccidentTypeNatural     AccidentType = "natural"
	AccidentTypeTheft       AccidentType = "theft"
	AccidentTypeGlass       AccidentType = "glass"
	AccidentTypeScratch     AccidentType = "scratch"
)

type LiabilityType string

const (
	LiabilityFull    LiabilityType = "full"
	LiabilityMajor   LiabilityType = "major"
	LiabilityEqual   LiabilityType = "equal"
	LiabilityMinor   LiabilityType = "minor"
	LiabilityNone    LiabilityType = "none"
)

type CaseStatus string

const (
	CaseStatusSubmitted      CaseStatus = "submitted"
	CaseStatusAssigned       CaseStatus = "assigned"
	CaseStatusInvestigating  CaseStatus = "investigating"
	CaseStatusInvestigated   CaseStatus = "investigated"
	CaseStatusAssessing      CaseStatus = "assessing"
	CaseStatusAssessed       CaseStatus = "assessed"
	CaseStatusApproved       CaseStatus = "approved"
	CaseStatusPaid           CaseStatus = "paid"
	CaseStatusReviewing      CaseStatus = "reviewing"
	CaseStatusRejected       CaseStatus = "rejected"
)

type RepairType string

const (
	RepairTypeRepair  RepairType = "repair"
	RepairTypeReplace RepairType = "replace"
)

type StatusHistory struct {
	ID         string     `json:"id"`
	CaseID     string     `json:"case_id"`
	FromStatus CaseStatus `json:"from_status"`
	ToStatus   CaseStatus `json:"to_status"`
	Operator   string     `json:"operator"`
	Reason     string     `json:"reason"`
	CreatedAt  time.Time  `json:"created_at"`
}

type InvestigationReport struct {
	CaseID            string        `json:"case_id"`
	InvestigatorID    string        `json:"investigator_id"`
	InvestigatorName  string        `json:"investigator_name"`
	CauseAnalysis     string        `json:"cause_analysis"`
	Liability         LiabilityType `json:"liability"`
	DamagedParts      string        `json:"damaged_parts"`
	HasValidReason    bool          `json:"has_valid_reason"`
	ReasonDescription string        `json:"reason_description"`
	InvestigatedAt    time.Time     `json:"investigated_at"`
}

type DamageItem struct {
	ID         string     `json:"id"`
	CaseID     string     `json:"case_id"`
	ItemName   string     `json:"item_name"`
	RepairType RepairType `json:"repair_type"`
	Amount     int64      `json:"amount"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Assessment struct {
	CaseID        string       `json:"case_id"`
	AssessorID    string       `json:"assessor_id"`
	AssessorName  string       `json:"assessor_name"`
	TotalAmount   int64        `json:"total_amount"`
	Items         []DamageItem `json:"items"`
	AssessedAt    time.Time    `json:"assessed_at"`
}

type Payout struct {
	CaseID            string    `json:"case_id"`
	StandardAmount    int64     `json:"standard_amount"`
	FinalAmount       int64     `json:"final_amount"`
	LiabilityRatio    float64   `json:"liability_ratio"`
	PolicyClaimCount  int       `json:"policy_claim_count"`
	HasTimeDiscount   bool      `json:"has_time_discount"`
	TimeDiscountRatio float64   `json:"time_discount_ratio"`
	HasFreqDiscount   bool      `json:"has_freq_discount"`
	FreqDiscountRatio float64   `json:"freq_discount_ratio"`
	CalculatedAt      time.Time `json:"calculated_at"`
}

type ClaimCase struct {
	ID                  string             `json:"id"`
	CaseNo              string             `json:"case_no"`
	PolicyNo            string             `json:"policy_no"`
	PolicyLimit         int64              `json:"policy_limit"`
	AccidentTime        time.Time          `json:"accident_time"`
	ReportTime          time.Time          `json:"report_time"`
	AccidentLocation    string             `json:"accident_location"`
	AccidentType        AccidentType       `json:"accident_type"`
	AccidentDescription string             `json:"accident_description"`
	EstimatedAmount     int64              `json:"estimated_amount"`
	InvestigatorID      string             `json:"investigator_id"`
	InvestigatorName    string             `json:"investigator_name"`
	AssessorID          string             `json:"assessor_id"`
	AssessorName        string             `json:"assessor_name"`
	Status              CaseStatus         `json:"status"`
	IsUnderReview       bool               `json:"is_under_review"`
	InvestigationReport *InvestigationReport `json:"investigation_report,omitempty"`
	Assessment          *Assessment        `json:"assessment,omitempty"`
	Payout              *Payout            `json:"payout,omitempty"`
	StatusHistory       []StatusHistory    `json:"status_history"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
}

type Policy struct {
	PolicyNo string `json:"policy_no"`
	Limit    int64  `json:"limit"`
}

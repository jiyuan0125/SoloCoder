package model

import "time"

type CCFlag string

const (
	CCFlagNone CCFlag = "none"
	CCFlagCC   CCFlag = "cc"
	CCFlagMCC  CCFlag = "mcc"
)

type SettlementStatus string

const (
	SettlementStatusDraft     SettlementStatus = "draft"
	SettlementStatusSubmitted SettlementStatus = "submitted"
	SettlementStatusApproved  SettlementStatus = "approved"
	SettlementStatusPublished SettlementStatus = "published"
)

type TodoType string

const (
	TodoTypeSubmitSettlement   TodoType = "submit_settlement"
	TodoTypeModifySettlement   TodoType = "modify_settlement"
)

type TodoStatus string

const (
	TodoStatusPending   TodoStatus = "pending"
	TodoStatusCompleted TodoStatus = "completed"
)

type DRGGroup struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Weight      float64   `json:"weight"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GroupingRule struct {
	ID              string    `json:"id"`
	DiagnosisPrefix string    `json:"diagnosis_prefix"`
	ProcedurePrefix string    `json:"procedure_prefix"`
	CCFlag          CCFlag    `json:"cc_flag"`
	AgeMin          int       `json:"age_min"`
	AgeMax          int       `json:"age_max"`
	Gender          string    `json:"gender"`
	DRGGroupCode    string    `json:"drg_group_code"`
	CreatedAt       time.Time `json:"created_at"`
}

type MedicalRecord struct {
	ID              string    `json:"id"`
	CaseNo          string    `json:"case_no"`
	HospitalID      string    `json:"hospital_id"`
	MainDiagnosis   string    `json:"main_diagnosis"`
	MainProcedure   string    `json:"main_procedure"`
	CCFlag          CCFlag    `json:"cc_flag"`
	Age             int       `json:"age"`
	Gender          string    `json:"gender"`
	ActualCost      int64     `json:"actual_cost"`
	DRGGroupCode    string    `json:"drg_group_code"`
	IsExcluded      bool      `json:"is_excluded"`
	ExclusionReason string    `json:"exclusion_reason"`
	SettlementID    string    `json:"settlement_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type PaymentParam struct {
	ID         string    `json:"id"`
	Rate       float64   `json:"rate"`
	Version    int       `json:"version"`
	IsActive   bool      `json:"is_active"`
	EffectiveAt time.Time `json:"effective_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type Settlement struct {
	ID               string             `json:"id"`
	HospitalID       string             `json:"hospital_id"`
	Period           string             `json:"period"`
	Status           SettlementStatus   `json:"status"`
	GroupSummaries   []GroupSummary     `json:"group_summaries"`
	TotalActualCost  int64              `json:"total_actual_cost"`
	TotalPayment     int64              `json:"total_payment"`
	Difference       int64              `json:"difference"`
	AuditComment     string             `json:"audit_comment"`
	CreatedAt        time.Time          `json:"created_at"`
	SubmittedAt      *time.Time         `json:"submitted_at"`
	ApprovedAt       *time.Time         `json:"approved_at"`
	PublishedAt      *time.Time         `json:"published_at"`
}

type GroupSummary struct {
	DRGGroupCode   string  `json:"drg_group_code"`
	DRGGroupName   string  `json:"drg_group_name"`
	CaseCount      int     `json:"case_count"`
	TotalActualCost int64  `json:"total_actual_cost"`
	EfficiencyIndex float64 `json:"efficiency_index"`
	StandardPayment int64  `json:"standard_payment"`
	AdjustedPayment int64  `json:"adjusted_payment"`
}

type Todo struct {
	ID            string     `json:"id"`
	Type          TodoType   `json:"type"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        TodoStatus `json:"status"`
	ReferenceID   string     `json:"reference_id"`
	HospitalID    string     `json:"hospital_id"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

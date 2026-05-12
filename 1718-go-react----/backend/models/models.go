package models

import "time"

type ClinicalPath struct {
	ID             int64      `json:"id"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	ICDCodes       string     `json:"icd_codes"`
	MinDays        int        `json:"min_days"`
	MaxDays        int        `json:"max_days"`
	MinCost        float64    `json:"min_cost"`
	MaxCost        float64    `json:"max_cost"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type PathStage struct {
	ID          int64       `json:"id"`
	PathID      int64       `json:"path_id"`
	Name        string      `json:"name"`
	Days        int         `json:"days"`
	Order       int         `json:"order"`
}

type OrderItem struct {
	ID          int64       `json:"id"`
	StageID     int64       `json:"stage_id"`
	PathID      int64       `json:"path_id"`
	Name        string      `json:"name"`
	Category    string      `json:"category"`
	IsRequired  bool        `json:"is_required"`
}

type Patient struct {
	ID             int64      `json:"id"`
	HospitalNo     string     `json:"hospital_no"`
	Name           string     `json:"name"`
	Diagnosis      string     `json:"diagnosis"`
	AdmissionDate  time.Time  `json:"admission_date"`
	Status         string     `json:"status"`
}

type PatientEnrollment struct {
	ID              int64      `json:"id"`
	PatientID       int64      `json:"patient_id"`
	PathID          int64      `json:"path_id"`
	PatientHospitalNo string   `json:"patient_hospital_no"`
	PatientName     string     `json:"patient_name"`
	Diagnosis       string     `json:"diagnosis"`
	EnrollDate      time.Time  `json:"enroll_date"`
	ActualStartDate time.Time  `json:"actual_start_date"`
	ActualEndDate   *time.Time `json:"actual_end_date,omitempty"`
	Status          string     `json:"status"`
	ExitDate        *time.Time `json:"exit_date,omitempty"`
	ExitReason      string     `json:"exit_reason,omitempty"`
	VariationCount  int        `json:"variation_count"`
	SuggestExit     bool       `json:"suggest_exit"`
	ActualDays      int        `json:"actual_days,omitempty"`
	ActualCost      float64    `json:"actual_cost,omitempty"`
}

type DailyOrder struct {
	ID                int64      `json:"id"`
	EnrollmentID      int64      `json:"enrollment_id"`
	PathDay           int        `json:"path_day"`
	OrderDate         time.Time  `json:"order_date"`
	StageID           int64      `json:"stage_id"`
	StageName         string     `json:"stage_name"`
	ItemID            int64      `json:"item_id"`
	ItemName          string     `json:"item_name"`
	Category          string     `json:"category"`
	IsRequired        bool       `json:"is_required"`
	IsExecuted        bool       `json:"is_executed"`
	ExecutedAt        *time.Time `json:"executed_at,omitempty"`
	IsVariation       bool       `json:"is_variation"`
}

type VariationRecord struct {
	ID              int64      `json:"id"`
	EnrollmentID    int64      `json:"enrollment_id"`
	PatientID       int64      `json:"patient_id"`
	PatientHospitalNo string   `json:"patient_hospital_no"`
	PathID          int64      `json:"path_id"`
	Date            time.Time  `json:"date"`
	Content         string     `json:"content"`
	Reason          string     `json:"reason"`
	Type            string     `json:"type"`
	Action          string     `json:"action"`
	CreatedAt       time.Time  `json:"created_at"`
}

type QualityMetrics struct {
	PathID          int64     `json:"path_id"`
	PathCode        string    `json:"path_code"`
	PathName        string    `json:"path_name"`
	Month           string    `json:"month"`
	TotalEligible   int       `json:"total_eligible"`
	EnrolledCount   int       `json:"enrolled_count"`
	CompletedCount  int       `json:"completed_count"`
	ExitedCount     int       `json:"exited_count"`
	VariationCount  int       `json:"variation_count"`
	EnrollRate      float64   `json:"enroll_rate"`
	CompleteRate    float64   `json:"complete_rate"`
	VariationRate   float64   `json:"variation_rate"`
	AvgStayDays     float64   `json:"avg_stay_days"`
	AvgCost         float64   `json:"avg_cost"`
	StandardMinDays int       `json:"standard_min_days"`
	StandardMaxDays int       `json:"standard_max_days"`
	StandardMinCost float64   `json:"standard_min_cost"`
	StandardMaxCost float64   `json:"standard_max_cost"`
	HasWarning      bool      `json:"has_warning"`
	WarningReason   string    `json:"warning_reason"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

package models

import (
	"time"

	"gorm.io/gorm"
)

type Department struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"uniqueIndex"`
}

type Patient struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	PatientID string    `json:"patient_id" gorm:"uniqueIndex"`
	Name      string    `json:"name"`
	Gender    string    `json:"gender"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InfectionType string

const (
	InfectionTypeCommunity InfectionType = "community"
	InfectionTypeNosocomial InfectionType = "nosocomial"
)

type InfectionSite string

const (
	SiteRespiratory      InfectionSite = "respiratory"
	SiteSurgicalIncision InfectionSite = "surgical_incision"
	SiteUrinaryTract     InfectionSite = "urinary_tract"
	SiteBloodstream      InfectionSite = "bloodstream"
	SiteDigestive        InfectionSite = "digestive"
	SiteSkinSoftTissue   InfectionSite = "skin_soft_tissue"
)

type InfectionCase struct {
	ID              uint          `json:"id" gorm:"primaryKey"`
	PatientID       string        `json:"patient_id"`
	PatientName     string        `json:"patient_name"`
	Gender          string        `json:"gender"`
	Age             int           `json:"age"`
	DepartmentID    uint          `json:"department_id"`
	Department      Department    `json:"department" gorm:"foreignKey:DepartmentID"`
	AdmissionDate   time.Time     `json:"admission_date"`
	InfectionDate   time.Time     `json:"infection_date"`
	InfectionSite   InfectionSite `json:"infection_site"`
	InfectionType   InfectionType `json:"infection_type"`
	Pathogen        string        `json:"pathogen"`
	DrugSensitivity bool          `json:"drug_sensitivity"`
	Status          string        `json:"status"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

type MeasureType string

const (
	MeasureIsolation        MeasureType = "isolation"
	MeasureHandHygiene      MeasureType = "hand_hygiene"
	MeasureEnvironmentClean MeasureType = "environment_clean"
	MeasureAntibioticAdjust MeasureType = "antibiotic_adjust"
	MeasureEquipmentSterile MeasureType = "equipment_sterile"
)

type PreventionMeasure struct {
	ID              uint            `json:"id" gorm:"primaryKey"`
	InfectionCaseID uint            `json:"infection_case_id"`
	InfectionCase   InfectionCase   `json:"infection_case" gorm:"foreignKey:InfectionCaseID"`
	MeasureType     MeasureType     `json:"measure_type"`
	DepartmentID    uint            `json:"department_id"`
	Department      Department      `json:"department" gorm:"foreignKey:DepartmentID"`
	Executor        string          `json:"executor"`
	ExecuteDate     time.Time       `json:"execute_date"`
	Status          string          `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `json:"-" gorm:"index"`
}

type TargetMonitoring struct {
	ID                   uint        `json:"id" gorm:"primaryKey"`
	DepartmentID         uint        `json:"department_id"`
	Department           Department  `json:"department" gorm:"foreignKey:DepartmentID"`
	Month                string      `json:"month"`
	HospitalizationDays  int         `json:"hospitalization_days"`
	VentilatorDays       int         `json:"ventilator_days"`
	VAPCases             int         `json:"vap_cases"`
	CentralLineDays      int         `json:"central_line_days"`
	CLABSICases          int         `json:"clabsi_cases"`
	CatheterDays         int         `json:"catheter_days"`
	CAUTICases          int         `json:"cauti_cases"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`
}

type AlertType string

const (
	AlertRateExceeded    AlertType = "rate_exceeded"
	AlertNeedIntervention AlertType = "need_intervention"
)

type Alert struct {
	ID           uint        `json:"id" gorm:"primaryKey"`
	DepartmentID uint        `json:"department_id"`
	Department   Department  `json:"department" gorm:"foreignKey:DepartmentID"`
	AlertType    AlertType   `json:"alert_type"`
	Month        string      `json:"month"`
	InfectionRate float64    `json:"infection_rate"`
	Threshold    float64     `json:"threshold"`
	Message      string      `json:"message"`
	Status       string      `json:"status"`
	Notified     bool        `json:"notified"`
	CreatedAt    time.Time   `json:"created_at"`
}

type ApprovalStatus string

const (
	ApprovalStatusSubmitted ApprovalStatus = "submitted"
	ApprovalStatusFirstReview ApprovalStatus = "first_review"
	ApprovalStatusSecondReview ApprovalStatus = "second_review"
	ApprovalStatusFinalReview ApprovalStatus = "final_review"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

type Report struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	ReportType     string         `json:"report_type"`
	Month          string         `json:"month"`
	Year           int            `json:"year"`
	HospitalRate   float64        `json:"hospital_rate"`
	DepartmentRates string        `json:"department_rates"`
	SiteDistribution string      `json:"site_distribution"`
	PathogenDistribution string   `json:"pathogen_distribution"`
	AntibioticUsage string       `json:"antibiotic_usage"`
	ApprovalStatus ApprovalStatus `json:"approval_status"`
	ApprovedBy     string         `json:"approved_by"`
	ApprovedAt     *time.Time     `json:"approved_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

type ApprovalHistory struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	ReportID   uint           `json:"report_id"`
	Status     ApprovalStatus `json:"status"`
	Operator   string         `json:"operator"`
	Comment    string         `json:"comment"`
	CreatedAt  time.Time      `json:"created_at"`
}

type PriceHistory struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	EntityType string   `json:"entity_type"`
	EntityID  uint      `json:"entity_id"`
	OldPrice  float64   `json:"old_price"`
	NewPrice  float64   `json:"new_price"`
	ChangedBy string    `json:"changed_by"`
	CreatedAt time.Time `json:"created_at"`
}

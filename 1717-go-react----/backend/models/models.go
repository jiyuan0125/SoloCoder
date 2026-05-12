package models

import (
	"time"

	"gorm.io/gorm"
)

type Department struct {
	ID   uint   `json:"ID" gorm:"primaryKey"`
	Name string `json:"Name" gorm:"uniqueIndex"`
}

type Patient struct {
	ID        uint      `json:"ID" gorm:"primaryKey"`
	PatientID string    `json:"PatientID" gorm:"uniqueIndex"`
	Name      string    `json:"Name"`
	Gender    string    `json:"Gender"`
	Age       int       `json:"Age"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}

type InfectionType string

const (
	InfectionTypeCommunity  InfectionType = "community"
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
	ID              uint          `json:"ID" gorm:"primaryKey"`
	PatientID       string        `json:"PatientID"`
	PatientName     string        `json:"PatientName"`
	Gender          string        `json:"Gender"`
	Age             int           `json:"Age"`
	DepartmentID    uint          `json:"DepartmentID"`
	Department      Department    `json:"Department" gorm:"foreignKey:DepartmentID"`
	AdmissionDate   time.Time     `json:"AdmissionDate"`
	InfectionDate   time.Time     `json:"InfectionDate"`
	InfectionSite   InfectionSite `json:"InfectionSite"`
	InfectionType   InfectionType `json:"InfectionType"`
	Pathogen        string        `json:"Pathogen"`
	DrugSensitivity bool          `json:"DrugSensitivity"`
	Status          string        `json:"Status"`
	CreatedAt       time.Time     `json:"CreatedAt"`
	UpdatedAt       time.Time     `json:"UpdatedAt"`
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
	ID              uint            `json:"ID" gorm:"primaryKey"`
	InfectionCaseID uint            `json:"InfectionCaseID"`
	InfectionCase   InfectionCase   `json:"InfectionCase" gorm:"foreignKey:InfectionCaseID"`
	MeasureType     MeasureType     `json:"MeasureType"`
	DepartmentID    uint            `json:"DepartmentID"`
	Department      Department      `json:"Department" gorm:"foreignKey:DepartmentID"`
	Executor        string          `json:"Executor"`
	ExecuteDate     time.Time       `json:"ExecuteDate"`
	Status          string          `json:"Status"`
	CreatedAt       time.Time       `json:"CreatedAt"`
	UpdatedAt       time.Time       `json:"UpdatedAt"`
	DeletedAt       gorm.DeletedAt  `json:"-" gorm:"index"`
}

type TargetMonitoring struct {
	ID                  uint           `json:"ID" gorm:"primaryKey"`
	DepartmentID        uint           `json:"DepartmentID"`
	Department          Department     `json:"Department" gorm:"foreignKey:DepartmentID"`
	Month               string         `json:"Month"`
	HospitalizationDays int            `json:"HospitalizationDays"`
	VentilatorDays      int            `json:"VentilatorDays"`
	VAPCases            int            `json:"VAPCases"`
	CentralLineDays     int            `json:"CentralLineDays"`
	CLABSICases         int            `json:"CLABSICases"`
	CatheterDays        int            `json:"CatheterDays"`
	CAUTICases          int            `json:"CAUTICases"`
	CreatedAt           time.Time      `json:"CreatedAt"`
	UpdatedAt           time.Time      `json:"UpdatedAt"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}

type AlertType string

const (
	AlertRateExceeded     AlertType = "rate_exceeded"
	AlertNeedIntervention AlertType = "need_intervention"
)

type Alert struct {
	ID            uint       `json:"ID" gorm:"primaryKey"`
	DepartmentID  uint       `json:"DepartmentID"`
	Department    Department `json:"Department" gorm:"foreignKey:DepartmentID"`
	AlertType     AlertType  `json:"AlertType"`
	Month         string     `json:"Month"`
	InfectionRate float64    `json:"InfectionRate"`
	Threshold     float64    `json:"Threshold"`
	Message       string     `json:"Message"`
	Status        string     `json:"Status"`
	Notified      bool       `json:"Notified"`
	CreatedAt     time.Time  `json:"CreatedAt"`
}

type ApprovalStatus string

const (
	ApprovalStatusSubmitted    ApprovalStatus = "submitted"
	ApprovalStatusFirstReview  ApprovalStatus = "first_review"
	ApprovalStatusSecondReview ApprovalStatus = "second_review"
	ApprovalStatusFinalReview  ApprovalStatus = "final_review"
	ApprovalStatusApproved     ApprovalStatus = "approved"
	ApprovalStatusRejected     ApprovalStatus = "rejected"
)

type Report struct {
	ID                   uint           `json:"ID" gorm:"primaryKey"`
	ReportType           string         `json:"ReportType"`
	Month                string         `json:"Month"`
	Year                 int            `json:"Year"`
	HospitalRate         float64        `json:"HospitalRate"`
	DepartmentRates      string         `json:"DepartmentRates"`
	SiteDistribution     string         `json:"SiteDistribution"`
	PathogenDistribution string         `json:"PathogenDistribution"`
	AntibioticUsage      string         `json:"AntibioticUsage"`
	ApprovalStatus       ApprovalStatus `json:"ApprovalStatus"`
	ApprovedBy           string         `json:"ApprovedBy"`
	ApprovedAt           *time.Time     `json:"ApprovedAt"`
	CreatedAt            time.Time      `json:"CreatedAt"`
	UpdatedAt            time.Time      `json:"UpdatedAt"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`
}

type ApprovalHistory struct {
	ID        uint           `json:"ID" gorm:"primaryKey"`
	ReportID  uint           `json:"ReportID"`
	Status    ApprovalStatus `json:"Status"`
	Operator  string         `json:"Operator"`
	Comment   string         `json:"Comment"`
	CreatedAt time.Time      `json:"CreatedAt"`
}

type PriceHistory struct {
	ID         uint      `json:"ID" gorm:"primaryKey"`
	EntityType string    `json:"EntityType"`
	EntityID   uint      `json:"EntityID"`
	OldPrice   float64   `json:"OldPrice"`
	NewPrice   float64   `json:"NewPrice"`
	ChangedBy  string    `json:"ChangedBy"`
	CreatedAt  time.Time `json:"CreatedAt"`
}

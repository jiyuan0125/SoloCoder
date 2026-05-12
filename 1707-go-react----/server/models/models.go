package models

import (
	"time"
	"gorm.io/gorm"
)

type DeviceCategory string

const (
	CategoryDiagnostic DeviceCategory = "diagnostic"
	CategoryTherapeutic DeviceCategory = "therapeutic"
	CategoryAuxiliary DeviceCategory = "auxiliary"
	CategoryMonitoring DeviceCategory = "monitoring"
)

func (c DeviceCategory) GetCalibrationInterval() time.Duration {
	switch c {
	case CategoryDiagnostic:
		return 365 * 24 * time.Hour / 2
	case CategoryTherapeutic:
		return 365 * 24 * time.Hour
	case CategoryAuxiliary:
		return 2 * 365 * 24 * time.Hour
	case CategoryMonitoring:
		return 365 * 24 * time.Hour / 4
	default:
		return 365 * 24 * time.Hour
	}
}

type DeviceStatus string

const (
	StatusInUse DeviceStatus = "in_use"
	StatusIdle DeviceStatus = "idle"
	StatusMaintenance DeviceStatus = "maintenance"
	StatusCalibrating DeviceStatus = "calibrating"
	StatusDisabled DeviceStatus = "disabled"
	StatusScrapped DeviceStatus = "scrapped"
)

type Device struct {
	ID uint `gorm:"primaryKey" json:"id"`
	AssetNumber string `gorm:"uniqueIndex;not null" json:"asset_number"`
	Name string `gorm:"not null" json:"name"`
	BrandModel string `gorm:"not null" json:"brand_model"`
	SerialNumber string `gorm:"not null" json:"serial_number"`
	Category DeviceCategory `gorm:"not null" json:"category"`
	Department string `gorm:"not null" json:"department"`
	Location string `gorm:"not null" json:"location"`
	PurchaseDate time.Time `gorm:"not null" json:"purchase_date"`
	PurchasePrice float64 `gorm:"not null" json:"purchase_price"`
	WarrantyExpiryDate time.Time `json:"warranty_expiry_date"`
	Status DeviceStatus `gorm:"default:in_use" json:"status"`
	LastCalibrationDate *time.Time `json:"last_calibration_date"`
	NextCalibrationDate *time.Time `json:"next_calibration_date"`
	MaintenancePlans []MaintenancePlan `gorm:"foreignKey:DeviceID" json:"maintenance_plans,omitempty"`
	WorkOrders []WorkOrder `gorm:"foreignKey:DeviceID" json:"work_orders,omitempty"`
	CalibrationRecords []CalibrationRecord `gorm:"foreignKey:DeviceID" json:"calibration_records,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type CalibrationAgency struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Name string `gorm:"not null" json:"name"`
	CertificationNumber string `gorm:"not null;unique" json:"certification_number"`
	ContactInfo string `gorm:"not null" json:"contact_info"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type CalibrationStatus string

const (
	CalibrationStatusPending CalibrationStatus = "pending"
	CalibrationStatusInProgress CalibrationStatus = "in_progress"
	CalibrationStatusCompleted CalibrationStatus = "completed"
)

type CalibrationResult string

const (
	CalibrationResultPass CalibrationResult = "pass"
	CalibrationResultFail CalibrationResult = "fail"
	CalibrationResultConditional CalibrationResult = "conditional"
)

type CalibrationRecord struct {
	ID uint `gorm:"primaryKey" json:"id"`
	DeviceID uint `gorm:"not null;index" json:"device_id"`
	Device Device `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	AgencyID uint `gorm:"not null;index" json:"agency_id"`
	Agency CalibrationAgency `gorm:"foreignKey:AgencyID" json:"agency,omitempty"`
	CalibrationDate time.Time `gorm:"not null" json:"calibration_date"`
	Result CalibrationResult `gorm:"not null" json:"result"`
	NextCalibrationDate time.Time `json:"next_calibration_date"`
	CertificateNumber string `json:"certificate_number"`
	LimitedFunctions string `json:"limited_functions"`
	Status CalibrationStatus `gorm:"default:pending" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type MaintenanceType string

const (
	MaintenanceTypeRoutine MaintenanceType = "routine"
	MaintenanceTypeDeep MaintenanceType = "deep"
)

type MaintenancePlanStatus string

const (
	MaintenancePlanStatusPending MaintenancePlanStatus = "pending"
	MaintenancePlanStatusOverdue MaintenancePlanStatus = "overdue"
	MaintenancePlanStatusCompleted MaintenancePlanStatus = "completed"
	MaintenancePlanStatusCancelled MaintenancePlanStatus = "cancelled"
)

type MaintenancePlan struct {
	ID uint `gorm:"primaryKey" json:"id"`
	DeviceID uint `gorm:"not null;index" json:"device_id"`
	Device Device `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	Type MaintenanceType `gorm:"not null" json:"type"`
	ScheduledDate time.Time `gorm:"not null" json:"scheduled_date"`
	Status MaintenancePlanStatus `gorm:"default:pending" json:"status"`
	Content string `json:"content"`
	ReplacedParts string `json:"replaced_parts"`
	DurationMinutes int `json:"duration_minutes"`
	CompletedDate *time.Time `json:"completed_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type WorkOrderPriority string

const (
	PriorityNormal WorkOrderPriority = "normal"
	PriorityUrgent WorkOrderPriority = "urgent"
	PriorityCritical WorkOrderPriority = "critical"
)

func (p WorkOrderPriority) GetResponseHours() float64 {
	switch p {
	case PriorityNormal:
		return 4.0
	case PriorityUrgent:
		return 2.0
	case PriorityCritical:
		return 0.5
	default:
		return 4.0
	}
}

type WorkOrderStatus string

const (
	WorkOrderStatusPending WorkOrderStatus = "pending"
	WorkOrderStatusInProgress WorkOrderStatus = "in_progress"
	WorkOrderStatusPendingAcceptance WorkOrderStatus = "pending_acceptance"
	WorkOrderStatusCompleted WorkOrderStatus = "completed"
	WorkOrderStatusClosed WorkOrderStatus = "closed"
)

type WorkOrder struct {
	ID uint `gorm:"primaryKey" json:"id"`
	DeviceID uint `gorm:"not null;index" json:"device_id"`
	Device Device `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	Description string `gorm:"not null" json:"description"`
	Priority WorkOrderPriority `gorm:"not null" json:"priority"`
	AssignedTo string `json:"assigned_to"`
	Status WorkOrderStatus `gorm:"default:pending" json:"status"`
	ReportTime time.Time `gorm:"not null" json:"report_time"`
	ResponseStartTime *time.Time `json:"response_start_time"`
	AcceptedTime *time.Time `json:"accepted_time"`
	CompletionTime *time.Time `json:"completion_time"`
	AcceptanceTime *time.Time `json:"acceptance_time"`
	CloseTime *time.Time `json:"close_time"`
	RepairContent string `json:"repair_content"`
	ReplacedParts string `json:"replaced_parts"`
	UpgradeCount int `gorm:"default:0" json:"upgrade_count"`
	Escalated bool `gorm:"default:false" json:"escalated"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

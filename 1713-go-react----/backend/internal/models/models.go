package models

import (
	"time"
)

type Enterprise struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	Name               string    `json:"name" gorm:"uniqueIndex"`
	UnifiedSocialCode  string    `json:"unified_social_code" gorm:"uniqueIndex"`
	Industry           string    `json:"industry"`
	Region             string    `json:"region"`
	ContactPerson      string    `json:"contact_person"`
	ContactPhone       string    `json:"contact_phone"`
	Address            string    `json:"address"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	HazardFactors      []HazardFactor     `json:"hazard_factors,omitempty" gorm:"foreignKey:EnterpriseID"`
	Workers            []Worker           `json:"workers,omitempty" gorm:"foreignKey:EnterpriseID"`
}

type HazardFactor struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	EnterpriseID       uint      `json:"enterprise_id"`
	Workshop           string    `json:"workshop"`
	PostName           string    `json:"post_name"`
	Category           string    `json:"category"`
	FactorName         string    `json:"factor_name"`
	HazardLevel        string    `json:"hazard_level"`
	ProtectiveMeasures string    `json:"protective_measures"`
	LastMonitorValue   float64   `json:"last_monitor_value"`
	LastMonitorDate    time.Time `json:"last_monitor_date"`
	ExceedLimit        bool      `json:"exceed_limit"`
	MonitorUnit        string    `json:"monitor_unit"`
	ExposureLimit      float64   `json:"exposure_limit"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Worker struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	EnterpriseID       uint      `json:"enterprise_id"`
	Name               string    `json:"name"`
	IDCard             string    `json:"id_card"`
	Gender             string    `json:"gender"`
	BirthDate          time.Time `json:"birth_date"`
	PostName           string    `json:"post_name"`
	Workshop           string    `json:"workshop"`
	EntryDate          time.Time `json:"entry_date"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	ExposedFactors     []WorkerHazardFactor `json:"exposed_factors,omitempty" gorm:"foreignKey:WorkerID"`
}

type WorkerHazardFactor struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	WorkerID     uint   `json:"worker_id"`
	FactorName   string `json:"factor_name"`
	Category     string `json:"category"`
}

type Examination struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	WorkerID          uint      `json:"worker_id"`
	EnterpriseID      uint      `json:"enterprise_id"`
	ExamType          string    `json:"exam_type"`
	ExamDate          time.Time `json:"exam_date"`
	ScheduledDate     time.Time `json:"scheduled_date"`
	Status            string    `json:"status"`
	FlowStatus        string    `json:"flow_status"`
	IsSuspected       bool      `json:"is_suspected"`
	HasAbnormal       bool      `json:"has_abnormal"`
	ExamCycleMonths   int       `json:"exam_cycle_months"`
	NextExamDate      time.Time `json:"next_exam_date"`
	Remarks           string    `json:"remarks"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	ExamItems         []ExamItemResult `json:"exam_items,omitempty" gorm:"foreignKey:ExaminationID"`
	Worker            *Worker          `json:"worker,omitempty" gorm:"foreignKey:WorkerID"`
	Enterprise        *Enterprise      `json:"enterprise,omitempty" gorm:"foreignKey:EnterpriseID"`
}

type ExamItemResult struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	ExaminationID uint   `json:"examination_id"`
	ItemCode      string `json:"item_code"`
	ItemName      string `json:"item_name"`
	Result        string `json:"result"`
	IsAbnormal    bool   `json:"is_abnormal"`
	IsSuspected   bool   `json:"is_suspected"`
	Remarks       string `json:"remarks"`
}

type TodoItem struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Assignee     string    `json:"assignee"`
	DueDate      time.Time `json:"due_date"`
	Status       string    `json:"status"`
	RelatedID    uint      `json:"related_id"`
	RelatedType  string    `json:"related_type"`
	Priority     string    `json:"priority"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type EnterpriseReport struct {
	ID                    uint      `json:"id" gorm:"primaryKey"`
	EnterpriseID          uint      `json:"enterprise_id"`
	Year                  int       `json:"year"`
	Content               string    `json:"content"`
	GeneratedAt           time.Time `json:"generated_at"`
	Enterprise            *Enterprise `json:"enterprise,omitempty" gorm:"foreignKey:EnterpriseID"`
}

type StatisticsCache struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"uniqueIndex"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

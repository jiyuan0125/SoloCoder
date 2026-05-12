package model

import (
	"time"

	"gorm.io/gorm"
)

type IndicatorCategory string

const (
	CategorySafety    IndicatorCategory = "safety"
	CategoryEfficiency IndicatorCategory = "efficiency"
	CategoryQuality   IndicatorCategory = "quality"
	CategoryExperience IndicatorCategory = "experience"
)

type Indicator struct {
	ID            uint              `gorm:"primaryKey" json:"id"`
	Code          string            `gorm:"uniqueIndex;not null" json:"code"`
	Name          string            `gorm:"not null" json:"name"`
	Formula       string            `json:"formula"`
	SourceDept    string            `json:"source_dept"`
	TargetValue   float64           `gorm:"not null" json:"target_value"`
	WarningValue  float64           `gorm:"not null" json:"warning_value"`
	Category      IndicatorCategory `json:"category"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `gorm:"index" json:"-"`
}

type IndicatorTargetHistory struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	IndicatorID    uint           `gorm:"not null" json:"indicator_id"`
	Indicator      Indicator      `gorm:"foreignKey:IndicatorID" json:"-"`
	OldTargetValue float64        `json:"old_target_value"`
	NewTargetValue float64        `json:"new_target_value"`
	Reason         string         `gorm:"not null" json:"reason"`
	ApprovedBy     string         `gorm:"not null" json:"approved_by"`
	EffectiveMonth string         `gorm:"not null" json:"effective_month"`
	CreatedAt      time.Time      `json:"created_at"`
}

type IndicatorData struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	IndicatorID uint        `gorm:"not null;index:idx_indicator_month,unique" json:"indicator_id"`
	Indicator   Indicator   `gorm:"foreignKey:IndicatorID" json:"indicator"`
	Month       string      `gorm:"not null;index:idx_indicator_month,unique" json:"month"`
	Value       float64     `gorm:"not null" json:"value"`
	IsTargetMet bool        `json:"is_target_met"`
	IsWarning   bool        `json:"is_warning"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type PDCAStatus string

const (
	StatusPlanning    PDCAStatus = "planning"
	StatusExecuting   PDCAStatus = "executing"
	StatusChecking    PDCAStatus = "checking"
	StatusCompleted   PDCAStatus = "completed"
	StatusClosed      PDCAStatus = "closed"
)

type PDCAPhase string

const (
	PhasePlan  PDCAPhase = "plan"
	PhaseDo    PDCAPhase = "do"
	PhaseCheck PDCAPhase = "check"
	PhaseAct   PDCAPhase = "act"
)

type PDCA struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Name            string     `gorm:"not null" json:"name"`
	IndicatorID     *uint      `json:"indicator_id"`
	Indicator       *Indicator `gorm:"foreignKey:IndicatorID" json:"indicator,omitempty"`
	Responsible     string     `gorm:"not null" json:"responsible"`
	StartDate       string     `gorm:"not null" json:"start_date"`
	CurrentPhase    PDCAPhase  `gorm:"default:plan" json:"current_phase"`
	Status          PDCAStatus `gorm:"default:planning" json:"status"`
	ParentPDCAID    *uint      `json:"parent_pdca_id"`
	ParentPDCA      *PDCA      `gorm:"foreignKey:ParentPDCAID" json:"parent_pdca,omitempty"`
	IsImproved      bool       `json:"is_improved"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type PDCAPhaseDetail struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	PDCAID    uint       `gorm:"not null;uniqueIndex:idx_pdca_phase" json:"pdca_id"`
	Phase     PDCAPhase  `gorm:"not null;uniqueIndex:idx_pdca_phase" json:"phase"`
	Content   string     `gorm:"not null" json:"content"`
	CompleteDate string   `json:"complete_date"`
	Evidence  string     `json:"evidence"`
	Completed bool       `json:"completed"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type TodoType string

const (
	TodoTypeImprovement TodoType = "improvement"
	TodoTypeAction      TodoType = "action"
)

type TodoStatus string

const (
	TodoStatusPending   TodoStatus = "pending"
	TodoStatusProgress  TodoStatus = "in_progress"
	TodoStatusCompleted TodoStatus = "completed"
	TodoStatusOverdue   TodoStatus = "overdue"
	TodoStatusArchived  TodoStatus = "archived"
)

type Todo struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Type        TodoType   `gorm:"not null" json:"type"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `json:"description"`
	IndicatorID *uint      `json:"indicator_id"`
	Indicator   *Indicator `gorm:"foreignKey:IndicatorID" json:"indicator,omitempty"`
	PDCAID      *uint      `json:"pdca_id"`
	PDCA        *PDCA      `gorm:"foreignKey:PDCAID" json:"pdca,omitempty"`
	MeetingID   *uint      `json:"meeting_id"`
	Responsible string     `gorm:"not null" json:"responsible"`
	Department  string     `json:"department"`
	DueDate     string     `json:"due_date"`
	Status      TodoStatus `gorm:"default:pending" json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Meeting struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Date      string         `gorm:"not null" json:"date"`
	Attendees string         `json:"attendees"`
	Indicators string        `json:"indicators"`
	Decisions string         `json:"decisions"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type ActionItem struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	MeetingID   uint       `gorm:"not null" json:"meeting_id"`
	Meeting     Meeting    `gorm:"foreignKey:MeetingID" json:"meeting,omitempty"`
	Content     string     `gorm:"not null" json:"content"`
	Responsible string     `gorm:"not null" json:"responsible"`
	DueDate     string     `json:"due_date"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

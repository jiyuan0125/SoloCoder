package models

import (
	"time"

	"gorm.io/gorm"
)

type Patient struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	BirthDate time.Time      `json:"birthDate"`
	Gender    string         `json:"gender"`
	Phone     string         `json:"phone"`
	Status    string         `gorm:"default:active" json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Plans     []Plan         `json:"plans,omitempty"`
}

type Plan struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	PatientID        uint           `gorm:"not null;index" json:"patientId"`
	Name             string         `gorm:"not null" json:"name"`
	AssessmentType   string         `gorm:"not null" json:"assessmentType"`
	StartDate        time.Time      `gorm:"not null" json:"startDate"`
	DurationWeeks    int            `gorm:"not null" json:"durationWeeks"`
	Status           string         `gorm:"default:active" json:"status"`
	Effectiveness    string         `gorm:"default:evaluating" json:"effectiveness"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	Patient          *Patient       `json:"patient,omitempty"`
	Exercises        []Exercise     `json:"exercises,omitempty"`
	Tasks            []Task         `json:"tasks,omitempty"`
	Assessments      []Assessment   `json:"assessments,omitempty"`
}

type Exercise struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	PlanID          uint           `gorm:"not null;index" json:"planId"`
	Name            string         `gorm:"not null" json:"name"`
	GoalDescription string         `json:"goalDescription"`
	FrequencyType   string         `gorm:"not null" json:"frequencyType"`
	FrequencyCount  int            `gorm:"not null" json:"frequencyCount"`
	DurationMinutes int            `gorm:"not null" json:"durationMinutes"`
	DifficultyLevel int            `gorm:"not null;default:1" json:"difficultyLevel"`
	Notes           string         `json:"notes"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	Plan            *Plan          `json:"plan,omitempty"`
	DifficultyLogs  []DifficultyLog `json:"difficultyLogs,omitempty"`
}

type Task struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	PlanID      uint           `gorm:"not null;index" json:"planId"`
	ExerciseID  uint           `gorm:"not null;index" json:"exerciseId"`
	TaskDate    time.Time      `gorm:"not null;index" json:"taskDate"`
	TaskTime    string         `json:"taskTime"`
	Status      string         `gorm:"default:pending" json:"status"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Plan        *Plan          `json:"plan,omitempty"`
	Exercise    *Exercise      `json:"exercise,omitempty"`
	Records     []TrainingRecord `json:"records,omitempty"`
}

type TrainingRecord struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	TaskID          uint           `gorm:"not null;index" json:"taskId"`
	ExerciseID      uint           `gorm:"not null;index" json:"exerciseId"`
	PlanID          uint           `gorm:"not null;index" json:"planId"`
	PatientID       uint           `gorm:"not null;index" json:"patientId"`
	CompletedAt     time.Time      `json:"completedAt"`
	ActualDuration  int            `json:"actualDuration"`
	QualityScore    int            `json:"qualityScore"`
	SubjectiveFeeling string       `json:"subjectiveFeeling"`
	Notes           string         `json:"notes"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	Task            *Task          `json:"task,omitempty"`
	Exercise        *Exercise      `json:"exercise,omitempty"`
}

type DifficultyLog struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ExerciseID    uint           `gorm:"not null;index" json:"exerciseId"`
	OldDifficulty int            `json:"oldDifficulty"`
	NewDifficulty int            `json:"newDifficulty"`
	AdjustReason  string         `json:"adjustReason"`
	AdjustedAt    time.Time      `json:"adjustedAt"`
	CreatedAt     time.Time      `json:"createdAt"`
}

type Assessment struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	PlanID              uint           `gorm:"not null;index" json:"planId"`
	PatientID           uint           `gorm:"not null;index" json:"patientId"`
	AssessmentType      string         `gorm:"not null" json:"assessmentType"`
	ScheduledDate       time.Time      `json:"scheduledDate"`
	AssessmentDate      time.Time      `json:"assessmentDate"`
	InitialScore        float64        `json:"initialScore"`
	CurrentScore        float64        `json:"currentScore"`
	ImprovementPercent  float64        `json:"improvementPercent"`
	RawData             string         `json:"rawData"`
	Notes               string         `json:"notes"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
	Plan                *Plan          `json:"plan,omitempty"`
	Indicators          []AssessmentIndicator `json:"indicators,omitempty"`
}

type AssessmentIndicator struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	AssessmentID uint           `gorm:"not null;index" json:"assessmentId"`
	Name         string         `gorm:"not null" json:"name"`
	Value        float64        `gorm:"not null" json:"value"`
	PreviousValue float64       `json:"previousValue"`
	Change       float64        `json:"change"`
	Unit         string         `json:"unit"`
	CreatedAt    time.Time      `json:"createdAt"`
}

const (
	FrequencyTypeDaily  = "daily"
	FrequencyTypeWeekly = "weekly"

	TaskStatusPending    = "pending"
	TaskStatusCompleted  = "completed"
	TaskStatusOverdue    = "overdue"
	TaskStatusAbandoned  = "abandoned"

	PlanStatusActive    = "active"
	PlanStatusCompleted = "completed"
	PlanStatusAdjusting = "adjusting"

	EffectivenessEvaluating = "evaluating"
	EffectivenessPoor       = "poor"
	EffectivenessGood       = "good"

	FeelingEasy    = "easy"
	FeelingNormal  = "normal"
	FeelingHard    = "hard"
	FeelingVeryHard = "very_hard"

	ReasonDifficultyUp   = "连续3次主观感受为轻松，自动提升难度"
	ReasonDifficultyDown = "连续2次主观感受为很吃力，自动降低难度"
)

var ValidSubjectiveFeelings = map[string]bool{
	FeelingEasy:    true,
	FeelingNormal:  true,
	FeelingHard:    true,
	FeelingVeryHard: true,
}

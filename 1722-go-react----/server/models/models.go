package models

import (
	"time"

	"gorm.io/gorm"
)

type KnowledgePoint struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"uniqueIndex:idx_name_subject;not null"`
	Subject   string         `json:"subject" gorm:"uniqueIndex:idx_name_subject;not null"`
	ParentID  *uint          `json:"parent_id"`
	Children  []KnowledgePoint `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Level     int            `json:"level"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type Question struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	QuestionNumber    string         `json:"question_number" gorm:"unique;not null"`
	Content           string         `json:"content" gorm:"type:text;not null"`
	Type              QuestionType   `json:"type" gorm:"type:varchar(20);not null"`
	KnowledgePointID  uint           `json:"knowledge_point_id" gorm:"not null"`
	KnowledgePoint    KnowledgePoint `json:"knowledge_point,omitempty" gorm:"foreignKey:KnowledgePointID"`
	Difficulty        int            `json:"difficulty" gorm:"not null;check:difficulty >= 1 AND difficulty <= 5"`
	CorrectAnswer     string         `json:"correct_answer" gorm:"type:text;not null"`
	Options           string         `json:"options,omitempty" gorm:"type:text"`
	Explanation       string         `json:"explanation,omitempty" gorm:"type:text"`
	Score             float64        `json:"score" gorm:"not null;default:1"`
	Status            QuestionStatus `json:"status" gorm:"type:varchar(20);not null;default:'draft'"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type QuestionType string

const (
	QuestionTypeSingleChoice QuestionType = "single_choice"
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeTrueFalse QuestionType = "true_false"
	QuestionTypeFillBlank QuestionType = "fill_blank"
	QuestionTypeEssay QuestionType = "essay"
)

type QuestionStatus string

const (
	QuestionStatusDraft QuestionStatus = "draft"
	QuestionStatusPendingReview QuestionStatus = "pending_review"
	QuestionStatusApproved QuestionStatus = "approved"
	QuestionStatusPublished QuestionStatus = "published"
	QuestionStatusOffline QuestionStatus = "offline"
	QuestionStatusRejected QuestionStatus = "rejected"
)

type Exam struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	StudentID      string         `json:"student_id" gorm:"not null;index"`
	Status         ExamStatus     `json:"status" gorm:"type:varchar(20);not null;default:'in_progress'"`
	CurrentQuestion int           `json:"current_question" gorm:"default:0"`
	CurrentDifficulty int         `json:"current_difficulty" gorm:"default:3"`
	ConsecutiveCorrect int        `json:"consecutive_correct" gorm:"default:0"`
	ConsecutiveWrong int          `json:"consecutive_wrong" gorm:"default:0"`
	TotalScore     float64        `json:"total_score" gorm:"default:0"`
	CorrectCount   int            `json:"correct_count" gorm:"default:0"`
	TotalQuestions int            `json:"total_questions" gorm:"default:20"`
	StartTime      *time.Time     `json:"start_time"`
	EndTime        *time.Time     `json:"end_time"`
	BillID         *uint          `json:"bill_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type ExamStatus string

const (
	ExamStatusInProgress ExamStatus = "in_progress"
	ExamStatusCompleted  ExamStatus = "completed"
	ExamStatusIncomplete ExamStatus = "incomplete"
)

type ExamQuestion struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	ExamID       uint           `json:"exam_id" gorm:"not null;index"`
	QuestionID   uint           `json:"question_id" gorm:"not null;index"`
	Question     Question       `json:"question,omitempty" gorm:"foreignKey:QuestionID"`
	OrderNumber  int            `json:"order_number" gorm:"not null"`
	DifficultyAtTime int        `json:"difficulty_at_time"`
	StudentAnswer string         `json:"student_answer,omitempty" gorm:"type:text"`
	IsCorrect    *bool          `json:"is_correct"`
	ScoreEarned  float64        `json:"score_earned" gorm:"default:0"`
	TimeSpent    int            `json:"time_spent" gorm:"default:0"`
	SubmittedAt  *time.Time     `json:"submitted_at"`
	NeedsGrading bool           `json:"needs_grading" gorm:"default:false"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type WrongAnswer struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	StudentID        string         `json:"student_id" gorm:"not null;index:idx_student_question,unique"`
	QuestionID       uint           `json:"question_id" gorm:"not null;index:idx_student_question,unique"`
	Question         Question       `json:"question,omitempty" gorm:"foreignKey:QuestionID"`
	ErrorCount       int            `json:"error_count" gorm:"default:1"`
	IsReviewed       bool           `json:"is_reviewed" gorm:"default:false"`
	IsPriority       bool           `json:"is_priority" gorm:"default:false"`
	LastWrongExamID  *uint          `json:"last_wrong_exam_id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type KnowledgeMastery struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	StudentID        string         `json:"student_id" gorm:"not null;index:idx_student_kp,unique"`
	KnowledgePointID uint           `json:"knowledge_point_id" gorm:"not null;index:idx_student_kp,unique"`
	KnowledgePoint   KnowledgePoint `json:"knowledge_point,omitempty" gorm:"foreignKey:KnowledgePointID"`
	TotalAnswered    int            `json:"total_answered" gorm:"default:0"`
	CorrectAnswered  int            `json:"correct_answered" gorm:"default:0"`
	MasteryLevel     MasteryLevel   `json:"mastery_level" gorm:"type:varchar(20);default:'weak'"`
	LastUpdated      time.Time      `json:"last_updated"`
}

type MasteryLevel string

const (
	MasteryLevelMastered MasteryLevel = "mastered"
	MasteryLevelReinforce MasteryLevel = "reinforce"
	MasteryLevelWeak MasteryLevel = "weak"
)

type Bill struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	StudentID       string         `json:"student_id" gorm:"not null;index"`
	TotalAmount     float64        `json:"total_amount" gorm:"not null"`
	RemainingAmount float64        `json:"remaining_amount" gorm:"not null"`
	Status          BillStatus     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type BillStatus string

const (
	BillStatusActive   BillStatus = "active"
	BillStatusExhausted BillStatus = "exhausted"
	BillStatusCancelled BillStatus = "cancelled"
)

type BillItem struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	BillID      uint           `json:"bill_id" gorm:"not null;index"`
	ExamID      *uint          `json:"exam_id"`
	Description string         `json:"description"`
	Amount      float64        `json:"amount" gorm:"not null"`
	IsCompleted bool           `json:"is_completed" gorm:"default:false"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

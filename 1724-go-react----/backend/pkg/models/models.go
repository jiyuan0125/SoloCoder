package models

import (
	"time"

	"gorm.io/gorm"
)

type CourseType string

const (
	CourseTypeTheory     CourseType = "theory"
	CourseTypeExperiment CourseType = "experiment"
	CourseTypeSport      CourseType = "sport"
)

type College struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"uniqueIndex;not null" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Teacher struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	CollegeID uint           `json:"college_id"`
	College   College        `json:"-"`
	Email     string         `gorm:"uniqueIndex" json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Student struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	StudentID string         `gorm:"uniqueIndex;not null" json:"student_id"`
	Email     string         `json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Course struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `json:"name"`
	CourseCode string         `gorm:"uniqueIndex;not null" json:"course_code"`
	TeacherID  uint           `json:"teacher_id"`
	Teacher    Teacher        `json:"-"`
	CourseType CourseType     `json:"course_type"`
	CollegeID  uint           `json:"college_id"`
	College    College        `json:"-"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

type QuestionType string

const (
	QuestionTypeGeneral QuestionType = "general"
	QuestionTypeSpecial QuestionType = "special"
)

type QuestionCategory string

const (
	CategoryAttitude    QuestionCategory = "attitude"
	CategoryContent     QuestionCategory = "content"
	CategoryMethod      QuestionCategory = "method"
	CategoryEffect      QuestionCategory = "effect"
	CategoryOverall     QuestionCategory = "overall"
	CategoryTheory      QuestionCategory = "theory"
	CategoryExperiment  QuestionCategory = "experiment"
	CategorySport       QuestionCategory = "sport"
)

type Question struct {
	ID            uint             `gorm:"primaryKey" json:"id"`
	Text          string           `json:"text"`
	QuestionType  QuestionType     `json:"question_type"`
	Category      QuestionCategory `json:"category"`
	CourseType    CourseType       `json:"course_type"`
	OrderIndex    int              `json:"order_index"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	DeletedAt     gorm.DeletedAt   `gorm:"index" json:"-"`
}

type EvaluationTaskStatus string

const (
	TaskStatusDraft     EvaluationTaskStatus = "draft"
	TaskStatusReview    EvaluationTaskStatus = "review"
	TaskStatusFinal     EvaluationTaskStatus = "final"
	TaskStatusPending   EvaluationTaskStatus = "pending"
	TaskStatusActive    EvaluationTaskStatus = "active"
	TaskStatusCompleted EvaluationTaskStatus = "completed"
	TaskStatusCancelled EvaluationTaskStatus = "cancelled"
)

type EvaluationTask struct {
	ID            uint                 `gorm:"primaryKey" json:"id"`
	Semester      string               `gorm:"uniqueIndex:idx_semester;not null" json:"semester"`
	StartDate     time.Time            `json:"start_date"`
	EndDate       time.Time            `json:"end_date"`
	Status        EvaluationTaskStatus `json:"status"`
	Courses       []Course             `gorm:"many2many:task_courses" json:"courses"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
	DeletedAt     gorm.DeletedAt       `gorm:"index" json:"-"`
}

type EvaluationSubmission struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	StudentID        uint           `gorm:"uniqueIndex:idx_student_course_task;not null" json:"student_id"`
	Student          Student        `json:"-"`
	CourseID         uint           `gorm:"uniqueIndex:idx_student_course_task;not null" json:"course_id"`
	Course           Course         `json:"-"`
	TaskID           uint           `gorm:"uniqueIndex:idx_student_course_task;not null" json:"task_id"`
	Task             EvaluationTask `json:"-"`
	SubmittedAt      time.Time      `json:"submitted_at"`
	Comment          string         `gorm:"type:text" json:"comment"`
	FilteredComment  string         `gorm:"type:text" json:"filtered_comment"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

type Answer struct {
	ID             uint                 `gorm:"primaryKey" json:"id"`
	SubmissionID   uint                 `gorm:"not null" json:"submission_id"`
	Submission     EvaluationSubmission `json:"-"`
	QuestionID     uint                 `gorm:"not null" json:"question_id"`
	Question       Question             `json:"-"`
	Score          int                  `gorm:"not null" json:"score"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	DeletedAt      gorm.DeletedAt       `gorm:"index" json:"-"`
}

type DataQuality string

const (
	DataQualityInsufficient DataQuality = "insufficient"
	DataQualityReference    DataQuality = "reference"
	DataQualityValid        DataQuality = "valid"
)

type CourseResult struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	TaskID        uint        `gorm:"not null" json:"task_id"`
	CourseID      uint        `gorm:"not null" json:"course_id"`
	CollegeID     uint        `json:"college_id"`
	TeacherID     uint        `json:"teacher_id"`
	CourseType    CourseType  `json:"course_type"`
	TotalStudents int         `json:"total_students"`
	SubmittedCount int        `json:"submitted_count"`
	SubmissionRate float64    `json:"submission_rate"`
	DataQuality   DataQuality `json:"data_quality"`
	AttitudeScore   float64   `json:"attitude_score"`
	ContentScore    float64   `json:"content_score"`
	MethodScore     float64   `json:"method_score"`
	EffectScore     float64   `json:"effect_score"`
	SpecialScore    float64   `json:"special_score"`
	GeneralAverage  float64   `json:"general_average"`
	OverallScore    float64   `json:"overall_score"`
	CollegeRank     int       `json:"college_rank"`
	TeacherRank     int       `json:"teacher_rank"`
	TypeRank        int       `json:"type_rank"`
	IsLowScore      bool      `json:"is_low_score"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type FeedbackStatus string

const (
	FeedbackStatusPending    FeedbackStatus = "pending"
	FeedbackStatusPlanFilled FeedbackStatus = "plan_filled"
	FeedbackStatusReviewing  FeedbackStatus = "reviewing"
	FeedbackStatusClosed     FeedbackStatus = "closed"
)

type FeedbackItem struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CourseResultID uint          `gorm:"not null" json:"course_result_id"`
	CourseResult CourseResult   `json:"-"`
	TeacherID    uint           `json:"teacher_id"`
	Teacher      Teacher        `json:"-"`
	Status       FeedbackStatus `json:"status"`
	ImprovementPlan string       `gorm:"type:text" json:"improvement_plan"`
	PlanDeadline time.Time      `json:"plan_deadline"`
	SupervisorID uint           `json:"supervisor_id"`
	NextReviewAt time.Time      `json:"next_review_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

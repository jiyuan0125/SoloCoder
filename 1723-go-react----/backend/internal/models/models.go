package models

import (
	"time"
)

type Difficulty string

const (
	DifficultyBeginner     Difficulty = "beginner"
	DifficultyElementary   Difficulty = "elementary"
	DifficultyIntermediate Difficulty = "intermediate"
	DifficultyAdvanced     Difficulty = "advanced"
)

type Domain string

const (
	DomainFrontend  Domain = "frontend"
	DomainBackend   Domain = "backend"
	DomainDatabase  Domain = "database"
	DomainAlgorithm Domain = "algorithm"
	DomainDevOps    Domain = "devops"
)

type Course struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	ExpectedHours   int        `json:"expected_hours"`
	Difficulty      Difficulty `json:"difficulty"`
	Domain          Domain     `json:"domain"`
	Units           []Unit     `json:"units"`
	Quiz            Quiz       `json:"quiz"`
	PrerequisiteIDs []string   `json:"prerequisite_ids"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Unit struct {
	ID          string    `json:"id"`
	CourseID    string    `json:"course_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Order       int       `json:"order"`
	CreatedAt   time.Time `json:"created_at"`
}

type Quiz struct {
	ID        string     `json:"id"`
	CourseID  string     `json:"course_id"`
	Questions []Question `json:"questions"`
	CreatedAt time.Time  `json:"created_at"`
}

type Question struct {
	ID            string   `json:"id"`
	QuizID        string   `json:"quiz_id"`
	Text          string   `json:"text"`
	Options       []string `json:"options"`
	CorrectIndex  int      `json:"correct_index"`
	Explanation   string   `json:"explanation"`
}

type Student struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Enrollment struct {
	ID         string    `json:"id"`
	StudentID  string    `json:"student_id"`
	CourseID   string    `json:"course_id"`
	Status     string    `json:"status"`
	StartedAt  time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type LearningProgress struct {
	ID                 string    `json:"id"`
	EnrollmentID       string    `json:"enrollment_id"`
	StudentID          string    `json:"student_id"`
	CourseID           string    `json:"course_id"`
	CompletedUnitIDs   []string  `json:"completed_unit_ids"`
	ProgressPercentage float64   `json:"progress_percentage"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type QuizAttempt struct {
	ID            string    `json:"id"`
	StudentID     string    `json:"student_id"`
	CourseID      string    `json:"course_id"`
	QuizID        string    `json:"quiz_id"`
	QuestionIDs   []string  `json:"question_ids"`
	Answers       []int     `json:"answers"`
	CorrectCount  int       `json:"correct_count"`
	TotalCount    int       `json:"total_count"`
	Passed        bool      `json:"passed"`
	Score         float64   `json:"score"`
	AttemptedAt   time.Time `json:"attempted_at"`
}

type LearningRecord struct {
	ID            string    `json:"id"`
	StudentID     string    `json:"student_id"`
	UnitID        string    `json:"unit_id"`
	CourseID      string    `json:"course_id"`
	TimeSpent     int       `json:"time_spent"`
	CompletedAt   time.Time `json:"completed_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type LearningAnalytics struct {
	StudentID              string  `json:"student_id"`
	TotalLearningHours     float64 `json:"total_learning_hours"`
	LearningSpeed          float64 `json:"learning_speed"`
	MasteryLevel           float64 `json:"mastery_level"`
	AverageQuizScore       float64 `json:"average_quiz_score"`
	CompletedCourses       int     `json:"completed_courses"`
	CurrentActiveCourses   int     `json:"current_active_courses"`
	ConsecutiveLearningDays int    `json:"consecutive_learning_days"`
}

type Achievement struct {
	ID           string    `json:"id"`
	StudentID    string    `json:"student_id"`
	Type         string    `json:"type"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	AchievedAt   time.Time `json:"achieved_at"`
}

type CoursePath struct {
	ID          string       `json:"id"`
	StudentID   string       `json:"student_id"`
	TargetGoal  string       `json:"target_goal"`
	PathNodes   []PathNode   `json:"path_nodes"`
	TotalCourses int         `json:"total_courses"`
	CreatedAt   time.Time    `json:"created_at"`
}

type PathNode struct {
	CourseID       string   `json:"course_id"`
	CourseName     string   `json:"course_name"`
	Level          int      `json:"level"`
	ParallelGroup  int      `json:"parallel_group"`
	Status         string   `json:"status"`
	Progress       float64  `json:"progress"`
}

type Product struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        float64   `json:"price"`
	Stock        int       `json:"stock"`
	SafetyStock  int       `json:"safety_stock"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PriceHistory struct {
	ID         string    `json:"id"`
	ProductID  string    `json:"product_id"`
	OldPrice   float64   `json:"old_price"`
	NewPrice   float64   `json:"new_price"`
	ChangedAt  time.Time `json:"changed_at"`
	ChangedBy  string    `json:"changed_by"`
}

type Order struct {
	ID             string    `json:"id"`
	StudentID      string    `json:"student_id"`
	OrderItems     []OrderItem `json:"order_items"`
	TotalAmount    float64   `json:"total_amount"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	ConfirmedAt    *time.Time `json:"confirmed_at"`
}

type OrderItem struct {
	ID         string  `json:"id"`
	OrderID    string  `json:"order_id"`
	ProductID  string  `json:"product_id"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
}

type PurchaseRequest struct {
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

package model

import "time"

type CourseType string

const (
	CourseTypeOnline   CourseType = "online"
	CourseTypeOffline  CourseType = "offline"
	CourseTypeHybrid   CourseType = "hybrid"
)

type Position string

const (
	PositionAssistant  Position = "助教"
	PositionLecturer   Position = "讲师"
	PositionAssociate  Position = "副教授"
	PositionProfessor  Position = "教授"
)

type EnrollmentStatus string

const (
	EnrollmentStatusEnrolled  EnrollmentStatus = "enrolled"
	EnrollmentStatusWithdrawn EnrollmentStatus = "withdrawn"
	EnrollmentStatusPassed    EnrollmentStatus = "passed"
	EnrollmentStatusFailed    EnrollmentStatus = "failed"
	EnrollmentStatusAbsent    EnrollmentStatus = "absent"
)

type TodoType string

const (
	TodoTypeCourseStart     TodoType = "course_start"
	TodoTypeResultAvailable TodoType = "result_available"
)

type Course struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Type           CourseType `json:"type"`
	Hours          int        `json:"hours"`
	Credits        int        `json:"credits"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        time.Time  `json:"end_date"`
	Instructor     string     `json:"instructor"`
	Description    string     `json:"description"`
	Capacity       int        `json:"capacity"`
	EnrolledCount  int        `json:"enrolled_count"`
	CreatedAt      time.Time  `json:"created_at"`
}

type Teacher struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	EmployeeID   string   `json:"employee_id"`
	College      string   `json:"college"`
	Position     Position `json:"position"`
	CreatedAt    time.Time `json:"created_at"`
}

type CollegeHistory struct {
	ID        string    `json:"id"`
	TeacherID string    `json:"teacher_id"`
	College   string    `json:"college"`
	StartDate time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

type Enrollment struct {
	ID         string           `json:"id"`
	CourseID   string           `json:"course_id"`
	TeacherID  string           `json:"teacher_id"`
	Status     EnrollmentStatus `json:"status"`
	Score      *int             `json:"score,omitempty"`
	College    string           `json:"college"`
	EnrolledAt time.Time        `json:"enrolled_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

type Todo struct {
	ID          string    `json:"id"`
	TeacherID   string    `json:"teacher_id"`
	CourseID    string    `json:"course_id"`
	Type        TodoType  `json:"type"`
	Message     string    `json:"message"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func GetCreditThreshold(position Position) int {
	switch position {
	case PositionAssistant:
		return 20
	case PositionLecturer:
		return 25
	case PositionAssociate:
		return 30
	case PositionProfessor:
		return 35
	default:
		return 20
	}
}

func CalculateCredits(hours int) int {
	if hours <= 0 {
		return 1
	}
	credits := hours / 8
	if credits < 1 {
		return 1
	}
	return credits
}

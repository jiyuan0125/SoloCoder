package models

import (
	"time"
)

type CourseType string

const (
	RequiredCourse CourseType = "required"
	ElectiveCourse CourseType = "elective"
)

type Course struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Credits    int        `json:"credits"`
	ValidUntil time.Time  `json:"valid_until"`
	Type       CourseType `json:"type"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type Employee struct {
	ID                         string            `json:"id"`
	Name                       string            `json:"name"`
	Department                 string            `json:"department"`
	CompletedRequiredCourses   map[string]bool   `json:"completed_required_courses"`
	CarryOverCredits           map[int]int       `json:"carry_over_credits"`
	CreatedAt                  time.Time         `json:"created_at"`
	UpdatedAt                  time.Time         `json:"updated_at"`
}

type CreditRecord struct {
	ID          string    `json:"id"`
	EmployeeID  string    `json:"employee_id"`
	CourseID    string    `json:"course_id"`
	Credits     int       `json:"credits"`
	CourseType  CourseType `json:"course_type"`
	CompletedAt time.Time `json:"completed_at"`
}

type CreditProgress struct {
	EmployeeID                string  `json:"employee_id"`
	EmployeeName              string  `json:"employee_name"`
	Department                string  `json:"department"`
	Year                      int     `json:"year"`
	RequiredCreditsCompleted  int     `json:"required_credits_completed"`
	RequiredCreditsRequired   int     `json:"required_credits_required"`
	ElectiveCreditsCompleted  int     `json:"elective_credits_completed"`
	ElectiveCreditsRequired   int     `json:"elective_credits_required"`
	CarryOverUsed             int     `json:"carry_over_used"`
	CarryOverAvailable        int     `json:"carry_over_available"`
	TotalCreditsCompleted     int     `json:"total_credits_completed"`
	TotalCreditsRequired      int     `json:"total_credits_required"`
	AllRequiredCoursesDone    bool    `json:"all_required_courses_done"`
	ProgressPercentage        float64 `json:"progress_percentage"`
}

type RankItem struct {
	Rank                     int     `json:"rank"`
	EmployeeID               string  `json:"employee_id"`
	EmployeeName             string  `json:"employee_name"`
	Department               string  `json:"department"`
	TotalCredits             int     `json:"total_credits"`
	RequiredCredits          int     `json:"required_credits"`
	ElectiveCredits          int     `json:"elective_credits"`
	AllRequiredCoursesDone   bool    `json:"all_required_courses_done"`
	ProgressPercentage       float64 `json:"progress_percentage"`
}

type YearlySummary struct {
	Year                      int     `json:"year"`
	RequiredCreditsCompleted  int     `json:"required_credits_completed"`
	ElectiveCreditsCompleted  int     `json:"elective_credits_completed"`
	CarryOverUsed             int     `json:"carry_over_used"`
	TotalCreditsCompleted     int     `json:"total_credits_completed"`
	RequiredCreditsRequired   int     `json:"required_credits_required"`
	ElectiveCreditsRequired   int     `json:"elective_credits_required"`
	TotalCreditsRequired      int     `json:"total_credits_required"`
	AllRequiredCoursesDone    bool    `json:"all_required_courses_done"`
	CarryOverToNextYear       int     `json:"carry_over_to_next_year"`
}

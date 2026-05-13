package models

import "time"

type Employee struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name" binding:"required"`
	TeamID    int64     `json:"team_id" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Quarter struct {
	ID        int64  `json:"id"`
	Year      int    `json:"year"`
	Quarter   int    `json:"quarter"`
	CreatedAt time.Time
}

type Review struct {
	ID              int64     `json:"id"`
	EmployeeID      int64     `json:"employee_id" binding:"required"`
	QuarterID       int64     `json:"quarter_id" binding:"required"`
	Quality         float64   `json:"quality" binding:"required"`
	Efficiency      float64   `json:"efficiency" binding:"required"`
	Collaboration   float64   `json:"collaboration" binding:"required"`
	Innovation      float64   `json:"innovation" binding:"required"`
	TotalScore      float64   `json:"total_score"`
	Level           string    `json:"level"`
	FinalLevel      string    `json:"final_level"`
	Status          string    `json:"status"`
	NeedImprovement bool      `json:"need_improvement"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AnnualReview struct {
	ID              int64     `json:"id"`
	EmployeeID      int64     `json:"employee_id"`
	Year            int       `json:"year"`
	AverageScore    float64   `json:"average_score"`
	Level           string    `json:"level"`
	CreatedAt       time.Time `json:"created_at"`
}

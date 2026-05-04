package common

import "time"

const (
	StatusPending  = "待审核"
	StatusAccepted = "已录取"
	StatusRejected = "已拒绝"
)

type EnrollmentPlan struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Grade          string    `json:"grade"`
	Classes        int       `json:"classes"`
	MaxPerClass    int       `json:"max_per_class"`
	TotalCapacity  int       `json:"total_capacity"`
	EnrolledCount  int       `json:"enrolled_count"`
	IsOpen         bool      `json:"is_open"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Registration struct {
	ID              string    `json:"id"`
	PlanID          string    `json:"plan_id"`
	StudentName     string    `json:"student_name"`
	IDCard          string    `json:"id_card"`
	ParentPhone     string    `json:"parent_phone"`
	Address         string    `json:"address"`
	Status          string    `json:"status"`
	RejectReason    string    `json:"reject_reason"`
	SubmittedAt     time.Time `json:"submitted_at"`
	ReviewedAt      time.Time `json:"reviewed_at"`
}

type CreatePlanRequest struct {
	Name        string `json:"name"`
	Grade       string `json:"grade"`
	Classes     int    `json:"classes"`
	MaxPerClass int    `json:"max_per_class"`
}

type UpdatePlanRequest struct {
	Classes     int `json:"classes"`
	MaxPerClass int `json:"max_per_class"`
}

type SubmitRegistrationRequest struct {
	PlanID      string `json:"plan_id"`
	StudentName string `json:"student_name"`
	IDCard      string `json:"id_card"`
	ParentPhone string `json:"parent_phone"`
	Address     string `json:"address"`
}

type ReviewRegistrationRequest struct {
	RegistrationID string `json:"registration_id"`
	Action         string `json:"action"`
	RejectReason   string `json:"reject_reason"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type QueryRegistrationsRequest struct {
	PlanID string `json:"plan_id"`
	Status string `json:"status"`
}

type QueryRegistrationByIDCardRequest struct {
	PlanID string `json:"plan_id"`
	IDCard string `json:"id_card"`
}

type ClosePlanRequest struct {
	PlanID string `json:"plan_id"`
}

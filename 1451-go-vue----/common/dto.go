package common

import "time"

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type AccessRequest struct {
	EmployeeID    string `json:"employee_id"`
	AccessPointID string `json:"access_point_id"`
	AccessType    string `json:"access_type"`
}

type CreateAccessPointRequest struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	AreaID     string `json:"area_id"`
	BuildingID string `json:"building_id"`
}

type CreateAccessRuleRequest struct {
	AccessPointID    string   `json:"access_point_id"`
	AllowedDepartments []string `json:"allowed_departments"`
	StartTime        string   `json:"start_time"`
	EndTime          string   `json:"end_time"`
}

type BatchAreaRuleRequest struct {
	AreaID           string   `json:"area_id"`
	AllowedDepartments []string `json:"allowed_departments"`
	StartTime        string   `json:"start_time"`
	EndTime          string   `json:"end_time"`
}

type CreateVisitorRequest struct {
	EmployeeID      string    `json:"employee_id"`
	VisitorName     string    `json:"visitor_name"`
	VisitorPhone    string    `json:"visitor_phone"`
	Purpose         string    `json:"purpose"`
	ExpectedArrival time.Time `json:"expected_arrival"`
}

type ReviewVisitorRequest struct {
	ReservationID string `json:"reservation_id"`
	Approve       bool   `json:"approve"`
}

type CheckInVisitorRequest struct {
	VisitorPhone string `json:"visitor_phone"`
}

type CreatePatrolRouteRequest struct {
	Name     string   `json:"name"`
	PointIDs []string `json:"point_ids"`
}

type CreatePatrolTaskRequest struct {
	RouteID    string    `json:"route_id"`
	AssigneeID string    `json:"assignee_id"`
	Frequency  string    `json:"frequency"`
	StartTime  time.Time `json:"start_time"`
}

type StartPatrolRequest struct {
	TaskID string `json:"task_id"`
}

type PatrolCheckInRequest struct {
	ExecutionID            string `json:"execution_id"`
	PointID                string `json:"point_id"`
	HasAnomaly             bool   `json:"has_anomaly"`
	AnomalyDesc            string `json:"anomaly_desc"`
	RelatedAccessPointID   string `json:"related_access_point_id"`
}

type FixAccessPointRequest struct {
	AccessPointID string `json:"access_point_id"`
}

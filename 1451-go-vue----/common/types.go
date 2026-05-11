package common

import "time"

type Employee struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Department string `json:"department"`
}

type AccessPoint struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	AreaID     string `json:"area_id"`
	BuildingID string `json:"building_id"`
	IsFault    bool   `json:"is_fault"`
	IsOpenMode bool  `json:"is_open_mode"`
}

type Area struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AccessRule struct {
	ID               string   `json:"id"`
	AccessPointID    string   `json:"access_point_id"`
	AllowedDepartments []string `json:"allowed_departments"`
	StartTime        string   `json:"start_time"`
	EndTime          string   `json:"end_time"`
	IsActive         bool     `json:"is_active"`
}

type AccessRecord struct {
	ID             string    `json:"id"`
	EmployeeID     string    `json:"employee_id"`
	EmployeeName   string    `json:"employee_name"`
	Department     string    `json:"department"`
	AccessTime     time.Time `json:"access_time"`
	AccessPointID  string    `json:"access_point_id"`
	AccessPointName string   `json:"access_point_name"`
	AccessType     string    `json:"access_type"`
}

type VisitorReservation struct {
	ID            string    `json:"id"`
	EmployeeID    string    `json:"employee_id"`
	EmployeeName  string    `json:"employee_name"`
	VisitorName   string    `json:"visitor_name"`
	VisitorPhone  string    `json:"visitor_phone"`
	Purpose       string    `json:"purpose"`
	ExpectedArrival time.Time `json:"expected_arrival"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	ReviewedAt    time.Time `json:"reviewed_at"`
	CheckInAt     time.Time `json:"check_in_at"`
}

type PatrolPoint struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PatrolRoute struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	PointIDs []string `json:"point_ids"`
}

type PatrolTask struct {
	ID         string `json:"id"`
	RouteID    string `json:"route_id"`
	AssigneeID string `json:"assignee_id"`
	Frequency  string `json:"frequency"`
	StartTime  time.Time `json:"start_time"`
}

type PatrolExecution struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	StartTime time.Time `json:"start_time"`
	Status    string    `json:"status"`
}

type PatrolCheckIn struct {
	ID           string    `json:"id"`
	ExecutionID  string    `json:"execution_id"`
	PointID      string    `json:"point_id"`
	CheckInTime  time.Time `json:"check_in_time"`
	HasAnomaly   bool      `json:"has_anomaly"`
	AnomalyDesc  string    `json:"anomaly_desc"`
	RelatedAccessPointID string `json:"related_access_point_id"`
}

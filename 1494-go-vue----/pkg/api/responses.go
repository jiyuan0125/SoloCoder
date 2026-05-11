package api

import "time"

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ZoneResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PlantResponse struct {
	ID           string       `json:"id"`
	ZoneID       string       `json:"zone_id"`
	Name         string       `json:"name"`
	Variety      string       `json:"variety"`
	PlantType    PlantType    `json:"plant_type"`
	PlantingDate time.Time    `json:"planting_date"`
	Location     string       `json:"location"`
	HealthStatus HealthStatus `json:"health_status"`
	Description  string       `json:"description"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type WorkerResponse struct {
	ID         string     `json:"id"`
	EmployeeID string     `json:"employee_id"`
	Name       string     `json:"name"`
	SkillLevel SkillLevel `json:"skill_level"`
	ZoneID     string     `json:"zone_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type MaintenancePlanResponse struct {
	ID         string                 `json:"id"`
	ZoneID     string                 `json:"zone_id"`
	PlantType  PlantType              `json:"plant_type"`
	Items      []MaintenancePlanItem  `json:"items"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

type TaskResponse struct {
	ID              string          `json:"id"`
	TaskType        TaskType        `json:"task_type"`
	ZoneID          string          `json:"zone_id"`
	PlantIDs        []string        `json:"plant_ids"`
	MaintenanceType MaintenanceType `json:"maintenance_type"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Status          TaskStatus      `json:"status"`
	EstimatedHours  float64         `json:"estimated_hours"`
	AssignedWorkerID string         `json:"assigned_worker_id,omitempty"`
	ActualHours     float64         `json:"actual_hours,omitempty"`
	MaterialCost    float64         `json:"material_cost,omitempty"`
	WorkerNotes     string          `json:"worker_notes,omitempty"`
	RejectCount     int             `json:"reject_count"`
	RejectReasons   []string        `json:"reject_reasons"`
	Priority        int             `json:"priority"`
	DueDate         time.Time       `json:"due_date"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type ZoneCostReport struct {
	ZoneID        string  `json:"zone_id"`
	ZoneName      string  `json:"zone_name"`
	MaterialCost  float64 `json:"material_cost"`
	LaborCost     float64 `json:"labor_cost"`
	TotalCost     float64 `json:"total_cost"`
	TaskCount     int     `json:"task_count"`
}

type CostReportResponse struct {
	Year         int              `json:"year"`
	Month        int              `json:"month"`
	ZoneReports  []ZoneCostReport `json:"zone_reports"`
	TotalCost    float64          `json:"total_cost"`
}

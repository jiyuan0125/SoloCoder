package api

import "time"

type CreateZoneRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreatePlantRequest struct {
	ZoneID        string       `json:"zone_id"`
	Name          string       `json:"name"`
	Variety       string       `json:"variety"`
	PlantType     PlantType    `json:"plant_type"`
	PlantingDate  time.Time    `json:"planting_date"`
	Location      string       `json:"location"`
	HealthStatus  HealthStatus `json:"health_status"`
	Description   string       `json:"description"`
}

type UpdatePlantRequest struct {
	Name          string       `json:"name"`
	Variety       string       `json:"variety"`
	PlantingDate  time.Time    `json:"planting_date"`
	Location      string       `json:"location"`
	HealthStatus  HealthStatus `json:"health_status"`
	Description   string       `json:"description"`
}

type CreateWorkerRequest struct {
	EmployeeID  string     `json:"employee_id"`
	Name        string     `json:"name"`
	SkillLevel  SkillLevel `json:"skill_level"`
	ZoneID      string     `json:"zone_id"`
}

type UpdateWorkerRequest struct {
	Name        string     `json:"name"`
	SkillLevel  SkillLevel `json:"skill_level"`
	ZoneID      string     `json:"zone_id"`
}

type MaintenancePlanItem struct {
	MaintenanceType MaintenanceType `json:"maintenance_type"`
	IntervalDays    int             `json:"interval_days"`
	EstimatedHours  float64         `json:"estimated_hours"`
}

type CreateMaintenancePlanRequest struct {
	ZoneID    string                 `json:"zone_id"`
	PlantType PlantType              `json:"plant_type"`
	Items     []MaintenancePlanItem  `json:"items"`
}

type CreateAdhocTaskRequest struct {
	ZoneID          string         `json:"zone_id"`
	PlantIDs        []string       `json:"plant_ids"`
	MaintenanceType MaintenanceType`json:"maintenance_type"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	EstimatedHours  float64        `json:"estimated_hours"`
	Priority        int            `json:"priority"`
}

type ExecuteTaskRequest struct {
	TaskID      string  `json:"task_id"`
	WorkerID    string  `json:"worker_id"`
	ActualHours float64 `json:"actual_hours"`
	Notes       string  `json:"notes"`
	MaterialCost float64 `json:"material_cost"`
}

type ReviewTaskRequest struct {
	TaskID  string `json:"task_id"`
	Approved bool   `json:"approved"`
	Reason  string `json:"reason"`
}

type GenerateTasksRequest struct {
	Date time.Time `json:"date"`
}

type CostReportRequest struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

package core

import (
	"time"

	"green-care-management/pkg/api"
)

type Zone struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Plant struct {
	ID           string
	ZoneID       string
	Name         string
	Variety      string
	PlantType    api.PlantType
	PlantingDate time.Time
	Location     string
	HealthStatus api.HealthStatus
	Description  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Worker struct {
	ID         string
	EmployeeID string
	Name       string
	SkillLevel api.SkillLevel
	ZoneID     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type MaintenancePlan struct {
	ID        string
	ZoneID    string
	PlantType api.PlantType
	Items     []api.MaintenancePlanItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Task struct {
	ID               string
	TaskType         api.TaskType
	ZoneID           string
	PlantIDs         []string
	MaintenanceType  api.MaintenanceType
	Title            string
	Description      string
	Status           api.TaskStatus
	EstimatedHours   float64
	AssignedWorkerID string
	ActualHours      float64
	MaterialCost     float64
	WorkerNotes      string
	RejectCount      int
	RejectReasons    []string
	Priority         int
	DueDate          time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

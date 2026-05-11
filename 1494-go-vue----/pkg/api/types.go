package api

type PlantType string

const (
	PlantTypeTree  PlantType = "tree"
	PlantTypeShrub PlantType = "shrub"
	PlantTypeLawn  PlantType = "lawn"
	PlantTypeFlower PlantType = "flower"
)

type HealthStatus string

const (
	HealthGood    HealthStatus = "good"
	HealthAverage HealthStatus = "average"
	HealthPoor    HealthStatus = "poor"
	HealthWithered HealthStatus = "withered"
)

type SkillLevel string

const (
	SkillJunior   SkillLevel = "junior"
	SkillIntermediate SkillLevel = "intermediate"
	SkillSenior   SkillLevel = "senior"
)

type TaskType string

const (
	TaskRegular TaskType = "regular"
	TaskAdhoc   TaskType = "adhoc"
	TaskEmergency TaskType = "emergency"
)

type TaskStatus string

const (
	TaskPending      TaskStatus = "pending"
	TaskInProgress   TaskStatus = "in_progress"
	TaskCompleted    TaskStatus = "completed"
	TaskReviewing    TaskStatus = "reviewing"
	TaskRejected     TaskStatus = "rejected"
	TaskApproved     TaskStatus = "approved"
)

type MaintenanceType string

const (
	MaintenanceWater   MaintenanceType = "water"
	MaintenancePrune   MaintenanceType = "prune"
	MaintenanceFertilize MaintenanceType = "fertilize"
	MaintenanceSpray   MaintenanceType = "spray"
	MaintenanceOther   MaintenanceType = "other"
)

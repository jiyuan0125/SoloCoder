package common

import "time"

type VehicleType string

const (
	VehicleTypeSedan    VehicleType = "小型轿车"
	VehicleTypeSUV      VehicleType = "SUV"
	VehicleTypeMPV      VehicleType = "MPV"
	VehicleTypeTruck    VehicleType = "货车"
	VehicleTypePassenger VehicleType = "客车"
)

type InspectionResult string

const (
	ResultPass    InspectionResult = "合格"
	ResultFail    InspectionResult = "不合格"
	ResultRecheck InspectionResult = "需复检"
)

type InspectionStep string

const (
	StepAppearance    InspectionStep = "外观检查"
	StepExhaust       InspectionStep = "尾气检测"
	StepSafety        InspectionStep = "安全性能检测"
	StepChassis       InspectionStep = "底盘检测"
)

type Vehicle struct {
	PlateNumber   string      `json:"plate_number"`
	VehicleType   VehicleType `json:"vehicle_type"`
	RegisterDate  time.Time   `json:"register_date"`
	LastInspectionDate time.Time `json:"last_inspection_date"`
}

type Station struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Workshops []Workshop `json:"workshops"`
	Schedules []Schedule `json:"schedules"`
}

type Workshop struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Schedule struct {
	Date      time.Time `json:"date"`
	TimeSlot  string    `json:"time_slot"`
	MaxVehicles int    `json:"max_vehicles"`
}

type Appointment struct {
	ID           string    `json:"id"`
	PlateNumber  string    `json:"plate_number"`
	StationID    string    `json:"station_id"`
	AppointmentDate time.Time `json:"appointment_date"`
	TimeSlot     string    `json:"time_slot"`
	CreatedAt    time.Time `json:"created_at"`
}

type InspectionProcess struct {
	ID                string             `json:"id"`
	AppointmentID     string             `json:"appointment_id"`
	PlateNumber       string             `json:"plate_number"`
	CurrentStep       InspectionStep     `json:"current_step"`
	Steps             []InspectionStepRecord `json:"steps"`
	Status            string             `json:"status"`
	RecheckCount      int                `json:"recheck_count"`
	TotalFeeFen       int                `json:"total_fee_fen"`
	CreatedAt         time.Time          `json:"created_at"`
}

type InspectionStepRecord struct {
	Step        InspectionStep   `json:"step"`
	Result      InspectionResult `json:"result"`
	FailItems   []FailItem       `json:"fail_items,omitempty"`
	CompletedAt time.Time        `json:"completed_at"`
	IsRecheck   bool             `json:"is_recheck"`
}

type FailItem struct {
	Item    string `json:"item"`
	Reason  string `json:"reason"`
}

type InspectionReport struct {
	ID              string               `json:"id"`
	ProcessID       string               `json:"process_id"`
	PlateNumber     string               `json:"plate_number"`
	VehicleType     VehicleType          `json:"vehicle_type"`
	RegisterDate    time.Time            `json:"register_date"`
	InspectionDate  time.Time            `json:"inspection_date"`
	Steps           []InspectionStepRecord `json:"steps"`
	OverallResult   string               `json:"overall_result"`
	TotalFeeFen     int                  `json:"total_fee_fen"`
	IssuedAt        time.Time            `json:"issued_at"`
}

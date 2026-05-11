package common

import "time"

type RegisterVehicleRequest struct {
	PlateNumber        string    `json:"plate_number"`
	VehicleType        VehicleType `json:"vehicle_type"`
	RegisterDate       time.Time `json:"register_date"`
	LastInspectionDate time.Time `json:"last_inspection_date"`
}

type CreateAppointmentRequest struct {
	PlateNumber     string    `json:"plate_number"`
	StationID       string    `json:"station_id"`
	AppointmentDate time.Time `json:"appointment_date"`
	TimeSlot        string    `json:"time_slot"`
}

type CreateStationRequest struct {
	Name      string     `json:"name"`
	Workshops []Workshop `json:"workshops"`
}

type AddScheduleRequest struct {
	StationID   string    `json:"station_id"`
	Date        time.Time `json:"date"`
	TimeSlot    string    `json:"time_slot"`
	MaxVehicles int       `json:"max_vehicles"`
}

type StartInspectionRequest struct {
	AppointmentID string `json:"appointment_id"`
}

type CompleteStepRequest struct {
	ProcessID string           `json:"process_id"`
	Result    InspectionResult `json:"result"`
	FailItems []FailItem       `json:"fail_items,omitempty"`
}

type StartRecheckRequest struct {
	ProcessID string `json:"process_id"`
}

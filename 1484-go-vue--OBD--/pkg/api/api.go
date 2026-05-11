package api

import "time"

type OBDPacket struct {
	DeviceID          string    `json:"device_id"`
	Timestamp         time.Time `json:"timestamp"`
	Latitude          float64   `json:"latitude"`
	Longitude         float64   `json:"longitude"`
	Speed             float64   `json:"speed"`
	EngineRPM         int       `json:"engine_rpm"`
	FuelConsumption   float64   `json:"fuel_consumption"`
	HardAccelerations int       `json:"hard_accelerations"`
	HardBrakes        int       `json:"hard_brakes"`
	HardTurns         int       `json:"hard_turns"`
}

type UploadDataRequest struct {
	DeviceID string       `json:"device_id"`
	Packets  []*OBDPacket `json:"packets"`
}

type UploadDataResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Uploaded  int    `json:"uploaded"`
	Deduplicated int `json:"deduplicated"`
}

type VehicleDailyStats struct {
	Date           time.Time `json:"date"`
	DeviceID       string    `json:"device_id"`
	Mileage        float64   `json:"mileage"`
	AvgFuel        float64   `json:"avg_fuel"`
	SafetyScore    int       `json:"safety_score"`
	TotalFuel      float64   `json:"total_fuel"`
}

type VehicleDailyStatsResponse struct {
	Success bool                `json:"success"`
	Data    []*VehicleDailyStats `json:"data"`
}

type FleetDailyStats struct {
	Date                  time.Time `json:"date"`
	TotalMileage          float64   `json:"total_mileage"`
	AvgFuelConsumption    float64   `json:"avg_fuel_consumption"`
	AbnormalEvents        int       `json:"abnormal_events"`
}

type FleetDailyStatsResponse struct {
	Success bool                `json:"success"`
	Data    []*FleetDailyStats `json:"data"`
}

type Alert struct {
	ID        string    `json:"id"`
	DeviceID  string    `json:"device_id"`
	Type      string    `json:"type"`
	Date      time.Time `json:"date"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type AlertQueryRequest struct {
	DeviceID string `json:"device_id,omitempty"`
	Status   string `json:"status,omitempty"`
}

type AlertResponse struct {
	Success bool     `json:"success"`
	Data    []*Alert `json:"data"`
}

type AlertResolveRequest struct {
	AlertIDs []string `json:"alert_ids"`
}

type CSVExportRequest struct {
	DeviceID  string `json:"device_id,omitempty"`
	YearMonth string `json:"year_month"`
}

type CSVExportResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	File    []byte `json:"-"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

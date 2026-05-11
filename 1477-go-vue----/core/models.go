package core

import "time"

type ChargerType string

const (
	FastCharger   ChargerType = "fast"
	SlowCharger   ChargerType = "slow"
)

type ChargerStatus string

const (
	StatusIdle       ChargerStatus = "idle"
	StatusCharging   ChargerStatus = "charging"
	StatusPaused     ChargerStatus = "paused"
	StatusFault      ChargerStatus = "fault"
	StatusOffline    ChargerStatus = "offline"
	StatusReserved   ChargerStatus = "reserved"
)

type Charger struct {
	ID             string        `json:"id"`
	StationID      string        `json:"station_id"`
	Code           string        `json:"code"`
	Type           ChargerType   `json:"type"`
	Status         ChargerStatus `json:"status"`
	MeterReading   float64       `json:"meter_reading"`
}

type Station struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Location   string    `json:"location"`
	Chargers   []*Charger `json:"chargers"`
}

type Reservation struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	StationID      string    `json:"station_id"`
	ChargerID      string    `json:"charger_id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Status         ReservationStatus `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	AutoCancelTime time.Time `json:"auto_cancel_time"`
}

type ReservationStatus string

const (
	ReservationActive    ReservationStatus = "active"
	ReservationCompleted ReservationStatus = "completed"
	ReservationCancelled ReservationStatus = "cancelled"
	ReservationInProgress ReservationStatus = "in_progress"
)

type ChargingSession struct {
	ID             string    `json:"id"`
	ReservationID  string    `json:"reservation_id"`
	UserID         string    `json:"user_id"`
	StationID      string    `json:"station_id"`
	ChargerID      string    `json:"charger_id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time,omitempty"`
	StartMeter     float64   `json:"start_meter"`
	EndMeter       float64   `json:"end_meter,omitempty"`
	EnergyUsed     float64   `json:"energy_used,omitempty"`
	Amount         int64     `json:"amount,omitempty"`
	Status         SessionStatus `json:"status"`
	PauseStartTime time.Time `json:"pause_start_time,omitempty"`
	TotalPauseDuration time.Duration `json:"total_pause_duration"`
}

type SessionStatus string

const (
	SessionCharging SessionStatus = "charging"
	SessionPaused   SessionStatus = "paused"
	SessionCompleted SessionStatus = "completed"
)

type ChargingRecord struct {
	ID             string    `json:"id"`
	ReservationID  string    `json:"reservation_id"`
	UserID         string    `json:"user_id"`
	StationID      string    `json:"station_id"`
	ChargerID      string    `json:"charger_id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	EnergyUsed     float64   `json:"energy_used"`
	Amount         int64     `json:"amount"`
	ChargingType   ChargerType `json:"charging_type"`
	Duration       time.Duration `json:"duration"`
	Month          string    `json:"month"`
}

type MonthlySummary struct {
	Month         string  `json:"month"`
	TotalSessions int     `json:"total_sessions"`
	TotalEnergy   float64 `json:"total_energy"`
	TotalAmount   int64   `json:"total_amount"`
}

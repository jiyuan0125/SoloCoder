package common

import "time"

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ListStationsResponse struct {
	Stations []*StationDTO `json:"stations"`
}

type StationDTO struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Location string        `json:"location"`
	Chargers []*ChargerDTO `json:"chargers"`
}

type ChargerDTO struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	MeterReading float64 `json:"meter_reading"`
}

type CreateReservationRequest struct {
	UserID    string    `json:"user_id"`
	StationID string    `json:"station_id"`
	ChargerID string    `json:"charger_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type CreateReservationResponse struct {
	Reservation *ReservationDTO `json:"reservation"`
}

type ReservationDTO struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	StationID string    `json:"station_id"`
	ChargerID string    `json:"charger_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
}

type StartChargingRequest struct {
	ReservationID string `json:"reservation_id"`
}

type StartChargingResponse struct {
	Session *SessionDTO `json:"session"`
}

type SessionDTO struct {
	ID                 string        `json:"id"`
	ReservationID      string        `json:"reservation_id"`
	UserID             string        `json:"user_id"`
	StationID          string        `json:"station_id"`
	ChargerID          string        `json:"charger_id"`
	StartTime          time.Time     `json:"start_time"`
	EndTime            time.Time     `json:"end_time,omitempty"`
	StartMeter         float64       `json:"start_meter"`
	EndMeter           float64       `json:"end_meter,omitempty"`
	EnergyUsed         float64       `json:"energy_used,omitempty"`
	Amount             int64         `json:"amount,omitempty"`
	Status             string        `json:"status"`
	TotalPauseDuration time.Duration `json:"total_pause_duration"`
}

type PauseChargingRequest struct {
	SessionID string `json:"session_id"`
}

type ResumeChargingRequest struct {
	SessionID string `json:"session_id"`
}

type EndChargingRequest struct {
	SessionID string  `json:"session_id"`
	EndMeter  float64 `json:"end_meter"`
}

type EndChargingResponse struct {
	Session *SessionDTO `json:"session"`
}

type ListChargingRecordsRequest struct {
	UserID string `json:"user_id"`
}

type ListChargingRecordsResponse struct {
	Records []*ChargingRecordDTO `json:"records"`
}

type ChargingRecordDTO struct {
	ID           string        `json:"id"`
	ReservationID string       `json:"reservation_id"`
	UserID       string        `json:"user_id"`
	StationID    string        `json:"station_id"`
	ChargerID    string        `json:"charger_id"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	EnergyUsed   float64       `json:"energy_used"`
	Amount       int64         `json:"amount"`
	ChargingType string        `json:"charging_type"`
	Duration     time.Duration `json:"duration"`
	Month        string        `json:"month"`
}

type GetMonthlySummaryRequest struct {
	UserID string `json:"user_id"`
}

type GetMonthlySummaryResponse struct {
	Summaries []*MonthlySummaryDTO `json:"summaries"`
}

type MonthlySummaryDTO struct {
	Month         string  `json:"month"`
	TotalSessions int     `json:"total_sessions"`
	TotalEnergy   float64 `json:"total_energy"`
	TotalAmount   int64   `json:"total_amount"`
}

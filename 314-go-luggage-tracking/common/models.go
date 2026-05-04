package common

import "time"

type LuggageStage struct {
	Stage       string    `json:"stage"`
	ScannedAt   time.Time `json:"scanned_at"`
	Operator    string    `json:"operator"`
	IsBackfilled bool     `json:"is_backfilled"`
}

type Luggage struct {
	TagNumber    string         `json:"tag_number"`
	FlightNumber string         `json:"flight_number"`
	Status       string         `json:"status"`
	Stages       []LuggageStage `json:"stages"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Completed    bool           `json:"completed"`
}

type ScanRequest struct {
	TagNumber    string `json:"tag_number"`
	FlightNumber string `json:"flight_number,omitempty"`
	Stage        string `json:"stage"`
	Operator     string `json:"operator"`
}

type BackfillRequest struct {
	TagNumber    string    `json:"tag_number"`
	Stage        string    `json:"stage"`
	Operator     string    `json:"operator"`
	ScannedAt    time.Time `json:"scanned_at,omitempty"`
}

type LuggageResponse struct {
	TagNumber    string         `json:"tag_number"`
	FlightNumber string         `json:"flight_number"`
	Status       string         `json:"status"`
	Stages       []LuggageStage `json:"stages"`
	CurrentStage string         `json:"current_stage"`
	CreatedAt    time.Time      `json:"created_at"`
}

type FlightLuggageResponse struct {
	FlightNumber string            `json:"flight_number"`
	Date         string            `json:"date"`
	Luggages     []LuggageResponse `json:"luggages"`
}

type OperationLog struct {
	ID        string    `json:"id"`
	Operator  string    `json:"operator"`
	Role      string    `json:"role"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

type AuditLog struct {
	ID        string    `json:"id"`
	Operator  string    `json:"operator"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
	Immutable bool      `json:"immutable"`
}

type FlightCancelRequest struct {
	FlightNumber string `json:"flight_number"`
	Date         string `json:"date"`
	Operator     string `json:"operator"`
}

type QueryLuggageRequest struct {
	TagNumber string `json:"tag_number"`
}

type QueryFlightRequest struct {
	FlightNumber string `json:"flight_number"`
	Date         string `json:"date,omitempty"`
}

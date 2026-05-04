package common

import "time"

type CreateTierRequest struct {
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Capacity int    `json:"capacity"`
}

type CreateEventRequest struct {
	Name     string              `json:"name"`
	Time     time.Time           `json:"time"`
	Location string              `json:"location"`
	Tiers    []CreateTierRequest `json:"tiers"`
}

type CreateEventResponse struct {
	Success bool   `json:"success"`
	EventID string `json:"event_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

type PurchaseTicketRequest struct {
	EventID  string `json:"event_id"`
	TierName string `json:"tier_name"`
	Quantity int    `json:"quantity"`
}

type PurchaseTicketResponse struct {
	Success     bool     `json:"success"`
	TicketNumbers []string `json:"ticket_numbers,omitempty"`
	Error       string   `json:"error,omitempty"`
}

type CheckInRequest struct {
	TicketNumber string `json:"ticket_number"`
}

type CheckInResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type RefundTicketRequest struct {
	TicketNumber string `json:"ticket_number"`
}

type RefundTicketResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type GetEventStatsResponse struct {
	Success bool        `json:"success"`
	Stats   *EventStats `json:"stats,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ListEventsResponse struct {
	Success bool      `json:"success"`
	Events  []Event   `json:"events,omitempty"`
	Error   string    `json:"error,omitempty"`
}

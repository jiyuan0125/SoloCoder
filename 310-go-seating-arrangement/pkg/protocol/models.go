package protocol

import "time"

type SeatStatus string

const (
	SeatStatusAvailable SeatStatus = "available"
	SeatStatusLocked    SeatStatus = "locked"
	SeatStatusSold      SeatStatus = "sold"
	SeatStatusAisle     SeatStatus = "aisle"
)

type Venue struct {
	ID       string              `json:"id"`
	Name     string              `json:"name"`
	Sections map[string]*Section `json:"sections"`
	CreatedAt time.Time          `json:"created_at"`
}

type Section struct {
	Name       string         `json:"name"`
	Rows       int            `json:"rows"`
	SeatsPerRow int           `json:"seats_per_row"`
	TotalSeats int            `json:"total_seats"`
	Aisles     map[string]bool `json:"aisles"`
}

type Session struct {
	ID          string              `json:"id"`
	VenueID     string              `json:"venue_id"`
	Name        string              `json:"name"`
	SeatStates  map[string]*SeatState `json:"seat_states"`
	CreatedAt   time.Time           `json:"created_at"`
	TotalSeats  int                 `json:"total_seats"`
}

type SeatState struct {
	SeatID       string        `json:"seat_id"`
	SectionName  string        `json:"section_name"`
	Row          string        `json:"row"`
	Number       int           `json:"number"`
	Status       SeatStatus    `json:"status"`
	LockedAt     *time.Time    `json:"locked_at"`
	OrderID      string        `json:"order_id"`
}

type Order struct {
	ID          string       `json:"id"`
	SessionID   string       `json:"session_id"`
	SectionName string       `json:"section_name"`
	SeatID      string       `json:"seat_id"`
	Status      OrderStatus  `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type OrderStatus string

const (
	OrderStatusCreated OrderStatus = "created"
	OrderStatusPaid    OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type CreateVenueRequest struct {
	Name     string           `json:"name"`
	Sections []SectionRequest `json:"sections"`
}

type SectionRequest struct {
	Name        string   `json:"name"`
	Rows        int      `json:"rows"`
	SeatsPerRow int      `json:"seats_per_row"`
	Aisles      []string `json:"aisles"`
}

type CreateSessionRequest struct {
	VenueID string `json:"venue_id"`
	Name    string `json:"name"`
}

type LockSeatRequest struct {
	SessionID   string `json:"session_id"`
	SectionName string `json:"section_name"`
	SeatID      string `json:"seat_id"`
}

type ConfirmOrderRequest struct {
	OrderID string `json:"order_id"`
}

type CancelOrderRequest struct {
	OrderID string `json:"order_id"`
}

type ReleaseSeatRequest struct {
	SessionID   string `json:"session_id"`
	SectionName string `json:"section_name"`
	SeatID      string `json:"seat_id"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type SessionStats struct {
	SessionID       string            `json:"session_id"`
	SessionName     string            `json:"session_name"`
	TotalSeats      int               `json:"total_seats"`
	SectionStats    map[string]*SectionStat `json:"section_stats"`
}

type SectionStat struct {
	SectionName    string `json:"section_name"`
	TotalSeats     int    `json:"total_seats"`
	AvailableSeats int    `json:"available_seats"`
	LockedSeats    int    `json:"locked_seats"`
	SoldSeats      int    `json:"sold_seats"`
}

type SeatDetail struct {
	SeatID      string     `json:"seat_id"`
	SectionName string     `json:"section_name"`
	Row         string     `json:"row"`
	Number      int        `json:"number"`
	Status      SeatStatus `json:"status"`
}

type OrderDetail struct {
	ID          string      `json:"id"`
	SessionID   string      `json:"session_id"`
	SectionName string      `json:"section_name"`
	SeatID      string      `json:"seat_id"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
}

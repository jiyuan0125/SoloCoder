package main

import (
	"sync"
	"time"
)

type Venue struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Sections  map[string]*Section `json:"sections"`
	TotalSeats int                `json:"total_seats"`
	CreatedAt time.Time           `json:"created_at"`
}

type Section struct {
	Name     string           `json:"name"`
	Rows     []*Row           `json:"rows"`
	SeatCount int             `json:"seat_count"`
}

type Row struct {
	RowNumber int             `json:"row_number"`
	Seats     []*SeatTemplate `json:"seats"`
}

type SeatTemplate struct {
	Index       int    `json:"index"`
	IsAisle     bool   `json:"is_aisle"`
	SeatNumber  string `json:"seat_number,omitempty"`
}

type Event struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	VenueID     string              `json:"venue_id"`
	SeatStatus  map[string]*SeatStatus `json:"seat_status"`
	CreatedAt   time.Time           `json:"created_at"`
}

type SeatStatus struct {
	SeatKey    string    `json:"seat_key"`
	Section    string    `json:"section"`
	RowNumber  int       `json:"row_number"`
	SeatIndex  int       `json:"seat_index"`
	Status     string    `json:"status"`
	OrderID    string    `json:"order_id,omitempty"`
	LockedAt   time.Time `json:"locked_at,omitempty"`
}

type Order struct {
	ID          string    `json:"id"`
	EventID     string    `json:"event_id"`
	Section     string    `json:"section"`
	RowNumber   int       `json:"row_number"`
	SeatIndex   int       `json:"seat_index"`
	SeatKey     string    `json:"seat_key"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	PaidAt      time.Time `json:"paid_at,omitempty"`
	CancelledAt time.Time `json:"cancelled_at,omitempty"`
}

type SeatLock struct {
	SeatKey    string
	OrderID    string
	LockedAt   time.Time
}

const (
	SeatStatusAvailable = "available"
	SeatStatusLocked    = "locked"
	SeatStatusSold      = "sold"

	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusCancelled = "cancelled"

	LockDuration = 10 * time.Minute
)

var (
	venues      = make(map[string]*Venue)
	events      = make(map[string]*Event)
	orders      = make(map[string]*Order)
	venueMutex  sync.RWMutex
	eventMutex  sync.RWMutex
	orderMutex  sync.RWMutex
)

func GenerateVenueKey(section string, rowNumber int, seatIndex int) string {
	return section + ":" + string(rune(rowNumber)) + ":" + string(rune(seatIndex))
}

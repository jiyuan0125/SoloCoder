package common

import (
	"time"

	"moving-platform/pkg/core"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type EstimateRequest struct {
	Items      core.ItemList `json:"items"`
	DistanceKM float64       `json:"distanceKM"`
}

type EstimateResponse struct {
	Estimate int64 `json:"estimate"`
}

type CreateBookingRequest struct {
	CustomerName string          `json:"customerName"`
	Phone        string          `json:"phone"`
	MoveDate     time.Time       `json:"moveDate"`
	TimeSlot     core.TimeSlot   `json:"timeSlot"`
	FromAddress  string          `json:"fromAddress"`
	ToAddress    string          `json:"toAddress"`
	DistanceKM   float64         `json:"distanceKM"`
	Items        core.ItemList   `json:"items"`
}

type CreateBookingResponse struct {
	BookingID string `json:"bookingId"`
	Estimate  int64  `json:"estimate"`
}

type GetBookingResponse struct {
	Booking *core.Booking `json:"booking"`
}

type ListBookingsResponse struct {
	Bookings []*core.Booking `json:"bookings"`
}

type CancelBookingResponse struct {
	CancellationFee int64 `json:"cancellationFee"`
}

type SettleBookingRequest struct {
	FinalItems    *core.ItemList `json:"finalItems,omitempty"`
	FinalDistanceKM float64      `json:"finalDistanceKM,omitempty"`
}

type SettleBookingResponse struct {
	FinalAmount int64 `json:"finalAmount"`
}

type AddReviewRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment,omitempty"`
}

type CheckAvailabilityRequest struct {
	MoveDate time.Time `json:"moveDate"`
}

type CheckAvailabilityResponse struct {
	AvailableSlots []core.TimeSlot `json:"availableSlots"`
}

func NewErrorResponse(err error) ErrorResponse {
	return ErrorResponse{Error: err.Error()}
}

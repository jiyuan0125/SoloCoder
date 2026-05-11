package core

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const MaxVehiclesPerSlot = 3

type BookingManager struct {
	mu          sync.RWMutex
	bookings    map[string]*Booking
	slotBookings map[string]int
}

func NewBookingManager() *BookingManager {
	return &BookingManager{
		bookings:    make(map[string]*Booking),
		slotBookings: make(map[string]int),
	}
}

func (bm *BookingManager) CreateBooking(booking *Booking) error {
	if booking == nil {
		return errors.New("booking cannot be nil")
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	slotKey := getSlotKey(booking.MoveDate, booking.TimeSlot)
	if bm.slotBookings[slotKey] >= MaxVehiclesPerSlot {
		return errors.New("no available vehicles for this time slot")
	}

	booking.ID = generateBookingID()
	booking.Status = StatusConfirmed
	booking.CreatedAt = time.Now()
	booking.UpdatedAt = time.Now()

	bm.bookings[booking.ID] = booking
	bm.slotBookings[slotKey]++

	return nil
}

func (bm *BookingManager) GetBooking(id string) (*Booking, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	booking, exists := bm.bookings[id]
	if !exists {
		return nil, errors.New("booking not found")
	}

	return booking, nil
}

func (bm *BookingManager) ListBookings() []*Booking {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	bookings := make([]*Booking, 0, len(bm.bookings))
	for _, booking := range bm.bookings {
		bookings = append(bookings, booking)
	}

	return bookings
}

func (bm *BookingManager) CancelBooking(id string) (int64, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	booking, exists := bm.bookings[id]
	if !exists {
		return 0, errors.New("booking not found")
	}

	if booking.Status != StatusConfirmed {
		return 0, errors.New("only confirmed bookings can be cancelled")
	}

	hoursUntilMove := time.Until(booking.MoveDate).Hours()
	cancellationFee := CalculateCancellationFee(booking.Estimate, hoursUntilMove)

	slotKey := getSlotKey(booking.MoveDate, booking.TimeSlot)
	bm.slotBookings[slotKey]--
	if bm.slotBookings[slotKey] <= 0 {
		delete(bm.slotBookings, slotKey)
	}

	booking.Status = StatusCancelled
	booking.UpdatedAt = time.Now()

	return cancellationFee, nil
}

func (bm *BookingManager) CompleteBooking(id string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	booking, exists := bm.bookings[id]
	if !exists {
		return errors.New("booking not found")
	}

	if booking.Status != StatusConfirmed {
		return errors.New("only confirmed bookings can be completed")
	}

	booking.Status = StatusCompleted
	booking.UpdatedAt = time.Now()

	return nil
}

func (bm *BookingManager) SettleBooking(id string, finalItems *ItemList, finalDistanceKM float64) (int64, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	booking, exists := bm.bookings[id]
	if !exists {
		return 0, errors.New("booking not found")
	}

	if booking.Status != StatusCompleted {
		return 0, errors.New("only completed bookings can be settled")
	}

	var finalAmount int64
	if finalItems != nil {
		booking.Items = *finalItems
	}
	if finalDistanceKM > 0 {
		booking.DistanceKM = finalDistanceKM
	}

	finalAmount = EstimatePrice(booking.Items, booking.DistanceKM)
	booking.FinalAmount = finalAmount
	booking.Status = StatusSettled
	booking.UpdatedAt = time.Now()

	return finalAmount, nil
}

func (bm *BookingManager) AddReview(id string, review *Review) error {
	if review == nil {
		return errors.New("review cannot be nil")
	}

	if review.Rating < 1 || review.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	booking, exists := bm.bookings[id]
	if !exists {
		return errors.New("booking not found")
	}

	if booking.Status != StatusCompleted && booking.Status != StatusSettled {
		return errors.New("can only add review to completed or settled bookings")
	}

	booking.Review = review
	booking.UpdatedAt = time.Now()

	return nil
}

func (bm *BookingManager) CheckAvailability(moveDate time.Time, slot TimeSlot) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	slotKey := getSlotKey(moveDate, slot)
	return bm.slotBookings[slotKey] < MaxVehiclesPerSlot
}

func (bm *BookingManager) GetAvailableSlots(moveDate time.Time) []TimeSlot {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	slots := []TimeSlot{SlotMorning, SlotAfternoon, SlotEvening}
	available := make([]TimeSlot, 0)

	for _, slot := range slots {
		slotKey := getSlotKey(moveDate, slot)
		if bm.slotBookings[slotKey] < MaxVehiclesPerSlot {
			available = append(available, slot)
		}
	}

	return available
}

func getSlotKey(date time.Time, slot TimeSlot) string {
	dateStr := date.Format("2006-01-02")
	return fmt.Sprintf("%s-%s", dateStr, slot)
}

func generateBookingID() string {
	return fmt.Sprintf("BK%010d", time.Now().UnixNano())
}

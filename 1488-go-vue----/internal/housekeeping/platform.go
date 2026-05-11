package housekeeping

import (
	"fmt"
	"time"
)

type Platform struct {
	AuntManager    *AuntManager
	BookingManager *BookingManager
}

func NewPlatform() *Platform {
	return &Platform{
		AuntManager:    NewAuntManager(),
		BookingManager: NewBookingManager(),
	}
}

func (p *Platform) RegisterAunt(aunt *Aunt) error {
	return p.AuntManager.RegisterAunt(aunt)
}

func (p *Platform) RegisterCustomer(customer *Customer) error {
	return p.BookingManager.RegisterCustomer(customer)
}

func (p *Platform) CreateBooking(req BookingRequest) (*Booking, error) {
	booking, err := p.BookingManager.CreateBooking(req)
	if err != nil {
		return nil, err
	}

	p.BookingManager.TryMatchBooking(p.AuntManager, booking.ID)
	p.BookingManager.MatchPendingBookings(p.AuntManager)

	return booking, nil
}

func (p *Platform) AddReview(auntID, bookingID, customerID string, rating float64, comment string) (*Review, error) {
	if rating < 1 || rating > 5 {
		return nil, fmt.Errorf("评分必须在1-5分之间")
	}

	aunt, err := p.AuntManager.GetAunt(auntID)
	if err != nil {
		return nil, err
	}

	booking, err := p.BookingManager.GetBooking(bookingID)
	if err != nil {
		return nil, err
	}

	if booking.Status != BookingStatusCompleted {
		return nil, fmt.Errorf("只能对已完成的订单进行评价")
	}

	review := Review{
		ID:         fmt.Sprintf("RV%d", time.Now().UnixNano()),
		AuntID:     auntID,
		BookingID:  bookingID,
		CustomerID: customerID,
		Rating:     rating,
		Comment:    comment,
		Time:       time.Now(),
	}

	AddReview(aunt, review)
	return &review, nil
}

func (p *Platform) GetMonthlyStats(auntID, month string) (*MonthlyStats, error) {
	_, err := p.AuntManager.GetAunt(auntID)
	if err != nil {
		return nil, err
	}

	bookings := p.BookingManager.GetBookingsByAunt(auntID)

	var totalBookings int
	var totalHours float64
	var totalRating float64
	var ratingCount int
	var uniqueCustomers = make(map[string]bool)
	var repeatCustomers = make(map[string]bool)

	for _, booking := range bookings {
		if booking.Status != BookingStatusCompleted {
			continue
		}

		bookingMonth := booking.ServiceDate.Format("2006-01")
		if bookingMonth != month {
			continue
		}

		totalBookings++
		totalHours += booking.ActualDuration

		uniqueCustomers[booking.CustomerID] = true
	}

	aunt, _ := p.AuntManager.GetAunt(auntID)
	for _, review := range aunt.Reviews {
		reviewMonth := review.Time.Format("2006-01")
		if reviewMonth == month {
			totalRating += review.Rating
			ratingCount++
		}
	}

	for customerID := range uniqueCustomers {
		customerBookings := p.BookingManager.GetBookingsByCustomer(customerID)
		if len(customerBookings) > 1 {
			repeatCustomers[customerID] = true
		}
	}

	stats := &MonthlyStats{
		AuntID:        auntID,
		Month:         month,
		TotalBookings: totalBookings,
		TotalHours:    totalHours,
	}

	if ratingCount > 0 {
		stats.AverageRating = totalRating / float64(ratingCount)
	}

	if len(uniqueCustomers) > 0 {
		stats.RepeatRate = float64(len(repeatCustomers)) / float64(len(uniqueCustomers))
	}

	return stats, nil
}

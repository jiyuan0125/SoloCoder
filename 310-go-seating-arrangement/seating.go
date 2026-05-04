package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrEventNotFound         = errors.New("event not found")
	ErrSeatAlreadyLocked     = errors.New("座位已被选中")
	ErrSeatAlreadySold       = errors.New("座位已售出")
	ErrSeatNotLocked         = errors.New("seat is not locked")
	ErrInvalidEventName      = errors.New("event name cannot be empty or longer than 50 characters")
	lockCleanupRunning       bool
	lockCleanupMutex         sync.Mutex
)

type CreateEventRequest struct {
	Name    string `json:"name"`
	VenueID string `json:"venue_id"`
}

type SectionSeatStats struct {
	SectionName string `json:"section_name"`
	TotalSeats  int    `json:"total_seats"`
	SoldSeats   int    `json:"sold_seats"`
	LockedSeats int    `json:"locked_seats"`
	Available   int    `json:"available"`
}

func CreateEvent(req CreateEventRequest) (*Event, error) {
	if req.Name == "" || len(req.Name) > 50 {
		return nil, ErrInvalidEventName
	}

	venue, err := GetVenue(req.VenueID)
	if err != nil {
		return nil, err
	}

	eventMutex.Lock()
	defer eventMutex.Unlock()

	event := &Event{
		ID:         generateID(),
		Name:       req.Name,
		VenueID:    req.VenueID,
		SeatStatus: make(map[string]*SeatStatus),
		CreatedAt:  time.Now(),
	}

	for sectionName, section := range venue.Sections {
		for _, row := range section.Rows {
			for _, seat := range row.Seats {
				if !seat.IsAisle {
					seatKey := generateSeatKey(sectionName, row.RowNumber, seat.Index)
					event.SeatStatus[seatKey] = &SeatStatus{
						SeatKey:   seatKey,
						Section:   sectionName,
						RowNumber: row.RowNumber,
						SeatIndex: seat.Index,
						Status:    SeatStatusAvailable,
					}
				}
			}
		}
	}

	events[event.ID] = event

	startLockCleanupIfNeeded()

	return event, nil
}

func generateSeatKey(section string, rowNumber int, seatIndex int) string {
	return fmt.Sprintf("%s:%d:%d", section, rowNumber, seatIndex)
}

func GetEvent(eventID string) (*Event, error) {
	eventMutex.RLock()
	defer eventMutex.RUnlock()

	event, exists := events[eventID]
	if !exists {
		return nil, ErrEventNotFound
	}
	return event, nil
}

func GetAllEvents() []*Event {
	eventMutex.RLock()
	defer eventMutex.RUnlock()

	result := make([]*Event, 0, len(events))
	for _, e := range events {
		result = append(result, e)
	}
	return result
}

func LockSeat(eventID string, section string, rowNumber int, seatIndex int, orderID string) (*SeatStatus, error) {
	eventMutex.Lock()
	defer eventMutex.Unlock()

	event, exists := events[eventID]
	if !exists {
		return nil, ErrEventNotFound
	}

	seatKey := generateSeatKey(section, rowNumber, seatIndex)
	seatStatus, exists := event.SeatStatus[seatKey]
	if !exists {
		return nil, ErrSeatNotFound
	}

	switch seatStatus.Status {
	case SeatStatusLocked:
		if !isLockExpired(seatStatus.LockedAt) {
			return nil, ErrSeatAlreadyLocked
		}
	case SeatStatusSold:
		return nil, ErrSeatAlreadySold
	}

	seatStatus.Status = SeatStatusLocked
	seatStatus.OrderID = orderID
	seatStatus.LockedAt = time.Now()

	return seatStatus, nil
}

func ReleaseSeat(eventID string, seatKey string) error {
	eventMutex.Lock()
	defer eventMutex.Unlock()

	event, exists := events[eventID]
	if !exists {
		return ErrEventNotFound
	}

	seatStatus, exists := event.SeatStatus[seatKey]
	if !exists {
		return ErrSeatNotFound
	}

	if seatStatus.Status != SeatStatusLocked {
		return ErrSeatNotLocked
	}

	seatStatus.Status = SeatStatusAvailable
	seatStatus.OrderID = ""
	seatStatus.LockedAt = time.Time{}

	return nil
}

func ConfirmSeatSale(eventID string, seatKey string) error {
	eventMutex.Lock()
	defer eventMutex.Unlock()

	event, exists := events[eventID]
	if !exists {
		return ErrEventNotFound
	}

	seatStatus, exists := event.SeatStatus[seatKey]
	if !exists {
		return ErrSeatNotFound
	}

	if seatStatus.Status != SeatStatusLocked {
		return ErrSeatNotLocked
	}

	seatStatus.Status = SeatStatusSold

	return nil
}

func isLockExpired(lockedAt time.Time) bool {
	return time.Since(lockedAt) > LockDuration
}

func GetEventSeatStats(eventID string) ([]SectionSeatStats, error) {
	eventMutex.RLock()
	defer eventMutex.RUnlock()

	event, exists := events[eventID]
	if !exists {
		return nil, ErrEventNotFound
	}

	venue, err := GetVenue(event.VenueID)
	if err != nil {
		return nil, err
	}

	stats := make([]SectionSeatStats, 0, len(venue.Sections))
	for sectionName, section := range venue.Sections {
		sectionStats := SectionSeatStats{
			SectionName: sectionName,
			TotalSeats:  section.SeatCount,
		}

		for _, seatStatus := range event.SeatStatus {
			if seatStatus.Section != sectionName {
				continue
			}

			switch seatStatus.Status {
			case SeatStatusSold:
				sectionStats.SoldSeats++
			case SeatStatusLocked:
				if !isLockExpired(seatStatus.LockedAt) {
					sectionStats.LockedSeats++
				}
			}
		}

		sectionStats.Available = sectionStats.TotalSeats - sectionStats.SoldSeats - sectionStats.LockedSeats
		stats = append(stats, sectionStats)
	}

	return stats, nil
}

func GetSeatedDetails(eventID string) ([]*SeatStatus, error) {
	eventMutex.RLock()
	defer eventMutex.RUnlock()

	event, exists := events[eventID]
	if !exists {
		return nil, ErrEventNotFound
	}

	details := make([]*SeatStatus, 0, len(event.SeatStatus))
	for _, seatStatus := range event.SeatStatus {
		details = append(details, seatStatus)
	}

	return details, nil
}

func GetUserViewableSeats(eventID string, sectionName string) ([]*UserSeat, error) {
	event, err := GetEvent(eventID)
	if err != nil {
		return nil, err
	}

	venue, err := GetVenue(event.VenueID)
	if err != nil {
		return nil, err
	}

	section, exists := venue.Sections[sectionName]
	if !exists {
		return nil, ErrSectionNotFound
	}

	eventMutex.RLock()
	defer eventMutex.RUnlock()

	var userSeats []*UserSeat
	for _, row := range section.Rows {
		for _, seatTemplate := range row.Seats {
			if seatTemplate.IsAisle {
				continue
			}

			seatKey := generateSeatKey(sectionName, row.RowNumber, seatTemplate.Index)
			seatStatus := event.SeatStatus[seatKey]

			status := seatStatus.Status
			if status == SeatStatusLocked && isLockExpired(seatStatus.LockedAt) {
				status = SeatStatusAvailable
			}

			userSeats = append(userSeats, &UserSeat{
				RowNumber:  row.RowNumber,
				SeatNumber: seatTemplate.SeatNumber,
				SeatIndex:  seatTemplate.Index,
				Status:     status,
			})
		}
	}

	return userSeats, nil
}

type UserSeat struct {
	RowNumber  int    `json:"row_number"`
	SeatNumber string `json:"seat_number"`
	SeatIndex  int    `json:"seat_index"`
	Status     string `json:"status"`
}

func AdminReleaseSeat(eventID string, seatKey string) error {
	return ReleaseSeat(eventID, seatKey)
}

func startLockCleanupIfNeeded() {
	lockCleanupMutex.Lock()
	defer lockCleanupMutex.Unlock()

	if lockCleanupRunning {
		return
	}

	lockCleanupRunning = true
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			cleanupExpiredLocks()
		}
	}()
}

func cleanupExpiredLocks() {
	eventMutex.Lock()
	defer eventMutex.Unlock()

	for _, event := range events {
		for _, seatStatus := range event.SeatStatus {
			if seatStatus.Status == SeatStatusLocked && isLockExpired(seatStatus.LockedAt) {
				seatStatus.Status = SeatStatusAvailable
				seatStatus.OrderID = ""
				seatStatus.LockedAt = time.Time{}
			}
		}
	}
}

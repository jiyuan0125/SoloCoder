package core

import (
	"errors"
	"sort"
	"sync"
)

type Manager struct {
	mu       sync.RWMutex
	bookings map[string]map[string]*Booking
	config   Config
}

type Config struct {
	CheckSameBookerConflict bool
}

type ConflictResult struct {
	Conflicts []*Booking
}

func NewManager(config Config) *Manager {
	return &Manager{
		bookings: make(map[string]map[string]*Booking),
		config:   config,
	}
}

func (m *Manager) AddBooking(booking *Booking) (*ConflictResult, error) {
	if booking == nil {
		return nil, errors.New("booking cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	resourceBookings, exists := m.bookings[booking.Resource]
	if !exists {
		resourceBookings = make(map[string]*Booking)
		m.bookings[booking.Resource] = resourceBookings
	}

	conflicts := m.findConflictsLocked(booking)
	if len(conflicts) > 0 {
		return &ConflictResult{Conflicts: conflicts}, errors.New("booking conflicts with existing bookings")
	}

	resourceBookings[booking.ID] = booking
	return &ConflictResult{}, nil
}

func (m *Manager) AddBookings(bookings []*Booking) (*ConflictResult, error) {
	if len(bookings) == 0 {
		return &ConflictResult{}, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range bookings {
		for j := i + 1; j < len(bookings); j++ {
			if bookings[i].Overlaps(bookings[j]) {
				return &ConflictResult{Conflicts: []*Booking{bookings[i], bookings[j]}},
					errors.New("batch contains conflicting bookings")
			}
			if m.config.CheckSameBookerConflict &&
				bookings[i].Booker == bookings[j].Booker &&
				bookings[i].Resource == bookings[j].Resource {
				return &ConflictResult{Conflicts: []*Booking{bookings[i], bookings[j]}},
					errors.New("batch contains conflicting bookings for same booker")
			}
		}
	}

	var allConflicts []*Booking
	for _, booking := range bookings {
		conflicts := m.findConflictsLocked(booking)
		allConflicts = append(allConflicts, conflicts...)
	}

	if len(allConflicts) > 0 {
		return &ConflictResult{Conflicts: allConflicts}, errors.New("some bookings conflict with existing bookings")
	}

	for _, booking := range bookings {
		resourceBookings, exists := m.bookings[booking.Resource]
		if !exists {
			resourceBookings = make(map[string]*Booking)
			m.bookings[booking.Resource] = resourceBookings
		}
		resourceBookings[booking.ID] = booking
	}

	return &ConflictResult{}, nil
}

func (m *Manager) findConflictsLocked(booking *Booking) []*Booking {
	var conflicts []*Booking
	resourceBookings, exists := m.bookings[booking.Resource]
	if !exists {
		return conflicts
	}

	for _, existing := range resourceBookings {
		if existing.ID == booking.ID {
			continue
		}

		if existing.Overlaps(booking) {
			if m.config.CheckSameBookerConflict {
				if existing.Booker == booking.Booker {
					conflicts = append(conflicts, existing)
				}
			} else {
				conflicts = append(conflicts, existing)
			}
		}
	}

	return conflicts
}

func (m *Manager) CheckConflict(resource string, interval *Interval) (*ConflictResult, error) {
	if interval == nil || interval.IsEmpty() {
		return &ConflictResult{}, nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var conflicts []*Booking
	resourceBookings, exists := m.bookings[resource]
	if !exists {
		return &ConflictResult{Conflicts: conflicts}, nil
	}

	for _, booking := range resourceBookings {
		if booking.OverlapsInterval(interval) {
			conflicts = append(conflicts, booking)
		}
	}

	return &ConflictResult{Conflicts: conflicts}, nil
}

func (m *Manager) DeleteBooking(bookingID, operator string, isAdmin bool) error {
	if bookingID == "" {
		return errors.New("booking ID cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for resource, resourceBookings := range m.bookings {
		booking, exists := resourceBookings[bookingID]
		if exists {
			if !isAdmin && booking.Booker != operator {
				return errors.New("permission denied: only the booker or admin can delete this booking")
			}

			delete(resourceBookings, bookingID)
			if len(resourceBookings) == 0 {
				delete(m.bookings, resource)
			}
			return nil
		}
	}

	return errors.New("booking not found")
}

func (m *Manager) ListBookings(resource string, interval *Interval) []*Booking {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Booking
	resourceBookings, exists := m.bookings[resource]
	if !exists {
		return result
	}

	for _, booking := range resourceBookings {
		if interval == nil {
			result = append(result, booking)
		} else if !interval.IsEmpty() && booking.IsWithinInterval(interval) {
			result = append(result, booking)
		}
	}

	sort.Sort(ByStartTime(result))
	return result
}

func (m *Manager) GetBooking(bookingID string) *Booking {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, resourceBookings := range m.bookings {
		if booking, exists := resourceBookings[bookingID]; exists {
			return booking
		}
	}
	return nil
}

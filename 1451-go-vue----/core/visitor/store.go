package visitor

import (
	"sync"

	"smart-park/common"
)

const (
	StatusPending    = "pending"
	StatusApproved   = "approved"
	StatusRejected   = "rejected"
	StatusCheckedIn  = "checked_in"
	StatusExpired    = "expired"
)

type Store struct {
	mu           sync.RWMutex
	reservations map[string]*common.VisitorReservation
}

func NewStore() *Store {
	return &Store{
		reservations: make(map[string]*common.VisitorReservation),
	}
}

func (s *Store) AddReservation(r *common.VisitorReservation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reservations[r.ID] = r
}

func (s *Store) GetReservation(id string) *common.VisitorReservation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.reservations[id]
}

func (s *Store) UpdateReservation(r *common.VisitorReservation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reservations[r.ID] = r
}

func (s *Store) GetReservationsByEmployee(employeeID string) []*common.VisitorReservation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*common.VisitorReservation
	for _, r := range s.reservations {
		if r.EmployeeID == employeeID {
			result = append(result, r)
		}
	}
	return result
}

func (s *Store) GetReservationByPhone(phone string) *common.VisitorReservation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.reservations {
		if r.VisitorPhone == phone {
			return r
		}
	}
	return nil
}

func (s *Store) GetAllReservations() []*common.VisitorReservation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*common.VisitorReservation
	for _, r := range s.reservations {
		result = append(result, r)
	}
	return result
}

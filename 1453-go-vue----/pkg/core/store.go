package core

import (
	"property-management/pkg/core/models"
	"sync"
)

type Store struct {
	mu            sync.RWMutex
	users         map[string]*models.User
	repairs       map[string]*models.Repair
	bills         map[string]*models.Bill
	payments      map[string]*models.Payment
	announcements map[string]*models.Announcement
}

func NewStore() *Store {
	return &Store{
		users:         make(map[string]*models.User),
		repairs:       make(map[string]*models.Repair),
		bills:         make(map[string]*models.Bill),
		payments:      make(map[string]*models.Payment),
		announcements: make(map[string]*models.Announcement),
	}
}

func (s *Store) Users() map[string]*models.User {
	return s.users
}

func (s *Store) SetUser(id string, user *models.User) {
	s.users[id] = user
}

func (s *Store) Repairs() map[string]*models.Repair {
	return s.repairs
}

func (s *Store) SetRepair(id string, repair *models.Repair) {
	s.repairs[id] = repair
}

func (s *Store) Bills() map[string]*models.Bill {
	return s.bills
}

func (s *Store) SetBill(id string, bill *models.Bill) {
	s.bills[id] = bill
}

func (s *Store) Payments() map[string]*models.Payment {
	return s.payments
}

func (s *Store) SetPayment(id string, payment *models.Payment) {
	s.payments[id] = payment
}

func (s *Store) Announcements() map[string]*models.Announcement {
	return s.announcements
}

func (s *Store) SetAnnouncement(id string, announcement *models.Announcement) {
	s.announcements[id] = announcement
}

func (s *Store) Lock() {
	s.mu.Lock()
}

func (s *Store) Unlock() {
	s.mu.Unlock()
}

func (s *Store) RLock() {
	s.mu.RLock()
}

func (s *Store) RUnlock() {
	s.mu.RUnlock()
}

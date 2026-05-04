package server

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"event-ticketing/pkg/common"
)

type Store struct {
	events  map[string]*common.Event
	tickets map[string]*common.Ticket
	mu      sync.RWMutex
	dataDir string
}

func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	store := &Store{
		events:  make(map[string]*common.Event),
		tickets: make(map[string]*common.Ticket),
		dataDir: dataDir,
	}

	if err := store.Load(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) CreateEvent(event *common.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events[event.ID] = event
	return s.Save()
}

func (s *Store) GetEvent(eventID string) (*common.Event, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, ok := s.events[eventID]
	return event, ok
}

func (s *Store) ListEvents() []common.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]common.Event, 0, len(s.events))
	for _, event := range s.events {
		events = append(events, *event)
	}
	return events
}

func (s *Store) PurchaseTickets(eventID string, tierName string, quantity int) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.events[eventID]
	if !ok {
		return nil, common.ErrEventNotFound
	}

	if event.Time.Before(time.Now()) {
		return nil, common.ErrEventEnded
	}

	var targetTier *common.Tier
	for i := range event.Tiers {
		if event.Tiers[i].Name == tierName {
			targetTier = &event.Tiers[i]
			break
		}
	}

	if targetTier == nil {
		return nil, common.ErrInvalidTier
	}

	if targetTier.Available < quantity {
		return nil, common.ErrInsufficientStock
	}

	targetTier.Available -= quantity

	ticketNumbers := make([]string, quantity)
	for i := 0; i < quantity; i++ {
		ticketNumber := common.GenerateTicketNumber()
		ticketNumbers[i] = ticketNumber
		s.tickets[ticketNumber] = &common.Ticket{
			TicketNumber: ticketNumber,
			EventID:      eventID,
			TierName:     tierName,
			Price:        targetTier.Price,
			PurchasedAt:  time.Now(),
			CheckedIn:    false,
			Refunded:     false,
		}
	}

	if err := s.Save(); err != nil {
		return nil, err
	}

	return ticketNumbers, nil
}

func (s *Store) CheckIn(ticketNumber string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ticket, ok := s.tickets[ticketNumber]
	if !ok {
		return common.ErrTicketNotFound
	}

	if ticket.Refunded {
		return common.ErrTicketRefunded
	}

	if ticket.CheckedIn {
		return common.ErrTicketAlreadyUsed
	}

	ticket.CheckedIn = true
	ticket.CheckedInAt = time.Now()

	return s.Save()
}

func (s *Store) RefundTicket(ticketNumber string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ticket, ok := s.tickets[ticketNumber]
	if !ok {
		return common.ErrTicketNotFound
	}

	if ticket.Refunded {
		return common.ErrTicketRefunded
	}

	if ticket.CheckedIn {
		return common.ErrTicketAlreadyUsed
	}

	event, ok := s.events[ticket.EventID]
	if !ok {
		return common.ErrEventNotFound
	}

	for i := range event.Tiers {
		if event.Tiers[i].Name == ticket.TierName {
			event.Tiers[i].Available++
			break
		}
	}

	ticket.Refunded = true
	ticket.RefundedAt = time.Now()

	return s.Save()
}

func (s *Store) GetTicket(ticketNumber string) (*common.Ticket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ticket, ok := s.tickets[ticketNumber]
	return ticket, ok
}

func (s *Store) GetEventStats(eventID string) (*common.EventStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, ok := s.events[eventID]
	if !ok {
		return nil, common.ErrEventNotFound
	}

	stats := &common.EventStats{
		EventID:   event.ID,
		EventName: event.Name,
		TierStats: make([]common.TierStats, len(event.Tiers)),
	}

	for i, tier := range event.Tiers {
		sold := tier.Capacity - tier.Available
		checkedIn := 0

		for _, ticket := range s.tickets {
			if ticket.EventID == eventID && ticket.TierName == tier.Name {
				if ticket.CheckedIn && !ticket.Refunded {
					checkedIn++
				}
			}
		}

		stats.TierStats[i] = common.TierStats{
			Name:      tier.Name,
			Price:     tier.Price,
			Capacity:  tier.Capacity,
			Available: tier.Available,
			Sold:      sold,
			CheckedIn: checkedIn,
		}
	}

	return stats, nil
}

type dataSnapshot struct {
	Events  map[string]*common.Event  `json:"events"`
	Tickets map[string]*common.Ticket `json:"tickets"`
}

func (s *Store) Save() error {
	eventsPath := fmt.Sprintf("%s/events.json", s.dataDir)
	ticketsPath := fmt.Sprintf("%s/tickets.json", s.dataDir)

	eventsData, err := json.MarshalIndent(s.events, "", "  ")
	if err != nil {
		return err
	}

	ticketsData, err := json.MarshalIndent(s.tickets, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(eventsPath, eventsData, 0644); err != nil {
		return err
	}

	if err := os.WriteFile(ticketsPath, ticketsData, 0644); err != nil {
		return err
	}

	return nil
}

func (s *Store) Load() error {
	eventsPath := fmt.Sprintf("%s/events.json", s.dataDir)
	ticketsPath := fmt.Sprintf("%s/tickets.json", s.dataDir)

	if _, err := os.Stat(eventsPath); err == nil {
		data, err := os.ReadFile(eventsPath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &s.events); err != nil {
			return err
		}
	}

	if _, err := os.Stat(ticketsPath); err == nil {
		data, err := os.ReadFile(ticketsPath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &s.tickets); err != nil {
			return err
		}
	}

	return nil
}

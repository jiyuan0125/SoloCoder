package server

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"seating-arrangement/pkg/protocol"
)

type StorageData struct {
	Venues   map[string]*VenueData   `json:"venues"`
	Sessions map[string]*SessionData `json:"sessions"`
	Orders   map[string]*OrderData   `json:"orders"`
}

type VenueData struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Sections  map[string]*SectionData `json:"sections"`
	CreatedAt int64                  `json:"created_at"`
}

type SectionData struct {
	Name        string            `json:"name"`
	Rows        int               `json:"rows"`
	SeatsPerRow int               `json:"seats_per_row"`
	TotalSeats  int               `json:"total_seats"`
	Aisles      map[string]bool   `json:"aisles"`
}

type SessionData struct {
	ID         string                  `json:"id"`
	VenueID    string                  `json:"venue_id"`
	Name       string                  `json:"name"`
	SeatStates map[string]*SeatStateData `json:"seat_states"`
	CreatedAt  int64                   `json:"created_at"`
	TotalSeats int                     `json:"total_seats"`
}

type SeatStateData struct {
	SeatID      string `json:"seat_id"`
	SectionName string `json:"section_name"`
	Row         string `json:"row"`
	Number      int    `json:"number"`
	Status      string `json:"status"`
	LockedAt    int64  `json:"locked_at"`
	OrderID     string `json:"order_id"`
}

type OrderData struct {
	ID          string `json:"id"`
	SessionID   string `json:"session_id"`
	SectionName string `json:"section_name"`
	SeatID      string `json:"seat_id"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type Storage struct {
	filePath string
	vm       *VenueManager
	sm       *SessionManager
	om       *OrderManager
	mu       sync.Mutex
}

func NewStorage(filePath string, vm *VenueManager, sm *SessionManager, om *OrderManager) *Storage {
	return &Storage{
		filePath: filePath,
		vm:       vm,
		sm:       sm,
		om:       om,
	}
}

func (s *Storage) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := &StorageData{
		Venues:   make(map[string]*VenueData),
		Sessions: make(map[string]*SessionData),
		Orders:   make(map[string]*OrderData),
	}

	for id, venue := range s.vm.GetVenues() {
		venueData := &VenueData{
			ID:        venue.ID,
			Name:      venue.Name,
			Sections:  make(map[string]*SectionData),
			CreatedAt: venue.CreatedAt.Unix(),
		}
		for sectionName, section := range venue.Sections {
			venueData.Sections[sectionName] = &SectionData{
				Name:        section.Name,
				Rows:        section.Rows,
				SeatsPerRow: section.SeatsPerRow,
				TotalSeats:  section.TotalSeats,
				Aisles:      section.Aisles,
			}
		}
		data.Venues[id] = venueData
	}

	for id, session := range s.sm.GetSessions() {
		sessionData := &SessionData{
			ID:         session.ID,
			VenueID:    session.VenueID,
			Name:       session.Name,
			SeatStates: make(map[string]*SeatStateData),
			CreatedAt:  session.CreatedAt.Unix(),
			TotalSeats: session.TotalSeats,
		}
		for seatID, seatState := range session.SeatStates {
			seatData := &SeatStateData{
				SeatID:      seatState.SeatID,
				SectionName: seatState.SectionName,
				Row:         seatState.Row,
				Number:      seatState.Number,
				Status:      string(seatState.Status),
				OrderID:     seatState.OrderID,
			}
			if seatState.LockedAt != nil {
				seatData.LockedAt = seatState.LockedAt.Unix()
			}
			sessionData.SeatStates[seatID] = seatData
		}
		data.Sessions[id] = sessionData
	}

	for id, order := range s.om.GetOrders() {
		data.Orders[id] = &OrderData{
			ID:          order.ID,
			SessionID:   order.SessionID,
			SectionName: order.SectionName,
			SeatID:      order.SeatID,
			Status:      string(order.Status),
			CreatedAt:   order.CreatedAt.Unix(),
			UpdatedAt:   order.UpdatedAt.Unix(),
		}
	}

	file, err := os.Create(s.filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	return nil
}

func (s *Storage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(s.filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var data StorageData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("反序列化数据失败: %w", err)
	}

	venues := make(map[string]*protocol.Venue)
	for id, vd := range data.Venues {
		venue := &protocol.Venue{
			ID:        vd.ID,
			Name:      vd.Name,
			Sections:  make(map[string]*protocol.Section),
			CreatedAt: time.Unix(vd.CreatedAt, 0),
		}
		for sectionName, sd := range vd.Sections {
			venue.Sections[sectionName] = &protocol.Section{
				Name:        sd.Name,
				Rows:        sd.Rows,
				SeatsPerRow: sd.SeatsPerRow,
				TotalSeats:  sd.TotalSeats,
				Aisles:      sd.Aisles,
			}
		}
		venues[id] = venue
	}
	s.vm.SetVenues(venues)

	sessions := make(map[string]*protocol.Session)
	for id, sd := range data.Sessions {
		session := &protocol.Session{
			ID:         sd.ID,
			VenueID:    sd.VenueID,
			Name:       sd.Name,
			SeatStates: make(map[string]*protocol.SeatState),
			CreatedAt:  time.Unix(sd.CreatedAt, 0),
			TotalSeats: sd.TotalSeats,
		}
		for seatID, ssd := range sd.SeatStates {
			seatState := &protocol.SeatState{
				SeatID:      ssd.SeatID,
				SectionName: ssd.SectionName,
				Row:         ssd.Row,
				Number:      ssd.Number,
				Status:      protocol.SeatStatus(ssd.Status),
				OrderID:     ssd.OrderID,
			}
			if ssd.LockedAt > 0 {
				lockedAt := time.Unix(ssd.LockedAt, 0)
				seatState.LockedAt = &lockedAt
			}
			session.SeatStates[seatID] = seatState
		}
		sessions[id] = session
	}
	s.sm.SetSessions(sessions)

	orders := make(map[string]*protocol.Order)
	for id, od := range data.Orders {
		orders[id] = &protocol.Order{
			ID:          od.ID,
			SessionID:   od.SessionID,
			SectionName: od.SectionName,
			SeatID:      od.SeatID,
			Status:      protocol.OrderStatus(od.Status),
			CreatedAt:   time.Unix(od.CreatedAt, 0),
			UpdatedAt:   time.Unix(od.UpdatedAt, 0),
		}
	}
	s.om.SetOrders(orders)

	return nil
}

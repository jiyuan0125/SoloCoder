package server

import (
	"errors"
	"sync"
	"time"

	"seating-arrangement/pkg/protocol"

	"github.com/google/uuid"
)

type VenueManager struct {
	venues map[string]*protocol.Venue
	mu     sync.RWMutex
}

func NewVenueManager() *VenueManager {
	return &VenueManager{
		venues: make(map[string]*protocol.Venue),
	}
}

func (vm *VenueManager) CreateVenue(req *protocol.CreateVenueRequest) (*protocol.Venue, error) {
	if req.Name == "" {
		return nil, errors.New("场馆名称不能为空")
	}

	if len(req.Sections) == 0 {
		return nil, errors.New("场馆至少需要包含一个区域")
	}

	sectionNames := make(map[string]bool)
	for _, s := range req.Sections {
		if s.Name == "" {
			return nil, errors.New("区域名称不能为空")
		}
		if sectionNames[s.Name] {
			return nil, errors.New("区域名称不能重复: " + s.Name)
		}
		sectionNames[s.Name] = true

		if s.Rows <= 0 {
			return nil, errors.New("区域排数必须大于0: " + s.Name)
		}
		if s.SeatsPerRow <= 0 {
			return nil, errors.New("每排座位数必须大于0: " + s.Name)
		}
	}

	venue := &protocol.Venue{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Sections:  make(map[string]*protocol.Section),
		CreatedAt: time.Now(),
	}

	for _, sReq := range req.Sections {
		aisles := make(map[string]bool)
		for _, aisle := range sReq.Aisles {
			aisles[aisle] = true
		}

		section := &protocol.Section{
			Name:       sReq.Name,
			Rows:       sReq.Rows,
			SeatsPerRow: sReq.SeatsPerRow,
			TotalSeats: sReq.Rows * sReq.SeatsPerRow,
			Aisles:     aisles,
		}
		venue.Sections[sReq.Name] = section
	}

	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.venues[venue.ID] = venue

	return venue, nil
}

func (vm *VenueManager) GetVenue(id string) (*protocol.Venue, bool) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	venue, exists := vm.venues[id]
	return venue, exists
}

func (vm *VenueManager) ListVenues() []*protocol.Venue {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	venues := make([]*protocol.Venue, 0, len(vm.venues))
	for _, v := range vm.venues {
		venues = append(venues, v)
	}
	return venues
}

func (vm *VenueManager) SetVenues(venues map[string]*protocol.Venue) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.venues = venues
}

func (vm *VenueManager) GetVenues() map[string]*protocol.Venue {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.venues
}

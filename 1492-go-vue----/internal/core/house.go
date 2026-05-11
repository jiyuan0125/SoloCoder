package core

import (
	"errors"
	"time"

	"renovation-management/internal/api"
)

func (s *Store) CreateHouse(req *api.CreateHouseRequest) (*api.HouseInfo, error) {
	if req.Area <= 0 {
		return nil, errors.New("area must be positive")
	}
	if req.Layout == "" {
		return nil, errors.New("layout is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	house := &api.HouseInfo{
		ID:          GenerateID("HOUSE"),
		Area:        req.Area,
		Layout:      req.Layout,
		Floor:       req.Floor,
		Orientation: req.Orientation,
		CreatedAt:   time.Now(),
	}

	s.houses[house.ID] = house
	return house, nil
}

func (s *Store) GetHouse(id string) (*api.HouseInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	house, ok := s.houses[id]
	if !ok {
		return nil, errors.New("house not found")
	}
	return house, nil
}

func (s *Store) ListHouses() []*api.HouseInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	houses := make([]*api.HouseInfo, 0, len(s.houses))
	for _, h := range s.houses {
		houses = append(houses, h)
	}
	return houses
}

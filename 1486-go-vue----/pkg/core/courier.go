package core

import (
	"errors"
	"locker/pkg/api"
)

func (s *System) CreateCourier(req api.CreateCourierRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Couriers[req.ID]; exists {
		return errors.New("courier already exists")
	}

	s.Couriers[req.ID] = &Courier{
		ID:       req.ID,
		Name:     req.Name,
		Balance:  req.Balance,
		LastUsed: make(map[string][]string),
	}
	return nil
}

func (s *System) GetCourier(req api.GetCourierRequest) (*api.CourierInfo, error) {
	s.mu.RLock()
	courier, exists := s.Couriers[req.ID]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("courier not found")
	}

	courier.mu.Lock()
	defer courier.mu.Unlock()

	return &api.CourierInfo{
		ID:      courier.ID,
		Name:    courier.Name,
		Balance: courier.Balance,
	}, nil
}

func (s *System) RechargeCourier(req api.RechargeCourierRequest) (*api.CourierInfo, error) {
	if req.Amount <= 0 {
		return nil, errors.New("recharge amount must be positive")
	}

	s.mu.RLock()
	courier, exists := s.Couriers[req.ID]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("courier not found")
	}

	courier.mu.Lock()
	defer courier.mu.Unlock()

	courier.Balance += req.Amount
	return &api.CourierInfo{
		ID:      courier.ID,
		Name:    courier.Name,
		Balance: courier.Balance,
	}, nil
}

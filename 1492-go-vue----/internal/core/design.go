package core

import (
	"errors"
	"time"

	"renovation-management/internal/api"
)

func (s *Store) CreateDesignScheme(req *api.CreateDesignRequest) (*api.DesignScheme, error) {
	if req.HouseID == "" {
		return nil, errors.New("house_id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.houses[req.HouseID]; !ok {
		return nil, errors.New("house not found")
	}

	scheme := &api.DesignScheme{
		ID:          GenerateID("SCHEME"),
		HouseID:     req.HouseID,
		Version:     req.Version,
		Spaces:      req.Spaces,
		Description: req.Description,
		Confirmed:   false,
		CreatedAt:   time.Now(),
	}

	s.schemes[scheme.ID] = scheme
	s.schemesByHouse[req.HouseID] = append(s.schemesByHouse[req.HouseID], scheme)

	return scheme, nil
}

func (s *Store) GetDesignScheme(id string) (*api.DesignScheme, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scheme, ok := s.schemes[id]
	if !ok {
		return nil, errors.New("design scheme not found")
	}
	return scheme, nil
}

func (s *Store) ListDesignSchemesByHouse(houseID string) []*api.DesignScheme {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.schemesByHouse[houseID]
}

func (s *Store) ConfirmDesignScheme(schemeID string) (*api.DesignScheme, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	scheme, ok := s.schemes[schemeID]
	if !ok {
		return nil, errors.New("design scheme not found")
	}

	scheme.Confirmed = true
	return scheme, nil
}

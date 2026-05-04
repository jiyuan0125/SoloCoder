package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Storage struct {
	claims map[string]*WarrantyClaim
	mu     sync.RWMutex
	path   string
}

func NewStorage(path string) *Storage {
	s := &Storage{
		claims: make(map[string]*WarrantyClaim),
		path:   path,
	}
	s.load()
	return s
}

func (s *Storage) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		fmt.Printf("Warning: failed to read storage file: %v\n", err)
		return
	}

	var claims []*WarrantyClaim
	if err := json.Unmarshal(data, &claims); err != nil {
		fmt.Printf("Warning: failed to parse storage file: %v\n", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, claim := range claims {
		s.claims[claim.ID] = claim
	}
}

func (s *Storage) save() error {
	s.mu.RLock()
	claims := make([]*WarrantyClaim, 0, len(s.claims))
	for _, claim := range s.claims {
		claims = append(claims, claim)
	}
	s.mu.RUnlock()

	data, err := json.MarshalIndent(claims, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func (s *Storage) SaveClaim(claim *WarrantyClaim) error {
	s.mu.Lock()
	s.claims[claim.ID] = claim
	s.mu.Unlock()
	return s.save()
}

func (s *Storage) GetClaimByID(id string) (*WarrantyClaim, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	claim, exists := s.claims[id]
	return claim, exists
}

func (s *Storage) GetClaimsByUserID(userID string) []*WarrantyClaim {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*WarrantyClaim
	for _, claim := range s.claims {
		if claim.UserID == userID {
			result = append(result, claim)
		}
	}
	return result
}

func (s *Storage) GetClaimsByStatus(status ClaimStatus) []*WarrantyClaim {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*WarrantyClaim
	for _, claim := range s.claims {
		if claim.Status == status {
			result = append(result, claim)
		}
	}
	return result
}

func (s *Storage) GetAllClaims() []*WarrantyClaim {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*WarrantyClaim, 0, len(s.claims))
	for _, claim := range s.claims {
		result = append(result, claim)
	}
	return result
}

func (s *Storage) HasActiveClaim(serialNumber string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, claim := range s.claims {
		if claim.SerialNumber == serialNumber {
			if claim.Status == StatusPending || claim.Status == StatusProcessing {
				return true
			}
		}
	}
	return false
}

func (s *Storage) UpdateClaim(claim *WarrantyClaim) error {
	return s.SaveClaim(claim)
}

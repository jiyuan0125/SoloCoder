package main

import (
	"encoding/json"
	"os"
	"sync"
)

const (
	BatchesFile   = "batches.json"
	ClaimsFile    = "claims.json"
	RedeemsFile   = "redeems.json"
)

type Storage struct {
	batches   map[string]*CouponBatch
	claims    map[string]*CouponClaim
	redeems   map[string]*RedeemRecord
	batchClaims map[string][]*CouponClaim
	userClaims  map[string][]*CouponClaim
	codeToClaim map[string]*CouponClaim
	mu        sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		batches:     make(map[string]*CouponBatch),
		claims:      make(map[string]*CouponClaim),
		redeems:     make(map[string]*RedeemRecord),
		batchClaims: make(map[string][]*CouponClaim),
		userClaims:  make(map[string][]*CouponClaim),
		codeToClaim: make(map[string]*CouponClaim),
	}
}

func (s *Storage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.loadBatches(); err != nil {
		return err
	}
	if err := s.loadClaims(); err != nil {
		return err
	}
	if err := s.loadRedeems(); err != nil {
		return err
	}
	return nil
}

func (s *Storage) loadBatches() error {
	data, err := os.ReadFile(BatchesFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var batches []*CouponBatch
	if err := json.Unmarshal(data, &batches); err != nil {
		return err
	}

	for _, batch := range batches {
		s.batches[batch.ID] = batch
	}
	return nil
}

func (s *Storage) loadClaims() error {
	data, err := os.ReadFile(ClaimsFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var claims []*CouponClaim
	if err := json.Unmarshal(data, &claims); err != nil {
		return err
	}

	for _, claim := range claims {
		s.claims[claim.ID] = claim
		s.batchClaims[claim.BatchID] = append(s.batchClaims[claim.BatchID], claim)
		s.userClaims[claim.UserID] = append(s.userClaims[claim.UserID], claim)
		s.codeToClaim[claim.CouponCode] = claim
	}
	return nil
}

func (s *Storage) loadRedeems() error {
	data, err := os.ReadFile(RedeemsFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var redeems []*RedeemRecord
	if err := json.Unmarshal(data, &redeems); err != nil {
		return err
	}

	for _, redeem := range redeems {
		s.redeems[redeem.ID] = redeem
	}
	return nil
}

func (s *Storage) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.saveBatches(); err != nil {
		return err
	}
	if err := s.saveClaims(); err != nil {
		return err
	}
	if err := s.saveRedeems(); err != nil {
		return err
	}
	return nil
}

func (s *Storage) saveBatches() error {
	batches := make([]*CouponBatch, 0, len(s.batches))
	for _, batch := range s.batches {
		batches = append(batches, batch)
	}

	data, err := json.MarshalIndent(batches, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(BatchesFile, data, 0644)
}

func (s *Storage) saveClaims() error {
	claims := make([]*CouponClaim, 0, len(s.claims))
	for _, claim := range s.claims {
		claims = append(claims, claim)
	}

	data, err := json.MarshalIndent(claims, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ClaimsFile, data, 0644)
}

func (s *Storage) saveRedeems() error {
	redeems := make([]*RedeemRecord, 0, len(s.redeems))
	for _, redeem := range s.redeems {
		redeems = append(redeems, redeem)
	}

	data, err := json.MarshalIndent(redeems, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(RedeemsFile, data, 0644)
}

func (s *Storage) CreateBatch(batch *CouponBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.batches[batch.ID] = batch
	return s.saveBatches()
}

func (s *Storage) GetBatch(id string) (*CouponBatch, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	batch, exists := s.batches[id]
	return batch, exists
}

func (s *Storage) GetBatchByName(name string) (*CouponBatch, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, batch := range s.batches {
		if batch.Name == name {
			return batch, true
		}
	}
	return nil, false
}

func (s *Storage) UpdateBatch(batch *CouponBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.batches[batch.ID] = batch
	return s.saveBatches()
}

func (s *Storage) GetAllBatches() []*CouponBatch {
	s.mu.RLock()
	defer s.mu.RUnlock()

	batches := make([]*CouponBatch, 0, len(s.batches))
	for _, batch := range s.batches {
		batches = append(batches, batch)
	}
	return batches
}

func (s *Storage) CreateClaim(claim *CouponClaim) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.claims[claim.ID] = claim
	s.batchClaims[claim.BatchID] = append(s.batchClaims[claim.BatchID], claim)
	s.userClaims[claim.UserID] = append(s.userClaims[claim.UserID], claim)
	s.codeToClaim[claim.CouponCode] = claim
	return s.saveClaims()
}

func (s *Storage) GetClaim(id string) (*CouponClaim, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	claim, exists := s.claims[id]
	return claim, exists
}

func (s *Storage) GetClaimByCode(code string) (*CouponClaim, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	claim, exists := s.codeToClaim[code]
	return claim, exists
}

func (s *Storage) GetUserClaims(userID string) []*CouponClaim {
	s.mu.RLock()
	defer s.mu.RUnlock()

	claims := make([]*CouponClaim, len(s.userClaims[userID]))
	copy(claims, s.userClaims[userID])
	return claims
}

func (s *Storage) GetBatchClaims(batchID string) []*CouponClaim {
	s.mu.RLock()
	defer s.mu.RUnlock()

	claims := make([]*CouponClaim, len(s.batchClaims[batchID]))
	copy(claims, s.batchClaims[batchID])
	return claims
}

func (s *Storage) UpdateClaim(claim *CouponClaim) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.claims[claim.ID] = claim
	return s.saveClaims()
}

func (s *Storage) DeleteClaim(claim *CouponClaim) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.claims, claim.ID)
	delete(s.codeToClaim, claim.CouponCode)

	claims := s.batchClaims[claim.BatchID]
	for i, c := range claims {
		if c.ID == claim.ID {
			s.batchClaims[claim.BatchID] = append(claims[:i], claims[i+1:]...)
			break
		}
	}

	claims = s.userClaims[claim.UserID]
	for i, c := range claims {
		if c.ID == claim.ID {
			s.userClaims[claim.UserID] = append(claims[:i], claims[i+1:]...)
			break
		}
	}

	return s.saveClaims()
}

func (s *Storage) CreateRedeem(redeem *RedeemRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.redeems[redeem.ID] = redeem
	return s.saveRedeems()
}

func (s *Storage) GetRedeem(id string) (*RedeemRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	redeem, exists := s.redeems[id]
	return redeem, exists
}

func (s *Storage) Lock() {
	s.mu.Lock()
}

func (s *Storage) Unlock() {
	s.mu.Unlock()
}

func (s *Storage) RLock() {
	s.mu.RLock()
}

func (s *Storage) RUnlock() {
	s.mu.RUnlock()
}

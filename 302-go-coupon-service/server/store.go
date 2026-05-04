package main

import (
	"coupon-service/pkg/api"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Store struct {
	batches     map[string]*api.CouponBatch
	claims      map[string]*api.CouponClaim
	redemptions map[string]*api.CouponRedemption
	mu          sync.RWMutex
	dataDir     string
}

func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	store := &Store{
		batches:     make(map[string]*api.CouponBatch),
		claims:      make(map[string]*api.CouponClaim),
		redemptions: make(map[string]*api.CouponRedemption),
		dataDir:     dataDir,
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	batchesFile := filepath.Join(s.dataDir, "batches.json")
	if _, err := os.Stat(batchesFile); err == nil {
		data, err := os.ReadFile(batchesFile)
		if err != nil {
			return err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &s.batches); err != nil {
				return err
			}
		}
	}

	claimsFile := filepath.Join(s.dataDir, "claims.json")
	if _, err := os.Stat(claimsFile); err == nil {
		data, err := os.ReadFile(claimsFile)
		if err != nil {
			return err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &s.claims); err != nil {
				return err
			}
		}
	}

	redemptionsFile := filepath.Join(s.dataDir, "redemptions.json")
	if _, err := os.Stat(redemptionsFile); err == nil {
		data, err := os.ReadFile(redemptionsFile)
		if err != nil {
			return err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &s.redemptions); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Store) save() error {
	batchesData, err := json.MarshalIndent(s.batches, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(s.dataDir, "batches.json"), batchesData, 0644); err != nil {
		return err
	}

	claimsData, err := json.MarshalIndent(s.claims, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(s.dataDir, "claims.json"), claimsData, 0644); err != nil {
		return err
	}

	redemptionsData, err := json.MarshalIndent(s.redemptions, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(s.dataDir, "redemptions.json"), redemptionsData, 0644); err != nil {
		return err
	}

	return nil
}

func (s *Store) CreateBatch(batch *api.CouponBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if batch.Name == "" {
		return errors.New("coupon batch name cannot be empty")
	}

	if batch.Discount <= 0 {
		return errors.New("discount must be greater than zero")
	}

	for _, existingBatch := range s.batches {
		if existingBatch.Name == batch.Name {
			return errors.New("coupon batch name already exists")
		}
	}

	s.batches[batch.ID] = batch
	return s.save()
}

func (s *Store) GetBatch(batchID string) (*api.CouponBatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	batch, exists := s.batches[batchID]
	if !exists {
		return nil, errors.New("batch not found")
	}
	return batch, nil
}

func (s *Store) UpdateBatch(batch *api.CouponBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existingBatch, exists := s.batches[batch.ID]
	if !exists {
		return errors.New("batch not found")
	}

	if existingBatch.IsAllClaimed {
		return errors.New("cannot update batch that has been fully claimed")
	}

	if batch.Name != existingBatch.Name {
		for _, b := range s.batches {
			if b.ID != batch.ID && b.Name == batch.Name {
				return errors.New("coupon batch name already exists")
			}
		}
	}

	s.batches[batch.ID] = batch
	return s.save()
}

func (s *Store) CreateClaim(claim *api.CouponClaim) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	batch, exists := s.batches[claim.BatchID]
	if !exists {
		return errors.New("batch not found")
	}

	if batch.IsAllClaimed || batch.Claimed >= batch.TotalQuantity {
		return errors.New("batch has been fully claimed")
	}

	userClaimCount := 0
	for _, c := range s.claims {
		if c.UserID == claim.UserID && c.BatchID == claim.BatchID {
			userClaimCount++
		}
	}

	if userClaimCount >= batch.PerUserLimit {
		return errors.New("user has exceeded the claim limit for this batch")
	}

	s.claims[claim.ID] = claim
	batch.Claimed++
	if batch.Claimed >= batch.TotalQuantity {
		batch.IsAllClaimed = true
	}
	s.batches[claim.BatchID] = batch

	return s.save()
}

func (s *Store) GetClaim(claimID string) (*api.CouponClaim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	claim, exists := s.claims[claimID]
	if !exists {
		return nil, errors.New("claim not found")
	}
	return claim, nil
}

func (s *Store) GetUserClaims(userID string, onlyAvailable bool) []*api.CouponClaim {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var claims []*api.CouponClaim
	for _, claim := range s.claims {
		if claim.UserID == userID {
			if !onlyAvailable || !claim.IsRedeemed {
				claims = append(claims, claim)
			}
		}
	}
	return claims
}

func (s *Store) CreateRedemption(redemption *api.CouponRedemption, claim *api.CouponClaim, batch *api.CouponBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	claim.IsRedeemed = true
	claim.RedeemTime = redemption.RedeemTime
	s.claims[claim.ID] = claim

	batch.Redeemed++
	s.batches[batch.ID] = batch

	s.redemptions[redemption.ID] = redemption

	return s.save()
}

func (s *Store) ReturnCoupon(claimID string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	claim, exists := s.claims[claimID]
	if !exists {
		return errors.New("claim not found")
	}

	if claim.UserID != userID {
		return errors.New("claim does not belong to the user")
	}

	if claim.IsRedeemed {
		return errors.New("cannot return a redeemed coupon")
	}

	batch, exists := s.batches[claim.BatchID]
	if !exists {
		return errors.New("batch not found")
	}

	delete(s.claims, claimID)
	batch.Claimed--
	if batch.Claimed < batch.TotalQuantity {
		batch.IsAllClaimed = false
	}
	s.batches[batch.ID] = batch

	return s.save()
}

func (s *Store) CheckBatchExpired(batch *api.CouponBatch) bool {
	now := time.Now()
	return now.After(batch.EndTime)
}

func (s *Store) CheckClaimValid(claim *api.CouponClaim, batch *api.CouponBatch, userID string, orderAmount float64) error {
	if claim.IsRedeemed {
		return errors.New("coupon has already been redeemed")
	}

	if claim.UserID != userID {
		return errors.New("coupon does not belong to the user")
	}

	now := time.Now()
	if now.Before(batch.StartTime) {
		return errors.New("coupon has not become valid yet")
	}

	endOfDay := time.Date(batch.EndTime.Year(), batch.EndTime.Month(), batch.EndTime.Day(), 23, 59, 59, 0, batch.EndTime.Location())
	if now.After(endOfDay) {
		return errors.New("coupon has expired")
	}

	if orderAmount < batch.Threshold {
		return errors.New("order amount does not meet the threshold")
	}

	return nil
}

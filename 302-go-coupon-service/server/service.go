package main

import (
	"coupon-service/pkg/api"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type Service struct {
	store      *Store
	claimMutex sync.Map
}

func NewService(store *Store) *Service {
	return &Service{
		store:      store,
		claimMutex: sync.Map{},
	}
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func parseTime(timeStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339,
	}

	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, timeStr, time.Local)
		if err == nil {
			if layout == "2006-01-02" {
				return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()), nil
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
}

func (s *Service) CreateBatch(req *api.CreateBatchRequest) (*api.CreateBatchResponse, error) {
	startTime, err := parseTime(req.StartTime)
	if err != nil {
		return &api.CreateBatchResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid start time: %v", err),
		}, nil
	}

	endTime, err := parseTime(req.EndTime)
	if err != nil {
		return &api.CreateBatchResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid end time: %v", err),
		}, nil
	}

	if startTime.After(endTime) {
		return &api.CreateBatchResponse{
			Success: false,
			Message: "Start time must be before end time",
		}, nil
	}

	batch := &api.CouponBatch{
		ID:            generateID(),
		Name:          req.Name,
		Discount:      req.Discount,
		Threshold:     req.Threshold,
		TotalQuantity: req.TotalQuantity,
		Claimed:       0,
		Redeemed:      0,
		StartTime:     startTime,
		EndTime:       endTime,
		PerUserLimit:  req.PerUserLimit,
		IsAllClaimed:  false,
	}

	if err := s.store.CreateBatch(batch); err != nil {
		return &api.CreateBatchResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &api.CreateBatchResponse{
		Success: true,
		Message: "Coupon batch created successfully",
		Batch:   *batch,
	}, nil
}

func (s *Service) UpdateBatch(req *api.UpdateBatchRequest) (*api.UpdateBatchResponse, error) {
	batch, err := s.store.GetBatch(req.BatchID)
	if err != nil {
		return &api.UpdateBatchResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	if batch.IsAllClaimed {
		return &api.UpdateBatchResponse{
			Success: false,
			Message: "Cannot update batch that has been fully claimed",
		}, nil
	}

	if req.Name != "" {
		batch.Name = req.Name
	}
	if req.Discount > 0 {
		batch.Discount = req.Discount
	}
	if req.Threshold >= 0 {
		batch.Threshold = req.Threshold
	}
	if req.PerUserLimit > 0 {
		batch.PerUserLimit = req.PerUserLimit
	}
	if req.StartTime != "" {
		startTime, err := parseTime(req.StartTime)
		if err != nil {
			return &api.UpdateBatchResponse{
				Success: false,
				Message: fmt.Sprintf("Invalid start time: %v", err),
			}, nil
		}
		batch.StartTime = startTime
	}
	if req.EndTime != "" {
		endTime, err := parseTime(req.EndTime)
		if err != nil {
			return &api.UpdateBatchResponse{
				Success: false,
				Message: fmt.Sprintf("Invalid end time: %v", err),
			}, nil
		}
		batch.EndTime = endTime
	}

	if batch.StartTime.After(batch.EndTime) {
		return &api.UpdateBatchResponse{
			Success: false,
			Message: "Start time must be before end time",
		}, nil
	}

	if err := s.store.UpdateBatch(batch); err != nil {
		return &api.UpdateBatchResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &api.UpdateBatchResponse{
		Success: true,
		Message: "Coupon batch updated successfully",
		Batch:   *batch,
	}, nil
}

func (s *Service) ClaimCoupon(req *api.ClaimCouponRequest) (*api.ClaimCouponResponse, error) {
	mutexKey := fmt.Sprintf("%s:%s", req.UserID, req.BatchID)
	mutex, _ := s.claimMutex.LoadOrStore(mutexKey, &sync.Mutex{})
	userBatchMutex := mutex.(*sync.Mutex)

	userBatchMutex.Lock()
	defer userBatchMutex.Unlock()

	batch, err := s.store.GetBatch(req.BatchID)
	if err != nil {
		return &api.ClaimCouponResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	if batch.IsAllClaimed || batch.Claimed >= batch.TotalQuantity {
		return &api.ClaimCouponResponse{
			Success: false,
			Message: "Batch has been fully claimed",
		}, nil
	}

	now := time.Now()
	endOfDay := time.Date(batch.EndTime.Year(), batch.EndTime.Month(), batch.EndTime.Day(), 23, 59, 59, 0, batch.EndTime.Location())
	if now.After(endOfDay) {
		return &api.ClaimCouponResponse{
			Success: false,
			Message: "Coupon batch has expired",
		}, nil
	}

	userClaims := s.store.GetUserClaims(req.UserID, false)
	userBatchClaimCount := 0
	for _, claim := range userClaims {
		if claim.BatchID == req.BatchID {
			userBatchClaimCount++
		}
	}

	if userBatchClaimCount >= batch.PerUserLimit {
		return &api.ClaimCouponResponse{
			Success: false,
			Message: "User has exceeded the claim limit for this batch",
		}, nil
	}

	claim := &api.CouponClaim{
		ID:         generateID(),
		UserID:     req.UserID,
		BatchID:    req.BatchID,
		ClaimTime:  time.Now(),
		IsRedeemed: false,
		RedeemTime: time.Time{},
	}

	if err := s.store.CreateClaim(claim); err != nil {
		return &api.ClaimCouponResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &api.ClaimCouponResponse{
		Success: true,
		Message: "Coupon claimed successfully",
		Claim:   *claim,
	}, nil
}

func (s *Service) RedeemCoupon(req *api.RedeemCouponRequest) (*api.RedeemCouponResponse, error) {
	claim, err := s.store.GetClaim(req.ClaimID)
	if err != nil {
		return &api.RedeemCouponResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	batch, err := s.store.GetBatch(claim.BatchID)
	if err != nil {
		return &api.RedeemCouponResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	if err := s.store.CheckClaimValid(claim, batch, req.UserID, req.OrderAmount); err != nil {
		return &api.RedeemCouponResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	redemption := &api.CouponRedemption{
		ID:           generateID(),
		ClaimID:      req.ClaimID,
		BatchID:      claim.BatchID,
		UserID:       req.UserID,
		OrderAmount:  req.OrderAmount,
		RedeemTime:   time.Now(),
		DiscountUsed: batch.Discount,
	}

	if err := s.store.CreateRedemption(redemption, claim, batch); err != nil {
		return &api.RedeemCouponResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &api.RedeemCouponResponse{
		Success:    true,
		Message:    "Coupon redeemed successfully",
		Redemption: *redemption,
	}, nil
}

func (s *Service) GetBatchStats(req *api.GetBatchStatsRequest) (*api.GetBatchStatsResponse, error) {
	batch, err := s.store.GetBatch(req.BatchID)
	if err != nil {
		return &api.GetBatchStatsResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	remaining := batch.TotalQuantity - batch.Claimed

	return &api.GetBatchStatsResponse{
		Success:   true,
		Message:   "Batch statistics retrieved successfully",
		Batch:     *batch,
		Claimed:   batch.Claimed,
		Redeemed:  batch.Redeemed,
		Remaining: remaining,
	}, nil
}

func (s *Service) GetUserCoupons(req *api.GetUserCouponsRequest) (*api.GetUserCouponsResponse, error) {
	claims := s.store.GetUserClaims(req.UserID, true)

	availableCoupons := make([]api.CouponClaim, 0, len(claims))
	for _, claim := range claims {
		if !claim.IsRedeemed {
			availableCoupons = append(availableCoupons, *claim)
		}
	}

	return &api.GetUserCouponsResponse{
		Success: true,
		Message: "User coupons retrieved successfully",
		Coupons: availableCoupons,
	}, nil
}

func (s *Service) ReturnCoupon(req *api.ReturnCouponRequest) (*api.ReturnCouponResponse, error) {
	if err := s.store.ReturnCoupon(req.ClaimID, req.UserID); err != nil {
		return &api.ReturnCouponResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &api.ReturnCouponResponse{
		Success: true,
		Message: "Coupon returned successfully",
	}, nil
}

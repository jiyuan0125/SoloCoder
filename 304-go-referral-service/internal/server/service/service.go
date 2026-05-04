package service

import (
	"errors"
	"time"

	"referral-service/internal/server/generator"
	"referral-service/internal/server/store"
)

var (
	ErrCodeAlreadyExists    = errors.New("user already has a referral code")
	ErrInvalidReferralCode  = errors.New("invalid referral code")
	ErrSelfReferral         = errors.New("cannot use your own referral code")
	ErrAlreadyBound         = errors.New("user already bound to a referrer")
	ErrNoReferral           = errors.New("user has no referral relationship")
	ErrOrderAlreadyCompleted = errors.New("first order already completed")
	ErrInvalidOrder         = errors.New("invalid order")
)

type Service struct {
	store     *store.Store
	generator *generator.Generator
}

func NewService(s *store.Store, g *generator.Generator) *Service {
	return &Service{
		store:     s,
		generator: g,
	}
}

func (s *Service) GenerateReferralCode(userID string) (string, error) {
	if existingCode, exists := s.store.GetReferralCodeByUser(userID); exists {
		return existingCode, nil
	}

	code := s.generator.GenerateUnique(s.store.IsReferralCodeExists)

	if err := s.store.SaveReferralCode(userID, code); err != nil {
		return "", err
	}

	return code, nil
}

func (s *Service) BindReferral(newUserID, referralCode string) error {
	referrerID, exists := s.store.GetUserByReferralCode(referralCode)
	if !exists {
		return ErrInvalidReferralCode
	}

	if referrerID == newUserID {
		return ErrSelfReferral
	}

	if _, exists := s.store.GetReferralRelationship(newUserID); exists {
		return ErrAlreadyBound
	}

	now := time.Now().Unix()
	return s.store.BindReferral(newUserID, referrerID, now)
}

func (s *Service) CompleteFirstOrder(userID, orderID string) (int64, error) {
	rel, exists := s.store.GetReferralRelationship(userID)
	if !exists {
		return 0, ErrNoReferral
	}

	if rel.FirstOrderDone {
		return 0, ErrOrderAlreadyCompleted
	}

	points := s.store.GetRewardPoints()

	if err := s.store.CompleteFirstOrder(userID, orderID, points); err != nil {
		return 0, err
	}

	return points, nil
}

func (s *Service) RefundFirstOrder(userID, orderID string) (int64, error) {
	rel, exists := s.store.GetReferralRelationship(userID)
	if !exists {
		return 0, ErrNoReferral
	}

	if !rel.FirstOrderDone || rel.OrderID != orderID {
		return 0, ErrInvalidOrder
	}

	return s.store.RefundFirstOrder(userID, orderID)
}

func (s *Service) GetUserStats(userID string) (string, int64, int64, int64) {
	code, _ := s.store.GetReferralCodeByUser(userID)
	
	userData, exists := s.store.GetUserData(userID)
	if !exists {
		return code, 0, 0, 0
	}

	return code, userData.TotalReferrals, userData.CompletedFirstOrders, userData.TotalPoints
}

func (s *Service) SetRewardPoints(points int64) error {
	return s.store.SetRewardPoints(points)
}

func (s *Service) GetRewardPoints() int64 {
	return s.store.GetRewardPoints()
}

func (s *Service) GetStats() (int64, int64, float64, int64) {
	totalReferrals, completedFirstOrders, totalPointsIssued := s.store.GetStats()
	
	var rate float64
	if totalReferrals > 0 {
		rate = float64(completedFirstOrders) / float64(totalReferrals) * 100
	}

	return totalReferrals, completedFirstOrders, rate, totalPointsIssued
}

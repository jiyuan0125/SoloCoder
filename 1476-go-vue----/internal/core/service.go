package core

import (
	"gasstation/internal/common"
	"sync"
	"time"
)

type RefuelService struct {
	store *Store
	mu    sync.Mutex
}

func NewRefuelService(store *Store) *RefuelService {
	return &RefuelService{
		store: store,
	}
}

func (s *RefuelService) ProcessRefuel(req common.RefuelRequest) (*common.RefuelRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	liters, err := RoundLiters(req.Liters)
	if err != nil {
		return nil, err
	}

	fuel, err := s.store.GetStationFuel(req.StationID, req.FuelCode)
	if err != nil {
		return nil, err
	}

	litersInt := LitersToInt64(liters)
	if fuel.Stock < litersInt {
		return nil, ErrInsufficientStock
	}

	originalAmount := int64(liters) * fuel.Price
	if liters > float64(int64(liters)) {
		originalAmount = int64(liters*100) * fuel.Price / 100
	}

	record := &common.RefuelRecord{
		ID:             s.store.GenerateID(),
		StationID:      req.StationID,
		FuelCode:       req.FuelCode,
		Liters:         liters,
		UnitPrice:      fuel.Price,
		OriginalAmount: originalAmount,
		CreatedAt:      time.Now(),
	}

	var member *common.Member
	if req.MemberPhone != nil && *req.MemberPhone != "" {
		m, err := s.store.GetMemberByPhone(*req.MemberPhone)
		if err != nil {
			member = nil
		} else {
			member = m
		}
	}

	if member != nil {
		levelStr := string(member.Level)
		record.MemberID = &member.ID
		record.MemberLevel = &levelStr

		discount := CalculateDiscount(member.Level, originalAmount)
		record.DiscountAmount = discount

		afterDiscount := originalAmount - discount

		if req.UsePoints != nil && *req.UsePoints > 0 && member.Points > 0 {
			pointsToUse := *req.UsePoints
			if pointsToUse > member.Points {
				pointsToUse = member.Points
			}

			deduction := PointsToDeduction(pointsToUse)
			actualPointsUsed := deduction

			if actualPointsUsed > afterDiscount {
				actualPointsUsed = (afterDiscount / 100) * 100
			}

			record.PointsUsed = actualPointsUsed
			record.PointsDeducted = actualPointsUsed / 100

			final := afterDiscount - actualPointsUsed
			if final < 0 {
				final = 0
			}
			record.FinalAmount = final

			pointsDelta := CalculatePointsFromAmount(record.FinalAmount) - actualPointsUsed
			s.store.UpdateMemberPointsAndLevel(member.ID, pointsDelta, record.FinalAmount)
		} else {
			record.FinalAmount = afterDiscount
			pointsEarned := CalculatePointsFromAmount(record.FinalAmount)
			s.store.UpdateMemberPointsAndLevel(member.ID, pointsEarned, record.FinalAmount)
		}
	} else {
		record.FinalAmount = originalAmount
	}

	if err := s.store.DeductStock(req.StationID, req.FuelCode, litersInt); err != nil {
		return nil, err
	}

	if err := s.store.AddRefuelRecord(record); err != nil {
		return nil, err
	}

	return record, nil
}

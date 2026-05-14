package service

import (
	"fmt"
	"groupbuy/models"
	"groupbuy/storage"
	"sort"
	"time"
)

type Service struct {
	store *storage.Storage
}

func NewService(store *storage.Storage) *Service {
	return &Service{store: store}
}

func DefaultTiers() []models.Tier {
	return []models.Tier{
		{MinPeople: 10, Discount: 0.9},
		{MinPeople: 50, Discount: 0.8},
		{MinPeople: 100, Discount: 0.7},
		{MinPeople: 500, Discount: 0.6},
	}
}

func (s *Service) CreateActivity(name string, basePrice int64, deadline time.Time, tiers []models.Tier) (*models.Activity, error) {
	if tiers == nil {
		tiers = DefaultTiers()
	}
	sort.Slice(tiers, func(i, j int) bool {
		return tiers[i].MinPeople < tiers[j].MinPeople
	})
	activity := models.Activity{
		ActivityID:        fmt.Sprintf("act_%d", time.Now().UnixNano()),
		Name:              name,
		BasePrice:         basePrice,
		Tiers:             tiers,
		Deadline:          deadline,
		Status:            "active",
		CreatedAt:         time.Now(),
		CurrentTier:       -1,
		ParticipantCount:  0,
	}
	if err := s.store.SaveActivity(activity); err != nil {
		return nil, err
	}
	return &activity, nil
}

func (s *Service) GetActivity(activityID string) (*models.Activity, error) {
	return s.store.GetActivity(activityID)
}

func (s *Service) ListActivities() ([]models.Activity, error) {
	return s.store.ListActivities()
}

func (s *Service) calculateCurrentTier(tiers []models.Tier, participantCount int) int {
	currentTierIndex := -1
	for i, tier := range tiers {
		if participantCount >= tier.MinPeople {
			currentTierIndex = i
		}
	}
	return currentTierIndex
}

func (s *Service) calculatePrice(basePrice int64, tierIndex int, tiers []models.Tier) int64 {
	if tierIndex < 0 || tierIndex >= len(tiers) {
		return basePrice
	}
	discount := tiers[tierIndex].Discount
	return int64(float64(basePrice) * discount)
}

func (s *Service) PlaceOrder(activityID, userID string) (*models.Order, error) {
	activity, err := s.store.GetActivity(activityID)
	if err != nil {
		return nil, err
	}
	if activity.Status != "active" {
		return nil, fmt.Errorf("活动已结束或已取消")
	}
	if time.Now().After(activity.Deadline) {
		return nil, fmt.Errorf("活动已截止")
	}
	if s.store.HasUserParticipated(activityID, userID) {
		return nil, fmt.Errorf("已参与该活动")
	}
	currentPrice := activity.BasePrice
	if activity.CurrentTier >= 0 {
		currentPrice = s.calculatePrice(activity.BasePrice, activity.CurrentTier, activity.Tiers)
	}
	order := models.Order{
		OrderID:    fmt.Sprintf("ord_%d", time.Now().UnixNano()),
		ActivityID: activityID,
		UserID:     userID,
		PaidPrice:  currentPrice,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}
	if err := s.store.SaveOrder(order); err != nil {
		return nil, err
	}
	order.Status = "paid"
	if err := s.store.SaveOrder(order); err != nil {
		return nil, err
	}
	activity.ParticipantCount++
	newTierIndex := s.calculateCurrentTier(activity.Tiers, activity.ParticipantCount)
	if newTierIndex != activity.CurrentTier && newTierIndex >= 0 {
		oldTierIndex := activity.CurrentTier
		activity.CurrentTier = newTierIndex
		if err := s.store.SaveActivity(*activity); err != nil {
			return nil, err
		}
		if oldTierIndex >= 0 {
			if err := s.processTierUpgrade(activityID, activity.BasePrice, oldTierIndex, newTierIndex, activity.Tiers); err != nil {
				return nil, err
			}
		} else {
			if err := s.processFirstTierUpgrade(activityID, activity.BasePrice, newTierIndex, activity.Tiers); err != nil {
				return nil, err
			}
		}
	} else {
		if err := s.store.SaveActivity(*activity); err != nil {
			return nil, err
		}
	}
	return &order, nil
}

func (s *Service) processFirstTierUpgrade(activityID string, basePrice int64, newTierIndex int, tiers []models.Tier) error {
	orders, err := s.store.GetOrdersByActivity(activityID)
	if err != nil {
		return err
	}
	newPrice := s.calculatePrice(basePrice, newTierIndex, tiers)
	tier := tiers[newTierIndex]
	for _, order := range orders {
		if order.Status != "paid" {
			continue
		}
		refundAmount := order.PaidPrice - newPrice
		if refundAmount <= 0 {
			continue
		}
		if err := s.refundToBalance(order.UserID, refundAmount); err != nil {
			return err
		}
		record := models.RefundRecord{
			RefundID:   fmt.Sprintf("ref_%d", time.Now().UnixNano()),
			UserID:     order.UserID,
			ActivityID: activityID,
			OrderID:    order.OrderID,
			Amount:     refundAmount,
			Reason:     fmt.Sprintf("阶梯升级至%d人%0.0f折", tier.MinPeople, tier.Discount*10),
			CreatedAt:  time.Now(),
		}
		if err := s.store.SaveRefundRecord(record); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) processTierUpgrade(activityID string, basePrice int64, oldTierIndex, newTierIndex int, tiers []models.Tier) error {
	orders, err := s.store.GetOrdersByActivity(activityID)
	if err != nil {
		return err
	}
	oldPrice := s.calculatePrice(basePrice, oldTierIndex, tiers)
	newPrice := s.calculatePrice(basePrice, newTierIndex, tiers)
	refundPerPerson := oldPrice - newPrice
	if refundPerPerson <= 0 {
		return nil
	}
	newTier := tiers[newTierIndex]
	for _, order := range orders {
		if order.Status != "paid" {
			continue
		}
		if err := s.refundToBalance(order.UserID, refundPerPerson); err != nil {
			return err
		}
		record := models.RefundRecord{
			RefundID:   fmt.Sprintf("ref_%d", time.Now().UnixNano()),
			UserID:     order.UserID,
			ActivityID: activityID,
			OrderID:    order.OrderID,
			Amount:     refundPerPerson,
			Reason:     fmt.Sprintf("阶梯升级至%d人%0.0f折", newTier.MinPeople, newTier.Discount*10),
			CreatedAt:  time.Now(),
		}
		if err := s.store.SaveRefundRecord(record); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) refundToBalance(userID string, amount int64) error {
	_, err := s.store.GetOrCreateUser(userID)
	if err != nil {
		return err
	}
	return s.store.UpdateUserBalance(userID, amount)
}

func (s *Service) SettleActivity(activityID string) error {
	activity, err := s.store.GetActivity(activityID)
	if err != nil {
		return err
	}
	if activity.Status != "active" {
		return fmt.Errorf("活动已结算或已取消")
	}
	now := time.Now()
	if !now.After(activity.Deadline) {
		return fmt.Errorf("活动尚未截止，不能结算")
	}
	minPeople := 10
	if activity.ParticipantCount < minPeople {
		return s.cancelActivity(activityID, activity.BasePrice)
	}
	return s.completeActivity(activityID, activity.BasePrice, activity.CurrentTier, activity.Tiers)
}

func (s *Service) cancelActivity(activityID string, basePrice int64) error {
	orders, err := s.store.GetOrdersByActivity(activityID)
	if err != nil {
		return err
	}
	for _, order := range orders {
		if order.Status != "paid" {
			continue
		}
		if err := s.refundToBalance(order.UserID, order.PaidPrice); err != nil {
			return err
		}
		record := models.RefundRecord{
			RefundID:   fmt.Sprintf("ref_%d", time.Now().UnixNano()),
			UserID:     order.UserID,
			ActivityID: activityID,
			OrderID:    order.OrderID,
			Amount:     order.PaidPrice,
			Reason:     "活动取消，全额退款",
			CreatedAt:  time.Now(),
		}
		if err := s.store.SaveRefundRecord(record); err != nil {
			return err
		}
		if err := s.store.UpdateOrderStatus(order.OrderID, "refunded"); err != nil {
			return err
		}
	}
	return s.store.WriteTransaction(func(store *storage.DataStore) error {
		a := store.Activities[activityID]
		a.Status = "cancelled"
		now := time.Now()
		a.SettledAt = &now
		store.Activities[activityID] = a
		return nil
	})
}

func (s *Service) completeActivity(activityID string, basePrice int64, currentTierIndex int, tiers []models.Tier) error {
	orders, err := s.store.GetOrdersByActivity(activityID)
	if err != nil {
		return err
	}
	finalPrice := basePrice
	if currentTierIndex >= 0 {
		finalPrice = s.calculatePrice(basePrice, currentTierIndex, tiers)
	}
	currentTier := "无阶梯"
	if currentTierIndex >= 0 {
		t := tiers[currentTierIndex]
		currentTier = fmt.Sprintf("%d人%0.0f折", t.MinPeople, t.Discount*10)
	}
	for _, order := range orders {
		if order.Status != "paid" {
			continue
		}
		refundAmount := order.PaidPrice - finalPrice
		if refundAmount > 0 {
			if err := s.refundToBalance(order.UserID, refundAmount); err != nil {
				return err
			}
			record := models.RefundRecord{
				RefundID:   fmt.Sprintf("ref_%d", time.Now().UnixNano()),
				UserID:     order.UserID,
				ActivityID: activityID,
				OrderID:    order.OrderID,
				Amount:     refundAmount,
				Reason:     fmt.Sprintf("结算：最终阶梯价%s", currentTier),
				CreatedAt:  time.Now(),
			}
			if err := s.store.SaveRefundRecord(record); err != nil {
				return err
			}
		}
		if err := s.store.UpdateOrderStatus(order.OrderID, "completed"); err != nil {
			return err
		}
	}
	return s.store.WriteTransaction(func(store *storage.DataStore) error {
		a := store.Activities[activityID]
		a.Status = "completed"
		now := time.Now()
		a.SettledAt = &now
		store.Activities[activityID] = a
		return nil
	})
}

func (s *Service) GetBalance(userID string) (int64, error) {
	user, err := s.store.GetOrCreateUser(userID)
	if err != nil {
		return 0, err
	}
	return user.Balance, nil
}

func (s *Service) Withdraw(userID string, amount int64) (netAmount int64, fee int64, err error) {
	user, err := s.store.GetOrCreateUser(userID)
	if err != nil {
		return 0, 0, err
	}
	if amount <= 0 {
		return 0, 0, fmt.Errorf("提现金额必须大于0")
	}
	if user.Balance < amount {
		return 0, 0, fmt.Errorf("余额不足")
	}
	fee = int64(float64(amount) * 0.01)
	minFee := int64(10)
	if fee < minFee {
		fee = minFee
	}
	if amount <= fee {
		return 0, 0, fmt.Errorf("提现金额不足以支付手续费")
	}
	netAmount = amount - fee
	if err := s.store.UpdateUserBalance(userID, -amount); err != nil {
		return 0, 0, err
	}
	record := models.WithdrawalRecord{
		WithdrawalID: fmt.Sprintf("wd_%d", time.Now().UnixNano()),
		UserID:       userID,
		Amount:       amount,
		Fee:          fee,
		NetAmount:    netAmount,
		CreatedAt:    time.Now(),
	}
	if err := s.store.SaveWithdrawalRecord(record); err != nil {
		return 0, 0, err
	}
	return netAmount, fee, nil
}

func (s *Service) GetRefundRecords(userID string) ([]models.RefundRecord, error) {
	return s.store.GetRefundRecordsByUser(userID)
}

func (s *Service) GetOrders(userID string) ([]models.Order, error) {
	return s.store.GetOrdersByUser(userID)
}

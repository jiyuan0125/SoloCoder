package core

import (
	"errors"
	"time"
)

var (
	ErrDailyClosingNotFound = errors.New("daily closing not found")
	ErrDailyClosingAlreadyExists = errors.New("daily closing already exists for this date")
)

func (s *Store) CreateDailyClosing(dateStr string) (*DailyClosing, error) {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.dailyClosings[dateStr]; exists {
		return nil, ErrDailyClosingAlreadyExists
	}

	orders := s.getOrdersByDateLocked(dateStr)

	paymentStatsMap := make(map[string]*DailyClosingPaymentStat)
	var totalAmount int64 = 0

	for _, order := range orders {
		stat, exists := paymentStatsMap[order.PaymentMethod]
		if !exists {
			stat = &DailyClosingPaymentStat{
				PaymentMethod: order.PaymentMethod,
				TotalAmount:   0,
				OrderCount:    0,
			}
			paymentStatsMap[order.PaymentMethod] = stat
		}

		stat.TotalAmount += order.PayableAmount
		stat.OrderCount++
		totalAmount += order.PayableAmount
	}

	paymentStats := make([]DailyClosingPaymentStat, 0, len(paymentStatsMap))
	for _, stat := range paymentStatsMap {
		paymentStats = append(paymentStats, *stat)
	}

	closingID := s.idGen.Generate()
	closing := &DailyClosing{
		ID:           closingID,
		Date:         dateStr,
		TotalAmount:  totalAmount,
		PaymentStats: paymentStats,
		OrderCount:   len(orders),
		ClosedAt:     time.Now(),
	}

	s.dailyClosings[dateStr] = closing

	return closing, nil
}

func (s *Store) GetDailyClosing(dateStr string) (*DailyClosing, error) {
	s.RLock()
	defer s.RUnlock()

	closing, exists := s.dailyClosings[dateStr]
	if !exists {
		return nil, ErrDailyClosingNotFound
	}

	closingCopy := *closing
	closingCopy.PaymentStats = make([]DailyClosingPaymentStat, len(closing.PaymentStats))
	copy(closingCopy.PaymentStats, closing.PaymentStats)

	return &closingCopy, nil
}

func (s *Store) IsDateClosed(dateStr string) bool {
	s.RLock()
	defer s.RUnlock()

	_, exists := s.dailyClosings[dateStr]
	return exists
}

func (s *Store) getOrdersByDateLocked(dateStr string) []*Order {
	var orders []*Order
	for _, order := range s.orders {
		orderDate := order.PaymentTime.Format("2006-01-02")
		if orderDate == dateStr {
			orders = append(orders, order)
		}
	}
	return orders
}

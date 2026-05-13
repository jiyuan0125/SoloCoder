package service

import (
	"database/sql"
	"errors"
	"time"

	"flashsale/internal/model"
	"flashsale/internal/repository"
)

var (
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderAlreadyPaid     = errors.New("order already paid")
	ErrOrderAlreadyCancelled = errors.New("order already cancelled")
	ErrOrderExpired          = errors.New("order has expired")
)

type OrderService struct {
	orderRepo    *repository.OrderRepository
	stockRepo    *repository.StockRepository
	activityRepo *repository.ActivityRepository
}

func NewOrderService(
	orderRepo *repository.OrderRepository,
	stockRepo *repository.StockRepository,
	activityRepo *repository.ActivityRepository,
) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		stockRepo:    stockRepo,
		activityRepo: activityRepo,
	}
}

func (s *OrderService) GetOrder(orderNo string) (*model.Order, error) {
	order, err := s.orderRepo.GetByOrderNo(orderNo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return order, nil
}

func (s *OrderService) PayOrder(orderNo string) error {
	order, err := s.GetOrder(orderNo)
	if err != nil {
		return err
	}

	if order.Status == model.OrderStatusPaid {
		return ErrOrderAlreadyPaid
	}
	if order.Status == model.OrderStatusCancelled {
		return ErrOrderAlreadyCancelled
	}

	if time.Now().After(order.ExpireAt) {
		return ErrOrderExpired
	}

	success, err := s.orderRepo.UpdateStatus(orderNo, model.OrderStatusPaid, true)
	if err != nil {
		return err
	}

	if !success {
		return ErrOrderAlreadyPaid
	}

	return nil
}

func (s *OrderService) CancelOrder(orderNo string) error {
	order, err := s.GetOrder(orderNo)
	if err != nil {
		return err
	}

	if order.Status == model.OrderStatusPaid {
		return ErrOrderAlreadyPaid
	}
	if order.Status == model.OrderStatusCancelled {
		return ErrOrderAlreadyCancelled
	}

	success, err := s.orderRepo.UpdateStatus(orderNo, model.OrderStatusCancelled, false)
	if err != nil {
		return err
	}

	if !success {
		return ErrOrderAlreadyPaid
	}

	if err := s.stockRepo.IncrementStock(order.ActivityID, 1); err != nil {
		return err
	}

	return nil
}

func (s *OrderService) CleanupExpiredOrders(now time.Time) ([]string, error) {
	expiredOrders, err := s.orderRepo.GetExpiredOrders(now)
	if err != nil {
		return nil, err
	}

	cancelledOrders := make([]string, 0)

	for _, order := range expiredOrders {
		success, err := s.orderRepo.UpdateStatus(order.OrderNo, model.OrderStatusCancelled, false)
		if err != nil {
			continue
		}

		if !success {
			continue
		}

		if err := s.stockRepo.IncrementStock(order.ActivityID, 1); err != nil {
			continue
		}

		cancelledOrders = append(cancelledOrders, order.OrderNo)
	}

	return cancelledOrders, nil
}

package service

import (
	"database/sql"
	"errors"
	"sync"
	"time"

	"flashsale/internal/model"
	"flashsale/internal/repository"
)

var (
	ErrActivityNotFound       = errors.New("activity not found")
	ErrFlashPriceTooHigh      = errors.New("flash price must be less than original price")
	ErrActivityNotStarted     = errors.New("activity has not started")
	ErrActivityEnded          = errors.New("activity has ended")
	ErrStockEmpty             = errors.New("stock is empty")
	ErrAlreadyPurchased       = errors.New("user has already purchased this activity")
	ErrPurchaseFailed         = errors.New("purchase failed, please try again")
)

type ActivityService struct {
	activityRepo *repository.ActivityRepository
	stockRepo    *repository.StockRepository
	orderRepo    *repository.OrderRepository
	mu         sync.Mutex
}

func NewActivityService(
	activityRepo *repository.ActivityRepository,
	stockRepo *repository.StockRepository,
	orderRepo *repository.OrderRepository,
) *ActivityService {
	return &ActivityService{
		activityRepo: activityRepo,
		stockRepo:    stockRepo,
		orderRepo:    orderRepo,
	}
}

func (s *ActivityService) CreateActivity(
	productName string,
	originalPrice, flashPrice float64,
	totalStock int,
	startTime time.Time,
) (*model.Activity, error) {
	if flashPrice >= originalPrice {
		return nil, ErrFlashPriceTooHigh
	}

	activity := &model.Activity{
		ProductName:  productName,
		OriginalPrice: originalPrice,
		FlashPrice:   flashPrice,
		TotalStock:   totalStock,
		StartTime:    startTime,
	}

	if err := s.activityRepo.Create(activity); err != nil {
		return nil, err
	}

	return activity, nil
}

func (s *ActivityService) GetActivity(id int64) (*model.Activity, error) {
	activity, err := s.activityRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrActivityNotFound
		}
		return nil, err
	}

	activity.Status = s.getActivityStatus(activity)
	return activity, nil
}

func (s *ActivityService) GetActivityWithStock(id int64) (*model.Activity, *model.Stock, error) {
	activity, err := s.GetActivity(id)
	if err != nil {
		return nil, nil, err
	}

	stock, err := s.stockRepo.GetByActivityID(id)
	if err != nil {
		return nil, nil, err
	}

	return activity, stock, nil
}

func (s *ActivityService) Purchase(activityID int64, userID string) (*model.Order, error) {
	activity, stock, err := s.GetActivityWithStock(activityID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if activity.Status == model.ActivityStatusNotStarted {
		return nil, ErrActivityNotStarted
	}

	if stock.AvailableStock <= 0 {
		return nil, ErrStockEmpty
	}

	exists, err := s.orderRepo.Exists(activityID, userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyPurchased
	}

	success, err := s.stockRepo.DecrementStock(activityID, 1)
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, ErrPurchaseFailed
	}

	orderNo := generateOrderNo()
	order := &model.Order{
		OrderNo:    orderNo,
		ActivityID: activityID,
		UserID:     userID,
		Status:     model.OrderStatusPending,
		Price:      activity.FlashPrice,
		ExpireAt:   now.Add(10 * time.Minute),
	}

	if err := s.orderRepo.Create(order); err != nil {
		s.stockRepo.IncrementStock(activityID, 1)
		return nil, err
	}

	orderFromDB, err := s.orderRepo.GetByOrderNo(orderNo)
	if err == nil {
		order.CreatedAt = orderFromDB.CreatedAt
	}

	return order, nil
}

func (s *ActivityService) getActivityStatus(activity *model.Activity) model.ActivityStatus {
	now := time.Now()

	if now.Before(activity.StartTime) {
		return model.ActivityStatusNotStarted
	}

	return model.ActivityStatusActive
}

func (s *ActivityService) ListActivities() ([]*model.Activity, error) {
	activities, err := s.activityRepo.ListActive()
	if err != nil {
		return nil, err
	}

	for _, activity := range activities {
		activity.Status = s.getActivityStatus(activity)
	}

	return activities, nil
}

func generateOrderNo() string {
	return "ORD" + time.Now().Format("20060102150405") + randomString(8)
}

func randomString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}

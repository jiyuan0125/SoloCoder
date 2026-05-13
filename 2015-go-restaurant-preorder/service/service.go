package service

import (
	"database/sql"
	"errors"
	"fmt"
	"restaurant-preorder/database"
	"restaurant-preorder/models"
	"time"
)

const (
	MaxTablesPerSlot   = 20
	MaxGuestsPerOrder  = 20
	MinGuestsPerOrder  = 1
	CancellationHours  = 2
	NoShowGraceMinutes = 30
	BlacklistDays      = 30
	MaxNoShows         = 3
	KitchenSummaryHours = 4
)

var (
	ErrInvalidTimeSlot     = errors.New("无效时段")
	ErrInvalidGuestCount   = errors.New("人数不符合要求")
	ErrSlotFull            = errors.New("该时段已满")
	ErrDuplicateBooking    = errors.New("该手机号当天已预订")
	ErrBlacklisted         = errors.New("该手机号已被拉黑")
	ErrTooLateToCancel     = errors.New("距就餐不足 2 小时")
	ErrAlreadyCancelled    = errors.New("订单已取消")
	ErrInvalidStatus       = errors.New("无效状态")
	ErrInvalidStatusTransition = errors.New("状态转换无效")
	ErrOrderNotFound       = errors.New("订单不存在")
)

func ValidateCreateOrder(req CreateOrderRequest) error {
	if req.TimeSlot != models.Lunch && req.TimeSlot != models.Dinner {
		return ErrInvalidTimeSlot
	}

	if req.GuestCount < MinGuestsPerOrder || req.GuestCount > MaxGuestsPerOrder {
		return ErrInvalidGuestCount
	}

	if req.ResourceType == "" {
		req.ResourceType = models.ResourceTable
	}
	if req.ResourceID == "" {
		req.ResourceID = "default"
	}

	return nil
}

func CreateOrder(req CreateOrderRequest) (*models.OrderWithItems, error) {
	if err := ValidateCreateOrder(req); err != nil {
		return nil, err
	}

	blacklisted, err := database.IsBlacklisted(req.Phone)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, ErrBlacklisted
	}

	hasOrder, err := database.HasOrderOnDate(req.Phone, req.DiningDate)
	if err != nil {
		return nil, err
	}
	if hasOrder {
		return nil, ErrDuplicateBooking
	}

	count, err := database.CountOrdersByDateAndSlot(
		req.DiningDate,
		req.TimeSlot,
		models.StatusCancelled,
		models.StatusNoShow,
	)
	if err != nil {
		return nil, err
	}
	if count >= MaxTablesPerSlot {
		return nil, ErrSlotFull
	}

	order := &models.Order{
		Phone:        req.Phone,
		DiningDate:   req.DiningDate,
		TimeSlot:     req.TimeSlot,
		GuestCount:   req.GuestCount,
		Status:       models.StatusPending,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
	}

	items := make([]models.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = models.OrderItem{
			DishName: item.DishName,
			Quantity: item.Quantity,
		}
	}

	orderID, err := database.CreateOrder(order, items)
	if err != nil {
		return nil, err
	}

	return database.GetOrderWithItems(orderID)
}

func CancelOrder(orderID int64) error {
	order, err := database.GetOrderByID(orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderNotFound
		}
		return err
	}

	if order.Status == models.StatusCancelled {
		return ErrAlreadyCancelled
	}

	diningStart, err := GetSlotStartTime(order.DiningDate, order.TimeSlot)
	if err != nil {
		return err
	}

	if time.Now().Add(time.Hour * time.Duration(CancellationHours)).After(diningStart) {
		return ErrTooLateToCancel
	}

	return database.UpdateOrderStatus(orderID, models.StatusCancelled)
}

func GetSlotStartTime(date string, slot models.TimeSlot) (time.Time, error) {
	layout := "2006-01-02 15:04"
	var timeStr string
	if slot == models.Lunch {
		timeStr = date + " 11:00"
	} else {
		timeStr = date + " 17:00"
	}
	return time.Parse(layout, timeStr)
}

func UpdateOrderStatus(orderID int64, newStatus models.OrderStatus) error {
	order, err := database.GetOrderByID(orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrOrderNotFound
		}
		return err
	}

	if order.Status == models.StatusCancelled || order.Status == models.StatusNoShow {
		return ErrInvalidStatus
	}

	if !IsValidStatusTransition(order.Status, newStatus) {
		return ErrInvalidStatusTransition
	}

	return database.UpdateOrderStatus(orderID, newStatus)
}

func IsValidStatusTransition(current, next models.OrderStatus) bool {
	currentIdx := -1
	nextIdx := -1

	for i, status := range models.StatusSequence {
		if status == current {
			currentIdx = i
		}
		if status == next {
			nextIdx = i
		}
	}

	if currentIdx == -1 || nextIdx == -1 {
		return false
	}

	return nextIdx == currentIdx+1
}

func CheckNoShows() error {
	now := time.Now()
	gracePeriod := time.Minute * time.Duration(NoShowGraceMinutes)

	orders, err := database.GetOrdersForNoShowCheck(now, gracePeriod)
	if err != nil {
		return err
	}

	for _, order := range orders {
		err := database.UpdateOrderStatus(order.ID, models.StatusNoShow)
		if err != nil {
			fmt.Printf("Failed to update order %d to no_show: %v\n", order.ID, err)
			continue
		}

		noShowCount, err := database.CountNoShows(order.Phone)
		if err != nil {
			fmt.Printf("Failed to count no_shows for %s: %v\n", order.Phone, err)
			continue
		}

		if noShowCount >= MaxNoShows {
			endTime := time.Now().AddDate(0, 0, BlacklistDays)
			err = database.AddToBlacklist(order.Phone, endTime)
			if err != nil {
				fmt.Printf("Failed to add %s to blacklist: %v\n", order.Phone, err)
			}
		}
	}

	return nil
}

func GetKitchenSummary(date string) ([]models.KitchenSummary, error) {
	lunch, dinner, err := database.GetKitchenSummary(date)
	if err != nil {
		return nil, err
	}

	return []models.KitchenSummary{*lunch, *dinner}, nil
}

func GetResourceSummary(resourceType models.ResourceType, resourceID string) (*models.ResourceSummary, error) {
	return database.GetResourceSummary(resourceType, resourceID)
}

func GetOrder(orderID int64) (*models.OrderWithItems, error) {
	return database.GetOrderWithItems(orderID)
}

func GetAllOrders() ([]models.Order, error) {
	return database.GetAllOrders()
}

type CreateOrderRequest struct {
	Phone        string              `json:"phone"`
	DiningDate   string              `json:"dining_date"`
	TimeSlot     models.TimeSlot     `json:"time_slot"`
	GuestCount   int                 `json:"guest_count"`
	ResourceType models.ResourceType `json:"resource_type,omitempty"`
	ResourceID   string              `json:"resource_id,omitempty"`
	Items        []OrderItemRequest  `json:"items"`
}

type OrderItemRequest struct {
	DishName string `json:"dish_name"`
	Quantity int    `json:"quantity"`
}

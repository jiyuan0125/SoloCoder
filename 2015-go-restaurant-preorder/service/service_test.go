package service

import (
	"fmt"
	"os"
	"restaurant-preorder/database"
	"restaurant-preorder/models"
	"testing"
)

func setupTestDB(t *testing.T) {
	err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
}

func TestValidateCreateOrder(t *testing.T) {
	req := CreateOrderRequest{
		Phone:      "13800138000",
		DiningDate: "2026-05-14",
		TimeSlot:   models.TimeSlot("invalid"),
		GuestCount: 4,
		Items: []OrderItemRequest{
			{DishName: "宫保鸡丁", Quantity: 1},
		},
	}

	err := ValidateCreateOrder(req)
	if err != ErrInvalidTimeSlot {
		t.Errorf("Expected ErrInvalidTimeSlot, got %v", err)
	}

	req.TimeSlot = models.Lunch
	req.GuestCount = 0
	err = ValidateCreateOrder(req)
	if err != ErrInvalidGuestCount {
		t.Errorf("Expected ErrInvalidGuestCount for 0 guests, got %v", err)
	}

	req.GuestCount = 25
	err = ValidateCreateOrder(req)
	if err != ErrInvalidGuestCount {
		t.Errorf("Expected ErrInvalidGuestCount for 25 guests, got %v", err)
	}

	req.GuestCount = 10
	err = ValidateCreateOrder(req)
	if err != nil {
		t.Errorf("Expected no error for valid order, got %v", err)
	}
}

func TestIsValidStatusTransition(t *testing.T) {
	if IsValidStatusTransition(models.StatusPending, models.StatusConfirmed) != true {
		t.Error("pending -> confirmed should be valid")
	}

	if IsValidStatusTransition(models.StatusConfirmed, models.StatusDining) != true {
		t.Error("confirmed -> dining should be valid")
	}

	if IsValidStatusTransition(models.StatusDining, models.StatusCompleted) != true {
		t.Error("dining -> completed should be valid")
	}

	if IsValidStatusTransition(models.StatusPending, models.StatusDining) != false {
		t.Error("pending -> dining should be invalid (skip step)")
	}

	if IsValidStatusTransition(models.StatusCompleted, models.StatusPending) != false {
		t.Error("completed -> pending should be invalid")
	}
}

func TestGetSlotStartTime(t *testing.T) {
	lunchStart, err := GetSlotStartTime("2026-05-14", models.Lunch)
	if err != nil {
		t.Errorf("Failed to get lunch start time: %v", err)
	}

	if lunchStart.Hour() != 11 || lunchStart.Minute() != 0 {
		t.Errorf("Expected lunch start at 11:00, got %02d:%02d", lunchStart.Hour(), lunchStart.Minute())
	}

	dinnerStart, err := GetSlotStartTime("2026-05-14", models.Dinner)
	if err != nil {
		t.Errorf("Failed to get dinner start time: %v", err)
	}

	if dinnerStart.Hour() != 17 || dinnerStart.Minute() != 0 {
		t.Errorf("Expected dinner start at 17:00, got %02d:%02d", dinnerStart.Hour(), dinnerStart.Minute())
	}
}

func TestCreateOrderBasic(t *testing.T) {
	setupTestDB(t)
	defer os.Remove("test.db")

	req := CreateOrderRequest{
		Phone:        "13800138000",
		DiningDate:   "2026-05-20",
		TimeSlot:     models.Lunch,
		GuestCount:   4,
		ResourceType: models.ResourceTable,
		ResourceID:   "table-01",
		Items: []OrderItemRequest{
			{DishName: "宫保鸡丁", Quantity: 2},
			{DishName: "米饭", Quantity: 4},
		},
	}

	order, err := CreateOrder(req)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	if order.ID == 0 {
		t.Error("Order ID should not be 0")
	}

	if order.Phone != "13800138000" {
		t.Errorf("Expected phone 13800138000, got %s", order.Phone)
	}

	if order.Status != models.StatusPending {
		t.Errorf("Expected status pending, got %s", order.Status)
	}

	if len(order.Items) != 2 {
		t.Errorf("Expected 2 order items, got %d", len(order.Items))
	}
}

func TestDuplicateBooking(t *testing.T) {
	setupTestDB(t)

	req := CreateOrderRequest{
		Phone:      "13800138001",
		DiningDate: "2026-05-21",
		TimeSlot:   models.Lunch,
		GuestCount: 4,
		Items: []OrderItemRequest{
			{DishName: "鱼香肉丝", Quantity: 1},
		},
	}

	_, err := CreateOrder(req)
	if err != nil {
		t.Fatalf("First order should succeed: %v", err)
	}

	req.TimeSlot = models.Dinner
	_, err = CreateOrder(req)
	if err != ErrDuplicateBooking {
		t.Errorf("Expected ErrDuplicateBooking for same phone on same day, got %v", err)
	}
}

func TestMaxTablesPerSlot(t *testing.T) {
	setupTestDB(t)

	for i := 0; i < MaxTablesPerSlot; i++ {
		req := CreateOrderRequest{
			Phone:      formatPhone(13800138100 + i),
			DiningDate: "2026-05-22",
			TimeSlot:   models.Lunch,
			GuestCount: 2,
			Items: []OrderItemRequest{
				{DishName: "测试菜品", Quantity: 1},
			},
		}

		_, err := CreateOrder(req)
		if err != nil {
			t.Fatalf("Order %d should succeed: %v", i+1, err)
		}
	}

	req := CreateOrderRequest{
		Phone:      "13900000000",
		DiningDate: "2026-05-22",
		TimeSlot:   models.Lunch,
		GuestCount: 2,
		Items: []OrderItemRequest{
			{DishName: "测试菜品", Quantity: 1},
		},
	}

	_, err := CreateOrder(req)
	if err != ErrSlotFull {
		t.Errorf("Expected ErrSlotFull when exceeding table limit, got %v", err)
	}
}

func formatPhone(num int) string {
	return fmt.Sprintf("138%04d0000", num)
}

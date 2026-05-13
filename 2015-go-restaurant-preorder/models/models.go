package models

import "time"

type TimeSlot string

const (
	Lunch  TimeSlot = "lunch"
	Dinner TimeSlot = "dinner"
)

type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusConfirmed  OrderStatus = "confirmed"
	StatusDining     OrderStatus = "dining"
	StatusCompleted  OrderStatus = "completed"
	StatusCancelled  OrderStatus = "cancelled"
	StatusNoShow     OrderStatus = "no_show"
)

var StatusSequence = []OrderStatus{
	StatusPending,
	StatusConfirmed,
	StatusDining,
	StatusCompleted,
}

type ResourceType string

const (
	ResourceTable    ResourceType = "table"
	ResourceKitchen  ResourceType = "kitchen"
	ResourceCustomer ResourceType = "customer"
)

type Order struct {
	ID           int64       `json:"id"`
	Phone        string      `json:"phone"`
	DiningDate   string      `json:"dining_date"`
	TimeSlot     TimeSlot    `json:"time_slot"`
	GuestCount   int         `json:"guest_count"`
	Status       OrderStatus `json:"status"`
	ResourceType ResourceType `json:"resource_type"`
	ResourceID   string      `json:"resource_id"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        int64     `json:"id"`
	OrderID   int64     `json:"order_id"`
	DishName  string    `json:"dish_name"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

type Blacklist struct {
	ID        int64     `json:"id"`
	Phone     string    `json:"phone"`
	EndTime   time.Time `json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderWithItems struct {
	Order
	Items []OrderItem `json:"items"`
}

type KitchenSummary struct {
	Date      string              `json:"date"`
	TimeSlot  TimeSlot            `json:"time_slot"`
	Dishes    map[string]int      `json:"dishes"`
	GeneratedAt time.Time         `json:"generated_at"`
}

type ResourceSummary struct {
	ResourceType ResourceType  `json:"resource_type"`
	ResourceID   string        `json:"resource_id"`
	TotalOrders  int           `json:"total_orders"`
	Orders       []Order       `json:"orders"`
}

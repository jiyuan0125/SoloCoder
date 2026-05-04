package common

import "time"

type CarStatus string

const (
	CarStatusAvailable CarStatus = "available"
	CarStatusMaintenance CarStatus = "maintenance"
)

type OrderStatus string

const (
	OrderStatusReserved OrderStatus = "reserved"
	OrderStatusPickedUp OrderStatus = "picked_up"
	OrderStatusReturned OrderStatus = "returned"
)

type Car struct {
	ID          string    `json:"id"`
	Brand       string    `json:"brand"`
	Model       string    `json:"model"`
	DailyRate   float64   `json:"daily_rate"`
	Deposit     float64   `json:"deposit"`
	Status      CarStatus `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Order struct {
	ID                  string      `json:"id"`
	CarID               string      `json:"car_id"`
	UserName            string      `json:"user_name"`
	UserPhone           string      `json:"user_phone"`
	PickupDate          time.Time   `json:"pickup_date"`
	ReturnDate          time.Time   `json:"return_date"`
	ActualReturnDate    *time.Time  `json:"actual_return_date,omitempty"`
	OriginalTotalCost   float64     `json:"original_total_cost"`
	TotalCost           float64     `json:"total_cost"`
	Deposit             float64     `json:"deposit"`
	DepositDeduction    float64     `json:"deposit_deduction"`
	Mileage             float64     `json:"mileage"`
	Condition           string      `json:"condition"`
	ViolationOrDamage   bool        `json:"violation_or_damage"`
	ViolationAmount     float64     `json:"violation_amount"`
	Status              OrderStatus `json:"status"`
	CreatedAt           time.Time   `json:"created_at"`
}

type CreateCarRequest struct {
	Brand     string  `json:"brand"`
	Model     string  `json:"model"`
	DailyRate float64 `json:"daily_rate"`
	Deposit   float64 `json:"deposit"`
}

type CreateCarResponse struct {
	Car *Car   `json:"car,omitempty"`
	Error string `json:"error,omitempty"`
}

type ListCarsResponse struct {
	Cars  []*Car `json:"cars,omitempty"`
	Error string `json:"error,omitempty"`
}

type CreateOrderRequest struct {
	CarID      string    `json:"car_id"`
	UserName   string    `json:"user_name"`
	UserPhone  string    `json:"user_phone"`
	PickupDate time.Time `json:"pickup_date"`
	ReturnDate time.Time `json:"return_date"`
}

type CreateOrderResponse struct {
	Order *Order `json:"order,omitempty"`
	Error string `json:"error,omitempty"`
}

type GetOrderRequest struct {
	OrderID string `json:"order_id"`
}

type GetOrderResponse struct {
	Order *Order `json:"order,omitempty"`
	Error string `json:"error,omitempty"`
}

type ListOrdersRequest struct {
	UserPhone string `json:"user_phone,omitempty"`
}

type ListOrdersResponse struct {
	Orders []*Order `json:"orders,omitempty"`
	Error  string   `json:"error,omitempty"`
}

type ReturnCarRequest struct {
	OrderID            string    `json:"order_id"`
	ActualReturnDate   time.Time `json:"actual_return_date"`
	Mileage            float64   `json:"mileage"`
	Condition          string    `json:"condition"`
	ViolationOrDamage  bool      `json:"violation_or_damage"`
	ViolationAmount    float64   `json:"violation_amount"`
}

type ReturnCarResponse struct {
	Order *Order `json:"order,omitempty"`
	Error string `json:"error,omitempty"`
}

type CheckAvailabilityRequest struct {
	CarID      string    `json:"car_id"`
	PickupDate time.Time `json:"pickup_date"`
	ReturnDate time.Time `json:"return_date"`
}

type CheckAvailabilityResponse struct {
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

type UpdateCarStatusRequest struct {
	CarID  string    `json:"car_id"`
	Status CarStatus `json:"status"`
}

type UpdateCarStatusResponse struct {
	Car   *Car   `json:"car,omitempty"`
	Error string `json:"error,omitempty"`
}

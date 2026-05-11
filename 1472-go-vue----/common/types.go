package common

type VehicleStatus string

const (
	VehicleStatusIdle       VehicleStatus = "idle"
	VehicleStatusOccupied   VehicleStatus = "occupied"
	VehicleStatusReserved   VehicleStatus = "reserved"
	VehicleStatusRest       VehicleStatus = "rest"
	VehicleStatusMaintenance VehicleStatus = "maintenance"
)

type Vehicle struct {
	PlateNumber     string        `json:"plate_number"`
	DriverName      string        `json:"driver_name"`
	DriverPhone     string        `json:"driver_phone"`
	Status          VehicleStatus `json:"status"`
	Latitude        float64       `json:"latitude"`
	Longitude       float64       `json:"longitude"`
}

type OrderStatus string

const (
	OrderStatusPending       OrderStatus = "pending"
	OrderStatusDispatched    OrderStatus = "dispatched"
	OrderStatusAccepted      OrderStatus = "accepted"
	OrderStatusInProgress    OrderStatus = "in_progress"
	OrderStatusCompleted     OrderStatus = "completed"
	OrderStatusNoVehicle     OrderStatus = "no_vehicle"
	OrderStatusCancelled     OrderStatus = "cancelled"
)

type Order struct {
	OrderID              string      `json:"order_id"`
	PassengerID          string      `json:"passenger_id"`
	PassengerPhone       string      `json:"passenger_phone"`
	PickupLocation       Location    `json:"pickup_location"`
	DestLocation         Location    `json:"dest_location"`
	Status               OrderStatus `json:"status"`
	VehiclePlate         string      `json:"vehicle_plate,omitempty"`
	EstimatedDistance    float64     `json:"estimated_distance"`
	EstimatedDuration    int64       `json:"estimated_duration"`
	ActualDistance       float64     `json:"actual_distance,omitempty"`
	ActualDuration       int64       `json:"actual_duration,omitempty"`
	LowSpeedMinutes      int         `json:"low_speed_minutes,omitempty"`
	EstimatedFare        float64     `json:"estimated_fare"`
	ActualFare           float64     `json:"actual_fare,omitempty"`
	IsLargeOrder         bool        `json:"is_large_order,omitempty"`
	CreateTime           int64       `json:"create_time"`
	DispatchAttempts     int         `json:"dispatch_attempts,omitempty"`
	AcceptTime           int64       `json:"accept_time,omitempty"`
	StartTime            int64       `json:"start_time,omitempty"`
	EndTime              int64       `json:"end_time,omitempty"`
	Rating               *Rating     `json:"rating,omitempty"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
}

type Rating struct {
	OrderID     string `json:"order_id"`
	Stars       int    `json:"stars"`
	Content     string `json:"content,omitempty"`
	CreateTime  int64  `json:"create_time"`
	HasComplaint bool   `json:"has_complaint"`
}

type Complaint struct {
	ComplaintID string `json:"complaint_id"`
	OrderID     string `json:"order_id"`
	Content     string `json:"content"`
	CreateTime  int64  `json:"create_time"`
	Handled     bool   `json:"handled"`
}

type Metrics struct {
	TodayOrderCount    int     `json:"today_order_count"`
	AvgAcceptTime      float64 `json:"avg_accept_time_seconds"`
	AvgTripFare        float64 `json:"avg_trip_fare"`
	GoodReviewRate     float64 `json:"good_review_rate"`
}

type CreateVehicleRequest struct {
	PlateNumber string  `json:"plate_number"`
	DriverName  string  `json:"driver_name"`
	DriverPhone string  `json:"driver_phone"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

type UpdateVehicleStatusRequest struct {
	Status VehicleStatus `json:"status"`
}

type UpdateVehicleLocationRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type EstimateFareRequest struct {
	PickupLocation Location `json:"pickup_location"`
	DestLocation   Location `json:"dest_location"`
}

type CreateOrderRequest struct {
	PassengerID    string   `json:"passenger_id"`
	PassengerPhone string   `json:"passenger_phone"`
	PickupLocation Location `json:"pickup_location"`
	DestLocation   Location `json:"dest_location"`
}

type AcceptOrderRequest struct {
	PlateNumber string `json:"plate_number"`
}

type StartTripRequest struct {
	ActualDistance float64 `json:"actual_distance"`
	ActualDuration int64   `json:"actual_duration"`
	LowSpeedMinutes int    `json:"low_speed_minutes"`
}

type SubmitRatingRequest struct {
	Stars   int    `json:"stars"`
	Content string `json:"content,omitempty"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

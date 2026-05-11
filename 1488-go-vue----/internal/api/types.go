package api

import "time"

type ServiceCategory string

const (
	ServiceDailyCleaning   ServiceCategory = "日常保洁"
	ServiceDeepCleaning    ServiceCategory = "深度清洁"
	ServiceNanny           ServiceCategory = "月嫂"
	ServiceChildcare       ServiceCategory = "育婴师"
	ServiceElderlyCare     ServiceCategory = "老人陪护"
)

type BookingStatus string

const (
	BookingStatusPending      BookingStatus = "待匹配"
	BookingStatusMatched      BookingStatus = "已匹配"
	BookingStatusInProgress   BookingStatus = "服务中"
	BookingStatusCompleted    BookingStatus = "已完成"
	BookingStatusCancelled    BookingStatus = "已取消"
)

type Aunt struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Age           int             `json:"age"`
	Phone         string          `json:"phone"`
	ServiceCategory ServiceCategory `json:"service_category"`
	YearsOfService int            `json:"years_of_service"`
	ServiceArea   string          `json:"service_area"`
	Rating        float64         `json:"rating"`
}

type Customer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type Review struct {
	ID         string    `json:"id"`
	AuntID     string    `json:"aunt_id"`
	BookingID  string    `json:"booking_id"`
	CustomerID string    `json:"customer_id"`
	Rating     float64   `json:"rating"`
	Comment    string    `json:"comment"`
	Time       time.Time `json:"time"`
}

type Booking struct {
	ID            string          `json:"id"`
	CustomerID    string          `json:"customer_id"`
	AuntID        string          `json:"aunt_id"`
	ServiceCategory ServiceCategory `json:"service_category"`
	ServiceDate   time.Time       `json:"service_date"`
	EstimatedDuration float64     `json:"estimated_duration"`
	ActualDuration   float64      `json:"actual_duration,omitempty"`
	Address       string          `json:"address"`
	Status        BookingStatus   `json:"status"`
	StartTime     time.Time       `json:"start_time,omitempty"`
	EndTime       time.Time       `json:"end_time,omitempty"`
	Price         int             `json:"price,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

type MonthlyStats struct {
	AuntID        string  `json:"aunt_id"`
	Month         string  `json:"month"`
	TotalBookings int     `json:"total_bookings"`
	TotalHours    float64 `json:"total_hours"`
	AverageRating float64 `json:"average_rating"`
	RepeatRate    float64 `json:"repeat_rate"`
}

type RegisterAuntRequest struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Age           int             `json:"age"`
	Phone         string          `json:"phone"`
	ServiceCategory ServiceCategory `json:"service_category"`
	YearsOfService int            `json:"years_of_service"`
	ServiceArea   string          `json:"service_area"`
}

type RegisterCustomerRequest struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type CreateBookingRequest struct {
	CustomerID       string          `json:"customer_id"`
	ServiceCategory  ServiceCategory `json:"service_category"`
	ServiceDate      time.Time       `json:"service_date"`
	EstimatedDuration float64        `json:"estimated_duration"`
	Address          string          `json:"address"`
}

type StartServiceRequest struct {
	BookingID string `json:"booking_id"`
}

type CompleteServiceRequest struct {
	BookingID      string  `json:"booking_id"`
	ActualDuration float64 `json:"actual_duration"`
}

type AddReviewRequest struct {
	AuntID     string  `json:"aunt_id"`
	BookingID  string  `json:"booking_id"`
	CustomerID string  `json:"customer_id"`
	Rating     float64 `json:"rating"`
	Comment    string  `json:"comment"`
}

type GetStatsRequest struct {
	AuntID string `json:"aunt_id"`
	Month  string `json:"month"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

func NewErrorResponse(err string) Response {
	return Response{
		Success: false,
		Error:   err,
	}
}

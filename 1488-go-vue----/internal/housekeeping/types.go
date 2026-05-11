package housekeeping

import (
	"sync"
	"time"
)

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
	ID            string
	Name          string
	Age           int
	Phone         string
	ServiceCategory ServiceCategory
	YearsOfService int
	ServiceArea   string
	Rating        float64
	Reviews       []Review
	mu            sync.RWMutex
}

type Customer struct {
	ID      string
	Name    string
	Phone   string
	Bookings []string
}

type Review struct {
	ID        string
	AuntID    string
	BookingID string
	CustomerID string
	Rating    float64
	Comment   string
	Time      time.Time
}

type Booking struct {
	ID            string
	CustomerID    string
	AuntID        string
	ServiceCategory ServiceCategory
	ServiceDate   time.Time
	EstimatedDuration float64
	ActualDuration   float64
	Address       string
	Status        BookingStatus
	StartTime     time.Time
	EndTime       time.Time
	Price         int
	CreatedAt     time.Time
	mu            sync.RWMutex
}

type BookingRequest struct {
	CustomerID       string
	ServiceCategory  ServiceCategory
	ServiceDate      time.Time
	EstimatedDuration float64
	Address          string
}

type MonthlyStats struct {
	AuntID          string
	Month           string
	TotalBookings   int
	TotalHours      float64
	AverageRating   float64
	RepeatRate      float64
}

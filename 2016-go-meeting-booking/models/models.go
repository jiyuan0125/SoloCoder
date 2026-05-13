package models

import (
	"time"
	"gorm.io/gorm"
)

type Room struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"uniqueIndex;not null" json:"name"`
	Capacity  int            `gorm:"not null" json:"capacity"`
	Equipment string         `json:"equipment"`
	Floor     int            `gorm:"not null" json:"floor"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Booking struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	RoomID          uint           `gorm:"not null;index:idx_room_time" json:"room_id"`
	Room            Room           `gorm:"foreignKey:RoomID" json:"-"`
	EmployeeID      string         `gorm:"not null;index" json:"employee_id"`
	Title           string         `json:"title"`
	Date            string         `gorm:"not null;index:idx_room_time;index:idx_emp_date" json:"date"`
	StartTime       string         `gorm:"not null" json:"start_time"`
	EndTime         string         `gorm:"not null" json:"end_time"`
	DurationMinutes int            `gorm:"not null" json:"duration_minutes"`
	IsRecurring     bool           `gorm:"not null;default:false" json:"is_recurring"`
	RecurringEndDate string        `gorm:"" json:"recurring_end_date"`
	RecurringWeekDay time.Weekday  `gorm:"" json:"recurring_week_day"`
	Cancelled       bool           `gorm:"not null;default:false" json:"cancelled"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type SkippedWeek struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	BookingID uint           `gorm:"not null;index" json:"booking_id"`
	Date      string         `gorm:"not null" json:"date"`
	Reason    string         `json:"reason"`
	CreatedAt time.Time      `json:"-"`
}

type TimeSlot struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type RoomOccupancy struct {
	RoomID    uint        `json:"room_id"`
	RoomName  string      `json:"room_name"`
	Occupied  []TimeSlot  `json:"occupied"`
	Available []TimeSlot  `json:"available"`
}

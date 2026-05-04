package types

import (
	"encoding/json"
	"time"
)

type CardType string

const (
	MonthlyCard   CardType = "monthly"
	QuarterlyCard CardType = "quarterly"
	YearlyCard    CardType = "yearly"
)

type CardStatus string

const (
	CardActive    CardStatus = "active"
	CardExpired   CardStatus = "expired"
	CardCancelled CardStatus = "cancelled"
)

type ClassStatus string

const (
	ClassScheduled ClassStatus = "scheduled"
	ClassCancelled ClassStatus = "cancelled"
	ClassCompleted ClassStatus = "completed"
)

type BookingStatus string

const (
	BookingConfirmed BookingStatus = "confirmed"
	BookingCancelled BookingStatus = "cancelled"
	BookingNoShow    BookingStatus = "no_show"
)

type Member struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Phone       string     `json:"phone"`
	Password    string     `json:"password,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CurrentCard *MembershipCard `json:"current_card,omitempty"`
}

type MembershipCard struct {
	ID        string     `json:"id"`
	MemberID  string     `json:"member_id"`
	Type      CardType   `json:"type"`
	StartDate time.Time  `json:"start_date"`
	EndDate   time.Time  `json:"end_date"`
	Price     float64    `json:"price"`
	Status    CardStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

type Class struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Weekday      time.Weekday `json:"weekday"`
	StartTime    string      `json:"start_time"`
	Duration     int         `json:"duration"`
	MaxCapacity  int         `json:"max_capacity"`
	MinCapacity  int         `json:"min_capacity"`
	Instructor   string      `json:"instructor"`
	Location     string      `json:"location"`
	CreatedAt    time.Time   `json:"created_at"`
}

type ClassInstance struct {
	ID        string      `json:"id"`
	ClassID   string      `json:"class_id"`
	Date      time.Time   `json:"date"`
	Status    ClassStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

type Booking struct {
	ID             string         `json:"id"`
	MemberID       string         `json:"member_id"`
	ClassInstanceID string         `json:"class_instance_id"`
	Status         BookingStatus `json:"status"`
	BookedAt       time.Time      `json:"booked_at"`
	CancelledAt    *time.Time     `json:"cancelled_at,omitempty"`
}

type Attendance struct {
	ID              string    `json:"id"`
	MemberID        string    `json:"member_id"`
	ClassInstanceID string    `json:"class_instance_id"`
	CheckInTime     time.Time `json:"check_in_time"`
	CreatedAt       time.Time `json:"created_at"`
}

type CardPrice struct {
	ID        string    `json:"id"`
	CardType  CardType  `json:"card_type"`
	Price     float64   `json:"price"`
	ValidDays int       `json:"valid_days"`
	UpdatedAt time.Time `json:"updated_at"`
}

// API请求/响应结构

type CreateMemberRequest struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type CreateMemberResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Member  *Member `json:"member,omitempty"`
}

type LoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Member  *Member `json:"member,omitempty"`
	Token   string `json:"token,omitempty"`
}

type PurchaseCardRequest struct {
	MemberID string   `json:"member_id"`
	CardType CardType `json:"card_type"`
}

type PurchaseCardResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Card    *MembershipCard `json:"card,omitempty"`
}

type RenewCardRequest struct {
	MemberID string   `json:"member_id"`
	CardType CardType `json:"card_type"`
}

type RenewCardResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Card    *MembershipCard `json:"card,omitempty"`
}

type GetMemberInfoRequest struct {
	MemberID string `json:"member_id"`
}

type GetMemberInfoResponse struct {
	Success          bool              `json:"success"`
	Message          string            `json:"message,omitempty"`
	Member           *Member           `json:"member,omitempty"`
	CurrentCard      *MembershipCard   `json:"current_card,omitempty"`
	ActiveBookings   []BookingDetail   `json:"active_bookings,omitempty"`
	AttendanceHistory []AttendanceDetail `json:"attendance_history,omitempty"`
}

type CreateClassRequest struct {
	Name        string      `json:"name"`
	Weekday     time.Weekday `json:"weekday"`
	StartTime   string      `json:"start_time"`
	Duration    int         `json:"duration"`
	MaxCapacity int         `json:"max_capacity"`
	MinCapacity int         `json:"min_capacity"`
	Instructor  string      `json:"instructor"`
	Location    string      `json:"location"`
}

type CreateClassResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Class   *Class `json:"class,omitempty"`
}

type ListClassesRequest struct {
	Weekday *time.Weekday `json:"weekday,omitempty"`
}

type ListClassesResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message,omitempty"`
	Classes []Class `json:"classes,omitempty"`
}

type BookClassRequest struct {
	MemberID    string `json:"member_id"`
	ClassID     string `json:"class_id"`
	Date        string `json:"date"`
}

type BookClassResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message,omitempty"`
	Booking *Booking `json:"booking,omitempty"`
}

type CancelBookingRequest struct {
	BookingID string `json:"booking_id"`
	MemberID  string `json:"member_id"`
}

type CancelBookingResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type CheckInRequest struct {
	MemberID  string `json:"member_id"`
	ClassID   string `json:"class_id"`
	Date      string `json:"date"`
}

type CheckInResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Attendance *Attendance `json:"attendance,omitempty"`
}

type GetClassBookingsRequest struct {
	ClassID  string `json:"class_id"`
	Date     string `json:"date"`
}

type GetClassBookingsResponse struct {
	Success         bool             `json:"success"`
	Message         string           `json:"message,omitempty"`
	Class           *Class           `json:"class,omitempty"`
	ClassInstance   *ClassInstance   `json:"class_instance,omitempty"`
	Bookings        []BookingDetail  `json:"bookings,omitempty"`
	AttendanceCount int              `json:"attendance_count,omitempty"`
}

type BookingDetail struct {
	Booking
	MemberName  string `json:"member_name"`
	MemberPhone string `json:"member_phone"`
	ClassName   string `json:"class_name"`
	Date        string `json:"date"`
}

type AttendanceDetail struct {
	Attendance
	MemberName  string `json:"member_name"`
	MemberPhone string `json:"member_phone"`
	ClassName   string `json:"class_name"`
	Date        string `json:"date"`
}

type SetCardPriceRequest struct {
	CardType  CardType `json:"card_type"`
	Price     float64  `json:"price"`
	ValidDays int      `json:"valid_days"`
}

type SetCardPriceResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Price   *CardPrice `json:"price,omitempty"`
}

type ListCardPricesRequest struct {
}

type ListCardPricesResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Prices  []CardPrice `json:"prices,omitempty"`
}

type GenerateClassInstancesRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type GenerateClassInstancesResponse struct {
	Success   bool           `json:"success"`
	Message   string         `json:"message,omitempty"`
	Instances []ClassInstance `json:"instances,omitempty"`
}

// 辅助函数

func (ct CardType) String() string {
	switch ct {
	case MonthlyCard:
		return "月卡"
	case QuarterlyCard:
		return "季卡"
	case YearlyCard:
		return "年卡"
	default:
		return string(ct)
	}
}

func WeekdayToString(w time.Weekday) string {
	weekdays := []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}
	if w >= 0 && w <= 6 {
		return weekdays[w]
	}
	return "未知"
}

// JSON序列化辅助
func (m *Member) MarshalJSON() ([]byte, error) {
	type MemberAlias Member
	return json.Marshal(&struct {
		*MemberAlias
		Password string `json:"password,omitempty"`
	}{
		MemberAlias: (*MemberAlias)(m),
		Password:    "",
	})
}

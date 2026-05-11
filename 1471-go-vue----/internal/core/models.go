package core

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type ScheduleStatus string

const (
	ScheduleStatusNotDeparted ScheduleStatus = "not_departed"
	ScheduleStatusDeparted    ScheduleStatus = "departed"
	ScheduleStatusArrived     ScheduleStatus = "arrived"
	ScheduleStatusCancelled   ScheduleStatus = "cancelled"
)

type TicketStatus string

const (
	TicketStatusSold       TicketStatus = "sold"
	TicketStatusChecked    TicketStatus = "checked"
	TicketStatusNotBoarded TicketStatus = "not_boarded"
	TicketStatusRefunded   TicketStatus = "refunded"
)

type RefundStatus string

const (
	RefundStatusPending RefundStatus = "pending"
	RefundStatusSuccess RefundStatus = "success"
	RefundStatusReject  RefundStatus = "rejected"
)

type ScheduleTemplate struct {
	ScheduleNo      string
	DepartureStation string
	ArrivalStation   string
	DepartureTime    string
	ArrivalTime      string
	BusType          string
	Price            int
	TotalSeats       int
}

type Schedule struct {
	ScheduleNo        string
	Date              string
	DepartureStation  string
	ArrivalStation    string
	DepartureTime     string
	ArrivalTime       string
	BusType           string
	Price             int
	TotalSeats        int
	SoldSeats         int
	Status            ScheduleStatus
	IsAlmostSoldOut   bool
	Archived          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Seat struct {
	SeatNo   int
	IsSold   bool
	TicketID string
}

type Ticket struct {
	TicketNo          string
	ScheduleNo        string
	Date              string
	SeatNo            int
	PassengerName     string
	Price             int
	Status            TicketStatus
	CheckedAt         *time.Time
	SoldAt            time.Time
	CancelledRefundAt *time.Time
}

type RefundRequest struct {
	ID             string
	TicketNo       string
	Status         RefundStatus
	RefundAmount   int
	RequestTime    time.Time
	ProcessedAt    *time.Time
	RejectReason   string
}

type CheckInRecord struct {
	ID          string
	TicketNo    string
	ScheduleNo  string
	Date        string
	SeatNo      int
	CheckedAt   time.Time
}

var (
	timeRegex = regexp.MustCompile(`^([01]?\d|2[0-3]):([0-5]\d)$`)
)

func ValidateTime(timeStr string) error {
	if !timeRegex.MatchString(timeStr) {
		return errors.New("时间格式错误，应为 HH:MM")
	}
	return nil
}

func CompareTime(time1, time2 string) (int, error) {
	if err := ValidateTime(time1); err != nil {
		return 0, err
	}
	if err := ValidateTime(time2); err != nil {
		return 0, err
	}
	
	h1, _ := strconv.Atoi(time1[:2])
	m1, _ := strconv.Atoi(time1[3:])
	h2, _ := strconv.Atoi(time2[:2])
	m2, _ := strconv.Atoi(time2[3:])
	
	t1 := h1*60 + m1
	t2 := h2*60 + m2
	
	if t1 < t2 {
		return -1, nil
	} else if t1 > t2 {
		return 1, nil
	}
	return 0, nil
}

func ValidateDepartureArrival(departure, arrival string) error {
	cmp, err := CompareTime(departure, arrival)
	if err != nil {
		return err
	}
	if cmp >= 0 {
		return errors.New("出发时间必须早于到达时间")
	}
	return nil
}

type Store struct {
	mu              sync.RWMutex
	schedules       map[string]*Schedule
	tickets         map[string]*Ticket
	refundRequests  map[string]*RefundRequest
	checkInRecords  map[string]*CheckInRecord
	scheduleSeats   map[string][]*Seat
	scheduleTemplates []*ScheduleTemplate
	ticketCounter   int64
	refundCounter   int64
	checkInCounter  int64
}

func NewStore() *Store {
	return &Store{
		schedules:       make(map[string]*Schedule),
		tickets:         make(map[string]*Ticket),
		refundRequests:  make(map[string]*RefundRequest),
		checkInRecords:  make(map[string]*CheckInRecord),
		scheduleSeats:   make(map[string][]*Seat),
		scheduleTemplates: make([]*ScheduleTemplate, 0),
	}
}

func (s *Store) GetScheduleKey(scheduleNo, date string) string {
	return fmt.Sprintf("%s_%s", scheduleNo, date)
}

func (s *Store) generateTicketNo() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ticketCounter++
	return fmt.Sprintf("TK%08d", s.ticketCounter)
}

func (s *Store) generateRefundID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refundCounter++
	return fmt.Sprintf("RF%08d", s.refundCounter)
}

func (s *Store) generateCheckInID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkInCounter++
	return fmt.Sprintf("CK%08d", s.checkInCounter)
}

var (
	ErrScheduleNotFound     = errors.New("班次不存在")
	ErrScheduleAlreadyExist = errors.New("班次已存在")
	ErrInvalidTime          = errors.New("时间格式错误")
	ErrDepartureAfterArrival= errors.New("出发时间必须早于到达时间")
	ErrScheduleNotModifiable= errors.New("班次状态不允许修改")
	ErrSeatAlreadySold      = errors.New("座位已售出")
	ErrSeatOutOfRange       = errors.New("座位号超出范围")
	ErrInvalidSeatCount     = errors.New("购票数量不能超过5张")
	ErrNoAvailableSeats     = errors.New("没有可用座位")
	ErrTicketNotFound       = errors.New("车票不存在")
	ErrAlreadyCheckedIn     = errors.New("已检票")
	ErrRefundNotAllowed     = errors.New("不允许退票")
	ErrAlreadyRefunded      = errors.New("已退票")
	ErrScheduleCancelled    = errors.New("班次已取消")
	ErrConcurrentUpdate     = errors.New("并发更新失败")
)

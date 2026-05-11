package housekeeping

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

const (
	maxDailyBookings    = 3
	minTransitionTime   = 30 * time.Minute
	validDurationMin    = 2.0
	validDurationMax    = 8.0
)

type BookingManager struct {
	bookings      map[string]*Booking
	pendingBookings []*Booking
	customers     map[string]*Customer
	auntBookings  map[string][]*Booking
	mu            sync.RWMutex
}

func NewBookingManager() *BookingManager {
	return &BookingManager{
		bookings:        make(map[string]*Booking),
		pendingBookings: make([]*Booking, 0),
		customers:       make(map[string]*Customer),
		auntBookings:    make(map[string][]*Booking),
	}
}

func isValidDuration(duration float64) bool {
	if duration < validDurationMin || duration > validDurationMax {
		return false
	}
	return duration == float64(int(duration/0.5))*0.5
}

func (bm *BookingManager) RegisterCustomer(customer *Customer) error {
	if customer.ID == "" {
		return fmt.Errorf("客户ID不能为空")
	}
	if customer.Name == "" {
		return fmt.Errorf("客户姓名不能为空")
	}
	if customer.Phone == "" {
		return fmt.Errorf("客户手机号不能为空")
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.customers[customer.ID]; exists {
		return fmt.Errorf("客户ID已存在")
	}

	if customer.Bookings == nil {
		customer.Bookings = make([]string, 0)
	}

	bm.customers[customer.ID] = customer
	return nil
}

func (bm *BookingManager) GetCustomer(id string) (*Customer, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	customer, exists := bm.customers[id]
	if !exists {
		return nil, fmt.Errorf("客户不存在")
	}
	return customer, nil
}

func (bm *BookingManager) GetAllCustomers() []*Customer {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	customers := make([]*Customer, 0, len(bm.customers))
	for _, customer := range bm.customers {
		customers = append(customers, customer)
	}
	return customers
}

func (bm *BookingManager) CreateBooking(req BookingRequest) (*Booking, error) {
	if req.CustomerID == "" {
		return nil, fmt.Errorf("客户ID不能为空")
	}
	if req.ServiceDate.Before(time.Now()) {
		return nil, fmt.Errorf("服务日期不能早于当前时间")
	}
	if !isValidDuration(req.EstimatedDuration) {
		return nil, fmt.Errorf("预约时长必须在2-8小时之间，且以半小时为单位")
	}
	if req.Address == "" {
		return nil, fmt.Errorf("服务地址不能为空")
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.customers[req.CustomerID]; !exists {
		return nil, fmt.Errorf("客户不存在")
	}

	booking := &Booking{
		ID:                fmt.Sprintf("BK%d", time.Now().UnixNano()),
		CustomerID:        req.CustomerID,
		ServiceCategory:   req.ServiceCategory,
		ServiceDate:       req.ServiceDate,
		EstimatedDuration: req.EstimatedDuration,
		Address:           req.Address,
		Status:            BookingStatusPending,
		CreatedAt:         time.Now(),
	}

	bm.bookings[booking.ID] = booking
	bm.customers[req.CustomerID].Bookings = append(bm.customers[req.CustomerID].Bookings, booking.ID)
	bm.pendingBookings = append(bm.pendingBookings, booking)

	return booking, nil
}

func (bm *BookingManager) GetBooking(id string) (*Booking, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	booking, exists := bm.bookings[id]
	if !exists {
		return nil, fmt.Errorf("订单不存在")
	}
	return booking, nil
}

func (bm *BookingManager) GetAllBookings() []*Booking {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	bookings := make([]*Booking, 0, len(bm.bookings))
	for _, booking := range bm.bookings {
		bookings = append(bookings, booking)
	}
	return bookings
}

func (bm *BookingManager) GetPendingBookings() []*Booking {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	pending := make([]*Booking, len(bm.pendingBookings))
	copy(pending, bm.pendingBookings)
	return pending
}

func (bm *BookingManager) GetBookingsByAunt(auntID string) []*Booking {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	bookings := make([]*Booking, 0)
	for _, booking := range bm.bookings {
		if booking.AuntID == auntID {
			bookings = append(bookings, booking)
		}
	}

	sort.Slice(bookings, func(i, j int) bool {
		return bookings[i].ServiceDate.Before(bookings[j].ServiceDate)
	})
	return bookings
}

func (bm *BookingManager) GetBookingsByCustomer(customerID string) []*Booking {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	bookings := make([]*Booking, 0)
	for _, booking := range bm.bookings {
		if booking.CustomerID == customerID {
			bookings = append(bookings, booking)
		}
	}

	sort.Slice(bookings, func(i, j int) bool {
		return bookings[i].CreatedAt.After(bookings[j].CreatedAt)
	})
	return bookings
}

func (bm *BookingManager) getAuntDailyBookings(auntID string, date time.Time) []*Booking {
	year, month, day := date.Date()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var dailyBookings []*Booking
	for _, booking := range bm.auntBookings[auntID] {
		if !booking.ServiceDate.Before(startOfDay) && booking.ServiceDate.Before(endOfDay) {
			dailyBookings = append(dailyBookings, booking)
		}
	}
	return dailyBookings
}

func (bm *BookingManager) hasTimeConflict(existing *Booking, newReq BookingRequest) bool {
	existingEnd := existing.StartTime.Add(time.Duration(existing.EstimatedDuration * float64(time.Hour)))
	existingEnd = existingEnd.Add(minTransitionTime)

	newStart := newReq.ServiceDate
	newEnd := newStart.Add(time.Duration(newReq.EstimatedDuration * float64(time.Hour)))

	return !newEnd.Before(existing.StartTime) && !newStart.After(existingEnd)
}

func (bm *BookingManager) isAuntAvailable(aunt *Aunt, req BookingRequest) bool {
	if aunt.ServiceCategory != req.ServiceCategory {
		return false
	}

	dailyBookings := bm.getAuntDailyBookings(aunt.ID, req.ServiceDate)
	if len(dailyBookings) >= maxDailyBookings {
		return false
	}

	for _, existing := range dailyBookings {
		if bm.hasTimeConflict(existing, req) {
			return false
		}
	}

	return true
}

func (bm *BookingManager) findAvailableAunt(am *AuntManager, req BookingRequest) *Aunt {
	candidates := am.FindAuntsByCategory(req.ServiceCategory)

	var availableAunts []*Aunt
	for _, aunt := range candidates {
		if bm.isAuntAvailable(aunt, req) {
			availableAunts = append(availableAunts, aunt)
		}
	}

	if len(availableAunts) == 0 {
		return nil
	}

	sort.Slice(availableAunts, func(i, j int) bool {
		return availableAunts[i].Rating > availableAunts[j].Rating
	})

	return availableAunts[0]
}

func (bm *BookingManager) TryMatchBooking(am *AuntManager, bookingID string) (*Booking, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	booking, exists := bm.bookings[bookingID]
	if !exists {
		return nil, fmt.Errorf("订单不存在")
	}

	if booking.Status != BookingStatusPending {
		return nil, fmt.Errorf("订单状态不是待匹配")
	}

	req := BookingRequest{
		CustomerID:       booking.CustomerID,
		ServiceCategory:  booking.ServiceCategory,
		ServiceDate:      booking.ServiceDate,
		EstimatedDuration: booking.EstimatedDuration,
		Address:          booking.Address,
	}

	aunt := bm.findAvailableAunt(am, req)
	if aunt == nil {
		return booking, nil
	}

	booking.AuntID = aunt.ID
	booking.Status = BookingStatusMatched
	bm.auntBookings[aunt.ID] = append(bm.auntBookings[aunt.ID], booking)

	for i, pending := range bm.pendingBookings {
		if pending.ID == booking.ID {
			bm.pendingBookings = append(bm.pendingBookings[:i], bm.pendingBookings[i+1:]...)
			break
		}
	}

	return booking, nil
}

func (bm *BookingManager) MatchPendingBookings(am *AuntManager) {
	for {
		matched := false
		pendingBookings := bm.GetPendingBookings()
		
		for _, booking := range pendingBookings {
			_, err := bm.TryMatchBooking(am, booking.ID)
			if err == nil && booking.Status == BookingStatusMatched {
				matched = true
			}
		}

		if !matched {
			break
		}
	}
}

func (bm *BookingManager) StartService(bookingID string) (*Booking, error) {
	booking, err := bm.GetBooking(bookingID)
	if err != nil {
		return nil, err
	}

	booking.mu.Lock()
	defer booking.mu.Unlock()

	if booking.Status != BookingStatusMatched {
		return nil, fmt.Errorf("订单状态不是已匹配")
	}

	booking.Status = BookingStatusInProgress
	booking.StartTime = time.Now()
	return booking, nil
}

func (bm *BookingManager) CompleteService(bookingID string, actualDuration float64) (*Booking, error) {
	booking, err := bm.GetBooking(bookingID)
	if err != nil {
		return nil, err
	}

	booking.mu.Lock()
	defer booking.mu.Unlock()

	if booking.Status != BookingStatusInProgress {
		return nil, fmt.Errorf("订单状态不是服务中")
	}

	if actualDuration < booking.EstimatedDuration {
		return nil, fmt.Errorf("实际服务时长不能少于预估时长")
	}

	booking.ActualDuration = actualDuration
	booking.EndTime = time.Now()
	booking.Price = CalculateBookingPrice(booking.ServiceCategory, booking.EstimatedDuration, actualDuration)
	booking.Status = BookingStatusCompleted

	return booking, nil
}

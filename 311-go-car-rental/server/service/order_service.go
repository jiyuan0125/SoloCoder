package service

import (
	"errors"
	"time"

	"carrental/common"
	"carrental/server/storage"
)

var (
	ErrUserNameRequired       = errors.New("用户名不能为空")
	ErrUserPhoneRequired      = errors.New("手机号不能为空")
	ErrPickupDateTooEarly     = errors.New("取车日期不能早于今天")
	ErrReturnDateTooEarly     = errors.New("还车日期必须晚于取车日期")
	ErrCarNotAvailable        = errors.New("车辆不可用或已被预订")
	ErrCarInMaintenance       = errors.New("车辆维护中，无法租赁")
	ErrOrderNotFound          = errors.New("订单不存在")
	ErrOrderAlreadyReturned   = errors.New("订单已还车")
	ErrDeductionExceedsDeposit = errors.New("扣款金额不能超过押金总额")
)

type OrderService struct {
	store      storage.Storage
	carService *CarService
}

func NewOrderService(store storage.Storage, carService *CarService) *OrderService {
	return &OrderService{
		store:      store,
		carService: carService,
	}
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func calculateDays(pickup, returnDate time.Time) int {
	pickupDay := truncateToDate(pickup)
	returnDay := truncateToDate(returnDate)
	days := int(returnDay.Sub(pickupDay).Hours()/24) + 1
	return days
}

func datesOverlap(start1, end1, start2, end2 time.Time) bool {
	s1 := truncateToDate(start1)
	e1 := truncateToDate(end1)
	s2 := truncateToDate(start2)
	e2 := truncateToDate(end2)

	return s1.Before(e2.AddDate(0, 0, 1)) && s2.Before(e1.AddDate(0, 0, 1))
}

func (s *OrderService) CheckAvailability(carID string, pickupDate, returnDate time.Time) (bool, error) {
	car, err := s.carService.GetCar(carID)
	if err != nil {
		return false, err
	}
	if car == nil {
		return false, ErrCarNotFound
	}
	if car.Status == common.CarStatusMaintenance {
		return false, ErrCarInMaintenance
	}

	orders, err := s.store.ListOrders()
	if err != nil {
		return false, err
	}

	for _, order := range orders {
		if order.CarID != carID {
			continue
		}

		if order.Status == common.OrderStatusReturned {
			continue
		}

		orderEndDate := order.ReturnDate
		if order.ActualReturnDate != nil {
			orderEndDate = *order.ActualReturnDate
		}

		if datesOverlap(pickupDate, returnDate, order.PickupDate, orderEndDate) {
			return false, nil
		}
	}

	return true, nil
}

func (s *OrderService) CreateOrder(req *common.CreateOrderRequest) (*common.Order, error) {
	if req.UserName == "" {
		return nil, ErrUserNameRequired
	}
	if req.UserPhone == "" {
		return nil, ErrUserPhoneRequired
	}

	today := truncateToDate(time.Now())
	pickupDay := truncateToDate(req.PickupDate)
	returnDay := truncateToDate(req.ReturnDate)

	if pickupDay.Before(today) {
		return nil, ErrPickupDateTooEarly
	}
	if !returnDay.After(pickupDay) {
		return nil, ErrReturnDateTooEarly
	}

	available, err := s.CheckAvailability(req.CarID, req.PickupDate, req.ReturnDate)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, ErrCarNotAvailable
	}

	car, err := s.carService.GetCar(req.CarID)
	if err != nil {
		return nil, err
	}
	if car == nil {
		return nil, ErrCarNotFound
	}

	days := calculateDays(req.PickupDate, req.ReturnDate)
	totalCost := float64(days) * car.DailyRate

	order := &common.Order{
		ID:                generateID(),
		CarID:             req.CarID,
		UserName:          req.UserName,
		UserPhone:         req.UserPhone,
		PickupDate:        req.PickupDate,
		ReturnDate:        req.ReturnDate,
		OriginalTotalCost: totalCost,
		TotalCost:         totalCost,
		Deposit:           car.Deposit,
		DepositDeduction:  0,
		Status:            common.OrderStatusReserved,
		CreatedAt:         time.Now(),
	}

	if err := s.store.SaveOrder(order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetOrder(id string) (*common.Order, error) {
	return s.store.GetOrder(id)
}

func (s *OrderService) ListOrders() ([]*common.Order, error) {
	return s.store.ListOrders()
}

func (s *OrderService) ListOrdersByPhone(phone string) ([]*common.Order, error) {
	return s.store.ListOrdersByPhone(phone)
}

func (s *OrderService) ReturnCar(req *common.ReturnCarRequest) (*common.Order, error) {
	order, err := s.store.GetOrder(req.OrderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	if order.Status == common.OrderStatusReturned {
		return nil, ErrOrderAlreadyReturned
	}

	actualReturnDay := truncateToDate(req.ActualReturnDate)
	expectedReturnDay := truncateToDate(order.ReturnDate)

	if actualReturnDay.After(expectedReturnDay) {
		overdueDays := int(actualReturnDay.Sub(expectedReturnDay).Hours()/24)

		car, err := s.carService.GetCar(order.CarID)
		if err != nil {
			return nil, err
		}
		if car == nil {
			return nil, ErrCarNotFound
		}

		overduePenalty := float64(overdueDays) * car.DailyRate * 1.5
		order.TotalCost = order.OriginalTotalCost + overduePenalty
	}

	depositDeduction := req.ViolationAmount
	if req.ViolationOrDamage && depositDeduction > 0 {
		if depositDeduction > order.Deposit {
			return nil, ErrDeductionExceedsDeposit
		}
		order.DepositDeduction = depositDeduction
	}

	actualReturnDate := req.ActualReturnDate
	order.ActualReturnDate = &actualReturnDate
	order.Mileage = req.Mileage
	order.Condition = req.Condition
	order.ViolationOrDamage = req.ViolationOrDamage
	order.ViolationAmount = req.ViolationAmount
	order.Status = common.OrderStatusReturned

	if err := s.store.UpdateOrder(order); err != nil {
		return nil, err
	}

	return order, nil
}

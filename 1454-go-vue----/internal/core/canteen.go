package core

import (
	"canteen/internal/common"
	"errors"
	"sync"
)

const (
	MinRechargeAmount = 1000
	MaxRechargeAmount = 50000
	MaxCreditLimit    = 10000
)

var (
	ErrEmployeeNotFound     = errors.New("employee not found")
	ErrDishNotFound         = errors.New("dish not found")
	ErrDishNotOnSale        = errors.New("dish is not on sale")
	ErrInvalidRechargeAmount = errors.New("recharge amount must be between 1000 and 50000 cents")
	ErrInsufficientBalance  = errors.New("insufficient balance and credit limit exceeded")
	ErrInvalidMealTime      = errors.New("breakfast voucher can only be used during breakfast")
	ErrEmployeeInactive     = errors.New("employee is inactive")
	ErrInvalidQuantity      = errors.New("quantity must be greater than 0")
)

type CanteenService struct {
	employees       map[string]*common.Employee
	employeeMu      map[string]*sync.Mutex
	employeesMu     sync.RWMutex
	dishes          map[string]*common.Dish
	dishesMu        sync.RWMutex
	rechargeRecords []*common.RechargeRecord
	rechargeMu      sync.RWMutex
	consumptions    map[string][]*common.ConsumptionRecord
	consumptionsMu  sync.RWMutex
}

func NewCanteenService() *CanteenService {
	return &CanteenService{
		employees:    make(map[string]*common.Employee),
		employeeMu:   make(map[string]*sync.Mutex),
		dishes:       make(map[string]*common.Dish),
		consumptions: make(map[string][]*common.ConsumptionRecord),
	}
}

func (s *CanteenService) getEmployeeMu(id string) *sync.Mutex {
	s.employeesMu.Lock()
	defer s.employeesMu.Unlock()
	if mu, exists := s.employeeMu[id]; exists {
		return mu
	}
	mu := &sync.Mutex{}
	s.employeeMu[id] = mu
	return mu
}

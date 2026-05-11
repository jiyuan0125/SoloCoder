package core

import (
	"canteen/internal/common"
	"errors"
	"time"

	"github.com/google/uuid"
)

var errNoItems = errors.New("no items to checkout")

func (s *CanteenService) Checkout(employeeID string, items []common.DishItem, mealType common.MealType) (*common.ConsumptionRecord, error) {
	if len(items) == 0 {
		return nil, errNoItems
	}

	mu := s.getEmployeeMu(employeeID)
	mu.Lock()
	defer mu.Unlock()

	s.employeesMu.RLock()
	employee, exists := s.employees[employeeID]
	s.employeesMu.RUnlock()

	if !exists {
		return nil, ErrEmployeeNotFound
	}
	if !employee.IsActive {
		return nil, ErrEmployeeInactive
	}

	s.dishesMu.RLock()
	consumptionItems := make([]common.ConsumptionItem, 0, len(items))
	totalAmount := 0

	for _, item := range items {
		if item.Quantity <= 0 {
			s.dishesMu.RUnlock()
			return nil, ErrInvalidQuantity
		}

		dish, exists := s.dishes[item.DishID]
		if !exists {
			s.dishesMu.RUnlock()
			return nil, ErrDishNotFound
		}
		if !dish.IsOnSale {
			s.dishesMu.RUnlock()
			return nil, ErrDishNotOnSale
		}

		snapshot := s.snapshotDish(dish)
		consumptionItems = append(consumptionItems, common.ConsumptionItem{
			Dish:     snapshot,
			Quantity: item.Quantity,
		})
		totalAmount += dish.Price * item.Quantity
	}
	s.dishesMu.RUnlock()

	isCredit := false
	if employee.Balance < totalAmount {
		creditNeeded := totalAmount - employee.Balance
		if employee.Credit+creditNeeded > MaxCreditLimit {
			return nil, ErrInsufficientBalance
		}
		employee.Credit += creditNeeded
		employee.Balance = 0
		isCredit = true
	} else {
		employee.Balance -= totalAmount
	}

	record := &common.ConsumptionRecord{
		ID:           uuid.New().String(),
		EmployeeID:   employeeID,
		EmployeeName: employee.Name,
		Items:        consumptionItems,
		TotalAmount:  totalAmount,
		MealType:     mealType,
		IsCredit:     isCredit,
		CreatedAt:    time.Now(),
	}

	s.consumptionsMu.Lock()
	s.consumptions[employeeID] = append(s.consumptions[employeeID], record)
	s.consumptionsMu.Unlock()

	return record, nil
}

func (s *CanteenService) GetConsumptionRecords(employeeID string) []*common.ConsumptionRecord {
	s.consumptionsMu.RLock()
	defer s.consumptionsMu.RUnlock()

	records, exists := s.consumptions[employeeID]
	if !exists {
		return []*common.ConsumptionRecord{}
	}

	result := make([]*common.ConsumptionRecord, len(records))
	copy(result, records)
	return result
}

func (s *CanteenService) GetConsumptionRecordsByDays(employeeID string, days int) []*common.ConsumptionRecord {
	if days <= 0 {
		days = 7
	}

	s.consumptionsMu.RLock()
	defer s.consumptionsMu.RUnlock()

	records, exists := s.consumptions[employeeID]
	if !exists {
		return []*common.ConsumptionRecord{}
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	result := make([]*common.ConsumptionRecord, 0)
	for _, r := range records {
		if r.CreatedAt.After(cutoff) {
			result = append(result, r)
		}
	}
	return result
}

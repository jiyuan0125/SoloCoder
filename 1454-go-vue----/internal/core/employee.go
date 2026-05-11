package core

import (
	"canteen/internal/common"
	"time"

	"github.com/google/uuid"
)

func (s *CanteenService) CreateEmployee(name string, gender common.Gender, weight int) *common.Employee {
	s.employeesMu.Lock()
	defer s.employeesMu.Unlock()

	id := uuid.New().String()
	employee := &common.Employee{
		ID:       id,
		Name:     name,
		Gender:   gender,
		Weight:   weight,
		Balance:  0,
		Credit:   0,
		IsActive: true,
	}

	s.employees[id] = employee
	return employee
}

func (s *CanteenService) GetEmployee(id string) (*common.Employee, error) {
	s.employeesMu.RLock()
	defer s.employeesMu.RUnlock()

	employee, exists := s.employees[id]
	if !exists {
		return nil, ErrEmployeeNotFound
	}
	return employee, nil
}

func (s *CanteenService) ListEmployees() []*common.Employee {
	s.employeesMu.RLock()
	defer s.employeesMu.RUnlock()

	employees := make([]*common.Employee, 0, len(s.employees))
	for _, emp := range s.employees {
		employees = append(employees, emp)
	}
	return employees
}

func (s *CanteenService) Recharge(employeeID string, amount int) (*common.RechargeRecord, error) {
	if amount < MinRechargeAmount || amount > MaxRechargeAmount {
		return nil, ErrInvalidRechargeAmount
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

	before := employee.Balance
	employee.Balance += amount

	record := &common.RechargeRecord{
		ID:         uuid.New().String(),
		EmployeeID: employeeID,
		Amount:     amount,
		Before:     before,
		After:      employee.Balance,
		CreatedAt:  time.Now(),
	}

	s.rechargeMu.Lock()
	s.rechargeRecords = append(s.rechargeRecords, record)
	s.rechargeMu.Unlock()

	return record, nil
}

func (s *CanteenService) GetRechargeRecords(employeeID string) []*common.RechargeRecord {
	s.rechargeMu.RLock()
	defer s.rechargeMu.RUnlock()

	records := make([]*common.RechargeRecord, 0)
	for _, r := range s.rechargeRecords {
		if r.EmployeeID == employeeID {
			records = append(records, r)
		}
	}
	return records
}

package core

import (
	"errors"
	"fmt"
	"merchant-mgmt-system/pkg/common"
	"time"

	"github.com/google/uuid"
)

func (s *Service) CreateContract(req *common.CreateContractRequest) (*common.Contract, error) {
	if req.CustomerID == "" {
		return nil, errors.New("客户ID不能为空")
	}
	if req.LeaseArea <= 0 {
		return nil, errors.New("租赁面积必须大于0")
	}
	if req.MonthlyRentRate <= 0 {
		return nil, errors.New("月租金单价必须大于0")
	}
	if req.EndDate.Before(req.StartDate) {
		return nil, errors.New("结束日期不能早于开始日期")
	}
	if req.FreeRentDays < 0 {
		return nil, errors.New("免租期天数不能为负数")
	}

	customer, ok := s.storage.GetCustomer(req.CustomerID)
	if !ok {
		return nil, errors.New("客户不存在")
	}

	if customer.Status == common.StatusSigned {
		existing := s.storage.GetContractsByCustomer(req.CustomerID)
		if len(existing) > 0 {
			return nil, errors.New("该客户已签约")
		}
	}

	totalRent := calculateTotalRent(
		req.StartDate,
		req.EndDate,
		req.MonthlyRentRate,
		req.LeaseArea,
		req.FreeRentDays,
	)

	now := time.Now()
	contract := &common.Contract{
		ID:              uuid.New().String(),
		CustomerID:      req.CustomerID,
		CustomerName:    customer.CustomerName,
		LeaseArea:       req.LeaseArea,
		MonthlyRentRate: req.MonthlyRentRate,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		FreeRentDays:    req.FreeRentDays,
		TotalRent:       totalRent,
		CreatedAt:       now,
	}

	s.storage.SaveContract(contract)
	s.updateCustomerStatus(req.CustomerID, common.StatusSigned)

	follows := s.storage.GetFollowsByCustomer(req.CustomerID)
	for _, f := range follows {
		if f.Status == common.FollowStatusOpen {
			f.Status = common.FollowStatusClose
			f.UpdatedAt = now
			s.storage.SaveFollow(f)
		}
	}

	return contract, nil
}

func (s *Service) ListContracts() ([]*common.Contract, error) {
	return s.storage.GetAllContracts(), nil
}

func calculateTotalRent(
	startDate, endDate time.Time,
	monthlyRate float64,
	area float64,
	freeRentDays int,
) float64 {
	if startDate.After(endDate) {
		return 0
	}

	baseMonthRent := monthlyRate * area

	remainingFreeDays := freeRentDays
	totalRent := 0.0

	currentMonth := startDate
	for !currentMonth.After(endDate) {
		monthStart := time.Date(currentMonth.Year(), currentMonth.Month(), 1, 0, 0, 0, 0, currentMonth.Location())
		monthEnd := monthStart.AddDate(0, 1, -1)

		effectiveStart := laterDate(startDate, monthStart)
		effectiveEnd := earlierDate(endDate, monthEnd)

		daysInMonth := monthEnd.Day()
		occupiedDays := int(effectiveEnd.Sub(effectiveStart).Hours()/24) + 1

		chargedDays := occupiedDays
		if remainingFreeDays > 0 {
			if remainingFreeDays >= occupiedDays {
				chargedDays = 0
				remainingFreeDays -= occupiedDays
			} else {
				chargedDays = occupiedDays - remainingFreeDays
				remainingFreeDays = 0
			}
		}

		if chargedDays > 0 {
			monthlyRent := baseMonthRent * float64(chargedDays) / float64(daysInMonth)
			totalRent += monthlyRent
		}

		currentMonth = currentMonth.AddDate(0, 1, 0)
	}

	return totalRent
}

func laterDate(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func earlierDate(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func ParseDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02",
		"2006/01/02",
		"2006-1-2",
		"2006/1/2",
	}
	var lastErr error
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, dateStr, time.Local)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, fmt.Errorf("无法解析日期: %s", lastErr)
}

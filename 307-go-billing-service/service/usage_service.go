package service

import (
	"billing-service/model"
	"billing-service/repository"
	"time"
)

type UsageService struct {
	usageRepo *repository.UsageRepository
}

func NewUsageService(usageRepo *repository.UsageRepository) *UsageService {
	return &UsageService{usageRepo: usageRepo}
}

func (s *UsageService) RecordSmsUsage(customerID uint, smsCount int) error {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	return s.usageRepo.UpdateSmsUsage(customerID, year, month, smsCount)
}

func (s *UsageService) RecordStorageUsage(customerID uint, storageGB float64) error {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	return s.usageRepo.UpdateStorageUsage(customerID, year, month, storageGB)
}

func (s *UsageService) GetUsage(customerID uint, year int, month int) (*model.Usage, error) {
	return s.usageRepo.GetByCustomerAndMonth(customerID, year, month)
}

func (s *UsageService) GetOrCreateUsage(customerID uint, year int, month int) (*model.Usage, error) {
	return s.usageRepo.GetOrCreate(customerID, year, month)
}

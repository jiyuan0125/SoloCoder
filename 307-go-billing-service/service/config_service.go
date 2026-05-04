package service

import (
	"billing-service/model"
	"billing-service/repository"
	"strconv"
)

type ConfigService struct {
	configRepo *repository.ConfigRepository
}

func NewConfigService(configRepo *repository.ConfigRepository) *ConfigService {
	return &ConfigService{configRepo: configRepo}
}

func (s *ConfigService) GetConfig(key string) (*model.Config, error) {
	return s.configRepo.Get(key)
}

func (s *ConfigService) SetConfig(key string, value string, description string) error {
	return s.configRepo.Set(key, value, description)
}

func (s *ConfigService) GetSmsUnitPrice() (float64, error) {
	return s.configRepo.GetSmsUnitPrice()
}

func (s *ConfigService) GetStorageUnitPrice() (float64, error) {
	return s.configRepo.GetStorageUnitPrice()
}

func (s *ConfigService) SetSmsUnitPrice(price float64) error {
	return s.configRepo.Set("sms.unit_price", strconv.FormatFloat(price, 'f', -1, 64), "短信单价（元/条）")
}

func (s *ConfigService) SetStorageUnitPrice(price float64) error {
	return s.configRepo.Set("storage.unit_price", strconv.FormatFloat(price, 'f', -1, 64), "云存储单价（元/GB/月）")
}

func (s *ConfigService) GetPaymentDueDays() (int, error) {
	return s.configRepo.GetPaymentDueDays()
}

func (s *ConfigService) GetSevereOverdueDays() (int, error) {
	return s.configRepo.GetSevereOverdueDays()
}

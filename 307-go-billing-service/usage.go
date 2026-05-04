package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SMSUsageRecord struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	Count       int       `json:"count"`
	RecordDate  time.Time `json:"record_date"`
	CreatedAt   time.Time `json:"created_at"`
}

type StorageUsageRecord struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	SizeGB      float64   `json:"size_gb"`
	RecordDate  time.Time `json:"record_date"`
	CreatedAt   time.Time `json:"created_at"`
}

type PricingConfig struct {
	SMSPricePerUnit     float64 `json:"sms_price_per_unit"`
	StoragePricePerGB   float64 `json:"storage_price_per_gb"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type UsageStore interface {
	SaveSMSUsage(record *SMSUsageRecord) error
	SaveStorageUsage(record *StorageUsageRecord) error
	GetSMSUsageForMonth(customerID string, year int, month time.Month) (int, error)
	GetMaxStorageUsageForMonth(customerID string, year int, month time.Month) (float64, error)
	GetSMSUsageRecordsForMonth(customerID string, year int, month time.Month) ([]*SMSUsageRecord, error)
	GetStorageUsageRecordsForMonth(customerID string, year int, month time.Month) ([]*StorageUsageRecord, error)
	SavePricingConfig(config *PricingConfig) error
	GetPricingConfig() (*PricingConfig, error)
}

type FileUsageStore struct {
	filePath          string
	smsRecords        []*SMSUsageRecord
	storageRecords    []*StorageUsageRecord
	pricingConfig     *PricingConfig
	mu                sync.RWMutex
}

func (s *FileUsageStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.smsRecords == nil {
		s.smsRecords = []*SMSUsageRecord{}
	}
	if s.storageRecords == nil {
		s.storageRecords = []*StorageUsageRecord{}
	}
	if s.pricingConfig == nil {
		s.pricingConfig = &PricingConfig{
			SMSPricePerUnit:   0.1,
			StoragePricePerGB: 10.0,
			UpdatedAt:         time.Now(),
		}
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var stored struct {
		SMSRecords     []*SMSUsageRecord   `json:"sms_records"`
		StorageRecords []*StorageUsageRecord `json:"storage_records"`
		PricingConfig  *PricingConfig      `json:"pricing_config"`
	}

	if err := json.Unmarshal(data, &stored); err != nil {
		return err
	}

	if stored.SMSRecords != nil {
		s.smsRecords = stored.SMSRecords
	}
	if stored.StorageRecords != nil {
		s.storageRecords = stored.StorageRecords
	}
	if stored.PricingConfig != nil {
		s.pricingConfig = stored.PricingConfig
	}

	return nil
}

func (s *FileUsageStore) saveToFile() error {
	stored := struct {
		SMSRecords     []*SMSUsageRecord   `json:"sms_records"`
		StorageRecords []*StorageUsageRecord `json:"storage_records"`
		PricingConfig  *PricingConfig      `json:"pricing_config"`
	}{
		SMSRecords:     s.smsRecords,
		StorageRecords: s.storageRecords,
		PricingConfig:  s.pricingConfig,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *FileUsageStore) SaveSMSUsage(record *SMSUsageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.smsRecords = append(s.smsRecords, record)
	return s.saveToFile()
}

func (s *FileUsageStore) SaveStorageUsage(record *StorageUsageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.storageRecords = append(s.storageRecords, record)
	return s.saveToFile()
}

func (s *FileUsageStore) GetSMSUsageForMonth(customerID string, year int, month time.Month) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	var total int
	for _, r := range s.smsRecords {
		if r.CustomerID == customerID && r.RecordDate.After(start) && r.RecordDate.Before(end) {
			total += r.Count
		}
	}
	return total, nil
}

func (s *FileUsageStore) GetMaxStorageUsageForMonth(customerID string, year int, month time.Month) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	var maxSize float64
	for _, r := range s.storageRecords {
		if r.CustomerID == customerID && r.RecordDate.After(start) && r.RecordDate.Before(end) {
			if r.SizeGB > maxSize {
				maxSize = r.SizeGB
			}
		}
	}
	return maxSize, nil
}

func (s *FileUsageStore) GetSMSUsageRecordsForMonth(customerID string, year int, month time.Month) ([]*SMSUsageRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	var result []*SMSUsageRecord
	for _, r := range s.smsRecords {
		if r.CustomerID == customerID && r.RecordDate.After(start) && r.RecordDate.Before(end) {
			result = append(result, r)
		}
	}
	return result, nil
}

func (s *FileUsageStore) GetStorageUsageRecordsForMonth(customerID string, year int, month time.Month) ([]*StorageUsageRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	var result []*StorageUsageRecord
	for _, r := range s.storageRecords {
		if r.CustomerID == customerID && r.RecordDate.After(start) && r.RecordDate.Before(end) {
			result = append(result, r)
		}
	}
	return result, nil
}

func (s *FileUsageStore) SavePricingConfig(config *PricingConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pricingConfig = config
	return s.saveToFile()
}

func (s *FileUsageStore) GetPricingConfig() (*PricingConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.pricingConfig == nil {
		return &PricingConfig{
			SMSPricePerUnit:   0.1,
			StoragePricePerGB: 10.0,
			UpdatedAt:         time.Now(),
		}, nil
	}
	return s.pricingConfig, nil
}

type UsageService struct {
	store         UsageStore
	customerStore CustomerStore
}

func NewUsageService(store UsageStore, customerStore CustomerStore) *UsageService {
	return &UsageService{
		store:         store,
		customerStore: customerStore,
	}
}

type RecordSMSUsageRequest struct {
	CustomerID string `json:"customer_id"`
	Count      int    `json:"count"`
	RecordDate string `json:"record_date"`
}

type RecordStorageUsageRequest struct {
	CustomerID string  `json:"customer_id"`
	SizeGB     float64 `json:"size_gb"`
	RecordDate string  `json:"record_date"`
}

type UpdatePricingConfigRequest struct {
	SMSPricePerUnit   float64 `json:"sms_price_per_unit"`
	StoragePricePerGB float64 `json:"storage_price_per_gb"`
}

func (s *UsageService) RecordSMSUsage(w http.ResponseWriter, r *http.Request) {
	var req RecordSMSUsageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CustomerID == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	if req.Count <= 0 {
		http.Error(w, "Count must be positive", http.StatusBadRequest)
		return
	}

	if _, err := s.customerStore.FindByID(req.CustomerID); err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	var recordDate time.Time
	if req.RecordDate != "" {
		var err error
		recordDate, err = time.Parse("2006-01-02", req.RecordDate)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	} else {
		recordDate = time.Now()
	}

	record := &SMSUsageRecord{
		ID:         uuid.New().String(),
		CustomerID: req.CustomerID,
		Count:      req.Count,
		RecordDate: recordDate,
		CreatedAt:  time.Now(),
	}

	if err := s.store.SaveSMSUsage(record); err != nil {
		http.Error(w, fmt.Sprintf("Failed to record SMS usage: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

func (s *UsageService) RecordStorageUsage(w http.ResponseWriter, r *http.Request) {
	var req RecordStorageUsageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CustomerID == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	if req.SizeGB < 0 {
		http.Error(w, "Size cannot be negative", http.StatusBadRequest)
		return
	}

	if _, err := s.customerStore.FindByID(req.CustomerID); err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	var recordDate time.Time
	if req.RecordDate != "" {
		var err error
		recordDate, err = time.Parse("2006-01-02", req.RecordDate)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	} else {
		recordDate = time.Now()
	}

	record := &StorageUsageRecord{
		ID:         uuid.New().String(),
		CustomerID: req.CustomerID,
		SizeGB:     req.SizeGB,
		RecordDate: recordDate,
		CreatedAt:  time.Now(),
	}

	if err := s.store.SaveStorageUsage(record); err != nil {
		http.Error(w, fmt.Sprintf("Failed to record storage usage: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

func (s *UsageService) GetSMSUsage(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer_id")
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	if customerID == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	var year int
	var month time.Month
	now := time.Now()

	if yearStr != "" {
		var err error
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			http.Error(w, "Invalid year", http.StatusBadRequest)
			return
		}
	} else {
		year = now.Year()
	}

	if monthStr != "" {
		m, err := strconv.Atoi(monthStr)
		if err != nil {
			http.Error(w, "Invalid month", http.StatusBadRequest)
			return
		}
		month = time.Month(m)
	} else {
		month = now.Month()
	}

	total, err := s.store.GetSMSUsageForMonth(customerID, year, month)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get SMS usage: %v", err), http.StatusInternalServerError)
		return
	}

	records, err := s.store.GetSMSUsageRecordsForMonth(customerID, year, month)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get SMS usage records: %v", err), http.StatusInternalServerError)
		return
	}

	customer, err := s.customerStore.FindByID(customerID)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	plan := getPlan(customer.CurrentPlan)
	quota := 0
	if plan != nil {
		quota = plan.SMSQuota
	}

	excess := 0
	if total > quota {
		excess = total - quota
	}

	result := map[string]interface{}{
		"customer_id": customerID,
		"year":        year,
		"month":       month,
		"total":       total,
		"quota":       quota,
		"excess":      excess,
		"records":     records,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *UsageService) GetStorageUsage(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer_id")
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	if customerID == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	var year int
	var month time.Month
	now := time.Now()

	if yearStr != "" {
		var err error
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			http.Error(w, "Invalid year", http.StatusBadRequest)
			return
		}
	} else {
		year = now.Year()
	}

	if monthStr != "" {
		m, err := strconv.Atoi(monthStr)
		if err != nil {
			http.Error(w, "Invalid month", http.StatusBadRequest)
			return
		}
		month = time.Month(m)
	} else {
		month = now.Month()
	}

	maxUsage, err := s.store.GetMaxStorageUsageForMonth(customerID, year, month)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get storage usage: %v", err), http.StatusInternalServerError)
		return
	}

	records, err := s.store.GetStorageUsageRecordsForMonth(customerID, year, month)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get storage usage records: %v", err), http.StatusInternalServerError)
		return
	}

	customer, err := s.customerStore.FindByID(customerID)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	plan := getPlan(customer.CurrentPlan)
	quota := 0.0
	if plan != nil {
		quota = float64(plan.StorageQuota)
	}

	excess := 0.0
	if maxUsage > quota {
		excess = maxUsage - quota
	}

	result := map[string]interface{}{
		"customer_id": customerID,
		"year":        year,
		"month":       month,
		"max_usage":   maxUsage,
		"quota":       quota,
		"excess":      excess,
		"records":     records,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *UsageService) UpdatePricingConfig(w http.ResponseWriter, r *http.Request) {
	var req UpdatePricingConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SMSPricePerUnit < 0 {
		http.Error(w, "SMS price cannot be negative", http.StatusBadRequest)
		return
	}

	if req.StoragePricePerGB < 0 {
		http.Error(w, "Storage price cannot be negative", http.StatusBadRequest)
		return
	}

	config := &PricingConfig{
		SMSPricePerUnit:   req.SMSPricePerUnit,
		StoragePricePerGB: req.StoragePricePerGB,
		UpdatedAt:         time.Now(),
	}

	if err := s.store.SavePricingConfig(config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update pricing config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (s *UsageService) GetPricingConfig(w http.ResponseWriter, r *http.Request) {
	config, err := s.store.GetPricingConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get pricing config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func calculateUsageCost(customerID string, year int, month time.Month, usageStore UsageStore, customerStore CustomerStore) (smsCost float64, storageCost float64, err error) {
	customer, err := customerStore.FindByID(customerID)
	if err != nil {
		return 0, 0, err
	}

	plan := getPlan(customer.CurrentPlan)
	if plan == nil {
		return 0, 0, fmt.Errorf("plan not found for customer")
	}

	smsUsage, err := usageStore.GetSMSUsageForMonth(customerID, year, month)
	if err != nil {
		return 0, 0, err
	}

	maxStorage, err := usageStore.GetMaxStorageUsageForMonth(customerID, year, month)
	if err != nil {
		return 0, 0, err
	}

	config, err := usageStore.GetPricingConfig()
	if err != nil {
		return 0, 0, err
	}

	smsExcess := 0
	if smsUsage > plan.SMSQuota {
		smsExcess = smsUsage - plan.SMSQuota
	}

	storageExcess := 0.0
	if maxStorage > float64(plan.StorageQuota) {
		storageExcess = maxStorage - float64(plan.StorageQuota)
	}

	smsCost = float64(smsExcess) * config.SMSPricePerUnit
	storageCost = storageExcess * config.StoragePricePerGB

	return smsCost, storageCost, nil
}

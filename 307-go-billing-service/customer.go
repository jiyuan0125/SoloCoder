package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PlanType string

const (
	PlanBasic    PlanType = "basic"
	PlanPro      PlanType = "pro"
	PlanEnterprise PlanType = "enterprise"
)

type Plan struct {
	Type        PlanType `json:"type"`
	MonthlyFee  float64  `json:"monthly_fee"`
	SMSQuota    int      `json:"sms_quota"`
	StorageQuota int     `json:"storage_quota"`
}

var plans = map[PlanType]Plan{
	PlanBasic: {
		Type:        PlanBasic,
		MonthlyFee:  99.0,
		SMSQuota:    1000,
		StorageQuota: 10,
	},
	PlanPro: {
		Type:        PlanPro,
		MonthlyFee:  299.0,
		SMSQuota:    5000,
		StorageQuota: 50,
	},
	PlanEnterprise: {
		Type:        PlanEnterprise,
		MonthlyFee:  899.0,
		SMSQuota:    20000,
		StorageQuota: 200,
	},
}

type Customer struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	CurrentPlan PlanType  `json:"current_plan"`
	CreatedAt   time.Time `json:"created_at"`
}

type PlanChangeRecord struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	OldPlan     PlanType  `json:"old_plan"`
	NewPlan     PlanType  `json:"new_plan"`
	ChangeDate  time.Time `json:"change_date"`
	CreatedAt   time.Time `json:"created_at"`
}

type CustomerStore interface {
	Save(customer *Customer) error
	FindByID(id string) (*Customer, error)
	FindAll() ([]*Customer, error)
	SavePlanChange(record *PlanChangeRecord) error
	FindPlanChangesByCustomer(customerID string) ([]*PlanChangeRecord, error)
	FindPlanChangesForMonth(customerID string, year int, month time.Month) ([]*PlanChangeRecord, error)
}

type FileCustomerStore struct {
	filePath        string
	customers       map[string]*Customer
	planChanges     []*PlanChangeRecord
	mu              sync.RWMutex
}

func (s *FileCustomerStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.customers == nil {
		s.customers = make(map[string]*Customer)
	}
	if s.planChanges == nil {
		s.planChanges = []*PlanChangeRecord{}
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var stored struct {
		Customers   map[string]*Customer   `json:"customers"`
		PlanChanges []*PlanChangeRecord    `json:"plan_changes"`
	}

	if err := json.Unmarshal(data, &stored); err != nil {
		return err
	}

	if stored.Customers != nil {
		s.customers = stored.Customers
	}
	if stored.PlanChanges != nil {
		s.planChanges = stored.PlanChanges
	}

	return nil
}

func (s *FileCustomerStore) saveToFile() error {
	stored := struct {
		Customers   map[string]*Customer   `json:"customers"`
		PlanChanges []*PlanChangeRecord    `json:"plan_changes"`
	}{
		Customers:   s.customers,
		PlanChanges: s.planChanges,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *FileCustomerStore) Save(customer *Customer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.customers[customer.ID] = customer
	return s.saveToFile()
}

func (s *FileCustomerStore) FindByID(id string) (*Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	customer, exists := s.customers[id]
	if !exists {
		return nil, errors.New("customer not found")
	}
	return customer, nil
}

func (s *FileCustomerStore) FindAll() ([]*Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Customer, 0, len(s.customers))
	for _, c := range s.customers {
		result = append(result, c)
	}
	return result, nil
}

func (s *FileCustomerStore) SavePlanChange(record *PlanChangeRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.planChanges = append(s.planChanges, record)
	return s.saveToFile()
}

func (s *FileCustomerStore) FindPlanChangesByCustomer(customerID string) ([]*PlanChangeRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*PlanChangeRecord
	for _, r := range s.planChanges {
		if r.CustomerID == customerID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (s *FileCustomerStore) FindPlanChangesForMonth(customerID string, year int, month time.Month) ([]*PlanChangeRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*PlanChangeRecord
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	for _, r := range s.planChanges {
		if r.CustomerID == customerID && r.ChangeDate.After(start) && r.ChangeDate.Before(end) {
			result = append(result, r)
		}
	}
	return result, nil
}

type CustomerService struct {
	store CustomerStore
}

func NewCustomerService(store CustomerStore) *CustomerService {
	return &CustomerService{store: store}
}

type CreateCustomerRequest struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	PlanType PlanType `json:"plan_type"`
}

type ChangePlanRequest struct {
	CustomerID string   `json:"customer_id"`
	NewPlan    PlanType `json:"new_plan"`
	ChangeDate string   `json:"change_date"`
}

func (s *CustomerService) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var req CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	if _, valid := plans[req.PlanType]; !valid {
		http.Error(w, "Invalid plan type", http.StatusBadRequest)
		return
	}

	customer := &Customer{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Email:       req.Email,
		CurrentPlan: req.PlanType,
		CreatedAt:   time.Now(),
	}

	if err := s.store.Save(customer); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create customer: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(customer)
}

func (s *CustomerService) GetCustomer(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/customers/")
	if path == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	customer, err := s.store.FindByID(path)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	planChanges, _ := s.store.FindPlanChangesByCustomer(customer.ID)

	result := struct {
		Customer    *Customer            `json:"customer"`
		Plan        *Plan                `json:"plan"`
		PlanChanges []*PlanChangeRecord `json:"plan_changes"`
	}{
		Customer:    customer,
		Plan:        getPlan(customer.CurrentPlan),
		PlanChanges: planChanges,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *CustomerService) ListCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := s.store.FindAll()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list customers: %v", err), http.StatusInternalServerError)
		return
	}

	type CustomerWithPlan struct {
		*Customer
		Plan *Plan `json:"plan"`
	}

	result := make([]CustomerWithPlan, 0, len(customers))
	for _, c := range customers {
		result = append(result, CustomerWithPlan{
			Customer: c,
			Plan:     getPlan(c.CurrentPlan),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *CustomerService) ChangePlan(w http.ResponseWriter, r *http.Request) {
	var req ChangePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CustomerID == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	if _, valid := plans[req.NewPlan]; !valid {
		http.Error(w, "Invalid plan type", http.StatusBadRequest)
		return
	}

	customer, err := s.store.FindByID(req.CustomerID)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	if customer.CurrentPlan == req.NewPlan {
		http.Error(w, "New plan must be different from current plan", http.StatusBadRequest)
		return
	}

	var changeDate time.Time
	if req.ChangeDate != "" {
		changeDate, err = time.Parse("2006-01-02", req.ChangeDate)
		if err != nil {
			http.Error(w, "Invalid change date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	} else {
		changeDate = time.Now()
	}

	oldPlan := customer.CurrentPlan
	customer.CurrentPlan = req.NewPlan

	record := &PlanChangeRecord{
		ID:         uuid.New().String(),
		CustomerID: customer.ID,
		OldPlan:    oldPlan,
		NewPlan:    req.NewPlan,
		ChangeDate: changeDate,
		CreatedAt:  time.Now(),
	}

	if err := s.store.Save(customer); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update customer: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.store.SavePlanChange(record); err != nil {
		http.Error(w, fmt.Sprintf("Failed to record plan change: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"customer":    customer,
		"plan_change": record,
	})
}

func getPlan(planType PlanType) *Plan {
	if p, ok := plans[planType]; ok {
		return &p
	}
	return nil
}

func getDaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}

func calculatePlanCostForMonth(customerID string, year int, month time.Month, store CustomerStore) (float64, error) {
	customer, err := store.FindByID(customerID)
	if err != nil {
		return 0, err
	}

	planChanges, err := store.FindPlanChangesForMonth(customerID, year, month)
	if err != nil {
		return 0, err
	}

	daysInMonth := getDaysInMonth(year, month)
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)

	type PlanPeriod struct {
		plan     PlanType
		startDay int
		endDay   int
	}

	var periods []PlanPeriod
	currentPlan := customer.CurrentPlan
	currentStartDay := 1

	sortedChanges := make([]*PlanChangeRecord, len(planChanges))
	copy(sortedChanges, planChanges)
	for i := range sortedChanges {
		for j := i + 1; j < len(sortedChanges); j++ {
			if sortedChanges[j].ChangeDate.Before(sortedChanges[i].ChangeDate) {
				sortedChanges[i], sortedChanges[j] = sortedChanges[j], sortedChanges[i]
			}
		}
	}

	for _, change := range sortedChanges {
		changeDay := change.ChangeDate.Day()
		if changeDay > currentStartDay && changeDay <= daysInMonth {
			periods = append(periods, PlanPeriod{
				plan:     change.OldPlan,
				startDay: currentStartDay,
				endDay:   changeDay - 1,
			})
			currentStartDay = changeDay
			currentPlan = change.NewPlan
		}
	}

	if currentStartDay <= daysInMonth {
		periods = append(periods, PlanPeriod{
			plan:     currentPlan,
			startDay: currentStartDay,
			endDay:   daysInMonth,
		})
	}

	if len(periods) == 0 {
		periods = append(periods, PlanPeriod{
			plan:     customer.CurrentPlan,
			startDay: 1,
			endDay:   daysInMonth,
		})
	}

	var totalCost float64
	for _, period := range periods {
		plan := getPlan(period.plan)
		if plan == nil {
			continue
		}
		days := period.endDay - period.startDay + 1
		dailyCost := plan.MonthlyFee / float64(daysInMonth)
		totalCost += dailyCost * float64(days)
	}

	return totalCost, nil
}

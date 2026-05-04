package model

import (
	"billing/pkg/api"
	"sync"
	"time"
)

type DataStore struct {
	mu           sync.RWMutex
	Plans        map[string]*api.Plan
	Customers    map[string]*api.Customer
	PlanChanges  map[string]*api.PlanChange
	UsageRecords map[string]*api.UsageRecord
	Bills        map[string]*api.Bill
	Pricing      *api.PricingConfig

	customerPlanChanges map[string][]*api.PlanChange
}

func NewDataStore() *DataStore {
	return &DataStore{
		Plans:               make(map[string]*api.Plan),
		Customers:           make(map[string]*api.Customer),
		PlanChanges:         make(map[string]*api.PlanChange),
		UsageRecords:        make(map[string]*api.UsageRecord),
		Bills:               make(map[string]*api.Bill),
		Pricing:             DefaultPricingConfig(),
		customerPlanChanges: make(map[string][]*api.PlanChange),
	}
}

func DefaultPricingConfig() *api.PricingConfig {
	return &api.PricingConfig{
		SmsPricePerUnit:    0.1,
		StoragePricePerGB:  5.0,
		PaymentDueDays:     15,
		SeriousOverdueDays: 60,
	}
}

func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[int(time.Now().UnixNano())%len(letters)]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}

func (s *DataStore) CreatePlan(plan *api.Plan) (*api.Plan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	plan.ID = generateID()
	s.Plans[plan.ID] = plan
	return plan, nil
}

func (s *DataStore) GetPlan(id string) (*api.Plan, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plan, ok := s.Plans[id]
	return plan, ok
}

func (s *DataStore) ListPlans() []*api.Plan {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plans := make([]*api.Plan, 0, len(s.Plans))
	for _, p := range s.Plans {
		plans = append(plans, p)
	}
	return plans
}

func (s *DataStore) CreateCustomer(customer *api.Customer) (*api.Customer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	customer.ID = generateID()
	customer.RegisteredAt = time.Now()
	customer.IsActive = true
	s.Customers[customer.ID] = customer
	return customer, nil
}

func (s *DataStore) GetCustomer(id string) (*api.Customer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	customer, ok := s.Customers[id]
	return customer, ok
}

func (s *DataStore) ListCustomers() []*api.Customer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	customers := make([]*api.Customer, 0, len(s.Customers))
	for _, c := range s.Customers {
		customers = append(customers, c)
	}
	return customers
}

func (s *DataStore) UpdateCustomerPlan(customerID string, newPlanID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	customer, ok := s.Customers[customerID]
	if !ok {
		return ErrCustomerNotFound
	}
	customer.CurrentPlanID = newPlanID
	return nil
}

func (s *DataStore) RecordPlanChange(change *api.PlanChange) (*api.PlanChange, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	change.ID = generateID()
	change.CreatedAt = time.Now()
	s.PlanChanges[change.ID] = change
	s.customerPlanChanges[change.CustomerID] = append(
		s.customerPlanChanges[change.CustomerID],
		change,
	)
	return change, nil
}

func (s *DataStore) GetCustomerPlanChanges(customerID string) []*api.PlanChange {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.customerPlanChanges[customerID]
}

func (s *DataStore) GetUsageRecord(customerID string, year, month int) (*api.UsageRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := usageKey(customerID, year, month)
	usage, ok := s.UsageRecords[key]
	return usage, ok
}

func (s *DataStore) CreateOrUpdateUsage(customerID string, year, month int, smsIncrement int, storageUpdate *float64) (*api.UsageRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := usageKey(customerID, year, month)
	usage, ok := s.UsageRecords[key]
	if !ok {
		usage = &api.UsageRecord{
			ID:             generateID(),
			CustomerID:     customerID,
			Year:           year,
			Month:          month,
			SmsCount:       0,
			StorageUsageGB: 0,
			LastUpdated:    time.Now(),
		}
		s.UsageRecords[key] = usage
	}

	usage.SmsCount += smsIncrement
	if storageUpdate != nil {
		usage.StorageUsageGB = *storageUpdate
	}
	usage.LastUpdated = time.Now()
	return usage, nil
}

func (s *DataStore) CreateBill(bill *api.Bill) (*api.Bill, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bill.ID = generateID()
	s.Bills[bill.ID] = bill
	return bill, nil
}

func (s *DataStore) GetBill(id string) (*api.Bill, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bill, ok := s.Bills[id]
	return bill, ok
}

func (s *DataStore) GetCustomerBill(customerID string, year, month int) (*api.Bill, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, b := range s.Bills {
		if b.CustomerID == customerID && b.Year == year && b.Month == month {
			return b, true
		}
	}
	return nil, false
}

func (s *DataStore) ListCustomerBills(customerID string) []*api.Bill {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bills := make([]*api.Bill, 0)
	for _, b := range s.Bills {
		if b.CustomerID == customerID {
			bills = append(bills, b)
		}
	}
	return bills
}

func (s *DataStore) ListAllBills() []*api.Bill {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bills := make([]*api.Bill, 0, len(s.Bills))
	for _, b := range s.Bills {
		bills = append(bills, b)
	}
	return bills
}

func (s *DataStore) UpdateBillStatus(billID string, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bill, ok := s.Bills[billID]
	if !ok {
		return ErrBillNotFound
	}
	bill.Status = status
	return nil
}

func (s *DataStore) UpdateBillPayment(billID string, payment *api.PaymentRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bill, ok := s.Bills[billID]
	if !ok {
		return ErrBillNotFound
	}
	bill.PaymentRecord = payment
	bill.Status = api.BillStatusPaid
	return nil
}

func (s *DataStore) GetPricing() *api.PricingConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &api.PricingConfig{
		SmsPricePerUnit:    s.Pricing.SmsPricePerUnit,
		StoragePricePerGB:  s.Pricing.StoragePricePerGB,
		PaymentDueDays:     s.Pricing.PaymentDueDays,
		SeriousOverdueDays: s.Pricing.SeriousOverdueDays,
	}
}

func (s *DataStore) UpdatePricing(req *api.UpdatePricingRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.SmsPricePerUnit != nil {
		s.Pricing.SmsPricePerUnit = *req.SmsPricePerUnit
	}
	if req.StoragePricePerGB != nil {
		s.Pricing.StoragePricePerGB = *req.StoragePricePerGB
	}
	if req.PaymentDueDays != nil {
		s.Pricing.PaymentDueDays = *req.PaymentDueDays
	}
	if req.SeriousOverdueDays != nil {
		s.Pricing.SeriousOverdueDays = *req.SeriousOverdueDays
	}
}

func (s *DataStore) HasUnpaidPreviousBill(customerID string, year, month int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, b := range s.Bills {
		if b.CustomerID != customerID {
			continue
		}
		if b.Year < year || (b.Year == year && b.Month < month) {
			if b.Status != api.BillStatusPaid {
				return true
			}
		}
	}
	return false
}

func usageKey(customerID string, year, month int) string {
	return customerID + ":" + string(rune(year)) + "-" + string(rune(month))
}

package service

import (
	"billing/internal/server/model"
	"billing/internal/server/store"
	"billing/pkg/api"
	"errors"
	"sort"
	"time"
)

type BillingService struct {
	dataStore      *model.DataStore
	persistentStore *store.PersistentStore
}

func NewBillingService(dataStore *model.DataStore, persistentStore *store.PersistentStore) *BillingService {
	return &BillingService{
		dataStore:       dataStore,
		persistentStore: persistentStore,
	}
}

func (s *BillingService) save() error {
	if s.persistentStore == nil {
		return nil
	}
	data := s.dataStore.ToStoredData()
	return s.persistentStore.Save(data)
}

func (s *BillingService) CreatePlan(req *api.CreatePlanRequest) (*api.Plan, error) {
	plan := &api.Plan{
		Name:           req.Name,
		MonthlyFee:     req.MonthlyFee,
		SmsQuota:       req.SmsQuota,
		StorageQuotaGB: req.StorageQuotaGB,
		Description:    req.Description,
	}

	created, err := s.dataStore.CreatePlan(plan)
	if err != nil {
		return nil, err
	}

	if err := s.save(); err != nil {
		return nil, err
	}
	return created, nil
}

func (s *BillingService) GetPlan(id string) (*api.Plan, error) {
	plan, ok := s.dataStore.GetPlan(id)
	if !ok {
		return nil, model.ErrPlanNotFound
	}
	return plan, nil
}

func (s *BillingService) ListPlans() []*api.Plan {
	return s.dataStore.ListPlans()
}

func (s *BillingService) CreateCustomer(req *api.CreateCustomerRequest) (*api.Customer, error) {
	if _, ok := s.dataStore.GetPlan(req.PlanID); !ok {
		return nil, model.ErrPlanNotFound
	}

	customer := &api.Customer{
		Name:          req.Name,
		Email:         req.Email,
		CurrentPlanID: req.PlanID,
	}

	created, err := s.dataStore.CreateCustomer(customer)
	if err != nil {
		return nil, err
	}

	if err := s.save(); err != nil {
		return nil, err
	}
	return created, nil
}

func (s *BillingService) GetCustomer(id string) (*api.Customer, error) {
	customer, ok := s.dataStore.GetCustomer(id)
	if !ok {
		return nil, model.ErrCustomerNotFound
	}
	return customer, nil
}

func (s *BillingService) ListCustomers() []*api.Customer {
	return s.dataStore.ListCustomers()
}

func (s *BillingService) ChangePlan(req *api.ChangePlanRequest) (*api.PlanChange, error) {
	if _, ok := s.dataStore.GetCustomer(req.CustomerID); !ok {
		return nil, model.ErrCustomerNotFound
	}

	if _, ok := s.dataStore.GetPlan(req.NewPlanID); !ok {
		return nil, model.ErrPlanNotFound
	}

	customer, _ := s.dataStore.GetCustomer(req.CustomerID)

	planChange := &api.PlanChange{
		CustomerID:    req.CustomerID,
		FromPlanID:    customer.CurrentPlanID,
		ToPlanID:      req.NewPlanID,
		EffectiveDate: req.EffectiveDate,
	}

	created, err := s.dataStore.RecordPlanChange(planChange)
	if err != nil {
		return nil, err
	}

	if err := s.dataStore.UpdateCustomerPlan(req.CustomerID, req.NewPlanID); err != nil {
		return nil, err
	}

	if err := s.save(); err != nil {
		return nil, err
	}
	return created, nil
}

func (s *BillingService) RecordUsage(req *api.RecordUsageRequest) (*api.UsageRecord, error) {
	if _, ok := s.dataStore.GetCustomer(req.CustomerID); !ok {
		return nil, model.ErrCustomerNotFound
	}

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	var storageUpdate *float64
	if req.StorageUpdate > 0 {
		storageUpdate = &req.StorageUpdate
	}

	usage, err := s.dataStore.CreateOrUpdateUsage(req.CustomerID, year, month, req.SmsIncrement, storageUpdate)
	if err != nil {
		return nil, err
	}

	if err := s.save(); err != nil {
		return nil, err
	}
	return usage, nil
}

func (s *BillingService) GetUsage(customerID string, year, month int) (*api.UsageRecord, error) {
	usage, ok := s.dataStore.GetUsageRecord(customerID, year, month)
	if !ok {
		return nil, model.ErrUsageNotFound
	}
	return usage, nil
}

func (s *BillingService) GenerateBill(req *api.GenerateBillRequest) (*api.Bill, error) {
	if req.CustomerID != "" {
		return s.generateSingleCustomerBill(req.CustomerID, req.Year, req.Month)
	}
	return nil, errors.New("batch generation not implemented yet")
}

func (s *BillingService) generateSingleCustomerBill(customerID string, year, month int) (*api.Bill, error) {
	if _, ok := s.dataStore.GetCustomer(customerID); !ok {
		return nil, model.ErrCustomerNotFound
	}

	if existing, _ := s.dataStore.GetCustomerBill(customerID, year, month); existing != nil {
		return nil, model.ErrBillAlreadyExists
	}

	customer, _ := s.dataStore.GetCustomer(customerID)

	pricing := s.dataStore.GetPricing()

	daysInMonth := daysIn(year, month)
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	lastDay := time.Date(year, time.Month(month), daysInMonth, 23, 59, 59, 0, time.Local)

	planBreakdown, planFee := s.calculatePlanFee(customer, year, month)

	smsOverageFee := s.calculateSmsOverage(customerID, year, month, pricing)

	storageOverageFee := s.calculateStorageOverage(customerID, year, month, pricing)

	totalAmount := planFee + smsOverageFee + storageOverageFee

	generatedAt := time.Now()
	dueDate := generatedAt.AddDate(0, 0, pricing.PaymentDueDays)

	billStatus := api.BillStatusUnpaid
	if s.dataStore.HasUnpaidPreviousBill(customerID, year, month) {
		billStatus = api.BillStatusOverdue
	}

	bill := &api.Bill{
		CustomerID:        customerID,
		Year:              year,
		Month:             month,
		PlanFee:           planFee,
		SmsOverageFee:     smsOverageFee,
		StorageOverageFee: storageOverageFee,
		TotalAmount:       totalAmount,
		Status:            billStatus,
		GeneratedAt:       generatedAt,
		DueDate:           dueDate,
		PlanBreakdown:     planBreakdown,
	}

	_ = firstDay
	_ = lastDay

	created, err := s.dataStore.CreateBill(bill)
	if err != nil {
		return nil, err
	}

	if err := s.save(); err != nil {
		return nil, err
	}
	return created, nil
}

func (s *BillingService) calculatePlanFee(customer *api.Customer, year, month int) ([]api.PlanProration, float64) {
	planChanges := s.dataStore.GetCustomerPlanChanges(customer.ID)

	daysInMonth := daysIn(year, month)
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	lastDay := time.Date(year, time.Month(month), daysInMonth, 23, 59, 59, 0, time.Local)

	type period struct {
		start time.Time
		end   time.Time
		planID string
	}

	periods := []period{}

	currentPlanID := customer.CurrentPlanID
	var currentStart time.Time = firstDay

	sort.Slice(planChanges, func(i, j int) bool {
		return planChanges[i].EffectiveDate.Before(planChanges[j].EffectiveDate)
	})

	for _, change := range planChanges {
		if change.EffectiveDate.After(lastDay) {
			continue
		}
		if change.EffectiveDate.Before(firstDay) {
			currentPlanID = change.ToPlanID
			continue
		}

		end := change.EffectiveDate.AddDate(0, 0, -1)
		if !end.Before(currentStart) {
			periods = append(periods, period{
				start:  currentStart,
				end:    end,
				planID: currentPlanID,
			})
		}
		currentStart = change.EffectiveDate
		currentPlanID = change.ToPlanID
	}

	if !currentStart.After(lastDay) {
		periods = append(periods, period{
			start:  currentStart,
			end:    lastDay,
			planID: currentPlanID,
		})
	}

	if len(periods) == 0 {
		periods = append(periods, period{
			start:  firstDay,
			end:    lastDay,
			planID: customer.CurrentPlanID,
		})
	}

	var breakdown []api.PlanProration
	totalFee := 0.0

	for _, p := range periods {
		days := countDaysInRange(p.start, p.end, year, month)
		if days <= 0 {
			continue
		}

		plan, ok := s.dataStore.GetPlan(p.planID)
		if !ok {
			continue
		}

		proratedFee := (plan.MonthlyFee / float64(daysInMonth)) * float64(days)

		breakdown = append(breakdown, api.PlanProration{
			PlanID:    p.planID,
			StartDate: p.start,
			EndDate:   p.end,
			Days:      days,
			Fee:       proratedFee,
		})

		totalFee += proratedFee
	}

	return breakdown, totalFee
}

func (s *BillingService) calculateSmsOverage(customerID string, year, month int, pricing *api.PricingConfig) float64 {
	usage, ok := s.dataStore.GetUsageRecord(customerID, year, month)
	if !ok {
		return 0
	}

	customer, ok := s.dataStore.GetCustomer(customerID)
	if !ok {
		return 0
	}

	plan, ok := s.dataStore.GetPlan(customer.CurrentPlanID)
	if !ok {
		return 0
	}

	if usage.SmsCount > plan.SmsQuota {
		overage := usage.SmsCount - plan.SmsQuota
		return float64(overage) * pricing.SmsPricePerUnit
	}
	return 0
}

func (s *BillingService) calculateStorageOverage(customerID string, year, month int, pricing *api.PricingConfig) float64 {
	usage, ok := s.dataStore.GetUsageRecord(customerID, year, month)
	if !ok {
		return 0
	}

	customer, ok := s.dataStore.GetCustomer(customerID)
	if !ok {
		return 0
	}

	plan, ok := s.dataStore.GetPlan(customer.CurrentPlanID)
	if !ok {
		return 0
	}

	if usage.StorageUsageGB > plan.StorageQuotaGB {
		overage := usage.StorageUsageGB - plan.StorageQuotaGB
		return overage * pricing.StoragePricePerGB
	}
	return 0
}

func (s *BillingService) GetBill(id string) (*api.Bill, error) {
	bill, ok := s.dataStore.GetBill(id)
	if !ok {
		return nil, model.ErrBillNotFound
	}
	return bill, nil
}

func (s *BillingService) GetCustomerBill(customerID string, year, month int) (*api.Bill, error) {
	bill, ok := s.dataStore.GetCustomerBill(customerID, year, month)
	if !ok {
		return nil, model.ErrBillNotFound
	}
	return bill, nil
}

func (s *BillingService) ListCustomerBills(customerID string) []*api.Bill {
	return s.dataStore.ListCustomerBills(customerID)
}

func (s *BillingService) ListAllBills() []*api.Bill {
	return s.dataStore.ListAllBills()
}

func (s *BillingService) MarkPaid(req *api.MarkPaidRequest) (*api.Bill, error) {
	bill, ok := s.dataStore.GetBill(req.BillID)
	if !ok {
		return nil, model.ErrBillNotFound
	}

	if bill.Status == api.BillStatusPaid {
		return bill, nil
	}

	payment := &api.PaymentRecord{
		BillID:   req.BillID,
		Operator: req.Operator,
		MarkedAt: time.Now(),
		Remark:   req.Remark,
		Amount:   req.Amount,
	}

	if err := s.dataStore.UpdateBillPayment(req.BillID, payment); err != nil {
		return nil, err
	}

	if err := s.save(); err != nil {
		return nil, err
	}

	updated, _ := s.dataStore.GetBill(req.BillID)
	return updated, nil
}

func (s *BillingService) GetPricing() *api.PricingConfig {
	return s.dataStore.GetPricing()
}

func (s *BillingService) UpdatePricing(req *api.UpdatePricingRequest) error {
	s.dataStore.UpdatePricing(req)
	return s.save()
}

func daysIn(year, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.Local).Day()
}

func countDaysInRange(start, end time.Time, year, month int) int {
	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	monthEnd := time.Date(year, time.Month(month)+1, 0, 23, 59, 59, 0, time.Local)

	if end.Before(monthStart) || start.After(monthEnd) {
		return 0
	}

	actualStart := maxTime(start, monthStart)
	actualEnd := minTime(end, monthEnd)

	days := 0
	for d := actualStart; !d.After(actualEnd); d = d.AddDate(0, 0, 1) {
		days++
	}
	return days
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

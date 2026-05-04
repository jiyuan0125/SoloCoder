package model

import (
	"billing/internal/server/store"
	"billing/pkg/api"
)

func (s *DataStore) ToStoredData() *store.StoredData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := &store.StoredData{}

	data.Plans = make([]*api.Plan, 0, len(s.Plans))
	for _, p := range s.Plans {
		data.Plans = append(data.Plans, p)
	}

	data.Customers = make([]*api.Customer, 0, len(s.Customers))
	for _, c := range s.Customers {
		data.Customers = append(data.Customers, c)
	}

	data.PlanChanges = make([]*api.PlanChange, 0, len(s.PlanChanges))
	for _, pc := range s.PlanChanges {
		data.PlanChanges = append(data.PlanChanges, pc)
	}

	data.UsageRecords = make([]*api.UsageRecord, 0, len(s.UsageRecords))
	for _, u := range s.UsageRecords {
		data.UsageRecords = append(data.UsageRecords, u)
	}

	data.Bills = make([]*api.Bill, 0, len(s.Bills))
	for _, b := range s.Bills {
		data.Bills = append(data.Bills, b)
	}

	data.Pricing = &api.PricingConfig{
		SmsPricePerUnit:    s.Pricing.SmsPricePerUnit,
		StoragePricePerGB:  s.Pricing.StoragePricePerGB,
		PaymentDueDays:     s.Pricing.PaymentDueDays,
		SeriousOverdueDays: s.Pricing.SeriousOverdueDays,
	}

	return data
}

func (s *DataStore) FromStoredData(data *store.StoredData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Plans = make(map[string]*api.Plan)
	for _, p := range data.Plans {
		s.Plans[p.ID] = p
	}

	s.Customers = make(map[string]*api.Customer)
	for _, c := range data.Customers {
		s.Customers[c.ID] = c
	}

	s.PlanChanges = make(map[string]*api.PlanChange)
	s.customerPlanChanges = make(map[string][]*api.PlanChange)
	for _, pc := range data.PlanChanges {
		s.PlanChanges[pc.ID] = pc
		s.customerPlanChanges[pc.CustomerID] = append(
			s.customerPlanChanges[pc.CustomerID],
			pc,
		)
	}

	s.UsageRecords = make(map[string]*api.UsageRecord)
	for _, u := range data.UsageRecords {
		key := usageKey(u.CustomerID, u.Year, u.Month)
		s.UsageRecords[key] = u
	}

	s.Bills = make(map[string]*api.Bill)
	for _, b := range data.Bills {
		s.Bills[b.ID] = b
	}

	if data.Pricing != nil {
		s.Pricing = &api.PricingConfig{
			SmsPricePerUnit:    data.Pricing.SmsPricePerUnit,
			StoragePricePerGB:  data.Pricing.StoragePricePerGB,
			PaymentDueDays:     data.Pricing.PaymentDueDays,
			SeriousOverdueDays: data.Pricing.SeriousOverdueDays,
		}
	} else {
		s.Pricing = DefaultPricingConfig()
	}
}

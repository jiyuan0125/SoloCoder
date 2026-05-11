package core

import (
	"math"
	"sync"
	"time"
	"recycling/api"
)

type RecordStore struct {
	mu      sync.RWMutex
	records map[string]*api.Record
}

func NewRecordStore() *RecordStore {
	return &RecordStore{
		records: make(map[string]*api.Record),
	}
}

func (rs *RecordStore) Create(record *api.Record) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.records[record.ID] = record
}

func (rs *RecordStore) GetByID(id string) (*api.Record, bool) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	r, exists := rs.records[id]
	if !exists {
		return nil, false
	}
	return cloneRecord(r), true
}

func (rs *RecordStore) GetAll() []api.Record {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	result := make([]api.Record, 0, len(rs.records))
	for _, r := range rs.records {
		result = append(result, *cloneRecord(r))
	}
	return result
}

func cloneRecord(r *api.Record) *api.Record {
	items := make([]api.RecordItemDetail, len(r.Items))
	for i, item := range r.Items {
		items[i] = item
	}
	return &api.Record{
		ID:           r.ID,
		DateTime:     r.DateTime,
		CustomerID:   r.CustomerID,
		CustomerName: r.CustomerName,
		CustomerType: r.CustomerType,
		Settlement:   r.Settlement,
		Items:        items,
		TotalAmount:  r.TotalAmount,
	}
}

func CalculateAmount(weight, price float64) float64 {
	baseWeight := 100.0
	premiumRate := 1.10

	if weight <= baseWeight {
		amount := weight * price
		return roundToCents(amount)
	}

	baseAmount := baseWeight * price
	excessWeight := weight - baseWeight
	excessAmount := excessWeight * price * premiumRate
	total := baseAmount + excessAmount
	return roundToCents(total)
}

func roundToCents(amount float64) float64 {
	return math.Round(amount*100) / 100
}

func (rs *RecordStore) GetRecordsByDateRange(start, end time.Time) []api.Record {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	var result []api.Record
	for _, r := range rs.records {
		if (r.DateTime.Equal(start) || r.DateTime.After(start)) &&
			(r.DateTime.Equal(end) || r.DateTime.Before(end)) {
			result = append(result, *cloneRecord(r))
		}
	}
	return result
}

func (rs *RecordStore) GetRecordsByMonth(year, month int) []api.Record {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	return rs.GetRecordsByDateRange(start, end)
}

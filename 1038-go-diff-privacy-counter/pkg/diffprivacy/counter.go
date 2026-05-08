package diffprivacy

import (
	"errors"
	"math"
	"sync"
)

type Record struct {
	Category string
}

type QueryResult struct {
	Count         int
	NoisyCount    float64
	FinalCount    int
	EpsilonUsed   float64
	IsExact       bool
}

type DiffPrivacyCounter struct {
	mu            sync.RWMutex
	records       []Record
	countsByCat   map[string]int
	budgetManager *BudgetManager
}

func NewDiffPrivacyCounter(initialBudget float64) *DiffPrivacyCounter {
	return &DiffPrivacyCounter{
		records:       make([]Record, 0),
		countsByCat:   make(map[string]int),
		budgetManager: NewBudgetManager(initialBudget),
	}
}

func (d *DiffPrivacyCounter) AddRecord(category string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.records = append(d.records, Record{Category: category})
	d.countsByCat[category]++
}

func (d *DiffPrivacyCounter) AddRecords(categories []string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, cat := range categories {
		d.records = append(d.records, Record{Category: cat})
		d.countsByCat[cat]++
	}
}

func (d *DiffPrivacyCounter) QueryCount(category string, epsilon float64) (*QueryResult, error) {
	if epsilon < 0 {
		return nil, errors.New("epsilon cannot be negative")
	}

	isExact := epsilon == 0

	if !isExact {
		if err := d.budgetManager.Consume(epsilon); err != nil {
			return nil, err
		}
	}

	d.mu.RLock()
	exactCount := d.countsByCat[category]
	d.mu.RUnlock()

	sensitivity := 1
	var noisyCount float64
	var finalCount int

	if isExact {
		noisyCount = float64(exactCount)
		finalCount = exactCount
	} else {
		noisyCount = AddLaplaceNoise(exactCount, epsilon, sensitivity)
		finalCount = int(math.Max(0, math.Round(noisyCount)))
	}

	return &QueryResult{
		Count:       exactCount,
		NoisyCount:  noisyCount,
		FinalCount:  finalCount,
		EpsilonUsed: epsilon,
		IsExact:     isExact,
	}, nil
}

func (d *DiffPrivacyCounter) RemainingBudget() float64 {
	return d.budgetManager.Remaining()
}

func (d *DiffPrivacyCounter) TotalBudget() float64 {
	return d.budgetManager.Total()
}

func (d *DiffPrivacyCounter) RechargeBudget(amount float64) error {
	return d.budgetManager.Recharge(amount)
}

func (d *DiffPrivacyCounter) AllCategories() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	categories := make([]string, 0, len(d.countsByCat))
	for cat := range d.countsByCat {
		categories = append(categories, cat)
	}
	return categories
}

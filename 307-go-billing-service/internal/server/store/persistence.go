package store

import (
	"billing/pkg/api"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type PersistentStore struct {
	dataDir      string
	plansFile    string
	customersFile string
	changesFile  string
	usageFile    string
	billsFile    string
	pricingFile  string
	mu           sync.Mutex
}

type StoredData struct {
	Plans               []*api.Plan         `json:"plans"`
	Customers           []*api.Customer     `json:"customers"`
	PlanChanges         []*api.PlanChange   `json:"plan_changes"`
	UsageRecords        []*api.UsageRecord  `json:"usage_records"`
	Bills               []*api.Bill          `json:"bills"`
	Pricing             *api.PricingConfig   `json:"pricing"`
}

func NewPersistentStore(dataDir string) (*PersistentStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	return &PersistentStore{
		dataDir:       dataDir,
		plansFile:     filepath.Join(dataDir, "plans.json"),
		customersFile: filepath.Join(dataDir, "customers.json"),
		changesFile:   filepath.Join(dataDir, "plan_changes.json"),
		usageFile:     filepath.Join(dataDir, "usage_records.json"),
		billsFile:     filepath.Join(dataDir, "bills.json"),
		pricingFile:   filepath.Join(dataDir, "pricing.json"),
	}, nil
}

func (p *PersistentStore) Save(data *StoredData) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	files := map[string]interface{}{
		p.plansFile:     data.Plans,
		p.customersFile: data.Customers,
		p.changesFile:   data.PlanChanges,
		p.usageFile:     data.UsageRecords,
		p.billsFile:     data.Bills,
		p.pricingFile:   data.Pricing,
	}

	for path, d := range files {
		if err := p.saveFile(path, d); err != nil {
			return err
		}
	}
	return nil
}

func (p *PersistentStore) saveFile(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (p *PersistentStore) Load() (*StoredData, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	data := &StoredData{}

	if err := p.loadFile(p.plansFile, &data.Plans); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err := p.loadFile(p.customersFile, &data.Customers); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err := p.loadFile(p.changesFile, &data.PlanChanges); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err := p.loadFile(p.usageFile, &data.UsageRecords); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err := p.loadFile(p.billsFile, &data.Bills); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err := p.loadFile(p.pricingFile, &data.Pricing); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return data, nil
}

func (p *PersistentStore) loadFile(path string, data interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(data)
}

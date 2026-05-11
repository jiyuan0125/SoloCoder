package core

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SolidWasteManager struct {
	records map[string]*SolidWasteRecord
	lock    sync.RWMutex
}

func NewSolidWasteManager() *SolidWasteManager {
	return &SolidWasteManager{
		records: make(map[string]*SolidWasteRecord),
	}
}

func (m *SolidWasteManager) AddRecord(name string, category WasteCategory, amount float64, storageLocation string, generatedAt time.Time) (*SolidWasteRecord, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	
	record := &SolidWasteRecord{
		ID:              uuid.New().String(),
		Name:            name,
		Category:        category,
		Amount:          amount,
		StorageLocation: storageLocation,
		GeneratedAt:     generatedAt,
		Status:          WasteStatusStored,
	}
	
	m.lock.Lock()
	m.records[record.ID] = record
	m.lock.Unlock()
	
	return record, nil
}

func (m *SolidWasteManager) DisposeRecord(id, method string) (*SolidWasteRecord, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	
	record, exists := m.records[id]
	if !exists {
		return nil, errors.New("record not found")
	}
	
	if record.Status == WasteStatusDisposed {
		return nil, errors.New("record already disposed")
	}
	
	now := time.Now()
	record.DisposalMethod = method
	record.DisposedAt = &now
	record.Status = WasteStatusDisposed
	
	return record, nil
}

func (m *SolidWasteManager) GetRecord(id string) (*SolidWasteRecord, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	
	record, exists := m.records[id]
	return record, exists
}

func (m *SolidWasteManager) CheckExpiration() []*Alarm {
	m.lock.RLock()
	defer m.lock.RUnlock()
	
	alarms := make([]*Alarm, 0)
	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)
	
	for _, record := range m.records {
		if record.Category == WasteCategoryHazardous &&
			record.Status == WasteStatusStored &&
			record.GeneratedAt.Before(oneYearAgo) {
			
			alarm := &Alarm{
				ID:         uuid.New().String(),
				Type:       AlarmTypeStorageExpired,
				Level:      AlarmLevelWarning,
				EntityType: EntityTypeSolidWaste,
				EntityID:   record.ID,
				RelatedData: map[string]interface{}{
					"name":      record.Name,
					"generatedAt": record.GeneratedAt.Format("2006-01-02"),
				},
				Message:  "危险废物贮存超过一年",
				CreatedAt: now,
			}
			alarms = append(alarms, alarm)
			
			record.Status = WasteStatusExpired
		}
	}
	
	return alarms
}

func (m *SolidWasteManager) GetRecords(category *WasteCategory, status *WasteStatus) []*SolidWasteRecord {
	m.lock.RLock()
	defer m.lock.RUnlock()
	
	records := make([]*SolidWasteRecord, 0)
	for _, record := range m.records {
		if category != nil && record.Category != *category {
			continue
		}
		if status != nil && record.Status != *status {
			continue
		}
		records = append(records, record)
	}
	return records
}

func (m *SolidWasteManager) UpdateRecord(id string, updates map[string]interface{}) (*SolidWasteRecord, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	
	record, exists := m.records[id]
	if !exists {
		return nil, errors.New("record not found")
	}
	
	if name, exists := updates["name"].(string); exists {
		record.Name = name
	}
	
	if storageLocation, exists := updates["storage_location"].(string); exists {
		record.StorageLocation = storageLocation
	}
	
	if disposalMethod, exists := updates["disposal_method"].(string); exists {
		record.DisposalMethod = disposalMethod
	}
	
	return record, nil
}

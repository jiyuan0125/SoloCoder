package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"luggage-tracking/common"
)

const (
	dataDir       = "data"
	luggageFile   = "luggages.json"
	operationFile = "operations.json"
	auditFile     = "audits.json"
)

type Store struct {
	luggages      map[string]*common.Luggage
	operationLogs []common.OperationLog
	auditLogs     []common.AuditLog
	mu            sync.RWMutex
}

func NewStore() *Store {
	store := &Store{
		luggages:      make(map[string]*common.Luggage),
		operationLogs: []common.OperationLog{},
		auditLogs:     []common.AuditLog{},
	}
	store.load()
	return store
}

func (s *Store) load() {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return
	}

	s.loadLuggages()
	s.loadOperations()
	s.loadAudits()
}

func (s *Store) loadLuggages() {
	path := filepath.Join(dataDir, luggageFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var luggages []*common.Luggage
	if err := json.Unmarshal(data, &luggages); err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, l := range luggages {
		s.luggages[l.TagNumber] = l
	}
}

func (s *Store) loadOperations() {
	path := filepath.Join(dataDir, operationFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var logs []common.OperationLog
	if err := json.Unmarshal(data, &logs); err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.operationLogs = logs
}

func (s *Store) loadAudits() {
	path := filepath.Join(dataDir, auditFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var logs []common.AuditLog
	if err := json.Unmarshal(data, &logs); err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.auditLogs = logs
}

func (s *Store) saveLuggages(luggages []*common.Luggage) {
	data, err := json.MarshalIndent(luggages, "", "  ")
	if err != nil {
		return
	}

	path := filepath.Join(dataDir, luggageFile)
	os.WriteFile(path, data, 0644)
}

func (s *Store) saveOperations(logs []common.OperationLog) {
	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return
	}

	path := filepath.Join(dataDir, operationFile)
	os.WriteFile(path, data, 0644)
}

func (s *Store) saveAudits(logs []common.AuditLog) {
	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return
	}

	path := filepath.Join(dataDir, auditFile)
	os.WriteFile(path, data, 0644)
}

func (s *Store) CreateLuggage(luggage *common.Luggage) error {
	s.mu.Lock()

	if existing, exists := s.luggages[luggage.TagNumber]; exists {
		if !existing.Completed {
			s.mu.Unlock()
			return common.ErrLuggageAlreadyExists
		}
	}

	s.luggages[luggage.TagNumber] = luggage

	luggages := make([]*common.Luggage, 0, len(s.luggages))
	for _, l := range s.luggages {
		luggages = append(luggages, l)
	}

	s.mu.Unlock()

	s.saveLuggages(luggages)
	return nil
}

func (s *Store) GetLuggage(tagNumber string) (*common.Luggage, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	luggage, exists := s.luggages[tagNumber]
	return luggage, exists
}

func (s *Store) UpdateLuggage(luggage *common.Luggage) error {
	s.mu.Lock()

	if _, exists := s.luggages[luggage.TagNumber]; !exists {
		s.mu.Unlock()
		return common.ErrLuggageNotFound
	}

	s.luggages[luggage.TagNumber] = luggage

	luggages := make([]*common.Luggage, 0, len(s.luggages))
	for _, l := range s.luggages {
		luggages = append(luggages, l)
	}

	s.mu.Unlock()

	s.saveLuggages(luggages)
	return nil
}

func (s *Store) GetLuggagesByFlight(flightNumber string, date string) []*common.Luggage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*common.Luggage
	for _, l := range s.luggages {
		if l.FlightNumber == flightNumber {
			if date == "" || l.CreatedAt.Format("2006-01-02") == date {
				result = append(result, l)
			}
		}
	}
	return result
}

func (s *Store) GetActiveLuggages() []*common.Luggage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*common.Luggage
	for _, l := range s.luggages {
		if !l.Completed && l.Status != common.StatusFlightCancel {
			result = append(result, l)
		}
	}
	return result
}

func (s *Store) AddOperationLog(log common.OperationLog) {
	s.mu.Lock()

	log.ID = generateID()
	log.CreatedAt = time.Now()
	s.operationLogs = append(s.operationLogs, log)

	logs := make([]common.OperationLog, len(s.operationLogs))
	copy(logs, s.operationLogs)

	s.mu.Unlock()

	s.saveOperations(logs)
}

func (s *Store) GetOperationLogs() []common.OperationLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logs := make([]common.OperationLog, len(s.operationLogs))
	copy(logs, s.operationLogs)
	return logs
}

func (s *Store) AddAuditLog(log common.AuditLog) {
	s.mu.Lock()

	log.ID = generateID()
	log.CreatedAt = time.Now()
	log.Immutable = true
	s.auditLogs = append(s.auditLogs, log)

	logs := make([]common.AuditLog, len(s.auditLogs))
	copy(logs, s.auditLogs)

	s.mu.Unlock()

	s.saveAudits(logs)
}

func (s *Store) GetAuditLogs() []common.AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logs := make([]common.AuditLog, len(s.auditLogs))
	copy(logs, s.auditLogs)
	return logs
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

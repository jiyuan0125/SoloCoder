package store

import (
	"quality-trace/pkg/common"
	"sync"
)

type Store struct {
	mu sync.RWMutex

	batches          map[string]*common.Batch
	batchesByDate    map[string]map[string]int
	processFlows     map[string]*common.ProcessFlow
	processRecords   []*common.ProcessRecord
	inspectionSpecs  []*common.InspectionSpec
	inspectionRecords []*common.InspectionRecord

	processRecordIDCounter   int64
	inspectionSpecIDCounter  int64
	inspectionRecordIDCounter int64
}

func New() *Store {
	return &Store{
		batches:         make(map[string]*common.Batch),
		batchesByDate:   make(map[string]map[string]int),
		processFlows:    make(map[string]*common.ProcessFlow),
		processRecords:  []*common.ProcessRecord{},
		inspectionSpecs: []*common.InspectionSpec{},
		inspectionRecords: []*common.InspectionRecord{},
	}
}

func (s *Store) GetNextBatchNumber(factoryCode string, date string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := factoryCode + "-" + date
	if _, exists := s.batchesByDate[key]; !exists {
		s.batchesByDate[key] = make(map[string]int)
		s.batchesByDate[key]["next"] = 1
	}
	num := s.batchesByDate[key]["next"]
	s.batchesByDate[key]["next"] = num + 1
	return num
}

func (s *Store) CreateBatch(batch *common.Batch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.batches[batch.BatchID]; exists {
		return ErrBatchExists
	}
	s.batches[batch.BatchID] = batch
	return nil
}

func (s *Store) GetBatch(batchID string) (*common.Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	batch, exists := s.batches[batchID]
	if !exists {
		return nil, ErrBatchNotFound
	}
	return batch, nil
}

func (s *Store) UpdateBatch(batch *common.Batch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.batches[batch.BatchID]; !exists {
		return ErrBatchNotFound
	}
	s.batches[batch.BatchID] = batch
	return nil
}

func (s *Store) ListBatches(filter func(*common.Batch) bool) []*common.Batch {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*common.Batch
	for _, batch := range s.batches {
		if filter == nil || filter(batch) {
			result = append(result, batch)
		}
	}
	return result
}

func (s *Store) CreateProcessFlow(flow *common.ProcessFlow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.processFlows[flow.ProductName] = flow
	return nil
}

func (s *Store) GetProcessFlow(productName string) (*common.ProcessFlow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	flow, exists := s.processFlows[productName]
	if !exists {
		return nil, ErrProcessFlowNotFound
	}
	return flow, nil
}

func (s *Store) CreateProcessRecord(record *common.ProcessRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.processRecordIDCounter++
	record.ID = s.processRecordIDCounter
	s.processRecords = append(s.processRecords, record)
	return nil
}

func (s *Store) UpdateProcessRecord(record *common.ProcessRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, r := range s.processRecords {
		if r.ID == record.ID {
			s.processRecords[i] = record
			return nil
		}
	}
	return ErrProcessRecordNotFound
}

func (s *Store) ListProcessRecords(filter func(*common.ProcessRecord) bool) []*common.ProcessRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*common.ProcessRecord
	for _, record := range s.processRecords {
		if filter == nil || filter(record) {
			result = append(result, record)
		}
	}
	return result
}

func (s *Store) GetLatestProcessRecordForBatch(batchID string) (*common.ProcessRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latest *common.ProcessRecord
	for _, record := range s.processRecords {
		if record.BatchID == batchID {
			if latest == nil || record.Sequence > latest.Sequence || (record.Sequence == latest.Sequence && record.IsRework && !latest.IsRework) {
				latest = record
			}
		}
	}
	return latest, latest != nil
}

func (s *Store) CreateInspectionSpec(spec *common.InspectionSpec) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.inspectionSpecIDCounter++
	spec.ID = s.inspectionSpecIDCounter
	s.inspectionSpecs = append(s.inspectionSpecs, spec)
	return nil
}

func (s *Store) GetInspectionSpecs(productName string, inspectionType common.InspectionType) []*common.InspectionSpec {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*common.InspectionSpec
	for _, spec := range s.inspectionSpecs {
		if spec.ProductName == productName && spec.InspectionType == inspectionType {
			result = append(result, spec)
		}
	}
	return result
}

func (s *Store) CreateInspectionRecord(record *common.InspectionRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.inspectionRecordIDCounter++
	record.ID = s.inspectionRecordIDCounter
	s.inspectionRecords = append(s.inspectionRecords, record)
	return nil
}

func (s *Store) ListInspectionRecords(filter func(*common.InspectionRecord) bool) []*common.InspectionRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*common.InspectionRecord
	for _, record := range s.inspectionRecords {
		if filter == nil || filter(record) {
			result = append(result, record)
		}
	}
	return result
}

func (s *Store) HasFinalInspection(batchID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, record := range s.inspectionRecords {
		if record.BatchID == batchID && record.InspectionType == common.InspectionTypeFinal {
			return true
		}
	}
	return false
}

func (s *Store) LockBatch(batchID string) func() {
	s.mu.Lock()
	return func() { s.mu.Unlock() }
}

package process

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"quality-trace/pkg/common"
	"quality-trace/pkg/core/batch"
	"quality-trace/pkg/core/store"
)

type Service struct {
	store    *store.Store
	batchSvc *batch.Service
	mu       sync.Mutex
}

func NewService(s *store.Store, batchSvc *batch.Service) *Service {
	return &Service{
		store:    s,
		batchSvc: batchSvc,
	}
}

func (s *Service) AddProcessFlow(req common.AddProcessFlowRequest) (*common.ProcessFlow, error) {
	if req.ProductName == "" || len(req.Processes) == 0 {
		return nil, fmt.Errorf("invalid request parameters")
	}

	sorted := make([]common.ProcessDef, len(req.Processes))
	copy(sorted, req.Processes)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Sequence < sorted[j].Sequence
	})

	for i := 1; i < len(sorted); i++ {
		if sorted[i].Sequence <= sorted[i-1].Sequence {
			return nil, fmt.Errorf("duplicate sequence numbers not allowed")
		}
	}

	flow := &common.ProcessFlow{
		ProductName: req.ProductName,
		Processes:   sorted,
	}

	err := s.store.CreateProcessFlow(flow)
	if err != nil {
		return nil, err
	}

	return flow, nil
}

func (s *Service) GetProcessFlow(productName string) (*common.ProcessFlow, error) {
	return s.store.GetProcessFlow(productName)
}

func (s *Service) getNextProcessDef(batch *common.Batch) (*common.ProcessDef, error) {
	flow, err := s.store.GetProcessFlow(batch.ProductName)
	if err != nil {
		return nil, err
	}

	latestRecord, hasRecord := s.store.GetLatestProcessRecordForBatch(batch.BatchID)

	if !hasRecord {
		return &flow.Processes[0], nil
	}

	if latestRecord.Result == common.ProcessResultFail {
		return &common.ProcessDef{
			Sequence: latestRecord.Sequence,
			ProcessName: latestRecord.ProcessName,
		}, nil
	}

	for i, proc := range flow.Processes {
		if proc.Sequence == latestRecord.Sequence {
			if i == len(flow.Processes)-1 {
				return nil, fmt.Errorf("all processes completed")
			}
			return &flow.Processes[i+1], nil
		}
	}

	return nil, fmt.Errorf("process not found in flow")
}

func (s *Service) StartProcess(req common.StartProcessRequest) (*common.ProcessRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := s.batchSvc.GetBatch(req.BatchID)
	if err != nil {
		return nil, err
	}

	if b.Status == common.BatchStatusScrapped {
		return nil, fmt.Errorf("batch is already scrapped")
	}
	if b.Status == common.BatchStatusPaused {
		return nil, fmt.Errorf("batch is paused, waiting for quality decision")
	}

	nextProc, err := s.getNextProcessDef(b)
	if err != nil {
		return nil, err
	}

	latest, hasLatest := s.store.GetLatestProcessRecordForBatch(b.BatchID)
	isRework := false
	if hasLatest && latest.Sequence == nextProc.Sequence && latest.Result == common.ProcessResultFail {
		isRework = true
	}

	if b.Status == common.BatchStatusCreated {
		_, err = s.batchSvc.UpdateBatchStatus(b.BatchID, common.BatchStatusInProgress)
		if err != nil {
			return nil, err
		}
	}

	record := &common.ProcessRecord{
		BatchID:     b.BatchID,
		Sequence:    nextProc.Sequence,
		ProcessName: nextProc.ProcessName,
		Operator:    req.Operator,
		StartTime:   req.StartTime,
		EndTime:     nil,
		Result:      "",
		IsRework:    isRework,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.store.CreateProcessRecord(record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (s *Service) CompleteProcess(req common.CompleteProcessRequest) (*common.ProcessRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := s.batchSvc.GetBatch(req.BatchID)
	if err != nil {
		return nil, err
	}

	if b.Status == common.BatchStatusScrapped {
		return nil, fmt.Errorf("batch is already scrapped")
	}

	records := s.store.ListProcessRecords(func(r *common.ProcessRecord) bool {
		return r.BatchID == req.BatchID && r.EndTime == nil
	})

	if len(records) == 0 {
		return nil, fmt.Errorf("no active process found for batch")
	}

	record := records[0]
	record.EndTime = &req.EndTime
	record.Result = req.Result
	record.UpdatedAt = time.Now()

	err = s.store.UpdateProcessRecord(record)
	if err != nil {
		return nil, err
	}

	if req.Result == common.ProcessResultFail {
		_, err = s.batchSvc.UpdateBatchStatus(b.BatchID, common.BatchStatusPaused)
		if err != nil {
			return nil, err
		}
	}

	return record, nil
}

func (s *Service) MakeReworkDecision(req common.ReworkDecisionRequest) (*common.Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := s.batchSvc.GetBatch(req.BatchID)
	if err != nil {
		return nil, err
	}

	if b.Status != common.BatchStatusPaused {
		return nil, fmt.Errorf("batch is not paused")
	}

	if req.Decision == "scrap" {
		return s.batchSvc.ScrapeBatch(b.BatchID)
	}

	if req.Decision == "rework" {
		b, err = s.batchSvc.IncrementReworkCount(b.BatchID)
		if err != nil {
			return nil, err
		}

		if b.ReworkCount > common.MaxReworkCount {
			return s.batchSvc.ScrapeBatch(b.BatchID)
		}

		_, err = s.batchSvc.UpdateBatchStatus(b.BatchID, common.BatchStatusInProgress)
		if err != nil {
			return nil, err
		}

		return s.batchSvc.GetBatch(b.BatchID)
	}

	return nil, fmt.Errorf("invalid decision: must be 'rework' or 'scrap'")
}

func (s *Service) ListProcessRecords(batchID string) []*common.ProcessRecord {
	return s.store.ListProcessRecords(func(r *common.ProcessRecord) bool {
		return r.BatchID == batchID
	})
}

func (s *Service) GetAllProcessRecords() []*common.ProcessRecord {
	return s.store.ListProcessRecords(nil)
}

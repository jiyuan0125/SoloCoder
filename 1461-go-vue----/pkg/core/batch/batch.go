package batch

import (
	"fmt"
	"math"
	"time"

	"quality-trace/pkg/common"
	"quality-trace/pkg/core/store"
)

type Service struct {
	store *store.Store
}

func NewService(s *store.Store) *Service {
	return &Service{store: s}
}

func (s *Service) generateBatchID(factoryCode string, date time.Time) string {
	dateStr := date.Format("20060102")
	seq := s.store.GetNextBatchNumber(factoryCode, dateStr)
	return fmt.Sprintf("%s%s%03d", factoryCode, dateStr, seq)
}

func (s *Service) CreateBatch(req common.CreateBatchRequest) (*common.Batch, error) {
	if req.ProductName == "" || req.FactoryCode == "" || req.PlanQuantity <= 0 {
		return nil, fmt.Errorf("invalid request parameters")
	}

	batchID := s.generateBatchID(req.FactoryCode, req.StartTime)

	batch := &common.Batch{
		BatchID:        batchID,
		ProductName:    req.ProductName,
		FactoryCode:    req.FactoryCode,
		PlanQuantity:   req.PlanQuantity,
		ActualQuantity: 0,
		StartTime:      req.StartTime,
		EndTime:        nil,
		Status:         common.BatchStatusCreated,
		IsAbnormal:     false,
		IsOneTimePass:  true,
		ReworkCount:    0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := s.store.CreateBatch(batch)
	if err != nil {
		return nil, err
	}

	return batch, nil
}

func (s *Service) GetBatch(batchID string) (*common.Batch, error) {
	return s.store.GetBatch(batchID)
}

func (s *Service) CompleteBatch(req common.CompleteBatchRequest) (*common.Batch, error) {
	batch, err := s.store.GetBatch(req.BatchID)
	if err != nil {
		return nil, err
	}

	if batch.Status == common.BatchStatusScrapped {
		return nil, fmt.Errorf("batch is already scrapped")
	}

	deviation := math.Abs(float64(req.ActualQuantity-batch.PlanQuantity)) / float64(batch.PlanQuantity)

	batch.ActualQuantity = req.ActualQuantity
	batch.EndTime = &req.EndTime
	batch.Status = common.BatchStatusCompleted
	batch.IsAbnormal = deviation > common.DeviationThreshold
	batch.UpdatedAt = time.Now()

	if batch.IsAbnormal {
		batch.Status = common.BatchStatusAbnormal
	}

	err = s.store.UpdateBatch(batch)
	if err != nil {
		return nil, err
	}

	return batch, nil
}

func (s *Service) ListBatches(filter common.ListBatchesRequest) []*common.Batch {
	return s.store.ListBatches(func(b *common.Batch) bool {
		if filter.Status != "" && b.Status != filter.Status {
			return false
		}
		if filter.ProductName != "" && b.ProductName != filter.ProductName {
			return false
		}
		if filter.StartTime != nil && b.StartTime.Before(*filter.StartTime) {
			return false
		}
		if filter.EndTime != nil && b.StartTime.After(*filter.EndTime) {
			return false
		}
		return true
	})
}

func (s *Service) ScrapeBatch(batchID string) (*common.Batch, error) {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return nil, err
	}

	batch.Status = common.BatchStatusScrapped
	batch.UpdatedAt = time.Now()

	err = s.store.UpdateBatch(batch)
	if err != nil {
		return nil, err
	}

	return batch, nil
}

func (s *Service) UpdateBatchStatus(batchID string, status common.BatchStatus) (*common.Batch, error) {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return nil, err
	}

	batch.Status = status
	batch.UpdatedAt = time.Now()

	err = s.store.UpdateBatch(batch)
	if err != nil {
		return nil, err
	}

	return batch, nil
}

func (s *Service) IncrementReworkCount(batchID string) (*common.Batch, error) {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return nil, err
	}

	batch.ReworkCount++
	batch.IsOneTimePass = false
	batch.UpdatedAt = time.Now()

	err = s.store.UpdateBatch(batch)
	if err != nil {
		return nil, err
	}

	return batch, nil
}

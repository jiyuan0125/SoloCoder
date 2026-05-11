package inspection

import (
	"fmt"
	"time"

	"quality-trace/pkg/common"
	"quality-trace/pkg/core/batch"
	"quality-trace/pkg/core/store"
)

type Service struct {
	store    *store.Store
	batchSvc *batch.Service
}

func NewService(s *store.Store, batchSvc *batch.Service) *Service {
	return &Service{
		store:    s,
		batchSvc: batchSvc,
	}
}

func (s *Service) AddInspectionSpec(req common.AddInspectionSpecRequest) (*common.InspectionSpec, error) {
	if req.ProductName == "" || req.MetricName == "" {
		return nil, fmt.Errorf("invalid request parameters")
	}

	if req.LowerLimit > req.UpperLimit {
		return nil, fmt.Errorf("lower limit cannot be greater than upper limit")
	}

	spec := &common.InspectionSpec{
		ProductName:    req.ProductName,
		InspectionType: req.InspectionType,
		MetricName:     req.MetricName,
		LowerLimit:     req.LowerLimit,
		UpperLimit:     req.UpperLimit,
	}

	err := s.store.CreateInspectionSpec(spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func (s *Service) GetInspectionSpecs(productName string, inspectionType common.InspectionType) []*common.InspectionSpec {
	return s.store.GetInspectionSpecs(productName, inspectionType)
}

func (s *Service) evaluateInspection(actualValue, lowerLimit, upperLimit float64) common.InspectionResult {
	if actualValue >= lowerLimit && actualValue <= upperLimit {
		return common.InspectionResultPass
	}
	return common.InspectionResultFail
}

func (s *Service) AddInspectionRecord(req common.AddInspectionRequest) (*common.InspectionRecord, error) {
	if req.BatchID == "" || req.MetricName == "" {
		return nil, fmt.Errorf("invalid request parameters")
	}

	b, err := s.batchSvc.GetBatch(req.BatchID)
	if err != nil {
		return nil, err
	}

	if b.Status == common.BatchStatusScrapped {
		return nil, fmt.Errorf("batch is already scrapped")
	}

	specs := s.store.GetInspectionSpecs(b.ProductName, req.InspectionType)

	var lowerLimit, upperLimit float64
	found := false
	for _, spec := range specs {
		if spec.MetricName == req.MetricName {
			lowerLimit = spec.LowerLimit
			upperLimit = spec.UpperLimit
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("inspection spec not found for metric: %s", req.MetricName)
	}

	result := s.evaluateInspection(req.ActualValue, lowerLimit, upperLimit)

	record := &common.InspectionRecord{
		BatchID:        req.BatchID,
		InspectionType: req.InspectionType,
		MetricName:     req.MetricName,
		ActualValue:    req.ActualValue,
		LowerLimit:     lowerLimit,
		UpperLimit:     upperLimit,
		Result:         result,
		InspectedBy:    req.InspectedBy,
		InspectedAt:    req.InspectedAt,
		CreatedAt:      time.Now(),
	}

	err = s.store.CreateInspectionRecord(record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (s *Service) ListInspectionRecords(filter common.ListInspectionRecordsRequest) []*common.InspectionRecord {
	return s.store.ListInspectionRecords(func(r *common.InspectionRecord) bool {
		if filter.BatchID != "" && r.BatchID != filter.BatchID {
			return false
		}
		if filter.InspectionType != "" && r.InspectionType != filter.InspectionType {
			return false
		}
		return true
	})
}

func (s *Service) GetAllInspectionRecords() []*common.InspectionRecord {
	return s.store.ListInspectionRecords(nil)
}

func (s *Service) HasFinalInspection(batchID string) bool {
	return s.store.HasFinalInspection(batchID)
}

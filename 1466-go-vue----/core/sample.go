package core

import (
	"errors"
	"fmt"
	"laboratory/common"
	"time"
)

func (s *Store) CreateSample(req *common.CreateSampleRequest) (*common.CreateSampleResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	isLarge := req.Quantity > 5 || req.Volume > LargeSampleVolume

	date := time.Now()
	sampleID := fmt.Sprintf("LB%s%03d", date.Format("20060102"), s.sampleSeq)
	s.sampleSeq++

	sample := &Sample{
		ID:            sampleID,
		Name:          req.Name,
		Quantity:      req.Quantity,
		Customer:      req.Customer,
		DeliveryDate:  req.DeliveryDate,
		Requirements:  req.Requirements,
		Volume:        req.Volume,
		IsLargeSample: isLarge,
		Status:        common.SampleStatusPending,
		StatusLogs: []common.StatusLogEntry{
			{
				Time:     time.Now(),
				Status:   common.SampleStatusPending,
				Operator: "system",
				Remark:   "样品创建",
			},
		},
	}

	s.samples[sampleID] = sample

	return &common.CreateSampleResponse{
		SampleID:      sampleID,
		IsLargeSample: isLarge,
	}, nil
}

func (s *Store) ListSamples(status common.SampleStatus) []*common.SampleInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*common.SampleInfo, 0)
	for _, sample := range s.samples {
		if status != "" && sample.Status != status {
			continue
		}
		result = append(result, sampleToInfo(sample))
	}
	return result
}

func (s *Store) GetSample(sampleID string) (*common.SampleInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sample, ok := s.samples[sampleID]
	if !ok {
		return nil, errors.New("样品不存在")
	}
	return sampleToInfo(sample), nil
}

func (s *Store) CreateCabinet(req *common.CreateCabinetRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.cabinets[req.ID]; ok {
		return errors.New("样品柜已存在")
	}

	s.cabinets[req.ID] = &Cabinet{
		ID:        req.ID,
		Capacity:  req.Capacity,
		Used:      0,
		IsSpecial: req.IsSpecial,
	}
	return nil
}

func (s *Store) ListCabinets() []*common.CabinetInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*common.CabinetInfo, 0)
	for _, cab := range s.cabinets {
		result = append(result, &common.CabinetInfo{
			ID:        cab.ID,
			Capacity:  cab.Capacity,
			Used:      cab.Used,
			IsSpecial: cab.IsSpecial,
		})
	}
	return result
}

func (s *Store) AssignCabinet(req *common.AssignCabinetRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sample, ok := s.samples[req.SampleID]
	if !ok {
		return errors.New("样品不存在")
	}

	cabinet, ok := s.cabinets[req.CabinetID]
	if !ok {
		return errors.New("样品柜不存在")
	}

	if sample.IsLargeSample && !cabinet.IsSpecial {
		return errors.New("大样必须分配到特殊存储柜")
	}

	if !sample.IsLargeSample && cabinet.IsSpecial {
		return errors.New("普通样品不能分配到特殊存储柜")
	}

	cabinet.mu.Lock()
	defer cabinet.mu.Unlock()

	if cabinet.Used+sample.Quantity > cabinet.Capacity {
		return errors.New("样品柜容量不足")
	}

	if sample.CabinetID != "" {
		if oldCab, ok := s.cabinets[sample.CabinetID]; ok {
			oldCab.mu.Lock()
			oldCab.Used -= sample.Quantity
			oldCab.mu.Unlock()
		}
	}

	cabinet.Used += sample.Quantity
	sample.CabinetID = req.CabinetID

	return nil
}

func (s *Store) UpdateSampleStatus(req *common.UpdateSampleStatusRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sample, ok := s.samples[req.SampleID]
	if !ok {
		return errors.New("样品不存在")
	}

	sample.Status = req.NewStatus
	sample.StatusLogs = append(sample.StatusLogs, common.StatusLogEntry{
		Time:     time.Now(),
		Status:   req.NewStatus,
		Operator: req.Operator,
		Remark:   req.Remark,
	})

	return nil
}

func sampleToInfo(s *Sample) *common.SampleInfo {
	return &common.SampleInfo{
		ID:            s.ID,
		Name:          s.Name,
		Quantity:      s.Quantity,
		Customer:      s.Customer,
		DeliveryDate:  s.DeliveryDate,
		Requirements:  s.Requirements,
		Volume:        s.Volume,
		IsLargeSample: s.IsLargeSample,
		Status:        s.Status,
		CabinetID:     s.CabinetID,
		StatusLogs:    append([]common.StatusLogEntry(nil), s.StatusLogs...),
	}
}

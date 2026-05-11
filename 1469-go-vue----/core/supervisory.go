package core

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"supervision-log-system/common"
)

type SupervisoryService struct {
	store *Store
}

func NewSupervisoryService(store *Store) *SupervisoryService {
	return &SupervisoryService{store: store}
}

func (s *SupervisoryService) Create(req *common.CreateSupervisoryRecordRequest) (*common.SupervisoryRecord, error) {
	if req.EndTime.Before(req.StartTime) {
		return nil, errors.New("结束时间不能早于开始时间")
	}

	duration := req.EndTime.Sub(req.StartTime)
	durationMinutes := int(duration.Minutes())
	isAbnormal := durationMinutes < 60

	record := &common.SupervisoryRecord{
		ID:               uuid.NewString(),
		ProjectPart:      req.ProjectPart,
		Process:          req.Process,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		DurationMinutes:  durationMinutes,
		IsAbnormal:       isAbnormal,
		ConstructionDesc: req.ConstructionDesc,
		IssuesFound:      req.IssuesFound,
		Supervisor:       req.Supervisor,
		Supplements:      make([]string, 0),
		CreatedAt:        time.Now(),
		Submitted:        false,
	}

	s.store.Lock()
	defer s.store.Unlock()
	s.store.GetSupervisoryRecords()[record.ID] = record

	return record, nil
}

func (s *SupervisoryService) Submit(id string) error {
	s.store.Lock()
	defer s.store.Unlock()

	record, exists := s.store.GetSupervisoryRecords()[id]
	if !exists {
		return errors.New("旁站记录不存在")
	}

	if record.Submitted {
		return errors.New("旁站记录已提交，不可重复提交")
	}

	record.Submitted = true
	return nil
}

func (s *SupervisoryService) AddSupplement(req *common.AddSupervisorySupplementRequest) error {
	s.store.Lock()
	defer s.store.Unlock()

	record, exists := s.store.GetSupervisoryRecords()[req.ID]
	if !exists {
		return errors.New("旁站记录不存在")
	}

	record.Supplements = append(record.Supplements, req.Supplement)
	return nil
}

func (s *SupervisoryService) GetByID(id string) (*common.SupervisoryRecord, error) {
	s.store.RLock()
	defer s.store.RUnlock()

	record, exists := s.store.GetSupervisoryRecords()[id]
	if !exists {
		return nil, errors.New("旁站记录不存在")
	}

	return record, nil
}

func (s *SupervisoryService) GetBySupervisorAndDate(supervisor string, date time.Time) []*common.SupervisoryRecord {
	s.store.RLock()
	defer s.store.RUnlock()

	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.AddDate(0, 0, 1)

	var results []*common.SupervisoryRecord
	for _, record := range s.store.GetSupervisoryRecords() {
		if record.Supervisor == supervisor &&
			record.StartTime.After(start) &&
			record.StartTime.Before(end) {
			results = append(results, record)
		}
	}

	return results
}

func (s *SupervisoryService) List() []*common.SupervisoryRecord {
	s.store.RLock()
	defer s.store.RUnlock()

	var results []*common.SupervisoryRecord
	for _, record := range s.store.GetSupervisoryRecords() {
		results = append(results, record)
	}

	return results
}

package core

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"supervision-log-system/common"
)

type AcceptanceService struct {
	store *Store
}

func NewAcceptanceService(store *Store) *AcceptanceService {
	return &AcceptanceService{store: store}
}

func (s *AcceptanceService) Create(req *common.CreateAcceptanceRecordRequest) (*common.AcceptanceRecord, error) {
	if len(req.Supervisors) < 2 {
		return nil, errors.New("验收至少需要两名监理人员参与")
	}

	record := &common.AcceptanceRecord{
		ID:             uuid.NewString(),
		ProjectPart:    req.ProjectPart,
		Items:          req.Items,
		Conclusion:     req.Conclusion,
		Status:         req.Status,
		Supervisors:    req.Supervisors,
		AcceptanceDate: req.AcceptanceDate,
		RecheckCount:   0,
		CreatedAt:      time.Now(),
	}

	s.store.Lock()
	defer s.store.Unlock()
	s.store.GetAcceptanceRecords()[record.ID] = record

	return record, nil
}

func (s *AcceptanceService) CreateRectificationNotice(req *common.CreateRectificationNoticeRequest) (*common.RectificationNotice, error) {
	s.store.RLock()
	_, exists := s.store.GetAcceptanceRecords()[req.AcceptanceID]
	s.store.RUnlock()
	if !exists {
		return nil, errors.New("验收记录不存在")
	}

	notice := &common.RectificationNotice{
		ID:           uuid.NewString(),
		AcceptanceID: req.AcceptanceID,
		Contractor:   req.Contractor,
		Requirements: req.Requirements,
		Deadline:     req.Deadline,
		Status:       "pending",
		CreatedAt:    time.Now(),
	}

	s.store.Lock()
	defer s.store.Unlock()

	s.store.GetRectificationNotices()[notice.ID] = notice
	s.store.GetAcceptanceRecords()[req.AcceptanceID].RectificationID = notice.ID
	s.store.GetAcceptanceRecords()[req.AcceptanceID].Status = common.AcceptanceStatusFailed

	return notice, nil
}

func (s *AcceptanceService) CompleteRectification(req *common.CompleteRectificationRequest) error {
	s.store.Lock()
	defer s.store.Unlock()

	notice, exists := s.store.GetRectificationNotices()[req.ID]
	if !exists {
		return errors.New("整改通知不存在")
	}

	if notice.Status != "pending" {
		return errors.New("整改通知状态不正确")
	}

	now := time.Now()
	notice.Status = "completed"
	notice.CompletedAt = &now

	acceptance, ok := s.store.GetAcceptanceRecords()[notice.AcceptanceID]
	if ok {
		acceptance.Status = common.AcceptanceStatusRectified
	}

	return nil
}

func (s *AcceptanceService) Recheck(req *common.RecheckAcceptanceRequest) (*common.AcceptanceRecord, error) {
	s.store.RLock()
	record, exists := s.store.GetAcceptanceRecords()[req.AcceptanceID]
	s.store.RUnlock()
	if !exists {
		return nil, errors.New("验收记录不存在")
	}

	if len(req.Supervisors) < 2 {
		return nil, errors.New("复查至少需要两名监理人员参与")
	}

	if record.RecheckCount >= 3 {
		s.store.Lock()
		record.Status = common.AcceptanceStatusReported
		s.store.Unlock()
		return nil, errors.New("复查次数已达三次，已上报建设单位处理")
	}

	s.store.Lock()
	defer s.store.Unlock()

	record.RecheckCount++
	record.Items = req.Items
	record.Conclusion = req.Conclusion
	record.Status = req.Status
	record.Supervisors = req.Supervisors
	record.AcceptanceDate = req.AcceptanceDate

	if record.RecheckCount >= 3 && req.Status != common.AcceptanceStatusPassed {
		record.Status = common.AcceptanceStatusReported
	}

	return record, nil
}

func (s *AcceptanceService) GetByID(id string) (*common.AcceptanceRecord, error) {
	s.store.RLock()
	defer s.store.RUnlock()

	record, exists := s.store.GetAcceptanceRecords()[id]
	if !exists {
		return nil, errors.New("验收记录不存在")
	}

	return record, nil
}

func (s *AcceptanceService) GetBySupervisorAndDate(supervisor string, date time.Time) []*common.AcceptanceRecord {
	s.store.RLock()
	defer s.store.RUnlock()

	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.AddDate(0, 0, 1)

	var results []*common.AcceptanceRecord
	for _, record := range s.store.GetAcceptanceRecords() {
		for _, s := range record.Supervisors {
			if s == supervisor &&
				record.AcceptanceDate.After(start) &&
				record.AcceptanceDate.Before(end) {
				results = append(results, record)
				break
			}
		}
	}

	return results
}

func (s *AcceptanceService) List() []*common.AcceptanceRecord {
	s.store.RLock()
	defer s.store.RUnlock()

	var results []*common.AcceptanceRecord
	for _, record := range s.store.GetAcceptanceRecords() {
		results = append(results, record)
	}

	return results
}

func (s *AcceptanceService) ListRectificationNotices() []*common.RectificationNotice {
	s.store.RLock()
	defer s.store.RUnlock()

	var results []*common.RectificationNotice
	for _, notice := range s.store.GetRectificationNotices() {
		results = append(results, notice)
	}

	return results
}

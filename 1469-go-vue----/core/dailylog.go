package core

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"supervision-log-system/common"
)

type DailyLogService struct {
	store             *Store
	supervisoryService *SupervisoryService
	acceptanceService  *AcceptanceService
}

func NewDailyLogService(
	store *Store,
	supervisoryService *SupervisoryService,
	acceptanceService *AcceptanceService,
) *DailyLogService {
	return &DailyLogService{
		store:              store,
		supervisoryService: supervisoryService,
		acceptanceService:  acceptanceService,
	}
}

func (s *DailyLogService) GenerateDraft(supervisor string, logDate time.Time) (*common.DailyLog, error) {
	key := s.getKey(supervisor, logDate)

	s.store.RLock()
	existing, exists := s.store.GetDailyLogs()[key]
	s.store.RUnlock()

	if exists {
		if existing.Submitted {
			return nil, errors.New("该日期的日志已提交，无法修改")
		}
		return existing, nil
	}

	supervisoryRecords := s.supervisoryService.GetBySupervisorAndDate(supervisor, logDate)
	acceptanceRecords := s.acceptanceService.GetBySupervisorAndDate(supervisor, logDate)

	log := &common.DailyLog{
		ID:                 uuid.NewString(),
		Supervisor:         supervisor,
		LogDate:            logDate,
		SupervisoryRecord:  supervisoryRecords,
		AcceptanceRecords:  acceptanceRecords,
		AdditionalContent:  "",
		Submitted:          false,
		CreatedAt:          time.Now(),
	}

	s.store.Lock()
	defer s.store.Unlock()
	s.store.GetDailyLogs()[key] = log

	return log, nil
}

func (s *DailyLogService) Submit(req *common.SubmitDailyLogRequest) (*common.DailyLog, error) {
	key := s.getKey(req.Supervisor, req.LogDate)

	s.store.Lock()
	defer s.store.Unlock()

	log, exists := s.store.GetDailyLogs()[key]
	if !exists {
		supervisoryRecords := s.supervisoryService.GetBySupervisorAndDate(req.Supervisor, req.LogDate)
		acceptanceRecords := s.acceptanceService.GetBySupervisorAndDate(req.Supervisor, req.LogDate)

		log = &common.DailyLog{
			ID:                 uuid.NewString(),
			Supervisor:         req.Supervisor,
			LogDate:            req.LogDate,
			SupervisoryRecord:  supervisoryRecords,
			AcceptanceRecords:  acceptanceRecords,
			AdditionalContent:  req.AdditionalContent,
			Submitted:          true,
			CreatedAt:          time.Now(),
		}

		s.store.GetDailyLogs()[key] = log
		return log, nil
	}

	if log.Submitted {
		return nil, errors.New("该日期的日志已提交，不可重复提交")
	}

	log.AdditionalContent = req.AdditionalContent
	log.Submitted = true

	return log, nil
}

func (s *DailyLogService) Get(supervisor string, logDate time.Time) (*common.DailyLog, error) {
	key := s.getKey(supervisor, logDate)

	s.store.RLock()
	defer s.store.RUnlock()

	log, exists := s.store.GetDailyLogs()[key]
	if !exists {
		return nil, errors.New("该日期的日志不存在")
	}

	return log, nil
}

func (s *DailyLogService) List() []*common.DailyLog {
	s.store.RLock()
	defer s.store.RUnlock()

	var results []*common.DailyLog
	for _, log := range s.store.GetDailyLogs() {
		results = append(results, log)
	}

	return results
}

func (s *DailyLogService) getKey(supervisor string, logDate time.Time) string {
	dateStr := logDate.Format("2006-01-02")
	return supervisor + "_" + dateStr
}

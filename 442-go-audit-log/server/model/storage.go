package model

import (
	"audit-log/common"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

type InMemoryStorage struct {
	logs         map[string]*common.AuditLog
	logsByMonth  map[string][]*common.AuditLog
	logsByUser   map[string][]*common.AuditLog
	logsByType   map[string][]*common.AuditLog
	logsByIP     map[string][]*common.AuditLog
	archivedLogs map[string][]*common.AuditLog

	userOperations     map[string][]time.Time
	userLoginFailures  map[string]int
	userLockedUntil    map[string]time.Time

	exportApprovals    map[string]*ExportApproval

	mu               sync.RWMutex
	totalCapacity    int64
	usedCapacity     int64
}

type ExportApproval struct {
	UserID     string
	UserName   string
	ApproverID string
	Approved   bool
	CreatedAt  time.Time
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		logs:              make(map[string]*common.AuditLog),
		logsByMonth:       make(map[string][]*common.AuditLog),
		logsByUser:        make(map[string][]*common.AuditLog),
		logsByType:        make(map[string][]*common.AuditLog),
		logsByIP:          make(map[string][]*common.AuditLog),
		archivedLogs:      make(map[string][]*common.AuditLog),
		userOperations:    make(map[string][]time.Time),
		userLoginFailures: make(map[string]int),
		userLockedUntil:   make(map[string]time.Time),
		exportApprovals:   make(map[string]*ExportApproval),
		totalCapacity:     1024 * 1024 * 1024,
		usedCapacity:      0,
	}
}

func generateMonthPartition(t time.Time) string {
	return t.Format("2006-01")
}

func generateLogID() string {
	return uuid.New().String()
}

func (s *InMemoryStorage) CreateLog(req *common.CreateLogRequest) (*common.AuditLog, error) {
	if req.UserID == "" {
		return nil, common.ErrMissingUserID
	}
	if req.OperationType == "" {
		return nil, common.ErrMissingOperationType
	}

	now := time.Now()
	log := &common.AuditLog{
		ID:             generateLogID(),
		UserID:         req.UserID,
		UserName:       req.UserName,
		IPAddress:      req.IPAddress,
		OperationTime:  now,
		OperationType:  req.OperationType,
		Description:    req.Description,
		Result:         req.Result,
		RecordIDs:      req.RecordIDs,
		BeforeValue:    req.BeforeValue,
		AfterValue:     req.AfterValue,
		Snapshot:       req.Snapshot,
		ExportRange:    req.ExportRange,
		Approver:       req.Approver,
		MonthPartition: generateMonthPartition(now),
	}

	logSize := s.estimateLogSize(log)
	if s.usedCapacity+logSize > s.totalCapacity {
		return nil, common.ErrStorageCapacity
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logs[log.ID] = log
	s.logsByMonth[log.MonthPartition] = append(s.logsByMonth[log.MonthPartition], log)
	s.logsByUser[log.UserID] = append(s.logsByUser[log.UserID], log)
	s.logsByType[log.OperationType] = append(s.logsByType[log.OperationType], log)
	s.logsByIP[log.IPAddress] = append(s.logsByIP[log.IPAddress], log)

	s.usedCapacity += logSize

	if req.Result == common.ResultFailed {
		s.archivedLogs[log.MonthPartition] = s.archivedLogs[log.MonthPartition]
	}

	return log, nil
}

func (s *InMemoryStorage) GetLogByID(id string) (*common.AuditLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if log, exists := s.logs[id]; exists {
		return log, nil
	}
	return nil, common.ErrLogNotFound
}

func (s *InMemoryStorage) QueryLogs(req *common.QueryLogsRequest) (*common.QueryLogsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var candidateLogs []*common.AuditLog

	if req.UserID != "" {
		candidateLogs = s.logsByUser[req.UserID]
	} else if req.OperationType != "" {
		candidateLogs = s.logsByType[req.OperationType]
	} else if req.IPAddress != "" {
		candidateLogs = s.logsByIP[req.IPAddress]
	} else {
		for _, logs := range s.logsByMonth {
			candidateLogs = append(candidateLogs, logs...)
		}
	}

	var filteredLogs []*common.AuditLog
	for _, log := range candidateLogs {
		if req.UserID != "" && log.UserID != req.UserID {
			continue
		}
		if req.OperationType != "" && log.OperationType != req.OperationType {
			continue
		}
		if req.IPAddress != "" && log.IPAddress != req.IPAddress {
			continue
		}
		if req.IsAbnormal && !log.IsAbnormal {
			continue
		}
		if !req.StartTime.IsZero() && log.OperationTime.Before(req.StartTime) {
			continue
		}
		if !req.EndTime.IsZero() && log.OperationTime.After(req.EndTime) {
			continue
		}
		filteredLogs = append(filteredLogs, log)
	}

	total := len(filteredLogs)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize

	if start >= total {
		return &common.QueryLogsResponse{
			Logs:     []*common.AuditLog{},
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		}, nil
	}

	if end > total {
		end = total
	}

	return &common.QueryLogsResponse{
		Logs:     filteredLogs[start:end],
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *InMemoryStorage) estimateLogSize(log *common.AuditLog) int64 {
	data, _ := json.Marshal(log)
	return int64(len(data))
}

func (s *InMemoryStorage) RecordUserOperation(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.userOperations[userID] = append(s.userOperations[userID], now)

	cutoff := now.Add(-time.Minute)
	var validOps []time.Time
	for _, op := range s.userOperations[userID] {
		if op.After(cutoff) {
			validOps = append(validOps, op)
		}
	}
	s.userOperations[userID] = validOps
}

func (s *InMemoryStorage) GetUserOperationCountInLastMinute(userID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.userOperations[userID])
}

func (s *InMemoryStorage) MarkLogAsAbnormal(logID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if log, exists := s.logs[logID]; exists {
		log.IsAbnormal = true
		log.AbnormalReason = reason
		return nil
	}
	return common.ErrLogNotFound
}

func (s *InMemoryStorage) GetStorageStatus() *common.StorageStatusResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logCount := int64(len(s.logs))
	archivedCount := int64(0)
	for _, logs := range s.archivedLogs {
		archivedCount += int64(len(logs))
	}

	usagePercent := float64(s.usedCapacity) / float64(s.totalCapacity)
	needAlert := usagePercent >= common.StorageAlertThreshold

	return &common.StorageStatusResponse{
		TotalCapacity: s.totalCapacity,
		UsedCapacity:  s.usedCapacity,
		UsagePercent:  usagePercent,
		NeedAlert:     needAlert,
		LogCount:      logCount,
		ArchivedCount: archivedCount,
	}
}

func (s *InMemoryStorage) GetTotalCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.logs)
}

func (s *InMemoryStorage) GetAllLogs() []*common.AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var all []*common.AuditLog
	for _, log := range s.logs {
		all = append(all, log)
	}
	return all
}

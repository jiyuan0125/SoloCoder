package model

import (
	"audit-log/common"
	"sync"
	"time"
)

type SecurityManager struct {
	storage *InMemoryStorage
	mu      sync.RWMutex
}

func NewSecurityManager(storage *InMemoryStorage) *SecurityManager {
	return &SecurityManager{
		storage: storage,
	}
}

func (sm *SecurityManager) RecordLoginAttempt(userID, userName, ipAddress string, success bool) (*common.LoginResponse, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.isUserLocked(userID) {
		return &common.LoginResponse{
			IsLocked:     true,
			LockUntil:    sm.storage.userLockedUntil[userID].Format(time.RFC3339),
			FailureCount: sm.storage.userLoginFailures[userID],
		}, common.ErrUserLocked
	}

	if success {
		delete(sm.storage.userLoginFailures, userID)
		return &common.LoginResponse{
			IsLocked:     false,
			FailureCount: 0,
		}, nil
	}

	sm.storage.userLoginFailures[userID]++

	failureCount := sm.storage.userLoginFailures[userID]
	if failureCount >= common.MaxLoginFailures {
		sm.storage.userLockedUntil[userID] = time.Now().Add(time.Minute * time.Duration(common.LockDurationMinutes))
		return &common.LoginResponse{
			IsLocked:     true,
			LockUntil:    sm.storage.userLockedUntil[userID].Format(time.RFC3339),
			FailureCount: failureCount,
		}, nil
	}

	return &common.LoginResponse{
		IsLocked:     false,
		FailureCount: failureCount,
	}, nil
}

func (sm *SecurityManager) isUserLocked(userID string) bool {
	lockUntil, exists := sm.storage.userLockedUntil[userID]
	if !exists {
		return false
	}
	if time.Now().After(lockUntil) {
		delete(sm.storage.userLockedUntil, userID)
		delete(sm.storage.userLoginFailures, userID)
		return false
	}
	return true
}

func (sm *SecurityManager) IsUserLocked(userID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.isUserLocked(userID)
}

func (sm *SecurityManager) CheckAbnormalBehavior(userID string) (isAbnormal bool, reason string) {
	count := sm.storage.GetUserOperationCountInLastMinute(userID)
	if count > common.MaxOperationsPerMinute {
		return true, "too many operations in one minute"
	}
	return false, ""
}

func (sm *SecurityManager) GetUserLockInfo(userID string) *common.LoginResponse {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.isUserLocked(userID) {
		return &common.LoginResponse{
			IsLocked:     true,
			LockUntil:    sm.storage.userLockedUntil[userID].Format(time.RFC3339),
			FailureCount: sm.storage.userLoginFailures[userID],
		}
	}

	return &common.LoginResponse{
		IsLocked:     false,
		FailureCount: sm.storage.userLoginFailures[userID],
	}
}

type ArchiveManager struct {
	storage *InMemoryStorage
	mu      sync.RWMutex
}

func NewArchiveManager(storage *InMemoryStorage) *ArchiveManager {
	return &ArchiveManager{
		storage: storage,
	}
}

func (am *ArchiveManager) ArchiveOldLogs() int {
	am.mu.Lock()
	defer am.mu.Unlock()

	cutoffDate := time.Now().AddDate(0, 0, -common.RetentionDays)
	archivedCount := 0

	for month, logs := range am.storage.logsByMonth {
		var activeLogs []*common.AuditLog
		var archivedLogs []*common.AuditLog

		for _, log := range logs {
			if log.OperationTime.Before(cutoffDate) {
				archivedLogs = append(archivedLogs, log)
				archivedCount++
			} else {
				activeLogs = append(activeLogs, log)
			}
		}

		if len(archivedLogs) > 0 {
			am.storage.logsByMonth[month] = activeLogs
			am.storage.archivedLogs[month] = append(am.storage.archivedLogs[month], archivedLogs...)
		}
	}

	return archivedCount
}

func (am *ArchiveManager) GetArchivedMonths() []*common.ArchiveInfo {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var result []*common.ArchiveInfo
	for month, logs := range am.storage.archivedLogs {
		if len(logs) == 0 {
			continue
		}
		result = append(result, &common.ArchiveInfo{
			MonthPartition: month,
			LogCount:       len(logs),
			CreatedAt:      logs[0].OperationTime,
			IsArchived:     true,
		})
	}
	return result
}

type StatisticsManager struct {
	storage *InMemoryStorage
	mu      sync.RWMutex
}

func NewStatisticsManager(storage *InMemoryStorage) *StatisticsManager {
	return &StatisticsManager{
		storage: storage,
	}
}

func (sm *StatisticsManager) GetStatistics(req *common.StatisticsRequest) (*common.StatisticsResponse, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	allLogs := sm.storage.GetAllLogs()

	filteredLogs := sm.filterLogsByTime(allLogs, req.StartTime, req.EndTime)

	return &common.StatisticsResponse{
		OperationFrequencyTrend:   sm.buildOperationFrequencyTrend(filteredLogs),
		AbnormalBehaviorList:      sm.buildAbnormalBehaviorList(filteredLogs),
		SensitiveOperationRanking: sm.buildSensitiveOperationRanking(filteredLogs),
	}, nil
}

func (sm *StatisticsManager) filterLogsByTime(logs []*common.AuditLog, startTime, endTime time.Time) []*common.AuditLog {
	var result []*common.AuditLog
	for _, log := range logs {
		if !startTime.IsZero() && log.OperationTime.Before(startTime) {
			continue
		}
		if !endTime.IsZero() && log.OperationTime.After(endTime) {
			continue
		}
		result = append(result, log)
	}
	return result
}

func (sm *StatisticsManager) buildOperationFrequencyTrend(logs []*common.AuditLog) *common.OperationFrequencyTrend {
	hourlyStats := make(map[string]*common.OperationTrendItem)

	for _, log := range logs {
		hourKey := log.OperationTime.Format("2006-01-02 15:00")
		if _, exists := hourlyStats[hourKey]; !exists {
			hourlyStats[hourKey] = &common.OperationTrendItem{
				TimePoint: hourKey,
			}
		}
		hourlyStats[hourKey].Count++
		if log.Result == common.ResultSuccess {
			hourlyStats[hourKey].SuccessCount++
		} else {
			hourlyStats[hourKey].FailedCount++
		}
	}

	var items []*common.OperationTrendItem
	for _, item := range hourlyStats {
		items = append(items, item)
	}

	return &common.OperationFrequencyTrend{
		Items: items,
	}
}

func (sm *StatisticsManager) buildAbnormalBehaviorList(logs []*common.AuditLog) *common.AbnormalBehaviorList {
	var items []*common.AbnormalBehaviorItem

	for _, log := range logs {
		if log.IsAbnormal {
			items = append(items, &common.AbnormalBehaviorItem{
				UserID:        log.UserID,
				UserName:      log.UserName,
				OperationTime: log.OperationTime,
				Reason:        log.AbnormalReason,
				LogID:         log.ID,
			})
		}
	}

	return &common.AbnormalBehaviorList{
		Items: items,
	}
}

func (sm *StatisticsManager) buildSensitiveOperationRanking(logs []*common.AuditLog) *common.SensitiveOperationRanking {
	opStats := make(map[string]*common.SensitiveOperationItem)

	for _, log := range logs {
		if common.SensitiveOperations[log.OperationType] {
			if _, exists := opStats[log.OperationType]; !exists {
				opStats[log.OperationType] = &common.SensitiveOperationItem{
					OperationType: log.OperationType,
				}
			}
			opStats[log.OperationType].Count++
			if log.Result != common.ResultSuccess {
				opStats[log.OperationType].FailedCount++
			}
		}
	}

	var items []*common.SensitiveOperationItem
	for _, item := range opStats {
		items = append(items, item)
	}

	return &common.SensitiveOperationRanking{
		Items: items,
	}
}

type ExportApprovalManager struct {
	storage *InMemoryStorage
	mu      sync.RWMutex
}

func NewExportApprovalManager(storage *InMemoryStorage) *ExportApprovalManager {
	return &ExportApprovalManager{
		storage: storage,
	}
}

func (eam *ExportApprovalManager) RequestApproval(req *common.ExportApprovalRequest) (*common.ExportApprovalResponse, error) {
	eam.mu.Lock()
	defer eam.mu.Unlock()

	key := req.UserID + ":" + time.Now().Format("2006010215")

	if req.Approval && req.ApproverID != "" {
		eam.storage.exportApprovals[key] = &ExportApproval{
			UserID:     req.UserID,
			UserName:   req.UserName,
			ApproverID: req.ApproverID,
			Approved:   true,
			CreatedAt:  time.Now(),
		}
		return &common.ExportApprovalResponse{Approved: true}, nil
	}

	return &common.ExportApprovalResponse{Approved: false}, common.ErrNeedApproval
}

func (eam *ExportApprovalManager) IsExportApproved(userID string) bool {
	eam.mu.RLock()
	defer eam.mu.RUnlock()

	key := userID + ":" + time.Now().Format("2006010215")
	if approval, exists := eam.storage.exportApprovals[key]; exists {
		if approval.Approved {
			return true
		}
	}
	return false
}

package main

import (
	"sync"
	"time"
)

type AuditResult string

const (
	ResultSuccess        AuditResult = "success"
	ResultUnauthorized   AuditResult = "unauthorized"
	ResultForbidden      AuditResult = "forbidden"
)

type AuditLog struct {
	Timestamp int64
	ClientID  string
	Path      string
	Result    AuditResult
}

const MaxAuditLogs = 10000

type AuditLogManager struct {
	mu    sync.RWMutex
	logs  []*AuditLog
	head  int
	count int
}

func NewAuditLogManager() *AuditLogManager {
	return &AuditLogManager{
		logs: make([]*AuditLog, MaxAuditLogs),
	}
}

func (alm *AuditLogManager) Log(clientID string, path string, result AuditResult) {
	alm.mu.Lock()
	defer alm.mu.Unlock()

	entry := &AuditLog{
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
		ClientID:  clientID,
		Path:      path,
		Result:    result,
	}

	alm.logs[alm.head] = entry
	alm.head = (alm.head + 1) % MaxAuditLogs
	if alm.count < MaxAuditLogs {
		alm.count++
	}
}

func (alm *AuditLogManager) List() []*AuditLog {
	alm.mu.RLock()
	defer alm.mu.RUnlock()

	result := make([]*AuditLog, 0, alm.count)

	if alm.count < MaxAuditLogs {
		for i := 0; i < alm.head; i++ {
			if alm.logs[i] != nil {
				result = append(result, alm.logs[i])
			}
		}
	} else {
		for i := alm.head; i < MaxAuditLogs; i++ {
			if alm.logs[i] != nil {
				result = append(result, alm.logs[i])
			}
		}
		for i := 0; i < alm.head; i++ {
			if alm.logs[i] != nil {
				result = append(result, alm.logs[i])
			}
		}
	}

	return result
}

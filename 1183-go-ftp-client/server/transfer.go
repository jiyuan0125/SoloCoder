package main

import (
	"context"
	"sync"
	"time"

	"ftp-client/common"
)

type transferJob struct {
	info   common.TransferInfo
	ctx    context.Context
	cancel context.CancelFunc
	config common.FTPConfig
}

type TransferManager struct {
	mu        sync.RWMutex
	transfers map[string]*transferJob
}

func NewTransferManager() *TransferManager {
	return &TransferManager{
		transfers: make(map[string]*transferJob),
	}
}

func (m *TransferManager) Add(id string, job *transferJob) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transfers[id] = job
}

func (m *TransferManager) Get(id string) (*transferJob, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, ok := m.transfers[id]
	return job, ok
}

func (m *TransferManager) List() []common.TransferInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]common.TransferInfo, 0, len(m.transfers))
	for _, job := range m.transfers {
		result = append(result, job.info)
	}
	return result
}

func (m *TransferManager) UpdateInfo(id string, info common.TransferInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job, ok := m.transfers[id]; ok {
		job.info = info
	}
}

func (m *TransferManager) UpdateProgress(id string, transferred int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job, ok := m.transfers[id]; ok {
		job.info.Transferred = transferred
	}
}

func (m *TransferManager) Cancel(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job, ok := m.transfers[id]; ok {
		if job.cancel != nil {
			job.cancel()
		}
		job.info.Status = common.StatusCancelled
		job.info.EndTime = time.Now()
		return true
	}
	return false
}

func (m *TransferManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.transfers, id)
}

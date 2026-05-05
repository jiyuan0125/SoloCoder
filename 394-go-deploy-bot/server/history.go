package main

import (
	"deploybot/common"
	"sync"
)

type HistoryManager struct {
	mu        sync.RWMutex
	history   []*common.Deployment
	maxLength int
}

func NewHistoryManager() *HistoryManager {
	return &HistoryManager{
		history:   make([]*common.Deployment, 0),
		maxLength: 100,
	}
}

func (hm *HistoryManager) Add(deployment *common.Deployment) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.history = append(hm.history, deployment)

	if len(hm.history) > hm.maxLength {
		hm.history = hm.history[len(hm.history)-hm.maxLength:]
	}

	Info("添加部署记录: %s, 状态: %s", deployment.ID, deployment.Status)
}

func (hm *HistoryManager) Update(deployment *common.Deployment) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	for i, d := range hm.history {
		if d.ID == deployment.ID {
			hm.history[i] = deployment
			Info("更新部署记录: %s, 状态: %s", deployment.ID, deployment.Status)
			return
		}
	}

	hm.history = append(hm.history, deployment)
}

func (hm *HistoryManager) Get(id string) *common.Deployment {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	for _, d := range hm.history {
		if d.ID == id {
			return d
		}
	}
	return nil
}

func (hm *HistoryManager) List() []*common.Deployment {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	result := make([]*common.Deployment, len(hm.history))
	for i := range hm.history {
		result[len(hm.history)-1-i] = hm.history[i]
	}
	return result
}

func (hm *HistoryManager) LastSuccess() *common.Deployment {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	for i := len(hm.history) - 1; i >= 0; i-- {
		if hm.history[i].Status == common.StatusSuccess && !hm.history[i].IsRollback {
			return hm.history[i]
		}
	}
	return nil
}

func (hm *HistoryManager) LastBackup() string {
	last := hm.LastSuccess()
	if last != nil && last.BackupPath != "" {
		return last.BackupPath
	}
	return ""
}

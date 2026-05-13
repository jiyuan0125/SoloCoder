package manager

import (
	"sync"
	"time"

	"registry/pkg/healthcheck"
	"registry/pkg/model"
	"registry/pkg/store"
)

const (
	heartbeatTimeout  = 30 * time.Second
	offlineRetention  = 24 * time.Hour
	checkInterval     = 5 * time.Second
)

type InstanceManager struct {
	store    *store.InstanceStore
	checker  *healthcheck.HealthChecker
	stopCh   chan struct{}
	wg       sync.WaitGroup
	once     sync.Once
}

func NewInstanceManager(s *store.InstanceStore, c *healthcheck.HealthChecker) *InstanceManager {
	return &InstanceManager{
		store:   s,
		checker: c,
		stopCh:  make(chan struct{}),
	}
}

func (m *InstanceManager) Start() {
	m.wg.Add(1)
	go m.periodicCheck()
}

func (m *InstanceManager) Stop() {
	m.once.Do(func() {
		close(m.stopCh)
	})
	m.wg.Wait()
}

func (m *InstanceManager) periodicCheck() {
	defer m.wg.Done()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkHeartbeats()
			m.checkDraining()
			m.cleanupOffline()
		}
	}
}

func (m *InstanceManager) checkHeartbeats() {
	now := time.Now()
	for _, inst := range m.store.ListAll() {
		status := inst.GetStatus()
		if status == model.StatusOffline {
			continue
		}

		if now.Sub(inst.LastHeartbeat) > heartbeatTimeout {
			m.markUnhealthy(inst)
		}
	}
}

func (m *InstanceManager) markUnhealthy(inst *model.Instance) {
	inst.HealthFail()
}

func (m *InstanceManager) checkDraining() {
	for _, inst := range m.store.ListAll() {
		if inst.GetStatus() == model.StatusDraining {
			if inst.GetActiveConnections() == 0 {
				inst.SetStatus(model.StatusOffline)
			}
		}
	}
}

func (m *InstanceManager) cleanupOffline() {
	now := time.Now()
	for _, inst := range m.store.ListAll() {
		if inst.GetStatus() == model.StatusOffline && inst.OfflineAt != nil {
			if now.Sub(*inst.OfflineAt) > offlineRetention {
				m.checker.Stop(inst.ID)
				m.store.Remove(inst.ID)
			}
		}
	}
}

func (m *InstanceManager) UpdateActiveConnections(id string, delta int) error {
	inst, ok := m.store.Get(id)
	if !ok {
		return nil
	}
	if delta > 0 {
		inst.IncrActiveConnections()
	} else if delta < 0 {
		inst.DecrActiveConnections()
	}
	return nil
}

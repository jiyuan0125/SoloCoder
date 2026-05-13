package stats

import (
	"context"
	"fmt"
	"sync"

	"connpool/internal/models"
	"connpool/internal/pool"
	"connpool/internal/storage"
)

type Manager struct {
	storage    *storage.SQLiteStorage
	poolMgr    *pool.Manager
	mu         sync.Mutex
}

func NewManager(storage *storage.SQLiteStorage, poolMgr *pool.Manager) *Manager {
	return &Manager{
		storage: storage,
		poolMgr: poolMgr,
	}
}

func (m *Manager) UpdateAndSaveStats(ctx context.Context) error {
	stats := m.poolMgr.GetStatistics()

	dbStats, err := m.storage.GetLatestStatistics(ctx)
	if err != nil {
		return fmt.Errorf("get db stats: %w", err)
	}

	if stats.TotalActive != dbStats.TotalActive ||
		stats.TotalIdle != dbStats.TotalIdle ||
		stats.TotalWaiting != dbStats.TotalWaiting ||
		stats.TotalMax != dbStats.TotalMax {
		return m.storage.SaveStatistics(ctx, stats)
	}

	return nil
}

func (m *Manager) AdjustQuotas(ctx context.Context, newTotal int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pools, err := m.storage.ListPools(ctx)
	if err != nil {
		return fmt.Errorf("list pools: %w", err)
	}

	if len(pools) == 0 {
		return nil
	}

	totalQuota := 0
	for _, p := range pools {
		if p.Quota > 0 {
			totalQuota += p.Quota
		} else {
			totalQuota += p.MaxConnections
		}
	}

	if totalQuota == 0 {
		totalQuota = 1
	}

	remaining := newTotal
	for i, p := range pools {
		var currentQuota int
		if p.Quota > 0 {
			currentQuota = p.Quota
		} else {
			currentQuota = p.MaxConnections
		}

		proportion := float64(currentQuota) / float64(totalQuota)
		allocated := int(float64(newTotal) * proportion)

		if i == len(pools)-1 {
			allocated = remaining
		}

		if allocated < 1 {
			allocated = 1
		}

		remaining -= allocated

		p.Quota = allocated
		p.MaxConnections = allocated

		if err := m.storage.UpdatePool(ctx, p); err != nil {
			return fmt.Errorf("update pool %s: %w", p.ID, err)
		}

		if poolObj, err := m.poolMgr.GetPool(p.ID); err == nil {
			poolObj.UpdateConfig(allocated)
		}
	}

	return m.UpdateAndSaveStats(ctx)
}

func (m *Manager) GetLatestStats(ctx context.Context) (*models.Statistics, error) {
	return m.storage.GetLatestStatistics(ctx)
}

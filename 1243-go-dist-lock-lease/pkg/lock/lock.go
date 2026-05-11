package lock

import (
	"sync"
	"time"
)

type Lock struct {
	Name         string
	HolderID     string
	ExpireMs     int64
	CreatedAt    time.Time
	LastRenewedAt time.Time
}

type LockManager struct {
	mu    sync.RWMutex
	locks map[string]*Lock
}

func NewLockManager() *LockManager {
	return &LockManager{
		locks: make(map[string]*Lock),
	}
}

func (lm *LockManager) isExpired(lock *Lock, now time.Time) bool {
	elapsed := now.Sub(lock.LastRenewedAt).Milliseconds()
	return elapsed >= lock.ExpireMs
}

func (lm *LockManager) remainingMs(lock *Lock, now time.Time) int64 {
	elapsed := now.Sub(lock.LastRenewedAt).Milliseconds()
	remaining := lock.ExpireMs - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (lm *LockManager) Acquire(name, holderID string, expireMs int64) bool {
	if expireMs <= 0 {
		expireMs = 10000
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	now := time.Now()

	if existing, exists := lm.locks[name]; exists {
		if !lm.isExpired(existing, now) {
			return false
		}
	}

	lock := &Lock{
		Name:          name,
		HolderID:      holderID,
		ExpireMs:      expireMs,
		CreatedAt:     now,
		LastRenewedAt: now,
	}
	lm.locks[name] = lock
	return true
}

func (lm *LockManager) Release(name, holderID string) bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lock, exists := lm.locks[name]
	if !exists {
		return false
	}

	if lock.HolderID != holderID {
		return false
	}

	delete(lm.locks, name)
	return true
}

func (lm *LockManager) Renew(name, holderID string) bool {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lock, exists := lm.locks[name]
	if !exists {
		return false
	}

	if lock.HolderID != holderID {
		return false
	}

	lock.LastRenewedAt = time.Now()
	return true
}

func (lm *LockManager) Status(name string) (*Lock, int64, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	now := time.Now()

	lock, exists := lm.locks[name]
	if !exists {
		return nil, 0, false
	}

	remaining := lm.remainingMs(lock, now)
	if remaining == 0 {
		return nil, 0, false
	}

	return lock, remaining, true
}

func (lm *LockManager) List() []*Lock {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	now := time.Now()
	result := make([]*Lock, 0, len(lm.locks))

	for _, lock := range lm.locks {
		if !lm.isExpired(lock, now) {
			result = append(result, lock)
		}
	}

	return result
}

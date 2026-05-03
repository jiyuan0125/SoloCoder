package distlock

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrNotOwner      = errors.New("not the lock owner")
	ErrLockNotHeld   = errors.New("lock not held")
	ErrInvalidLease  = errors.New("invalid lease time")
)

type LockStats struct {
	OwnerGoroutineID uint64
	WaitQueueLength  int
}

type waiter struct {
	goroutineID uint64
	notify      chan struct{}
}

type keyLock struct {
	mu           sync.Mutex
	lease        *Lease
	watchdog     *Watchdog
	waitQueue    []*waiter
}

type LockManager struct {
	locks sync.Map
}

func NewLockManager() *LockManager {
	return &LockManager{}
}

func (lm *LockManager) getOrCreateKeyLock(key string) *keyLock {
	if kl, ok := lm.locks.Load(key); ok {
		return kl.(*keyLock)
	}
	
	newKL := &keyLock{
		waitQueue: make([]*waiter, 0),
	}
	
	actual, _ := lm.locks.LoadOrStore(key, newKL)
	return actual.(*keyLock)
}

func (lm *LockManager) Lock(key string, lease time.Duration) error {
	if lease <= 0 {
		return ErrInvalidLease
	}
	
	goroutineID := getGoroutineID()
	kl := lm.getOrCreateKeyLock(key)
	
	kl.mu.Lock()
	
	if kl.lease != nil {
		if kl.lease.OwnerID() == goroutineID {
			kl.lease.IncrementCount()
			kl.mu.Unlock()
			return nil
		}
		
		w := &waiter{
			goroutineID: goroutineID,
			notify:      make(chan struct{}),
		}
		kl.waitQueue = append(kl.waitQueue, w)
		kl.mu.Unlock()
		
		<-w.notify
		
		kl.mu.Lock()
	}
	
	newLease := NewLease(goroutineID, lease)
	kl.lease = newLease
	
	wd := NewWatchdog(key, newLease, lm.onWatchdogFailure)
	kl.watchdog = wd
	wd.Start()
	
	kl.mu.Unlock()
	return nil
}

func (lm *LockManager) TryLock(key string, lease time.Duration) (bool, error) {
	if lease <= 0 {
		return false, ErrInvalidLease
	}
	
	goroutineID := getGoroutineID()
	kl := lm.getOrCreateKeyLock(key)
	
	kl.mu.Lock()
	defer kl.mu.Unlock()
	
	if kl.lease != nil {
		if kl.lease.OwnerID() == goroutineID {
			kl.lease.IncrementCount()
			return true, nil
		}
		return false, nil
	}
	
	newLease := NewLease(goroutineID, lease)
	kl.lease = newLease
	
	wd := NewWatchdog(key, newLease, lm.onWatchdogFailure)
	kl.watchdog = wd
	wd.Start()
	
	return true, nil
}

func (lm *LockManager) Unlock(key string) error {
	goroutineID := getGoroutineID()
	kl := lm.getOrCreateKeyLock(key)
	
	kl.mu.Lock()
	defer kl.mu.Unlock()
	
	if kl.lease == nil {
		return ErrLockNotHeld
	}
	
	if kl.lease.OwnerID() != goroutineID {
		return ErrNotOwner
	}
	
	count := kl.lease.DecrementCount()
	if count > 0 {
		return nil
	}
	
	if kl.watchdog != nil {
		kl.watchdog.Stop()
		kl.watchdog = nil
	}
	
	kl.lease = nil
	
	if len(kl.waitQueue) > 0 {
		firstWaiter := kl.waitQueue[0]
		kl.waitQueue = kl.waitQueue[1:]
		close(firstWaiter.notify)
	}
	
	return nil
}

func (lm *LockManager) Stats(key string) LockStats {
	kl := lm.getOrCreateKeyLock(key)
	
	kl.mu.Lock()
	defer kl.mu.Unlock()
	
	stats := LockStats{
		OwnerGoroutineID: 0,
		WaitQueueLength:  len(kl.waitQueue),
	}
	
	if kl.lease != nil {
		stats.OwnerGoroutineID = kl.lease.OwnerID()
	}
	
	return stats
}

func (lm *LockManager) onWatchdogFailure(key string) {
	kl := lm.getOrCreateKeyLock(key)
	
	kl.mu.Lock()
	defer kl.mu.Unlock()
	
	if kl.watchdog != nil {
		kl.watchdog.Stop()
		kl.watchdog = nil
	}
	
	kl.lease = nil
	
	if len(kl.waitQueue) > 0 {
		firstWaiter := kl.waitQueue[0]
		kl.waitQueue = kl.waitQueue[1:]
		close(firstWaiter.notify)
	}
}

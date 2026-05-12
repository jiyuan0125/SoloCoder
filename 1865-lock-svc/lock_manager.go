package main

import (
	"errors"
	"sync"
	"time"
)

const (
	MaxQueueSize   = 10
	WaitTimeout    = 30 * time.Second
	CleanupInterval = 1 * time.Second
)

var (
	ErrQueueFull    = errors.New("wait queue is full")
	ErrWaitTimeout  = errors.New("wait timeout")
	ErrNotHolder    = errors.New("not the lock holder")
	ErrLockNotFound = errors.New("lock not found")
	ErrLockExpired  = errors.New("lock has expired")
)

type LockStatus string

const (
	StatusFree     LockStatus = "free"
	StatusLocked   LockStatus = "locked"
	StatusExpired  LockStatus = "expired"
)

type Waiter struct {
	clientID string
	timeout  time.Duration
	notify   chan struct{}
}

type Lock struct {
	resource   string
	clientID   string
	expireAt   time.Time
	status     LockStatus
	waitQueue  []*Waiter
	queueMutex sync.Mutex
}

type LockStatusResponse struct {
	Resource          string     `json:"resource"`
	ClientID          string     `json:"client_id"`
	Status            LockStatus `json:"status"`
	RemainingTimeout  int        `json:"remaining_timeout_seconds"`
	WaitQueueLength   int        `json:"wait_queue_length"`
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

func (lm *LockManager) getOrCreateLock(resource string) *Lock {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lock, exists := lm.locks[resource]; exists {
		return lock
	}

	lock := &Lock{
		resource:  resource,
		status:    StatusFree,
		waitQueue: make([]*Waiter, 0, MaxQueueSize),
	}
	lm.locks[resource] = lock
	return lock
}

func (lm *LockManager) Acquire(resource string, clientID string, timeout time.Duration) error {
	lock := lm.getOrCreateLock(resource)

	lm.mu.Lock()

	if lock.status == StatusExpired || (lock.status == StatusLocked && time.Now().After(lock.expireAt)) {
		lock.status = StatusFree
		lock.clientID = ""
	}

	if lock.status == StatusFree {
		lock.clientID = clientID
		lock.expireAt = time.Now().Add(timeout)
		lock.status = StatusLocked
		lm.mu.Unlock()
		return nil
	}

	if lock.clientID == clientID {
		lock.expireAt = time.Now().Add(timeout)
		lm.mu.Unlock()
		return nil
	}

	if len(lock.waitQueue) >= MaxQueueSize {
		lm.mu.Unlock()
		return ErrQueueFull
	}

	waiter := &Waiter{
		clientID: clientID,
		timeout:  timeout,
		notify:   make(chan struct{}, 1),
	}
	lock.waitQueue = append(lock.waitQueue, waiter)
	lm.mu.Unlock()

	select {
	case <-waiter.notify:
		return nil
	case <-time.After(WaitTimeout):
		lm.mu.Lock()
		for i, w := range lock.waitQueue {
			if w == waiter {
				lock.waitQueue = append(lock.waitQueue[:i], lock.waitQueue[i+1:]...)
				break
			}
		}
		lm.mu.Unlock()
		return ErrWaitTimeout
	}
}

func (lm *LockManager) Release(resource string, clientID string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lock, exists := lm.locks[resource]
	if !exists {
		return ErrLockNotFound
	}

	if lock.clientID != clientID {
		return ErrNotHolder
	}

	lock.status = StatusFree
	lock.clientID = ""

	if len(lock.waitQueue) > 0 {
		waiter := lock.waitQueue[0]
		lock.waitQueue = lock.waitQueue[1:]

		lock.clientID = waiter.clientID
		lock.expireAt = time.Now().Add(waiter.timeout)
		lock.status = StatusLocked

		select {
		case waiter.notify <- struct{}{}:
		default:
		}
	}

	return nil
}

func (lm *LockManager) Renew(resource string, clientID string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lock, exists := lm.locks[resource]
	if !exists {
		return ErrLockNotFound
	}

	if lock.status != StatusLocked {
		return ErrLockExpired
	}

	if lock.clientID != clientID {
		return ErrNotHolder
	}

	if time.Now().After(lock.expireAt) {
		lock.status = StatusFree
		lock.clientID = ""
		return ErrLockExpired
	}

	remaining := time.Until(lock.expireAt)
	lock.expireAt = time.Now().Add(remaining)

	return nil
}

func (lm *LockManager) GetAllStatuses() []LockStatusResponse {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	statuses := make([]LockStatusResponse, 0, len(lm.locks))

	for _, lock := range lm.locks {
		var remaining int
		if lock.status == StatusLocked && time.Now().Before(lock.expireAt) {
			remaining = int(time.Until(lock.expireAt).Seconds())
		} else {
			remaining = 0
		}

		statuses = append(statuses, LockStatusResponse{
			Resource:         lock.resource,
			ClientID:         lock.clientID,
			Status:           lock.status,
			RemainingTimeout: remaining,
			WaitQueueLength:  len(lock.waitQueue),
		})
	}

	return statuses
}

func (lm *LockManager) cleanupExpiredLocks() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		lm.mu.Lock()
		now := time.Now()

		for resource, lock := range lm.locks {
			if lock.status == StatusLocked && now.After(lock.expireAt) {
				lock.status = StatusExpired

				if len(lock.waitQueue) > 0 {
					waiter := lock.waitQueue[0]
					lock.waitQueue = lock.waitQueue[1:]

					lock.clientID = waiter.clientID
					lock.expireAt = time.Now().Add(waiter.timeout)
					lock.status = StatusLocked

					select {
					case waiter.notify <- struct{}{}:
					default:
					}
				} else {
					lock.clientID = ""
					lock.status = StatusFree
				}
			}

			if lock.status == StatusFree && len(lock.waitQueue) == 0 {
				delete(lm.locks, resource)
			}
		}

		lm.mu.Unlock()
	}
}

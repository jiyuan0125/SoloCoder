package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	MaxTimeoutSeconds  = 300
	MaxWaiters     = 20
	CleanupInterval = 1 * time.Second
)

type LockState int

const (
	Idle LockState = iota
	Locked
	Expired
)

type Waiter struct {
	clientID  string
	timeout   time.Duration
	readyChan chan bool
}

type Lock struct {
	state      LockState
	owner      string
	expireTime time.Time
	waiters    []*Waiter
	mu         sync.Mutex
	cond       *sync.Cond
}

type LockManager struct {
	locks map[string]*Lock
	mu    sync.RWMutex
}

func NewLockManager() *LockManager {
	lm := &LockManager{
		locks: make(map[string]*Lock),
	}
	go lm.startCleanup()
	return lm
}

func (lm *LockManager) getOrCreateLock(resource string) *Lock {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	if l, ok := lm.locks[resource]; ok {
		return l
	}
	lock := &Lock{
		state:   Idle,
		waiters: make([]*Waiter, 0, 0),
	}
	lock.cond = sync.NewCond(&lock.mu)
	lm.locks[resource] = lock
	return lock
}

func (lm *LockManager) getLock(resource string) (*Lock, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	l, ok := lm.locks[resource]
	return l, ok
}

func (lm *LockManager) startCleanup() {
	ticker := time.NewTicker(CleanupInterval)
	for range ticker.C {
		lm.cleanupExpired()
	}
}

func (lm *LockManager) cleanupExpired() {
	lm.mu.Lock()
	now := time.Now()
	for resource, lock := range lm.locks {
		lock.mu.Lock()
		if lock.state == Locked && now.After(lock.expireTime) {
			lock.state = Expired
			lock.owner = ""
			if len(lock.waiters) > 0 {
				lm.assignToNextWaiter(resource, lock)
			}
		}
		lock.mu.Unlock()
	}
	lm.mu.Unlock()
}

func (lm *LockManager) assignToNextWaiter(resource string, lock *Lock) {
	if len(lock.waiters) == 0 {
		return
	}
	
	waiter := lock.waiters[0]
	lock.waiters = lock.waiters[1:]
	lock.state = Locked
	lock.owner = waiter.clientID
	lock.expireTime = time.Now().Add(waiter.timeout)
	select {
	case waiter.readyChan <- true:
	default:
	}
}

func (lm *LockManager) acquire(resource string, clientID string, timeout time.Duration, waitTimeout time.Duration) (bool, error) {
	if timeout > MaxTimeoutSeconds*time.Second {
		timeout = MaxTimeoutSeconds * time.Second
	}

	lock := lm.getOrCreateLock(resource)
	lock.mu.Lock()

	if lock.state == Locked && lock.owner == clientID {
		lock.mu.Unlock()
		return false, errors.New("lock already held by this client")
	}

	if lock.state == Idle || lock.state == Expired {
		lock.state = Locked
		lock.owner = clientID
		lock.expireTime = time.Now().Add(timeout)
		lock.mu.Unlock()
		return true, nil
	}

	if waitTimeout == 0 {
		lock.mu.Unlock()
		return false, nil
	}

	if len(lock.waiters) >= MaxWaiters {
		lock.mu.Unlock()
		return false, errors.New("queue full")
	}

	waiter := &Waiter{
		clientID:  clientID,
		timeout:   timeout,
		readyChan: make(chan bool, 1),
	}
	lock.waiters = append(lock.waiters, waiter)
	lock.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()

	select {
	case <-waiter.readyChan:
		return true, nil
	case <-ctx.Done():
		lock.mu.Lock()
		for i, w := range lock.waiters {
			if w == waiter {
				lock.waiters = append(lock.waiters[:i], lock.waiters[i+1:]...)
				break
			}
		}
		lock.mu.Unlock()
		return false, nil
	}
}

func (lm *LockManager) release(resource string, clientID string) error {
	lock, ok := lm.getLock(resource)
	if !ok {
		return errors.New("lock not found")
	}
	
	lock.mu.Lock()
	defer lock.mu.Unlock()
	
	if lock.state == Expired {
		return errors.New("lock already expired")
	}
	
	if lock.owner != clientID {
		return errors.New("not owner")
	}
	
	if len(lock.waiters) > 0 {
		lm.assignToNextWaiter(resource, lock)
	} else {
		lock.state = Idle
		lock.owner = ""
	}
	return nil
}

func (lm *LockManager) renew(resource string, clientID string, timeout time.Duration) (bool, error) {
	if timeout > MaxTimeoutSeconds*time.Second {
		timeout = MaxTimeoutSeconds * time.Second
	}
	
	lock, ok := lm.getLock(resource)
	if !ok {
		return false, errors.New("lock not found")
	}
	
	lock.mu.Lock()
	defer lock.mu.Unlock()
	
	if lock.state == Locked && lock.owner == clientID {
		lock.expireTime = time.Now().Add(timeout)
		return true, nil
	}
	
	return false, nil
}

type LockStatus struct {
	Resource      string  `json:"resource"`
	State         string  `json:"state"`
	Owner         string  `json:"owner,omitempty"`
	RemainingTime float64 `json:"remaining_time"`
	WaitQueueLen  int     `json:"wait_queue_len"`
}

func (lm *LockManager) status() []LockStatus {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	
	now := time.Now()
	statuses := make([]LockStatus, 0, len(lm.locks))
	
	for resource, lock := range lm.locks {
		lock.mu.Lock()
		stateStr := "idle"
		remaining := 0.0
		owner := ""
		
		switch lock.state {
		case Locked:
			stateStr = "locked"
			if now.Before(lock.expireTime) {
				remaining = lock.expireTime.Sub(now).Seconds()
			}
			owner = lock.owner
		case Expired:
			stateStr = "expired"
		}
		
		statuses = append(statuses, LockStatus{
			Resource:      resource,
			State:         stateStr,
			Owner:         owner,
			RemainingTime: remaining,
			WaitQueueLen:  len(lock.waiters),
		})
		lock.mu.Unlock()
	}
	
	return statuses
}

func main() {
	manager := NewLockManager()
	
	http.HandleFunc("/acquire", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		resource := r.FormValue("resource")
		clientID := r.FormValue("client_id")
		timeoutStr := r.FormValue("timeout")
		waitTimeoutStr := r.FormValue("wait_timeout")
		
		if resource == "" || clientID == "" || timeoutStr == "" {
			http.Error(w, "missing required fields", http.StatusBadRequest)
			return
		}
		
		timeout, err := strconv.Atoi(timeoutStr)
		if err != nil {
			http.Error(w, "invalid timeout", http.StatusBadRequest)
			return
		}
		
		waitTimeout := 0
		if waitTimeoutStr != "" {
			waitTimeout, err = strconv.Atoi(waitTimeoutStr)
			if err != nil {
				http.Error(w, "invalid wait_timeout", http.StatusBadRequest)
				return
			}
		}
		
		success, err := manager.acquire(resource, clientID, time.Duration(timeout)*time.Second, time.Duration(waitTimeout)*time.Second)
		if err != nil {
			if err.Error() == "queue full" {
				http.Error(w, err.Error(), http.StatusTooManyRequests)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": success})
	})
	
	http.HandleFunc("/release", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		resource := r.FormValue("resource")
		clientID := r.FormValue("client_id")
		
		if resource == "" || clientID == "" {
			http.Error(w, "missing required fields", http.StatusBadRequest)
			return
		}
		
		err := manager.release(resource, clientID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	
	http.HandleFunc("/renew", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		resource := r.FormValue("resource")
		clientID := r.FormValue("client_id")
		timeoutStr := r.FormValue("timeout")
		
		if resource == "" || clientID == "" || timeoutStr == "" {
			http.Error(w, "missing required fields", http.StatusBadRequest)
			return
		}
		
		timeout, err := strconv.Atoi(timeoutStr)
		if err != nil {
			http.Error(w, "invalid timeout", http.StatusBadRequest)
			return
		}
		
		success, err := manager.renew(resource, clientID, time.Duration(timeout)*time.Second)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		if !success {
			http.Error(w, "lock not held", http.StatusBadRequest)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		statuses := manager.status()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statuses)
	})
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	fmt.Printf("Server listening on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

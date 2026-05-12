package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Waiter struct {
	clientID   string
	timeoutSec int
	notifyChan chan struct{}
}

type Lock struct {
	mu         sync.Mutex
	holder     string
	expiresAt  time.Time
	timeoutSec int
	waitQueue  *list.List
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

func (lm *LockManager) getLock(resource string) *Lock {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	if lock, exists := lm.locks[resource]; exists {
		return lock
	}
	lock := &Lock{
		waitQueue: list.New(),
	}
	lm.locks[resource] = lock
	return lock
}

func (lm *LockManager) acquire(resource, clientID string, timeoutSec, waitTimeoutSec int) bool {
	lock := lm.getLock(resource)
	lock.mu.Lock()

	now := time.Now()
	if lock.holder == "" || now.After(lock.expiresAt) {
		lock.holder = clientID
		lock.timeoutSec = timeoutSec
		lock.expiresAt = now.Add(time.Duration(timeoutSec) * time.Second)
		lock.mu.Unlock()
		return true
	}

	if lock.holder == clientID {
		lock.timeoutSec = timeoutSec
		lock.expiresAt = now.Add(time.Duration(timeoutSec) * time.Second)
		lock.mu.Unlock()
		return true
	}

	if waitTimeoutSec <= 0 {
		lock.mu.Unlock()
		return false
	}

	waiter := &Waiter{
		clientID:   clientID,
		timeoutSec: timeoutSec,
		notifyChan: make(chan struct{}, 1),
	}
	element := lock.waitQueue.PushBack(waiter)
	lock.mu.Unlock()

	timer := time.NewTimer(time.Duration(waitTimeoutSec) * time.Second)
	defer timer.Stop()

	select {
	case <-waiter.notifyChan:
		lock.mu.Lock()
		if lock.holder == clientID {
			lock.timeoutSec = timeoutSec
			lock.expiresAt = time.Now().Add(time.Duration(timeoutSec) * time.Second)
			lock.mu.Unlock()
			return true
		}
		lock.mu.Unlock()
		return false
	case <-timer.C:
		lock.mu.Lock()
		lock.waitQueue.Remove(element)
		lock.mu.Unlock()
		return false
	}
}

func (lm *LockManager) release(resource, clientID string) bool {
	lock := lm.getLock(resource)
	lock.mu.Lock()
	defer lock.mu.Unlock()

	if lock.holder != clientID {
		return false
	}

	lock.holder = ""
	lock.expiresAt = time.Time{}
	lock.timeoutSec = 0

	for lock.waitQueue.Len() > 0 {
		front := lock.waitQueue.Front()
		waiter := front.Value.(*Waiter)
		lock.waitQueue.Remove(front)

		lock.holder = waiter.clientID
		lock.timeoutSec = waiter.timeoutSec
		lock.expiresAt = time.Now().Add(time.Duration(waiter.timeoutSec) * time.Second)
		waiter.notifyChan <- struct{}{}
		return true
	}

	return true
}

func (lm *LockManager) renew(resource, clientID string, timeoutSec int) bool {
	lock := lm.getLock(resource)
	lock.mu.Lock()
	defer lock.mu.Unlock()

	now := time.Now()
	if lock.holder == clientID && now.Before(lock.expiresAt) {
		lock.timeoutSec = timeoutSec
		lock.expiresAt = now.Add(time.Duration(timeoutSec) * time.Second)
		return true
	}
	return false
}

type LockStatus struct {
	Resource    string `json:"resource"`
	Holder      string `json:"holder"`
	Remaining   int64  `json:"remaining_seconds"`
	WaitQueueLen int   `json:"wait_queue_length"`
}

func (lm *LockManager) getStatus() []LockStatus {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	statuses := make([]LockStatus, 0, len(lm.locks))
	for resource, lock := range lm.locks {
		lock.mu.Lock()
		var holder string
		var remaining int64
		waitQueueLen := lock.waitQueue.Len()

		now := time.Now()
		if lock.holder == "" || now.After(lock.expiresAt) {
			holder = ""
			remaining = 0
		} else {
			holder = lock.holder
			remaining = int64(lock.expiresAt.Sub(now).Seconds())
			if remaining < 0 {
				remaining = 0
			}
		}
		lock.mu.Unlock()

		statuses = append(statuses, LockStatus{
			Resource:     resource,
			Holder:       holder,
			Remaining:    remaining,
			WaitQueueLen: waitQueueLen,
		})
	}
	return statuses
}

type AcquireRequest struct {
	Resource     string `json:"resource"`
	ClientID     string `json:"client_id"`
	Timeout      int    `json:"timeout"`
	WaitTimeout  int    `json:"wait_timeout"`
}

type ReleaseRequest struct {
	Resource string `json:"resource"`
	ClientID string `json:"client_id"`
}

type RenewRequest struct {
	Resource string `json:"resource"`
	ClientID string `json:"client_id"`
	Timeout  int    `json:"timeout"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

func handleAcquire(lm *LockManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
			return
		}

		var req AcquireRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid request body"})
			return
		}

		if req.Resource == "" || req.ClientID == "" || req.Timeout <= 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "resource, client_id and timeout (positive) are required"})
			return
		}

		success := lm.acquire(req.Resource, req.ClientID, req.Timeout, req.WaitTimeout)
		w.Header().Set("Content-Type", "application/json")
		if success {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(SuccessResponse{Success: true})
		} else {
			w.WriteHeader(http.StatusRequestTimeout)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to acquire lock"})
		}
	}
}

func handleRelease(lm *LockManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
			return
		}

		var req ReleaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid request body"})
			return
		}

		if req.Resource == "" || req.ClientID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "resource and client_id are required"})
			return
		}

		success := lm.release(req.Resource, req.ClientID)
		w.Header().Set("Content-Type", "application/json")
		if success {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(SuccessResponse{Success: true})
		} else {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "not the lock holder"})
		}
	}
}

func handleRenew(lm *LockManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
			return
		}

		var req RenewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid request body"})
			return
		}

		if req.Resource == "" || req.ClientID == "" || req.Timeout <= 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "resource, client_id and timeout (positive) are required"})
			return
		}

		success := lm.renew(req.Resource, req.ClientID, req.Timeout)
		w.Header().Set("Content-Type", "application/json")
		if success {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(SuccessResponse{Success: true})
		} else {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to renew lock"})
		}
	}
}

func handleStatus(lm *LockManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
			return
		}

		statuses := lm.getStatus()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(statuses)
	}
}

func main() {
	lm := NewLockManager()

	http.HandleFunc("/locks/acquire", handleAcquire(lm))
	http.HandleFunc("/locks/release", handleRelease(lm))
	http.HandleFunc("/locks/renew", handleRenew(lm))
	http.HandleFunc("/locks", handleStatus(lm))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	_, err := strconv.Atoi(port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid PORT: %s\n", port)
		os.Exit(1)
	}

	fmt.Printf("distributed lock service starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

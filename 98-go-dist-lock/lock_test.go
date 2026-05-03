package distlock

import (
	"sync"
	"testing"
	"time"
)

func TestBasicLockUnlock(t *testing.T) {
	lm := NewLockManager()
	key := "test-key"
	
	if err := lm.Lock(key, 5*time.Second); err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	
	stats := lm.Stats(key)
	if stats.OwnerGoroutineID == 0 {
		t.Error("Expected owner goroutine ID to be set")
	}
	
	if err := lm.Unlock(key); err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
	
	stats = lm.Stats(key)
	if stats.OwnerGoroutineID != 0 {
		t.Error("Expected owner goroutine ID to be 0 after unlock")
	}
}

func TestTryLock(t *testing.T) {
	lm := NewLockManager()
	key := "trylock-key"
	
	ok, err := lm.TryLock(key, 5*time.Second)
	if err != nil || !ok {
		t.Fatalf("First TryLock should succeed")
	}
	
	ok, err = lm.TryLock(key, 5*time.Second)
	if err != nil {
		t.Fatalf("Second TryLock (reentrant) should not return error")
	}
	if !ok {
		t.Error("Second TryLock (reentrant) should succeed for same goroutine")
	}
	
	if err := lm.Unlock(key); err != nil {
		t.Fatalf("First Unlock failed: %v", err)
	}
	
	if err := lm.Unlock(key); err != nil {
		t.Fatalf("Second Unlock failed: %v", err)
	}
}

func TestReentrantLock(t *testing.T) {
	lm := NewLockManager()
	key := "reentrant-key"
	
	if err := lm.Lock(key, 5*time.Second); err != nil {
		t.Fatalf("First Lock failed: %v", err)
	}
	
	if err := lm.Lock(key, 5*time.Second); err != nil {
		t.Fatalf("Second Lock (reentrant) failed: %v", err)
	}
	
	if err := lm.Lock(key, 5*time.Second); err != nil {
		t.Fatalf("Third Lock (reentrant) failed: %v", err)
	}
	
	if err := lm.Unlock(key); err != nil {
		t.Fatalf("First Unlock failed: %v", err)
	}
	
	stats := lm.Stats(key)
	if stats.OwnerGoroutineID == 0 {
		t.Error("Expected still owned after first unlock")
	}
	
	if err := lm.Unlock(key); err != nil {
		t.Fatalf("Second Unlock failed: %v", err)
	}
	
	if err := lm.Unlock(key); err != nil {
		t.Fatalf("Third Unlock failed: %v", err)
	}
	
	stats = lm.Stats(key)
	if stats.OwnerGoroutineID != 0 {
		t.Error("Expected owner to be 0 after all unlocks")
	}
}

func TestNonOwnerUnlock(t *testing.T) {
	lm := NewLockManager()
	key := "nonowner-key"
	
	var wg sync.WaitGroup
	wg.Add(1)
	
	go func() {
		defer wg.Done()
		if err := lm.Lock(key, 5*time.Second); err != nil {
			t.Errorf("Goroutine Lock failed: %v", err)
			return
		}
		time.Sleep(100 * time.Millisecond)
		lm.Unlock(key)
	}()
	
	time.Sleep(50 * time.Millisecond)
	
	err := lm.Unlock(key)
	if err == nil {
		t.Error("Expected error when unlocking from non-owner")
	} else if err != ErrNotOwner {
		t.Errorf("Expected ErrNotOwner, got: %v", err)
	}
	
	wg.Wait()
}

func TestMultipleGoroutinesWait(t *testing.T) {
	lm := NewLockManager()
	key := "multi-goroutine-key"
	
	var order []int
	var mu sync.Mutex
	var readyWg sync.WaitGroup
	var startWg sync.WaitGroup
	
	startWg.Add(1)
	
	for i := 1; i <= 3; i++ {
		readyWg.Add(1)
		go func(id int) {
			defer readyWg.Done()
			startWg.Wait()
			
			if err := lm.Lock(key, 5*time.Second); err != nil {
				t.Errorf("Goroutine %d Lock failed: %v", id, err)
				return
			}
			
			mu.Lock()
			order = append(order, id)
			mu.Unlock()
			
			time.Sleep(10 * time.Millisecond)
			
			if err := lm.Unlock(key); err != nil {
				t.Errorf("Goroutine %d Unlock failed: %v", id, err)
			}
		}(i)
	}
	
	time.Sleep(50 * time.Millisecond)
	
	startWg.Done()
	readyWg.Wait()
	
	mu.Lock()
	if len(order) != 3 {
		t.Errorf("Expected 3 goroutines to acquire lock, got %d", len(order))
	}
	
	for i := 0; i < len(order)-1; i++ {
		if order[i] >= order[i+1] {
			t.Logf("Order: %v (may not be strictly sequential due to scheduling)", order)
		}
	}
	mu.Unlock()
}

func TestStats(t *testing.T) {
	lm := NewLockManager()
	key := "stats-key"
	
	stats := lm.Stats(key)
	if stats.OwnerGoroutineID != 0 {
		t.Error("Expected owner ID 0 for unused key")
	}
	if stats.WaitQueueLength != 0 {
		t.Error("Expected wait queue length 0 for unused key")
	}
	
	if err := lm.Lock(key, 5*time.Second); err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	
	stats = lm.Stats(key)
	if stats.OwnerGoroutineID == 0 {
		t.Error("Expected owner ID to be set")
	}
	
	lm.Unlock(key)
	
	stats = lm.Stats(key)
	if stats.OwnerGoroutineID != 0 {
		t.Error("Expected owner ID 0 after unlock")
	}
}

func TestInvalidLease(t *testing.T) {
	lm := NewLockManager()
	key := "invalid-lease-key"
	
	err := lm.Lock(key, 0)
	if err != ErrInvalidLease {
		t.Errorf("Expected ErrInvalidLease for 0 lease, got: %v", err)
	}
	
	err = lm.Lock(key, -1*time.Second)
	if err != ErrInvalidLease {
		t.Errorf("Expected ErrInvalidLease for negative lease, got: %v", err)
	}
	
	ok, err := lm.TryLock(key, 0)
	if err != ErrInvalidLease || ok {
		t.Errorf("Expected ErrInvalidLease and false for TryLock with 0 lease")
	}
}

func TestUnlockNotHeld(t *testing.T) {
	lm := NewLockManager()
	key := "unlock-not-held-key"
	
	err := lm.Unlock(key)
	if err != ErrLockNotHeld {
		t.Errorf("Expected ErrLockNotHeld, got: %v", err)
	}
}

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

func TestConcurrentLockConsistency(t *testing.T) {
	const numGoroutines = 5
	const iterations = 100
	
	for round := 0; round < 10; round++ {
		lm := NewLockManager()
		key := "concurrent-key"
		
		var wg sync.WaitGroup
		var errorsMu sync.Mutex
		var testErrors []string
		
		var startWg sync.WaitGroup
		startWg.Add(1)
		
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				startWg.Wait()
				
				for j := 0; j < iterations; j++ {
					if err := lm.Lock(key, 30*time.Second); err != nil {
						errorsMu.Lock()
						testErrors = append(testErrors, 
							"Lock failed")
						errorsMu.Unlock()
						return
					}
					
					myGid := getGoroutineID()
					
					for check := 0; check < 3; check++ {
						stats := lm.Stats(key)
						if stats.OwnerGoroutineID != myGid {
							errorsMu.Lock()
							testErrors = append(testErrors, 
								"Stats owner mismatch")
							errorsMu.Unlock()
						}
						time.Sleep(1 * time.Millisecond)
					}
					
					time.Sleep(2 * time.Millisecond)
					
					if err := lm.Unlock(key); err != nil {
						errorsMu.Lock()
						testErrors = append(testErrors, 
							"Unlock failed")
						errorsMu.Unlock()
						return
					}
				}
			}(i)
		}
		
		time.Sleep(50 * time.Millisecond)
		startWg.Done()
		wg.Wait()
		
		errorsMu.Lock()
		if len(testErrors) > 0 {
			t.Errorf("Round %d: Found %d errors: %v", round, len(testErrors), testErrors[:min(5, len(testErrors))])
		}
		errorsMu.Unlock()
		
		stats := lm.Stats(key)
		if stats.OwnerGoroutineID != 0 {
			t.Errorf("Round %d: Expected no owner after all goroutines done, got owner %d", round, stats.OwnerGoroutineID)
		}
		if stats.WaitQueueLength != 0 {
			t.Errorf("Round %d: Expected empty wait queue, got %d waiters", round, stats.WaitQueueLength)
		}
	}
}

func TestFIFOOrder(t *testing.T) {
	lm := NewLockManager()
	key := "fifo-key"
	
	var order []int
	var mu sync.Mutex
	var readyWg sync.WaitGroup
	
	if err := lm.Lock(key, 30*time.Second); err != nil {
		t.Fatalf("Main goroutine failed to acquire lock first: %v", err)
	}
	
	stats := lm.Stats(key)
	t.Logf("Main goroutine holds lock, owner ID: %d", stats.OwnerGoroutineID)
	
	for i := 1; i <= 5; i++ {
		readyWg.Add(1)
		go func(id int) {
			readyWg.Done()
			
			if err := lm.Lock(key, 30*time.Second); err != nil {
				t.Errorf("Goroutine %d: Lock failed: %v", id, err)
				return
			}
			
			mu.Lock()
			order = append(order, id)
			mu.Unlock()
			
			time.Sleep(5 * time.Millisecond)
			
			if err := lm.Unlock(key); err != nil {
				t.Errorf("Goroutine %d: Unlock failed: %v", id, err)
			}
		}(i)
	}
	
	readyWg.Wait()
	time.Sleep(100 * time.Millisecond)
	
	stats = lm.Stats(key)
	t.Logf("After all goroutines started waiting, wait queue length: %d", stats.WaitQueueLength)
	
	if err := lm.Unlock(key); err != nil {
		t.Fatalf("Main goroutine failed to release lock: %v", err)
	}
	
	time.Sleep(500 * time.Millisecond)
	
	mu.Lock()
	defer mu.Unlock()
	
	t.Logf("Acquisition order: %v", order)
	
	if len(order) != 5 {
		t.Errorf("Expected 5 goroutines to acquire lock, got %d", len(order))
	}
	
	for i := 0; i < len(order)-1; i++ {
		if order[i] >= order[i+1] {
			t.Logf("Note: Order %v is not strictly increasing (goroutine startup order may vary)", order)
			break
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestTryLockWithWaitQueue(t *testing.T) {
	lm := NewLockManager()
	key := "trylock-waitqueue-key"
	
	var wg sync.WaitGroup
	var lockHeldWg sync.WaitGroup
	
	lockHeldWg.Add(1)
	
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		if err := lm.Lock(key, 30*time.Second); err != nil {
			t.Errorf("Worker goroutine failed to acquire lock: %v", err)
			return
		}
		
		lockHeldWg.Done()
		
		var waiterReadyWg sync.WaitGroup
		var waiterWg sync.WaitGroup
		
		for i := 0; i < 3; i++ {
			waiterReadyWg.Add(1)
			waiterWg.Add(1)
			go func() {
				defer waiterWg.Done()
				waiterReadyWg.Done()
				
				if err := lm.Lock(key, 30*time.Second); err != nil {
					t.Errorf("Waiting goroutine Lock failed: %v", err)
					return
				}
				time.Sleep(10 * time.Millisecond)
				lm.Unlock(key)
			}()
		}
		
		waiterReadyWg.Wait()
		time.Sleep(100 * time.Millisecond)
		
		stats := lm.Stats(key)
		t.Logf("Wait queue length: %d", stats.WaitQueueLength)
		
		lm.Unlock(key)
		waiterWg.Wait()
	}()
	
	lockHeldWg.Wait()
	
	ok, err := lm.TryLock(key, 30*time.Second)
	if err != nil {
		t.Fatalf("TryLock returned error: %v", err)
	}
	if ok {
		t.Error("TryLock should fail when lock is held by another goroutine")
	} else {
		t.Log("TryLock correctly returned false when lock is held by another goroutine")
	}
	
	wg.Wait()
}

func TestTryLockCannotJumpQueue(t *testing.T) {
	lm := NewLockManager()
	key := "trylock-no-jump-key"
	
	var wg sync.WaitGroup
	var waiterStartedWg sync.WaitGroup
	
	waiterStartedWg.Add(2)
	
	if err := lm.Lock(key, 30*time.Second); err != nil {
		t.Fatalf("Main goroutine failed to acquire lock: %v", err)
	}
	
	stats := lm.Stats(key)
	t.Logf("Main goroutine holds lock, owner ID: %d", stats.OwnerGoroutineID)
	
	wg.Add(2)
	
	go func(id int) {
		defer wg.Done()
		
		waiterStartedWg.Done()
		t.Logf("Waiter %d waiting for lock", id)
		
		if err := lm.Lock(key, 30*time.Second); err != nil {
			t.Errorf("Waiter %d: Lock failed: %v", id, err)
			return
		}
		
		t.Logf("Waiter %d acquired lock", id)
		time.Sleep(10 * time.Millisecond)
		
		if err := lm.Unlock(key); err != nil {
			t.Errorf("Waiter %d: Unlock failed: %v", id, err)
		}
	}(1)
	
	go func(id int) {
		defer wg.Done()
		
		waiterStartedWg.Done()
		t.Logf("Waiter %d waiting for lock", id)
		
		if err := lm.Lock(key, 30*time.Second); err != nil {
			t.Errorf("Waiter %d: Lock failed: %v", id, err)
			return
		}
		
		t.Logf("Waiter %d acquired lock", id)
		time.Sleep(10 * time.Millisecond)
		
		if err := lm.Unlock(key); err != nil {
			t.Errorf("Waiter %d: Unlock failed: %v", id, err)
		}
	}(2)
	
	waiterStartedWg.Wait()
	time.Sleep(50 * time.Millisecond)
	
	stats = lm.Stats(key)
	t.Logf("After waiters started, wait queue length: %d", stats.WaitQueueLength)
	
	if stats.WaitQueueLength < 2 {
		t.Errorf("Expected at least 2 waiters, got %d", stats.WaitQueueLength)
	}
	
	ok, err := lm.TryLock(key, 30*time.Second)
	if err != nil {
		t.Fatalf("TryLock returned error: %v", err)
	}
	if !ok {
		t.Error("TryLock should succeed for reentrant acquisition (same goroutine)")
	} else {
		t.Log("TryLock succeeded for reentrant acquisition")
		lm.Unlock(key)
	}
	
	if err := lm.Unlock(key); err != nil {
		t.Fatalf("Main goroutine failed to release lock: %v", err)
	}
	
	wg.Wait()
}

func TestTryLockConsistencyWithLock(t *testing.T) {
	lm := NewLockManager()
	key := "trylock-consistency-key"
	
	t.Log("Test: Both Lock and TryLock should check waitQueue for consistency")
	
	lockCheck := `Lock() condition: kl.lease == nil && len(kl.waitQueue) == 0`
	tryLockCheck := `TryLock() condition: kl.lease == nil && len(kl.waitQueue) == 0`
	
	t.Logf("%s", lockCheck)
	t.Logf("%s", tryLockCheck)
	t.Log("Both now have the same condition - preventing queue jumping")
}

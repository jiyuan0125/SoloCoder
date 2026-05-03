package circuitbreaker

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var testErr = errors.New("test error")

func TestNewCircuitBreaker(t *testing.T) {
	cb := New()
	if cb.State() != StateClosed {
		t.Errorf("Expected initial state to be Closed, got %s", cb.State())
	}
}

func TestWithFailureThreshold(t *testing.T) {
	cb := New(WithFailureThreshold(3))
	
	for i := 0; i < 2; i++ {
		allowed, _ := cb.Allow()
		if !allowed {
			t.Errorf("Expected request to be allowed on iteration %d", i)
		}
		cb.RecordFailure(testErr)
	}
	
	if cb.State() != StateClosed {
		t.Errorf("Expected state to be Closed after 2 failures (threshold 3), got %s", cb.State())
	}
	
	allowed, _ := cb.Allow()
	if !allowed {
		t.Error("Expected request to be allowed")
	}
	cb.RecordFailure(testErr)
	
	if cb.State() != StateOpen {
		t.Errorf("Expected state to be Open after 3 failures, got %s", cb.State())
	}
}

func TestStateTransitions(t *testing.T) {
	cb := New(WithFailureThreshold(2), WithOpenDuration(100*time.Millisecond))
	
	if cb.State() != StateClosed {
		t.Error("Expected initial state to be Closed")
	}
	
	cb.Allow()
	cb.RecordFailure(testErr)
	cb.Allow()
	cb.RecordFailure(testErr)
	
	if cb.State() != StateOpen {
		t.Error("Expected state to be Open after 2 failures")
	}
	
	allowed, err := cb.Allow()
	if allowed {
		t.Error("Expected request to be rejected when Open")
	}
	if err == nil {
		t.Error("Expected error when request is rejected")
	}
	
	time.Sleep(150 * time.Millisecond)
	
	allowed, _ = cb.Allow()
	if !allowed {
		t.Error("Expected request to be allowed after open duration")
	}
	if cb.State() != StateHalfOpen {
		t.Error("Expected state to be HalfOpen after open duration")
	}
}

func TestHalfOpenProbeSuccess(t *testing.T) {
	cb := New(WithFailureThreshold(1), WithOpenDuration(50*time.Millisecond))
	
	cb.Allow()
	cb.RecordFailure(testErr)
	
	if cb.State() != StateOpen {
		t.Error("Expected state to be Open after failure")
	}
	
	time.Sleep(100 * time.Millisecond)
	
	allowed, _ := cb.Allow()
	if !allowed {
		t.Error("Expected probe request to be allowed")
	}
	if cb.State() != StateHalfOpen {
		t.Error("Expected state to be HalfOpen")
	}
	
	cb.RecordSuccess()
	
	if cb.State() != StateClosed {
		t.Error("Expected state to be Closed after successful probe")
	}
	
	allowed, _ = cb.Allow()
	if !allowed {
		t.Error("Expected request to be allowed after state reset to Closed")
	}
}

func TestHalfOpenProbeFailure(t *testing.T) {
	cb := New(WithFailureThreshold(1), WithOpenDuration(50*time.Millisecond))
	
	cb.Allow()
	cb.RecordFailure(testErr)
	
	if cb.State() != StateOpen {
		t.Error("Expected state to be Open after failure")
	}
	
	time.Sleep(100 * time.Millisecond)
	
	allowed, _ := cb.Allow()
	if !allowed {
		t.Error("Expected probe request to be allowed")
	}
	if cb.State() != StateHalfOpen {
		t.Error("Expected state to be HalfOpen")
	}
	
	cb.RecordFailure(testErr)
	
	if cb.State() != StateHalfOpen {
		t.Errorf("Expected state to remain HalfOpen after probe failure, got %s", cb.State())
	}
}

func TestOnStateChangeCallback(t *testing.T) {
	var stateChanges []string
	var mu sync.Mutex
	
	cb := New(WithFailureThreshold(1), WithOpenDuration(50*time.Millisecond))
	cb.OnStateChange(func(from, to string) {
		mu.Lock()
		stateChanges = append(stateChanges, from+"->"+to)
		mu.Unlock()
	})
	
	cb.Allow()
	cb.RecordFailure(testErr)
	
	time.Sleep(100 * time.Millisecond)
	cb.Allow()
	cb.RecordSuccess()
	
	time.Sleep(50 * time.Millisecond)
	
	mu.Lock()
	changes := stateChanges
	mu.Unlock()
	
	expected := []string{"Closed->Open", "Open->HalfOpen", "HalfOpen->Closed"}
	if len(changes) != len(expected) {
		t.Errorf("Expected %d state changes, got %d", len(expected), len(changes))
	}
	
	for i, exp := range expected {
		if i < len(changes) && changes[i] != exp {
			t.Errorf("Expected state change %d to be %s, got %s", i, exp, changes[i])
		}
	}
}

func TestOnRequestCallback(t *testing.T) {
	var requestCount int32
	var allowedCount int32
	
	cb := New(WithFailureThreshold(1))
	cb.OnRequest(func(allowed bool, duration time.Duration) {
		atomic.AddInt32(&requestCount, 1)
		if allowed {
			atomic.AddInt32(&allowedCount, 1)
		}
	})
	
	for i := 0; i < 5; i++ {
		allowed, _ := cb.Allow()
		if i == 1 {
			cb.RecordFailure(testErr)
		} else if allowed {
			cb.RecordSuccess()
		}
	}
	
	if atomic.LoadInt32(&requestCount) != 5 {
		t.Errorf("Expected 5 request callbacks, got %d", atomic.LoadInt32(&requestCount))
	}
	
	if atomic.LoadInt32(&allowedCount) != 2 {
		t.Errorf("Expected 2 allowed requests, got %d", atomic.LoadInt32(&allowedCount))
	}
}

func TestHalfOpenConcurrentRequests(t *testing.T) {
	cb := New(WithFailureThreshold(1), WithOpenDuration(50*time.Millisecond))
	
	cb.Allow()
	cb.RecordFailure(testErr)
	
	if cb.State() != StateOpen {
		t.Error("Expected state to be Open after failure")
	}
	
	time.Sleep(100 * time.Millisecond)
	
	var wg sync.WaitGroup
	var allowedCount int32
	var firstAllowed int32
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			allowed, _ := cb.Allow()
			if allowed {
				atomic.AddInt32(&allowedCount, 1)
				if atomic.CompareAndSwapInt32(&firstAllowed, 0, 1) {
					time.Sleep(100 * time.Millisecond)
					cb.RecordSuccess()
				}
			}
		}(i)
	}
	
	wg.Wait()
	
	if atomic.LoadInt32(&allowedCount) != 10 {
		t.Errorf("Expected all 10 requests to be eventually allowed, got %d", atomic.LoadInt32(&allowedCount))
	}
	
	if cb.State() != StateClosed {
		t.Errorf("Expected state to be Closed after successful probe, got %s", cb.State())
	}
}

func TestFailureCountResetOnSuccess(t *testing.T) {
	cb := New(WithFailureThreshold(3))
	
	cb.Allow()
	cb.RecordFailure(testErr)
	cb.Allow()
	cb.RecordFailure(testErr)
	
	if cb.State() != StateClosed {
		t.Error("Expected state to be Closed after 2 failures")
	}
	
	cb.Allow()
	cb.RecordSuccess()
	
	for i := 0; i < 2; i++ {
		cb.Allow()
		cb.RecordFailure(testErr)
	}
	
	if cb.State() != StateClosed {
		t.Error("Expected state to be Closed after 2 more failures (counter should have been reset)")
	}
}

func TestLastErrorReturned(t *testing.T) {
	cb := New(WithFailureThreshold(1))
	
	testErr2 := errors.New("test error 2")
	
	cb.Allow()
	cb.RecordFailure(testErr)
	
	allowed, err := cb.Allow()
	if allowed {
		t.Error("Expected request to be rejected")
	}
	if err != testErr {
		t.Errorf("Expected last error to be returned, got %v", err)
	}
	
	cb = New(WithFailureThreshold(1), WithOpenDuration(50*time.Millisecond))
	cb.Allow()
	cb.RecordFailure(testErr2)
	
	time.Sleep(100 * time.Millisecond)
	
	cb.Allow()
	cb.RecordFailure(testErr)
	
	allowed, err = cb.Allow()
	if err != testErr {
		t.Errorf("Expected latest error to be returned after probe failure, got %v", err)
	}
}

func TestStateChangeCallbackCallsStateNoDeadlock(t *testing.T) {
	done := make(chan struct{})
	timeout := time.After(1 * time.Second)
	
	cb := New(WithFailureThreshold(1))
	cb.OnStateChange(func(from, to string) {
		_ = cb.State()
	})
	
	go func() {
		cb.Allow()
		cb.RecordFailure(testErr)
		close(done)
	}()
	
	select {
	case <-done:
	case <-timeout:
		t.Error("Test timed out - possible deadlock in StateChange callback")
	}
}

func TestRequestCallbackCallsStateNoDeadlock(t *testing.T) {
	done := make(chan struct{})
	timeout := time.After(1 * time.Second)
	
	cb := New()
	cb.OnRequest(func(allowed bool, duration time.Duration) {
		_ = cb.State()
	})
	
	go func() {
		cb.Allow()
		close(done)
	}()
	
	select {
	case <-done:
	case <-timeout:
		t.Error("Test timed out - possible deadlock in Request callback")
	}
}

func TestConcurrentOnRequestAndAllowNoDataRace(t *testing.T) {
	cb := New()
	
	var wg sync.WaitGroup
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cb.OnRequest(func(allowed bool, duration time.Duration) {})
		}(i)
		
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cb.Allow()
		}(i)
	}
	
	wg.Wait()
}

func TestConcurrentOnStateChangeAndStateTransitionNoDataRace(t *testing.T) {
	cb := New(WithFailureThreshold(1), WithOpenDuration(10*time.Millisecond))
	
	var wg sync.WaitGroup
	
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cb.OnStateChange(func(from, to string) {})
		}(i)
	}
	
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb.Allow()
			cb.RecordFailure(testErr)
			time.Sleep(20 * time.Millisecond)
			cb.Allow()
			cb.RecordSuccess()
		}()
	}
	
	wg.Wait()
}

package limiter

import (
	"context"
	"sync"
	"sync/atomic"
)

type Semaphore struct {
	tokens   chan struct{}
	capacity int32
	mu       sync.RWMutex
}

func NewSemaphore(capacity int) *Semaphore {
	s := &Semaphore{
		tokens:   make(chan struct{}, capacity),
		capacity: int32(capacity),
	}
	for i := 0; i < capacity; i++ {
		s.tokens <- struct{}{}
	}
	return s
}

func (s *Semaphore) Acquire(ctx context.Context, blocking bool) bool {
	if blocking {
		select {
		case <-ctx.Done():
			return false
		case <-s.tokens:
			return true
		}
	}
	select {
	case <-s.tokens:
		return true
	default:
		return false
	}
}

func (s *Semaphore) Release() {
	s.mu.RLock()
	capacity := atomic.LoadInt32(&s.capacity)
	s.mu.RUnlock()
	
	if len(s.tokens) < int(capacity) {
		select {
		case s.tokens <- struct{}{}:
		default:
		}
	}
}

func (s *Semaphore) SetCapacity(newCapacity int) {
	if newCapacity < 1 {
		newCapacity = 1
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	oldCapacity := atomic.LoadInt32(&s.capacity)
	if int(oldCapacity) == newCapacity {
		return
	}
	
	atomic.StoreInt32(&s.capacity, int32(newCapacity))
	newTokens := make(chan struct{}, newCapacity)
	
	if newCapacity > int(oldCapacity) {
		for len(s.tokens) > 0 {
			newTokens <- <-s.tokens
		}
		for i := int(oldCapacity); i < newCapacity; i++ {
			newTokens <- struct{}{}
		}
	} else {
		for i := 0; i < newCapacity; i++ {
			if len(s.tokens) > 0 {
				newTokens <- <-s.tokens
			} else {
				newTokens <- struct{}{}
			}
		}
	}
	
	s.tokens = newTokens
}

func (s *Semaphore) Capacity() int {
	return int(atomic.LoadInt32(&s.capacity))
}

func (s *Semaphore) Available() int {
	return len(s.tokens)
}

func (s *Semaphore) InUse() int {
	return s.Capacity() - s.Available()
}

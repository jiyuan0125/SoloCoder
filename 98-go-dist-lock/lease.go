package distlock

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrLeaseExpired = errors.New("lease expired")
)

type Lease struct {
	mu         sync.RWMutex
	ownerID    uint64
	leaseTime  time.Duration
	expireTime time.Time
	count      int
}

func NewLease(ownerID uint64, leaseTime time.Duration) *Lease {
	return &Lease{
		ownerID:    ownerID,
		leaseTime:  leaseTime,
		expireTime: time.Now().Add(leaseTime),
		count:      1,
	}
}

func (l *Lease) OwnerID() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.ownerID
}

func (l *Lease) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.count
}

func (l *Lease) LeaseTime() time.Duration {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.leaseTime
}

func (l *Lease) IsExpired() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return time.Now().After(l.expireTime)
}

func (l *Lease) TimeToExpire() time.Duration {
	l.mu.RLock()
	defer l.mu.RUnlock()
	remaining := time.Until(l.expireTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (l *Lease) Renew() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if time.Now().After(l.expireTime) {
		return ErrLeaseExpired
	}
	
	l.expireTime = time.Now().Add(l.leaseTime)
	return nil
}

func (l *Lease) IncrementCount() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.count++
}

func (l *Lease) DecrementCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.count--
	return l.count
}

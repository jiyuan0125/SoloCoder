package jwtcore

import (
	"sync"
	"time"
)

type blacklistEntry struct {
	TokenID   string
	ExpiresAt time.Time
}

type Blacklist struct {
	mu           sync.RWMutex
	tokenIDs     map[string]time.Time
	userTokens   map[string]time.Time
	stopCleanup  chan struct{}
	cleanupDone  chan struct{}
}

func NewBlacklist(cleanupInterval time.Duration) *Blacklist {
	b := &Blacklist{
		tokenIDs:   make(map[string]time.Time),
		userTokens: make(map[string]time.Time),
		stopCleanup:  make(chan struct{}),
		cleanupDone:  make(chan struct{}),
	}

	if cleanupInterval <= 0 {
		cleanupInterval = time.Hour
	}

	go b.startCleanupLoop(cleanupInterval)

	return b
}

func (b *Blacklist) startCleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.cleanupExpired()
		case <-b.stopCleanup:
			close(b.cleanupDone)
			return
		}
	}
}

func (b *Blacklist) cleanupExpired() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()

	for tokenID, expiresAt := range b.tokenIDs {
		if expiresAt.Before(now) {
			delete(b.tokenIDs, tokenID)
		}
	}

	for userID, expiresAt := range b.userTokens {
		if expiresAt.Before(now) {
			delete(b.userTokens, userID)
		}
	}
}

func (b *Blacklist) Stop() {
	select {
	case <-b.stopCleanup:
		return
	default:
		close(b.stopCleanup)
		<-b.cleanupDone
	}
}

func (b *Blacklist) Revoke(tokenID string, expiresAt time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.tokenIDs[tokenID] = expiresAt
}

func (b *Blacklist) RevokeUserTokens(userID string, expiresAt time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.userTokens[userID] = expiresAt
}

func (b *Blacklist) IsRevoked(tokenID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if _, revoked := b.tokenIDs[tokenID]; revoked {
		return true
	}

	return false
}

func (b *Blacklist) IsUserRevoked(userID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	_, revoked := b.userTokens[userID]
	return revoked
}

func (b *Blacklist) Size() (int, int) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.tokenIDs), len(b.userTokens)
}

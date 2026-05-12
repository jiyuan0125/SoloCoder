package main

import (
	"fmt"
	"sync"
	"time"
)

type KeyStatus string

const (
	StatusActive   KeyStatus = "active"
	StatusRotating KeyStatus = "rotating"
	StatusExpired  KeyStatus = "expired"
)

type KeyEntry struct {
	ID        string
	Secret    []byte
	Status    KeyStatus
	CreatedAt int64
	ExpiresAt int64
}

type KeyManager struct {
	mu            sync.RWMutex
	keys          map[string]*KeyEntry
	rotateSeconds int64
}

func NewKeyManager(rotateSeconds int64) *KeyManager {
	km := &KeyManager{
		keys:          make(map[string]*KeyEntry),
		rotateSeconds: rotateSeconds,
	}
	km.generateInitialKey()
	go km.cleanupRoutine()
	return km
}

func (km *KeyManager) generateInitialKey() {
	keyID := GenerateKeyID()
	km.keys[keyID] = &KeyEntry{
		ID:        keyID,
		Secret:    GenerateRandomKey(),
		Status:    StatusActive,
		CreatedAt: NowUnix(),
		ExpiresAt: 0,
	}
}

func (km *KeyManager) GetActiveKey() *KeyEntry {
	km.mu.RLock()
	defer km.mu.RUnlock()

	for _, k := range km.keys {
		if k.Status == StatusActive {
			return k
		}
	}
	return nil
}

func (km *KeyManager) GetKey(kid string) *KeyEntry {
	km.mu.RLock()
	defer km.mu.RUnlock()

	k, ok := km.keys[kid]
	if !ok {
		return nil
	}
	if k.Status == StatusExpired {
		return nil
	}
	return k
}

func (km *KeyManager) Rotate() (*KeyEntry, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	active := km.findActiveLocked()
	if active == nil {
		return nil, fmt.Errorf("no active key")
	}

	active.Status = StatusRotating
	active.ExpiresAt = NowUnix() + km.rotateSeconds

	newKeyID := GenerateKeyID()
	newKey := &KeyEntry{
		ID:        newKeyID,
		Secret:    GenerateRandomKey(),
		Status:    StatusActive,
		CreatedAt: NowUnix(),
		ExpiresAt: 0,
	}
	km.keys[newKeyID] = newKey

	return newKey, nil
}

func (km *KeyManager) findActiveLocked() *KeyEntry {
	for _, k := range km.keys {
		if k.Status == StatusActive {
			return k
		}
	}
	return nil
}

func (km *KeyManager) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		km.cleanupExpired()
	}
}

func (km *KeyManager) cleanupExpired() {
	km.mu.Lock()
	defer km.mu.Unlock()

	now := NowUnix()
	for id, k := range km.keys {
		if k.Status == StatusRotating && k.ExpiresAt > 0 && k.ExpiresAt <= now {
			k.Status = StatusExpired
			delete(km.keys, id)
		}
	}
}

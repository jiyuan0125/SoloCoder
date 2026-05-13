package store

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type KeyStatus string

const (
	StatusActive   KeyStatus = "active"
	StatusRevoked  KeyStatus = "revoked"
)

type APIKey struct {
	ID        string
	Secret    string
	RateLimit int
	Status    KeyStatus
	CreatedAt time.Time
}

type KeyStore struct {
	mu   sync.RWMutex
	keys map[string]*APIKey
}

func NewKeyStore() *KeyStore {
	return &KeyStore{
		keys: make(map[string]*APIKey),
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *KeyStore) Create(rateLimit int) (*APIKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := &APIKey{
		ID:        generateID(),
		Secret:    generateSecret(),
		RateLimit: rateLimit,
		Status:    StatusActive,
		CreatedAt: time.Now(),
	}
	s.keys[key.ID] = key
	return key, nil
}

func (s *KeyStore) Revoke(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if key, ok := s.keys[id]; ok {
		key.Status = StatusRevoked
		return true
	}
	return false
}

func (s *KeyStore) Get(id string) (*APIKey, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, ok := s.keys[id]
	if !ok {
		return nil, false
	}
	copy := *key
	return &copy, true
}

func (s *KeyStore) List() []*APIKey {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*APIKey, 0, len(s.keys))
	for _, k := range s.keys {
		copy := *k
		result = append(result, &copy)
	}
	return result
}

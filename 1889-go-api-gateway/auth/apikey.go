package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
)

type APIKeyStore struct {
	mu    sync.RWMutex
	keys  map[string]*APIKey
	concs map[string]int
	limit int
}

type APIKey struct {
	ID    string
	Key   string
	Valid bool
}

func NewAPIKeyStore() *APIKeyStore {
	return &APIKeyStore{
		keys:  make(map[string]*APIKey),
		concs: make(map[string]int),
		limit: 5,
	}
}

func (s *APIKeyStore) Create(kid string) (*APIKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.keys[kid]; exists {
		return nil, errors.New("key id already exists")
	}
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}
	key := hex.EncodeToString(keyBytes)
	ak := &APIKey{ID: kid, Key: key, Valid: true}
	s.keys[kid] = ak
	return ak, nil
}

func (s *APIKeyStore) Revoke(kid string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if k, ok := s.keys[kid]; ok {
		k.Valid = false
	}
}

func (s *APIKeyStore) Validate(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, k := range s.keys {
		if k.Key == key && k.Valid {
			return true
		}
	}
	return false
}

func (s *APIKeyStore) GetIDByKey(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id, k := range s.keys {
		if k.Key == key {
			return id, true
		}
	}
	return "", false
}

func (s *APIKeyStore) TryAcquire(kid string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.concs[kid] >= s.limit {
		return false
	}
	s.concs[kid]++
	return true
}

func (s *APIKeyStore) Release(kid string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.concs[kid] > 0 {
		s.concs[kid]--
	}
}

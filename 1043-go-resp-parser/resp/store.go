package resp

import (
	"sync"
	"time"
)

type storeEntry struct {
	Value      string
	ExpireAt   time.Time
	HasExpiry  bool
}

type Store struct {
	mu    sync.RWMutex
	data  map[string]storeEntry
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]storeEntry),
	}
}

func (s *Store) cleanupExpired() {
	now := time.Now()
	for key, entry := range s.data {
		if entry.HasExpiry && entry.ExpireAt.Before(now) {
			delete(s.data, key)
		}
	}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = storeEntry{Value: value}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.data[key]
	if !ok {
		return "", false
	}
	if entry.HasExpiry && time.Now().After(entry.ExpireAt) {
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		s.mu.RLock()
		return "", false
	}
	return entry.Value, true
}

func (s *Store) Del(keys ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, key := range keys {
		if _, ok := s.data[key]; ok {
			delete(s.data, key)
			count++
		}
	}
	return count
}

func (s *Store) Exists(keys ...string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.cleanupExpired()
	count := 0
	for _, key := range keys {
		if _, ok := s.data[key]; ok {
			count++
		}
	}
	return count
}

func (s *Store) Incr(key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.data[key]
	var current int64 = 0
	if ok {
		if entry.HasExpiry && time.Now().After(entry.ExpireAt) {
			delete(s.data, key)
		} else {
			var err error
			current, err = parseInt64(entry.Value)
			if err != nil {
				return 0, err
			}
		}
	}
	current++
	s.data[key] = storeEntry{Value: formatInt64(current)}
	return current, nil
}

func (s *Store) Expire(key string, seconds int64) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.data[key]
	if !ok {
		return 0
	}
	if entry.HasExpiry && time.Now().After(entry.ExpireAt) {
		delete(s.data, key)
		return 0
	}
	entry.HasExpiry = true
	entry.ExpireAt = time.Now().Add(time.Duration(seconds) * time.Second)
	s.data[key] = entry
	return 1
}

func (s *Store) TTL(key string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.data[key]
	if !ok {
		return -2
	}
	if !entry.HasExpiry {
		return -1
	}
	now := time.Now()
	if now.After(entry.ExpireAt) {
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		s.mu.RLock()
		return -2
	}
	ttl := int64(entry.ExpireAt.Sub(now).Seconds())
	if ttl < 0 {
		return 0
	}
	return ttl
}

func (s *Store) GetAllKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.cleanupExpired()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

func parseInt64(s string) (int64, error) {
	var result int64
	var sign int64 = 1
	i := 0
	if len(s) == 0 {
		return 0, NewProtocolError(0, "invalid integer: empty string")
	}
	if s[0] == '-' {
		sign = -1
		i++
	} else if s[0] == '+' {
		i++
	}
	if i == len(s) {
		return 0, NewProtocolError(0, "invalid integer: %s", s)
	}
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, NewProtocolError(0, "invalid integer: %s", s)
		}
		result = result*10 + int64(c-'0')
	}
	return sign * result, nil
}

func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

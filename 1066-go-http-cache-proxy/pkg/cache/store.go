package cache

import (
	"net/http"
	"sync"
	"time"
)

const (
	DefaultTTL       = 60
	MaxHeuristicTTL  = 24 * 3600
)

type CacheEntry struct {
	Key              string
	VaryKey          string
	VaryRequestHeaders http.Header
	StatusCode       int
	Header           http.Header
	Body             []byte
	RequestTime      time.Time
	ResponseTime     time.Time
	ExpiresAt        time.Time
	ETag             *ETag
	LastModified     time.Time
}

type CacheStore struct {
	mu         sync.RWMutex
	entries    map[string]map[string]*CacheEntry
	maxSize    int64
	currentSize int64
	defaultTTL int
}

func NewCacheStore(maxSize int64, defaultTTL int) *CacheStore {
	if maxSize <= 0 {
		maxSize = 100 * 1024 * 1024
	}
	if defaultTTL <= 0 {
		defaultTTL = DefaultTTL
	}
	return &CacheStore{
		entries:    make(map[string]map[string]*CacheEntry),
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
	}
}

func (s *CacheStore) Get(key, varyKey string) (*CacheEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, ok := s.entries[key]
	if !ok {
		return nil, false
	}

	entry, ok := entries[varyKey]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry, true
}

func (s *CacheStore) GetAllVaryEntries(key string) map[string]*CacheEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, ok := s.entries[key]
	if !ok {
		return nil
	}

	result := make(map[string]*CacheEntry)
	for k, v := range entries {
		result[k] = v
	}

	return result
}

func (s *CacheStore) Set(key, varyKey string, resp *http.Response, body []byte, requestTime, responseTime time.Time, reqHeader http.Header) {
	cc := ParseCacheControl(resp.Header)
	if cc.ShouldNotCache() {
		return
	}

	ttl := CalculateTTL(resp, cc, s.defaultTTL)
	if ttl < 0 {
		return
	}

	entrySize := int64(len(body)) + int64(headerSize(resp.Header))
	if entrySize > s.maxSize {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for s.currentSize+entrySize > s.maxSize {
		s.evictOldest()
	}

	var etag *ETag
	if etagValue := resp.Header.Get("ETag"); etagValue != "" {
		etag = ParseETag(etagValue)
	}

	lastModified, _ := ParseLastModified(resp.Header)

	header := make(http.Header)
	for k, v := range resp.Header {
		header[k] = append([]string(nil), v...)
	}

	varyFields := ParseVary(resp.Header)
	varyRequestHeaders := make(http.Header)
	for _, field := range varyFields {
		if values := reqHeader.Values(field); len(values) > 0 {
			varyRequestHeaders[field] = append([]string(nil), values...)
		}
	}

	entry := &CacheEntry{
		Key:                key,
		VaryKey:            varyKey,
		VaryRequestHeaders: varyRequestHeaders,
		StatusCode:         resp.StatusCode,
		Header:             header,
		Body:               append([]byte(nil), body...),
		RequestTime:        requestTime,
		ResponseTime:       responseTime,
		ExpiresAt:          responseTime.Add(time.Duration(ttl) * time.Second),
		ETag:               etag,
		LastModified:       lastModified,
	}

	if s.entries[key] == nil {
		s.entries[key] = make(map[string]*CacheEntry)
	}

	if existing, ok := s.entries[key][varyKey]; ok {
		s.currentSize -= int64(len(existing.Body))
	}

	s.entries[key][varyKey] = entry
	s.currentSize += entrySize
}

func (s *CacheStore) Update(key, varyKey string, resp *http.Response, requestTime, responseTime time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, ok := s.entries[key]
	if !ok {
		return
	}

	entry, ok := entries[varyKey]
	if !ok {
		return
	}

	cc := ParseCacheControl(resp.Header)
	ttl := CalculateTTL(resp, cc, s.defaultTTL)

	if etagValue := resp.Header.Get("ETag"); etagValue != "" {
		entry.ETag = ParseETag(etagValue)
	}

	if lastModified, ok := ParseLastModified(resp.Header); ok {
		entry.LastModified = lastModified
	}

	for k, v := range resp.Header {
		entry.Header[k] = append([]string(nil), v...)
	}

	entry.RequestTime = requestTime
	entry.ResponseTime = responseTime
	entry.ExpiresAt = responseTime.Add(time.Duration(ttl) * time.Second)
}

func (s *CacheStore) Delete(key, varyKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, ok := s.entries[key]
	if !ok {
		return
	}

	if entry, ok := entries[varyKey]; ok {
		s.currentSize -= int64(len(entry.Body))
		delete(entries, varyKey)
	}

	if len(entries) == 0 {
		delete(s.entries, key)
	}
}

func (s *CacheStore) evictOldest() {
	var oldestKey, oldestVaryKey string
	var oldestTime time.Time

	for key, entries := range s.entries {
		for varyKey, entry := range entries {
			if oldestTime.IsZero() || entry.ExpiresAt.Before(oldestTime) {
				oldestKey = key
				oldestVaryKey = varyKey
				oldestTime = entry.ExpiresAt
			}
		}
	}

	if oldestKey != "" {
		if entry, ok := s.entries[oldestKey][oldestVaryKey]; ok {
			s.currentSize -= int64(len(entry.Body))
		}
		delete(s.entries[oldestKey], oldestVaryKey)
		if len(s.entries[oldestKey]) == 0 {
			delete(s.entries, oldestKey)
		}
	}
}

func CalculateTTL(resp *http.Response, cc *CacheDirectives, defaultTTL int) int {
	maxAge := cc.GetMaxAge()
	if maxAge >= 0 {
		return maxAge
	}

	if expires := resp.Header.Get("Expires"); expires != "" {
		if t, err := http.ParseTime(expires); err == nil {
			date, ok := ParseDate(resp.Header)
			if !ok {
				date = time.Now()
			}
			ttl := int(t.Sub(date).Seconds())
			if ttl > 0 {
				return ttl
			}
			return 0
		}
	}

	if lastModified, ok := ParseLastModified(resp.Header); ok {
		date, ok := ParseDate(resp.Header)
		if !ok {
			date = time.Now()
		}
		diff := date.Sub(lastModified).Seconds()
		ttl := int(diff * 0.1)
		if ttl > MaxHeuristicTTL {
			ttl = MaxHeuristicTTL
		}
		if ttl > 0 {
			return ttl
		}
	}

	return defaultTTL
}

func (e *CacheEntry) CalculateAge() int {
	date, ok := ParseDate(e.Header)
	if !ok {
		date = e.ResponseTime
	}

	age := int(time.Now().Sub(date).Seconds())
	if headerAge := e.Header.Get("Age"); headerAge != "" {
		if a, err := parseAge(headerAge); err == nil {
			age += a
		}
	}

	if age < 0 {
		return 0
	}
	return age
}

func parseAge(value string) (int, error) {
	return 0, nil
}

func headerSize(h http.Header) int {
	size := 0
	for k, v := range h {
		size += len(k)
		for _, val := range v {
			size += len(val)
		}
	}
	return size
}

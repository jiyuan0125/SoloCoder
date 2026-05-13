package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CacheEntry struct {
	URL         string
	FilePath    string
	Size        int64
	ExpiresAt   time.Time
	LastAccess  time.Time
	CreatedAt   time.Time
	TTLSeconds  int64
	StatusCode  int
	ContentType string
}

type DailyStats struct {
	Date     string
	Hits     int64
	Misses   int64
	Total    int64
	HitRate  float64
}

type Storage struct {
	cacheDir      string
	indexPath     string
	statsPath     string
	mu            sync.RWMutex
	cacheIndex    map[string]*CacheEntry
	hitStats      map[string]map[string]int64
	defaultTTL    time.Duration
	maxCacheSize  int64
	currentSize   int64
}

const (
	DefaultMaxCacheSize = 1 * 1024 * 1024 * 1024
	DefaultTTL          = 5 * time.Minute
)

var ErrNotFound = errors.New("cache entry not found")

func NewStorage(cacheDir string, defaultTTL time.Duration, maxCacheSize int64) (*Storage, error) {
	if defaultTTL <= 0 {
		defaultTTL = DefaultTTL
	}
	if maxCacheSize <= 0 {
		maxCacheSize = DefaultMaxCacheSize
	}

	s := &Storage{
		cacheDir:     cacheDir,
		indexPath:    filepath.Join(cacheDir, "index.json"),
		statsPath:    filepath.Join(cacheDir, "stats.json"),
		defaultTTL:   defaultTTL,
		maxCacheSize: maxCacheSize,
		cacheIndex:   make(map[string]*CacheEntry),
		hitStats:     make(map[string]map[string]int64),
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	if err := s.loadIndex(); err != nil {
		return nil, err
	}

	if err := s.loadStats(); err != nil {
		return nil, err
	}

	s.calculateCurrentSize()

	return s, nil
}

func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.saveIndexLocked(); err != nil {
		return err
	}
	return s.saveStatsLocked()
}

func (s *Storage) loadIndex() error {
	if _, err := os.Stat(s.indexPath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(s.indexPath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	var entries []*CacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	for _, entry := range entries {
		s.cacheIndex[entry.URL] = entry
	}
	return nil
}

func (s *Storage) loadStats() error {
	if _, err := os.Stat(s.statsPath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(s.statsPath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &s.hitStats)
}

func (s *Storage) calculateCurrentSize() {
	s.currentSize = 0
	for _, entry := range s.cacheIndex {
		s.currentSize += entry.Size
	}
}

func (s *Storage) Get(url string) (*CacheEntry, error) {
	s.mu.RLock()
	entry, exists := s.cacheIndex[url]
	s.mu.RUnlock()

	if !exists {
		return nil, ErrNotFound
	}

	if _, err := os.Stat(entry.FilePath); os.IsNotExist(err) {
		s.mu.Lock()
		delete(s.cacheIndex, url)
		s.currentSize -= entry.Size
		s.saveIndexLocked()
		s.mu.Unlock()
		return nil, ErrNotFound
	}

	return entry, nil
}

func (s *Storage) GetWithStaleCheck(url string) (entry *CacheEntry, isExpired bool, err error) {
	s.mu.RLock()
	entry, exists := s.cacheIndex[url]
	s.mu.RUnlock()

	if !exists {
		return nil, false, ErrNotFound
	}

	if _, err := os.Stat(entry.FilePath); os.IsNotExist(err) {
		s.mu.Lock()
		delete(s.cacheIndex, url)
		s.currentSize -= entry.Size
		s.saveIndexLocked()
		s.mu.Unlock()
		return nil, false, ErrNotFound
	}

	isExpired = time.Now().After(entry.ExpiresAt)
	return entry, isExpired, nil
}

func (s *Storage) Put(url string, body io.Reader, ttl time.Duration, statusCode int, contentType string) (*CacheEntry, error) {
	if ttl <= 0 {
		ttl = s.defaultTTL
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	hash := sha256.Sum256([]byte(url))
	fileName := hex.EncodeToString(hash[:])
	filePath := filepath.Join(s.cacheDir, fileName)

	if oldEntry, exists := s.cacheIndex[url]; exists {
		os.Remove(oldEntry.FilePath)
		s.currentSize -= oldEntry.Size
	}

	tempFile, err := os.CreateTemp(s.cacheDir, "tmp_")
	if err != nil {
		return nil, err
	}

	size, err := io.Copy(tempFile, body)
	if err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, err
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempFile.Name())
		return nil, err
	}

	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		os.Remove(tempFile.Name())
		return nil, err
	}

	entry := &CacheEntry{
		URL:         url,
		FilePath:    filePath,
		Size:        size,
		ExpiresAt:   time.Now().Add(ttl),
		LastAccess:  time.Now(),
		CreatedAt:   time.Now(),
		TTLSeconds:  int64(ttl.Seconds()),
		StatusCode:  statusCode,
		ContentType: contentType,
	}

	s.currentSize += size
	s.cacheIndex[url] = entry

	s.evictIfNeededLocked()

	if err := s.saveIndexLocked(); err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *Storage) UpdateLastAccess(url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.cacheIndex[url]
	if !exists {
		return ErrNotFound
	}

	entry.LastAccess = time.Now()
	return s.saveIndexLocked()
}

func (s *Storage) Delete(url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.cacheIndex[url]
	if !exists {
		return ErrNotFound
	}

	if err := os.Remove(entry.FilePath); err != nil && !os.IsNotExist(err) {
		return err
	}

	s.currentSize -= entry.Size
	delete(s.cacheIndex, url)

	return s.saveIndexLocked()
}

func (s *Storage) RecordHit(url string, hit bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	today := time.Now().Format("2006-01-02")

	if _, exists := s.hitStats[today]; !exists {
		s.hitStats[today] = map[string]int64{
			"hits":   0,
			"misses": 0,
		}
	}

	if hit {
		s.hitStats[today]["hits"]++
	} else {
		s.hitStats[today]["misses"]++
	}

	return s.saveStatsLocked()
}

func (s *Storage) GetStats(days int) []DailyStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]DailyStats, 0, days)

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		stats := DailyStats{
			Date:   date,
			Hits:   0,
			Misses: 0,
		}

		if dayStats, exists := s.hitStats[date]; exists {
			stats.Hits = dayStats["hits"]
			stats.Misses = dayStats["misses"]
		}

		stats.Total = stats.Hits + stats.Misses
		if stats.Total > 0 {
			stats.HitRate = float64(stats.Hits) / float64(stats.Total)
		}

		result = append(result, stats)
	}

	return result
}

func (s *Storage) evictIfNeededLocked() {
	if s.currentSize <= s.maxCacheSize {
		return
	}

	var lruEntries []*CacheEntry
	for _, entry := range s.cacheIndex {
		lruEntries = append(lruEntries, entry)
	}

	for i := 0; i < len(lruEntries); i++ {
		for j := i + 1; j < len(lruEntries); j++ {
			if lruEntries[i].LastAccess.After(lruEntries[j].LastAccess) {
				lruEntries[i], lruEntries[j] = lruEntries[j], lruEntries[i]
			}
		}
	}

	for _, entry := range lruEntries {
		if s.currentSize <= s.maxCacheSize {
			break
		}

		os.Remove(entry.FilePath)
		s.currentSize -= entry.Size
		delete(s.cacheIndex, entry.URL)
	}
}

func (s *Storage) saveIndexLocked() error {
	var entries []*CacheEntry
	for _, entry := range s.cacheIndex {
		entries = append(entries, entry)
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	tempPath := s.indexPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, s.indexPath)
}

func (s *Storage) saveStatsLocked() error {
	data, err := json.Marshal(s.hitStats)
	if err != nil {
		return err
	}

	tempPath := s.statsPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, s.statsPath)
}

func (s *Storage) ReadBody(entry *CacheEntry) (io.ReadCloser, error) {
	file, err := os.Open(entry.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache file: %w", err)
	}
	return file, nil
}

func (s *Storage) GetStatsForURL(url string) (hits, misses int64, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return 0, 0, nil
}

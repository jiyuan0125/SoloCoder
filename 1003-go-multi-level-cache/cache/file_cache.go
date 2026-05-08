package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type fileItem struct {
	Key      string          `json:"key"`
	Value    json.RawMessage `json:"value"`
	ExpireAt time.Time       `json:"expire_at"`
}

type FileCache struct {
	dir           string
	cleanInterval time.Duration
	stopChan      chan struct{}
	once          sync.Once
	cleaning      int32
	mu            sync.RWMutex
}

func NewFileCache(dir string, cleanInterval time.Duration) *FileCache {
	if cleanInterval <= 0 {
		cleanInterval = 60 * time.Second
	}
	return &FileCache{
		dir:           dir,
		cleanInterval: cleanInterval,
		stopChan:      make(chan struct{}),
	}
}

func (f *FileCache) Start() error {
	if err := os.MkdirAll(f.dir, 0755); err != nil {
		return err
	}

	f.once.Do(func() {
		go f.cleanLoop()
	})
	return nil
}

func (f *FileCache) Stop() {
	close(f.stopChan)
}

func (f *FileCache) Get(key string) (json.RawMessage, time.Time, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	filename := f.keyToFilename(key)
	path := filepath.Join(f.dir, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, time.Time{}, false
	}

	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		os.Remove(path)
		return nil, time.Time{}, false
	}

	if item.Key != key {
		return nil, time.Time{}, false
	}

	if !item.ExpireAt.IsZero() && time.Now().After(item.ExpireAt) {
		os.Remove(path)
		return nil, time.Time{}, false
	}

	return item.Value, item.ExpireAt, true
}

func (f *FileCache) Set(key string, value json.RawMessage, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var expireAt time.Time
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}

	item := fileItem{
		Key:      key,
		Value:    value,
		ExpireAt: expireAt,
	}

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	filename := f.keyToFilename(key)
	path := filepath.Join(f.dir, filename)
	return os.WriteFile(path, data, 0644)
}

func (f *FileCache) Delete(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	filename := f.keyToFilename(key)
	path := filepath.Join(f.dir, filename)

	if err := os.Remove(path); err != nil {
		return false
	}
	return true
}

func (f *FileCache) DeletePrefix(prefix string) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	count := 0
	files, err := os.ReadDir(f.dir)
	if err != nil {
		return 0
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(f.dir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var item fileItem
		if err := json.Unmarshal(data, &item); err != nil {
			continue
		}

		if len(item.Key) >= len(prefix) && item.Key[:len(prefix)] == prefix {
			os.Remove(path)
			count++
		}
	}
	return count
}

func (f *FileCache) Clear() {
	f.mu.Lock()
	defer f.mu.Unlock()

	files, _ := os.ReadDir(f.dir)
	for _, file := range files {
		if !file.IsDir() {
			os.Remove(filepath.Join(f.dir, file.Name()))
		}
	}
}

func (f *FileCache) Count() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	files, err := os.ReadDir(f.dir)
	if err != nil {
		return 0
	}

	count := 0
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			count++
		}
	}
	return count
}

func (f *FileCache) cleanLoop() {
	ticker := time.NewTicker(f.cleanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			f.cleanExpired()
		case <-f.stopChan:
			return
		}
	}
}

func (f *FileCache) cleanExpired() {
	if !atomic.CompareAndSwapInt32(&f.cleaning, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&f.cleaning, 0)

	f.mu.RLock()
	defer f.mu.RUnlock()

	files, err := os.ReadDir(f.dir)
	if err != nil {
		return
	}

	now := time.Now()
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(f.dir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var item fileItem
		if err := json.Unmarshal(data, &item); err != nil {
			os.Remove(path)
			continue
		}

		if !item.ExpireAt.IsZero() && now.After(item.ExpireAt) {
			os.Remove(path)
		}
	}
}

func (f *FileCache) keyToFilename(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:]) + ".json"
}

func (f *FileCache) WarmUp(memory *MemoryCache, loader func(key string, value json.RawMessage)) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	files, err := os.ReadDir(f.dir)
	if err != nil {
		return
	}

	now := time.Now()
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(f.dir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var item fileItem
		if err := json.Unmarshal(data, &item); err != nil {
			continue
		}

		if !item.ExpireAt.IsZero() && now.After(item.ExpireAt) {
			continue
		}

		loader(item.Key, item.Value)
	}
}

func (f *FileCache) CountWarmUp() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	files, err := os.ReadDir(f.dir)
	if err != nil {
		return 0
	}

	count := 0
	now := time.Now()
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(f.dir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var item fileItem
		if err := json.Unmarshal(data, &item); err != nil {
			continue
		}

		if !item.ExpireAt.IsZero() && now.After(item.ExpireAt) {
			continue
		}

		count++
	}
	return count
}

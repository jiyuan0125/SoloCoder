package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type cacheItem struct {
	key      string
	value    string
	expiresAt time.Time
	element  *list.Element
}

type subscriber struct {
	callback func(string)
}

type Cache struct {
	mu          sync.RWMutex
	data        map[string]*cacheItem
	lru         *list.List
	subscribers map[string][]*subscriber

	capacityBytes int64
	usedBytes     int64

	statsHits   int64
	statsMisses int64
}

const defaultCapacityBytes int64 = 100 * 1024 * 1024

func NewCache(capacityBytes int64) *Cache {
	if capacityBytes <= 0 {
		capacityBytes = defaultCapacityBytes
	}
	return &Cache{
		data:          make(map[string]*cacheItem),
		lru:           list.New(),
		subscribers:   make(map[string][]*subscriber),
		capacityBytes: capacityBytes,
	}
}

func (c *Cache) itemSize(key, value string) int64 {
	return int64(len(key) + len(value))
}

func (c *Cache) isExpired(item *cacheItem) bool {
	if item.expiresAt.IsZero() {
		return false
	}
	return time.Now().After(item.expiresAt)
}

func (c *Cache) evictLRU() {
	itemsToEvict := c.lru.Len() / 10
	if itemsToEvict < 1 {
		itemsToEvict = 1
	}
	for i := 0; i < itemsToEvict; i++ {
		if c.lru.Len() == 0 {
			break
		}
		elem := c.lru.Back()
		if elem == nil {
			break
		}
		item := elem.Value.(*cacheItem)
		c.removeItemInternal(item.key)
	}
}

func (c *Cache) removeItemInternal(key string) {
	item, exists := c.data[key]
	if !exists {
		return
	}
	c.usedBytes -= c.itemSize(item.key, item.value)
	c.lru.Remove(item.element)
	delete(c.data, key)
}

func (c *Cache) Set(key, value string, ttlSeconds int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, exists := c.data[key]; exists {
		c.usedBytes -= c.itemSize(existing.key, existing.value)
		c.lru.Remove(existing.element)
		delete(c.data, key)
	}

	size := c.itemSize(key, value)
	for c.usedBytes+size > c.capacityBytes {
		if c.lru.Len() == 0 {
			break
		}
		c.evictLRU()
	}

	item := &cacheItem{
		key:   key,
		value: value,
	}
	if ttlSeconds > 0 {
		item.expiresAt = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	}

	item.element = c.lru.PushFront(item)
	c.data[key] = item
	c.usedBytes += size
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.data[key]
	if !exists {
		c.statsMisses++
		return "", false
	}

	if c.isExpired(item) {
		c.removeItemInternal(key)
		c.statsMisses++
		return "", false
	}

	c.lru.MoveToFront(item.element)
	c.statsHits++
	return item.value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	exists := false
	if item, ok := c.data[key]; ok {
		if !c.isExpired(item) {
			exists = true
		}
		c.removeItemInternal(key)
	}
	subs := c.subscribers[key]
	c.mu.Unlock()

	if exists && len(subs) > 0 {
		for _, sub := range subs {
			sub.callback(key)
		}
	}
}

func (c *Cache) Subscribe(key string, callback func(string)) *subscriber {
	c.mu.Lock()
	defer c.mu.Unlock()

	sub := &subscriber{callback: callback}
	c.subscribers[key] = append(c.subscribers[key], sub)
	return sub
}

func (c *Cache) Unsubscribe(key string, sub *subscriber) {
	c.mu.Lock()
	defer c.mu.Unlock()

	subs := c.subscribers[key]
	for i, s := range subs {
		if s == sub {
			c.subscribers[key] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}

type Stats struct {
	HitRate     float64 `json:"hit_rate"`
	KeyCount    int     `json:"key_count"`
	MemoryBytes int64   `json:"memory_bytes"`
}

func (c *Cache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.statsHits + c.statsMisses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.statsHits) / float64(total)
	}

	return Stats{
		HitRate:     hitRate,
		KeyCount:    len(c.data),
		MemoryBytes: c.usedBytes,
	}
}

func (c *Cache) CleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	batchSize := 100
	count := 0

	for key, item := range c.data {
		if count >= batchSize {
			break
		}
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			c.removeItemInternal(key)
			count++
		}
	}
}

func (c *Cache) StartCleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			c.CleanupExpired()
		}
	}()
}

type Handler struct {
	cache *Cache
}

func NewHandler(cache *Cache) *Handler {
	return &Handler{cache: cache}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/cache/stats" {
		h.handleStats(w, r)
		return
	}

	if strings.HasPrefix(path, "/cache/") {
		rest := path[len("/cache/"):]
		subscribeSuffix := "/subscribe"
		if strings.HasSuffix(rest, subscribeSuffix) {
			key := strings.TrimSuffix(rest, subscribeSuffix)
			if key != "" {
				h.handleSubscribe(w, r, key)
				return
			}
		} else if rest != "" {
			h.handleCache(w, r, rest)
			return
		}
	}

	http.NotFound(w, r)
}

func (h *Handler) handleCache(w http.ResponseWriter, r *http.Request, key string) {
	switch r.Method {
	case http.MethodPost:
		h.handleSet(w, r, key)
	case http.MethodGet:
		h.handleGet(w, r, key)
	case http.MethodDelete:
		h.handleDelete(w, r, key)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleSet(w http.ResponseWriter, r *http.Request, key string) {
	ttlStr := r.URL.Query().Get("ttl")
	if ttlStr == "" {
		http.Error(w, "ttl parameter required", http.StatusBadRequest)
		return
	}

	ttl, err := strconv.Atoi(ttlStr)
	if err != nil || ttl < 0 {
		http.Error(w, "invalid ttl", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	h.cache.Set(key, string(body), ttl)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request, key string) {
	value, ok := h.cache.Get(key)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request, key string) {
	h.cache.Delete(key)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.cache.Stats()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) handleSubscribe(w http.ResponseWriter, r *http.Request, key string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	notify := make(chan string, 1)
	sub := h.cache.Subscribe(key, func(k string) {
		select {
		case notify <- k:
		default:
		}
	})
	defer h.cache.Unsubscribe(key, sub)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case k := <-notify:
			fmt.Fprintf(w, "data: %s deleted\n\n", k)
			flusher.Flush()
		}
	}
}

func getCapacityFromEnv() int64 {
	capStr := os.Getenv("CAPACITY_BYTES")
	if capStr == "" {
		return defaultCapacityBytes
	}
	cap, err := strconv.ParseInt(capStr, 10, 64)
	if err != nil || cap <= 0 {
		return defaultCapacityBytes
	}
	return cap
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	capacity := getCapacityFromEnv()
	cache := NewCache(capacity)
	cache.StartCleanupLoop()

	handler := NewHandler(cache)

	fmt.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

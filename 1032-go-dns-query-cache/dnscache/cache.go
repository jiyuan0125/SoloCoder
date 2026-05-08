package dnscache

import (
	"dns-cache-service/common"
	"sync"
	"time"
)

type cacheKey struct {
	Domain string
	Type   common.RecordType
}

type cacheEntry struct {
	Domain    string
	Type      common.RecordType
	Records   []common.DNSRecord
	ExpiresAt time.Time
	IsNXDomain bool
}

type Cache struct {
	mu      sync.RWMutex
	entries map[cacheKey]*cacheEntry
	config  *Config
}

func NewCache(config *Config) *Cache {
	if config == nil {
		config = DefaultConfig()
	}
	return &Cache{
		entries: make(map[cacheKey]*cacheEntry),
		config:  config,
	}
}

func (c *Cache) Get(domain string, recordType common.RecordType) (*cacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := cacheKey{Domain: domain, Type: recordType}
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry, true
}

func (c *Cache) Set(domain string, recordType common.RecordType, records []common.DNSRecord, isNXDomain bool) {
	if len(records) > 0 && records[0].TTL == 0 {
		return
	}

	ttl := c.config.DefaultTTL
	if len(records) > 0 {
		ttl = records[0].TTL
	}

	if isNXDomain {
		ttl = int(float64(ttl) * c.config.NXDomainTTLMultiplier)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey{Domain: domain, Type: recordType}
	c.entries[key] = &cacheEntry{
		Domain:    domain,
		Type:      recordType,
		Records:   records,
		ExpiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
		IsNXDomain: isNXDomain,
	}
}

func (c *Cache) Delete(domain string, recordType common.RecordType) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey{Domain: domain, Type: recordType}
	delete(c.entries, key)
}

func (c *Cache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}

func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

func (c *Cache) GetAllEntries() []*cacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := make([]*cacheEntry, 0, len(c.entries))
	for _, entry := range c.entries {
		if !time.Now().After(entry.ExpiresAt) {
			entries = append(entries, entry)
		}
	}
	return entries
}

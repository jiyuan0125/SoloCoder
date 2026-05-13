package tagstore

import (
	"strings"
	"sync"
)

type TagStore struct {
	mu       sync.RWMutex
	counters map[string]int64
}

func NewTagStore() *TagStore {
	return &TagStore{
		counters: make(map[string]int64),
	}
}

func (s *TagStore) Increment(tags map[string]string) {
	if len(tags) == 0 {
		return
	}
	key := tagsToKey(tags)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[key]++
}

func (s *TagStore) List() []TagCount {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]TagCount, 0, len(s.counters))
	for key, count := range s.counters {
		tags := keyToTags(key)
		result = append(result, TagCount{
			Tags:  tags,
			Count: count,
		})
	}
	return result
}

func (s *TagStore) Get(tags map[string]string) (int64, bool) {
	key := tagsToKey(tags)
	s.mu.RLock()
	defer s.mu.RUnlock()
	count, exists := s.counters[key]
	return count, exists
}

type TagCount struct {
	Tags  map[string]string `json:"tags"`
	Count int64             `json:"count"`
}

func tagsToKey(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	stableSortStrings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+tags[k])
	}
	return strings.Join(parts, ",")
}

func keyToTags(key string) map[string]string {
	if key == "" {
		return nil
	}
	parts := strings.Split(key, ",")
	tags := make(map[string]string, len(parts))
	for _, part := range parts {
		idx := strings.Index(part, "=")
		if idx > 0 {
			k := part[:idx]
			v := part[idx+1:]
			tags[k] = v
		}
	}
	return tags
}

func stableSortStrings(keys []string) {
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
}

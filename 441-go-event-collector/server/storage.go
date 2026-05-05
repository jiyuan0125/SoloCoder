package server

import (
	"sort"
	"sync"
	"time"

	"event-collector/common"
)

type MemoryStorage struct {
	mu     sync.RWMutex
	events []*common.Event
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		events: make([]*common.Event, 0),
	}
}

func (s *MemoryStorage) Store(event *common.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
}

func (s *MemoryStorage) Query(eventName, userID string, startTime, endTime int64, page, pageSize int) ([]common.Event, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]*common.Event, 0)

	for _, event := range s.events {
		if eventName != "" && event.Name != eventName {
			continue
		}
		if userID != "" && event.UserID != userID {
			continue
		}
		if startTime > 0 && event.Timestamp < startTime {
			continue
		}
		if endTime > 0 && event.Timestamp > endTime {
			continue
		}
		filtered = append(filtered, event)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp > filtered[j].Timestamp
	})

	total := len(filtered)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []common.Event{}, total
	}

	if end > total {
		end = total
	}

	result := make([]common.Event, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, *filtered[i])
	}

	return result, total
}

func (s *MemoryStorage) GetAllEvents() []*common.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*common.Event, len(s.events))
	copy(result, s.events)
	return result
}

func (s *MemoryStorage) GetEventsByTimeRange(startTime, endTime int64) []*common.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*common.Event, 0)
	for _, event := range s.events {
		if event.Timestamp >= startTime && event.Timestamp <= endTime {
			result = append(result, event)
		}
	}
	return result
}

func (s *MemoryStorage) GetTodayEvents() []*common.Event {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location()).Unix()

	return s.GetEventsByTimeRange(startOfDay, endOfDay)
}

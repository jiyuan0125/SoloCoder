package main

import (
	"strconv"
	"sync"
	"time"

	"progress-tracker/common"
	"progress-tracker/progress"
)

type progressEntry struct {
	tracker     *progress.Tracker
	description string
	createdAt   time.Time
}

type ProgressStore struct {
	mu       sync.RWMutex
	entries  map[string]*progressEntry
	nextID   int64
}

func NewProgressStore() *ProgressStore {
	return &ProgressStore{
		entries: make(map[string]*progressEntry),
		nextID:  1,
	}
}

func (s *ProgressStore) Create(total int64, description string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.generateID()
	tracker := progress.New(total)

	s.entries[id] = &progressEntry{
		tracker:     tracker,
		description: description,
		createdAt:   time.Now(),
	}

	return id
}

func (s *ProgressStore) Get(id string) (*common.ProgressStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.entries[id]
	if !exists {
		return nil, false
	}

	return s.toStatus(id, entry), true
}

func (s *ProgressStore) Update(id string, delta int64) bool {
	s.mu.RLock()
	entry, exists := s.entries[id]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	entry.tracker.Update(delta)
	return true
}

func (s *ProgressStore) Set(id string, current int64) bool {
	s.mu.RLock()
	entry, exists := s.entries[id]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	entry.tracker.Set(current)
	return true
}

func (s *ProgressStore) Cancel(id string) bool {
	s.mu.RLock()
	entry, exists := s.entries[id]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	entry.tracker.Cancel()
	return true
}

func (s *ProgressStore) Close(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.entries[id]
	if !exists {
		return false
	}

	entry.tracker.Close()
	delete(s.entries, id)
	return true
}

func (s *ProgressStore) List() []common.ProgressStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statuses := make([]common.ProgressStatus, 0, len(s.entries))
	for id, entry := range s.entries {
		statuses = append(statuses, *s.toStatus(id, entry))
	}
	return statuses
}

func (s *ProgressStore) AddSubTask(parentID string, total int64, weight float64) (string, bool) {
	s.mu.RLock()
	parentEntry, exists := s.entries[parentID]
	s.mu.RUnlock()

	if !exists {
		return "", false
	}

	var subTracker *progress.Tracker
	if weight > 0 {
		subTracker = parentEntry.tracker.AddSubTaskWithWeight(total, weight)
	} else {
		subTracker = parentEntry.tracker.AddSubTask(total)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.generateID()
	s.entries[id] = &progressEntry{
		tracker:     subTracker,
		description: "Subtask of " + parentID,
		createdAt:   time.Now(),
	}

	return id, true
}

func (s *ProgressStore) generateID() string {
	id := s.nextID
	s.nextID++
	return "p_" + strconv.FormatInt(id, 10)
}

func (s *ProgressStore) toStatus(id string, entry *progressEntry) *common.ProgressStatus {
	tracker := entry.tracker
	remaining := tracker.RemainingTime()
	
	select {
	case <-tracker.Done():
	default:
		remaining = tracker.RemainingTime()
	}

	return &common.ProgressStatus{
		ID:          id,
		Current:     tracker.Current(),
		Total:       tracker.Total(),
		Percentage:  tracker.Percentage(),
		Remaining:   remaining,
		Description: entry.description,
		IsDone:      tracker.IsDone(),
		IsCancelled: tracker.IsCancelled(),
		IsClosed:    tracker.IsClosed(),
		CreatedAt:   entry.createdAt,
	}
}

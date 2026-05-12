package main

import (
	"time"
)

func NewStatsStore() *StatsStore {
	return &StatsStore{
		stats:   make(map[StatsKey]*Stats),
		counter: make(map[StatsKey]*Counter),
	}
}

func (s *StatsStore) GetCounter(key StatsKey) *Counter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.counter[key]
}

func (s *StatsStore) GetOrCreateCounter(key StatsKey) *Counter {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.counter[key] == nil {
		s.counter[key] = &Counter{
			FixedWindows: make(map[int64]int64),
			SlidingLogs:  make([]time.Time, 0),
		}
	}
	return s.counter[key]
}

func (s *StatsStore) RecordTrigger(key StatsKey, dimension Dimension) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stats[key] == nil {
		s.stats[key] = &Stats{
			RuleID:    key.RuleID,
			Dimension: dimension,
		}
	}
	s.stats[key].TriggerCount++
	s.stats[key].LastTriggerTime = time.Now()
}

func (s *StatsStore) GetAllStats() []Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Stats, 0, len(s.stats))
	for _, st := range s.stats {
		result = append(result, *st)
	}
	return result
}

func (s *StatsStore) GetCurrentCount(key StatsKey, rule Rule, now time.Time) (int64, time.Duration) {
	if rule.Mode == ModeFixed {
		return s.getFixedWindowCount(key, rule, now)
	}
	return s.getSlidingWindowCount(key, rule, now)
}

func (s *StatsStore) getFixedWindowCount(key StatsKey, rule Rule, now time.Time) (int64, time.Duration) {
	windowSec := int64(rule.WindowSeconds)
	windowStart := now.Unix() / windowSec * windowSec
	windowEnd := windowStart + windowSec
	retryAfter := time.Duration(windowEnd-now.Unix()) * time.Second

	counter := s.GetCounter(key)
	if counter == nil {
		return 0, retryAfter
	}

	return counter.FixedWindows[windowStart], retryAfter
}

func (s *StatsStore) getSlidingWindowCount(key StatsKey, rule Rule, now time.Time) (int64, time.Duration) {
	counter := s.GetCounter(key)
	if counter == nil {
		return 0, time.Duration(rule.WindowSeconds) * time.Second
	}

	windowStart := now.Add(-time.Duration(rule.WindowSeconds) * time.Second)
	count := int64(0)
	for _, t := range counter.SlidingLogs {
		if t.After(windowStart) {
			count++
		}
	}

	retryAfter := time.Duration(rule.WindowSeconds) * time.Second
	return count, retryAfter
}

func (s *StatsStore) IncrementCount(key StatsKey, rule Rule, now time.Time) bool {
	counter := s.GetOrCreateCounter(key)

	if rule.Mode == ModeFixed {
		windowSec := int64(rule.WindowSeconds)
		windowStart := now.Unix() / windowSec * windowSec
		newCount := counter.FixedWindows[windowStart] + 1
		counter.FixedWindows[windowStart] = newCount

		oldWindowStart := windowStart - windowSec
		delete(counter.FixedWindows, oldWindowStart)

		return newCount > int64(rule.MaxRequests)
	}

	counter.SlidingLogs = append(counter.SlidingLogs, now)

	windowStart := now.Add(-time.Duration(rule.WindowSeconds) * time.Second)
	validLogs := make([]time.Time, 0, len(counter.SlidingLogs))
	for _, t := range counter.SlidingLogs {
		if t.After(windowStart) {
			validLogs = append(validLogs, t)
		}
	}
	counter.SlidingLogs = validLogs

	return int64(len(validLogs)) > int64(rule.MaxRequests)
}

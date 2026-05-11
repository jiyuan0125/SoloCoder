package core

import (
	"sync"

	"supervision-log-system/common"
)

type Store struct {
	mu sync.RWMutex

	supervisoryRecords map[string]*common.SupervisoryRecord
	acceptanceRecords  map[string]*common.AcceptanceRecord
	rectificationNotices map[string]*common.RectificationNotice
	issues           map[string]*common.Issue
	dailyLogs        map[string]*common.DailyLog
}

func NewStore() *Store {
	return &Store{
		supervisoryRecords: make(map[string]*common.SupervisoryRecord),
		acceptanceRecords:  make(map[string]*common.AcceptanceRecord),
		rectificationNotices: make(map[string]*common.RectificationNotice),
		issues:           make(map[string]*common.Issue),
		dailyLogs:        make(map[string]*common.DailyLog),
	}
}

func (s *Store) Lock() {
	s.mu.Lock()
}

func (s *Store) Unlock() {
	s.mu.Unlock()
}

func (s *Store) RLock() {
	s.mu.RLock()
}

func (s *Store) RUnlock() {
	s.mu.RUnlock()
}

func (s *Store) GetSupervisoryRecords() map[string]*common.SupervisoryRecord {
	return s.supervisoryRecords
}

func (s *Store) GetAcceptanceRecords() map[string]*common.AcceptanceRecord {
	return s.acceptanceRecords
}

func (s *Store) GetRectificationNotices() map[string]*common.RectificationNotice {
	return s.rectificationNotices
}

func (s *Store) GetIssues() map[string]*common.Issue {
	return s.issues
}

func (s *Store) GetDailyLogs() map[string]*common.DailyLog {
	return s.dailyLogs
}

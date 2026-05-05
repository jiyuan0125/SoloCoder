package server

import (
	"fmt"
	"sync"
	"time"
	"userbehavior/internal/shared"
)

type Store struct {
	mu           sync.RWMutex
	behaviors    map[string]map[string][]*shared.Behavior
	sessions     map[string]*shared.Session
	userSessions map[string]map[string]bool
	userBehaviors map[string][]*shared.Behavior
}

func NewStore() *Store {
	return &Store{
		behaviors:     make(map[string]map[string][]*shared.Behavior),
		sessions:      make(map[string]*shared.Session),
		userSessions:  make(map[string]map[string]bool),
		userBehaviors: make(map[string][]*shared.Behavior),
	}
}

func (s *Store) getDateKey(t time.Time) string {
	return t.Format("2006-01-02")
}

func (s *Store) getHourKey(t time.Time) string {
	return t.Format("2006-01-02-15")
}

func (s *Store) getMinuteKey(t time.Time) string {
	return t.Format("2006-01-02-15-04")
}

func (s *Store) TrackBehavior(behavior *shared.Behavior) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dateKey := s.getDateKey(behavior.Timestamp)

	if _, exists := s.behaviors[dateKey]; !exists {
		s.behaviors[dateKey] = make(map[string][]*shared.Behavior)
	}

	userMinuteKey := fmt.Sprintf("%s-%s", behavior.UserID, s.getMinuteKey(behavior.Timestamp))
	s.behaviors[dateKey][userMinuteKey] = append(s.behaviors[dateKey][userMinuteKey], behavior)

	if _, exists := s.userBehaviors[behavior.UserID]; !exists {
		s.userBehaviors[behavior.UserID] = make([]*shared.Behavior, 0)
	}
	s.userBehaviors[behavior.UserID] = append(s.userBehaviors[behavior.UserID], behavior)

	return nil
}

func (s *Store) GetOrCreateSession(userID, sessionID string) (*shared.Session, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	if sessionID != "" {
		if session, exists := s.sessions[sessionID]; exists {
			if now.Sub(session.LastActivity) < shared.SessionTimeout {
				session.LastActivity = now
				return session, false, nil
			}
			session.IsActive = false
		}
	}

	newSessionID := sessionID
	if newSessionID == "" {
		newSessionID = fmt.Sprintf("sess-%s-%d", userID, now.UnixNano())
	}

	newSession := &shared.Session{
		ID:           newSessionID,
		UserID:       userID,
		StartTime:    now,
		LastActivity: now,
		IsActive:     true,
	}

	s.sessions[newSessionID] = newSession

	if _, exists := s.userSessions[userID]; !exists {
		s.userSessions[userID] = make(map[string]bool)
	}
	s.userSessions[userID][newSessionID] = true

	return newSession, true, nil
}

func (s *Store) CheckDeduplication(userID, pageKey string, timestamp time.Time, behaviorType shared.BehaviorType) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if behaviorType != shared.BehaviorTypePageView {
		return false
	}

	dateKey := s.getDateKey(timestamp)
	minuteKey := s.getMinuteKey(timestamp)

	if dayBehaviors, exists := s.behaviors[dateKey]; exists {
		userMinuteKey := fmt.Sprintf("%s-%s", userID, minuteKey)
		if behaviors, exists := dayBehaviors[userMinuteKey]; exists {
			for _, b := range behaviors {
				if b.Type == shared.BehaviorTypePageView && 
				   b.PageView != nil && 
				   b.PageView.PageKey == pageKey &&
				   !b.IsFiltered {
					return true
				}
			}
		}
	}

	return false
}

func (s *Store) GetBehaviorsByDateRange(start, end time.Time) []*shared.Behavior {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*shared.Behavior, 0)

	startDate := start.Truncate(24 * time.Hour)
	endDate := end.Truncate(24 * time.Hour).Add(24 * time.Hour)

	for d := startDate; d.Before(endDate); d = d.Add(24 * time.Hour) {
		dateKey := s.getDateKey(d)
		if dayBehaviors, exists := s.behaviors[dateKey]; exists {
			for _, behaviors := range dayBehaviors {
				for _, b := range behaviors {
					if !b.IsFiltered && b.Timestamp.After(start) && b.Timestamp.Before(end) {
						result = append(result, b)
					}
				}
			}
		}
	}

	return result
}

func (s *Store) GetUserBehaviors(userID string) []*shared.Behavior {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.userBehaviors[userID]
}

func (s *Store) GetAllUsers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]string, 0, len(s.userBehaviors))
	for userID := range s.userBehaviors {
		users = append(users, userID)
	}
	return users
}

func (s *Store) GetActiveSessions() []*shared.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	active := make([]*shared.Session, 0)
	for _, session := range s.sessions {
		if session.IsActive && now.Sub(session.LastActivity) < shared.SessionTimeout {
			active = append(active, session)
		}
	}
	return active
}

func (s *Store) GetUserSessions(userID string) []*shared.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*shared.Session, 0)
	if sessionIDs, exists := s.userSessions[userID]; exists {
		for sessionID := range sessionIDs {
			if session, exists := s.sessions[sessionID]; exists {
				sessions = append(sessions, session)
			}
		}
	}
	return sessions
}

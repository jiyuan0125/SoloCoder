package server

import (
	"sync"
	"time"

	"github.com/comment-moderator/pkg/common"
)

type Store struct {
	comments        map[string]*common.Comment
	commentsByUser  map[string][]*common.Comment
	users           map[string]*common.User
	moderators      map[string]*common.Moderator
	sensitiveWords  map[string]*common.SensitiveWord
	auditLogs       []*common.AuditLog
	pendingQueue    []*common.Comment
	mu              sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		comments:        make(map[string]*common.Comment),
		commentsByUser:  make(map[string][]*common.Comment),
		users:           make(map[string]*common.User),
		moderators:      make(map[string]*common.Moderator),
		sensitiveWords:  make(map[string]*common.SensitiveWord),
		auditLogs:       make([]*common.AuditLog, 0),
		pendingQueue:    make([]*common.Comment, 0),
	}
}

func (s *Store) GetOrCreateUser(userID string) *common.User {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		user = &common.User{
			ID:                  userID,
			ConsecutiveRejects:  0,
			LastRejectTime:      time.Time{},
			Comments24h:         0,
			Last24hResetTime:    time.Now(),
			UnderManualReview:   false,
			ManualReviewEndTime: time.Time{},
		}
		s.users[userID] = user
	}
	return user
}

func (s *Store) GetUser(userID string) *common.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.users[userID]
}

func (s *Store) UpdateUser(user *common.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[user.ID] = user
}

func (s *Store) AddComment(comment *common.Comment) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.comments[comment.ID] = comment
	s.commentsByUser[comment.UserID] = append(s.commentsByUser[comment.UserID], comment)

	if comment.Status == common.StatusPending || comment.Status == common.StatusManualReview {
		s.pendingQueue = append(s.pendingQueue, comment)
	}
}

func (s *Store) GetComment(commentID string) *common.Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.comments[commentID]
}

func (s *Store) UpdateComment(comment *common.Comment) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldStatus := s.comments[comment.ID].Status
	s.comments[comment.ID] = comment

	if (oldStatus == common.StatusPending || oldStatus == common.StatusManualReview) &&
		(comment.Status != common.StatusPending && comment.Status != common.StatusManualReview) {
		s.removeFromPendingQueue(comment.ID)
	}

	if (oldStatus != common.StatusPending && oldStatus != common.StatusManualReview) &&
		(comment.Status == common.StatusPending || comment.Status == common.StatusManualReview) {
		s.pendingQueue = append(s.pendingQueue, comment)
	}
}

func (s *Store) removeFromPendingQueue(commentID string) {
	for i, c := range s.pendingQueue {
		if c.ID == commentID {
			s.pendingQueue = append(s.pendingQueue[:i], s.pendingQueue[i+1:]...)
			break
		}
	}
}

func (s *Store) GetPendingComments(limit, offset int) []*common.Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if offset >= len(s.pendingQueue) {
		return []*common.Comment{}
	}

	end := offset + limit
	if end > len(s.pendingQueue) {
		end = len(s.pendingQueue)
	}

	return s.pendingQueue[offset:end]
}

func (s *Store) GetPendingCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.pendingQueue)
}

func (s *Store) AddModerator(moderator *common.Moderator) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.moderators[moderator.ID] = moderator
}

func (s *Store) GetModerator(moderatorID string) *common.Moderator {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.moderators[moderatorID]
}

func (s *Store) GetAllModerators() map[string]*common.Moderator {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*common.Moderator)
	for k, v := range s.moderators {
		result[k] = v
	}
	return result
}

func (s *Store) GetLeastBusyModerator() *common.Moderator {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.moderators) == 0 {
		return nil
	}

	today := common.GetTodayString()
	var leastBusy *common.Moderator
	minDaily := -1

	for _, m := range s.moderators {
		daily := m.DailyCount[today]
		if minDaily == -1 || daily < minDaily {
			minDaily = daily
			leastBusy = m
		}
	}

	return leastBusy
}

func (s *Store) AddSensitiveWord(word *common.SensitiveWord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sensitiveWords[word.Word] = word
}

func (s *Store) RemoveSensitiveWord(word string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sensitiveWords[word]; exists {
		delete(s.sensitiveWords, word)
		return true
	}
	return false
}

func (s *Store) GetSensitiveWords() []*common.SensitiveWord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*common.SensitiveWord, 0, len(s.sensitiveWords))
	for _, v := range s.sensitiveWords {
		result = append(result, v)
	}
	return result
}

func (s *Store) CheckSensitiveWords(content string) (bool, *common.SensitiveWord) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	highestLevel := common.LevelMild
	var matchedWord *common.SensitiveWord

	for word, sw := range s.sensitiveWords {
		if containsSubstring(content, word) {
			if isHigherSeverity(sw.Level, highestLevel) {
				highestLevel = sw.Level
				matchedWord = sw
			}
		}
	}

	return matchedWord != nil, matchedWord
}

func (s *Store) AddAuditLog(log *common.AuditLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.auditLogs = append(s.auditLogs, log)
}

func (s *Store) GetAuditLogs() []*common.AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*common.AuditLog, len(s.auditLogs))
	copy(result, s.auditLogs)
	return result
}

func (s *Store) GetCommentsByUser(userID string) []*common.Comment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.commentsByUser[userID]
}

func containsSubstring(s, substr string) bool {
	if len(substr) == 0 || len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func isHigherSeverity(level1, level2 common.SensitiveWordLevel) bool {
	severity := map[common.SensitiveWordLevel]int{
		common.LevelSevere: 3,
		common.LevelMedium: 2,
		common.LevelMild:   1,
	}
	return severity[level1] > severity[level2]
}

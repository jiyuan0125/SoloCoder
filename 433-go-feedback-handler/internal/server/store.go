package server

import (
	"go-feedback-handler/pkg/protocol"
	"sync"
	"time"
)

type Store struct {
	mu           sync.RWMutex
	feedbacks    map[string]*protocol.Feedback
	tags         map[string]*protocol.Tag
	userLimits   map[string]*protocol.UserLimitInfo
	notifications map[string]*protocol.Notification
	reports      map[string]*protocol.MonthlyReport
}

func NewStore() *Store {
	s := &Store{
		feedbacks:     make(map[string]*protocol.Feedback),
		tags:          make(map[string]*protocol.Tag),
		userLimits:    make(map[string]*protocol.UserLimitInfo),
		notifications: make(map[string]*protocol.Notification),
		reports:       make(map[string]*protocol.MonthlyReport),
	}
	s.initSystemTags()
	return s
}

func (s *Store) initSystemTags() {
	for _, st := range protocol.SystemTags {
		tag := &protocol.Tag{
			ID:        generateID(),
			Name:      st.Name,
			Color:     st.Color,
			IsSystem:  true,
			CreatedAt: time.Now(),
		}
		s.tags[tag.ID] = tag
	}
}

func (s *Store) CreateFeedback(fb *protocol.Feedback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.feedbacks[fb.ID] = fb
}

func (s *Store) GetFeedback(id string) *protocol.Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.feedbacks[id]
}

func (s *Store) UpdateFeedback(fb *protocol.Feedback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.feedbacks[fb.ID] = fb
}

func (s *Store) ListAllFeedbacks() []*protocol.Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*protocol.Feedback, 0, len(s.feedbacks))
	for _, fb := range s.feedbacks {
		result = append(result, fb)
	}
	return result
}

func (s *Store) GetFeedbacksByUser(userID string) []*protocol.Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*protocol.Feedback
	for _, fb := range s.feedbacks {
		if fb.UserID == userID {
			result = append(result, fb)
		}
	}
	return result
}

func (s *Store) GetFeedbacksByHandler(handlerID string) []*protocol.Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*protocol.Feedback
	for _, fb := range s.feedbacks {
		if fb.HandlerID == handlerID {
			result = append(result, fb)
		}
	}
	return result
}

func (s *Store) GetFeedbacksByStatus(status protocol.FeedbackStatus) []*protocol.Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*protocol.Feedback
	for _, fb := range s.feedbacks {
		if fb.Status == status {
			result = append(result, fb)
		}
	}
	return result
}

func (s *Store) GetFeedbacksByType(ftype protocol.FeedbackType) []*protocol.Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*protocol.Feedback
	for _, fb := range s.feedbacks {
		if fb.Type == ftype {
			result = append(result, fb)
		}
	}
	return result
}

func (s *Store) GetFeedbacksByTag(tagID string) []*protocol.Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*protocol.Feedback
	for _, fb := range s.feedbacks {
		for _, id := range fb.TagIDs {
			if id == tagID {
				result = append(result, fb)
				break
			}
		}
	}
	return result
}

func (s *Store) GetFeedbacksInReview() []*protocol.Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*protocol.Feedback
	for _, fb := range s.feedbacks {
		if fb.InReviewQueue {
			result = append(result, fb)
		}
	}
	return result
}

func (s *Store) CreateTag(tag *protocol.Tag) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tags[tag.ID] = tag
}

func (s *Store) GetTag(id string) *protocol.Tag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tags[id]
}

func (s *Store) GetTagByName(name string) *protocol.Tag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, tag := range s.tags {
		if tag.Name == name {
			return tag
		}
	}
	return nil
}

func (s *Store) ListAllTags() []*protocol.Tag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*protocol.Tag, 0, len(s.tags))
	for _, tag := range s.tags {
		result = append(result, tag)
	}
	return result
}

func (s *Store) DeleteTag(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tags, id)
}

func (s *Store) GetOrCreateUserLimit(userID string) *protocol.UserLimitInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	limit, exists := s.userLimits[userID]
	if !exists {
		limit = &protocol.UserLimitInfo{
			UserID:            userID,
			InvalidCloseCount: 0,
			IsUnderReview:     false,
		}
		s.userLimits[userID] = limit
	}
	return limit
}

func (s *Store) UpdateUserLimit(limit *protocol.UserLimitInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userLimits[limit.UserID] = limit
}

func (s *Store) ListAllUserLimits() []*protocol.UserLimitInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*protocol.UserLimitInfo, 0, len(s.userLimits))
	for _, limit := range s.userLimits {
		result = append(result, limit)
	}
	return result
}

func (s *Store) CreateNotification(notif *protocol.Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifications[notif.ID] = notif
}

func (s *Store) GetNotification(id string) *protocol.Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.notifications[id]
}

func (s *Store) ListNotificationsByUser(userID string) []*protocol.Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*protocol.Notification
	for _, notif := range s.notifications {
		if notif.UserID == userID {
			result = append(result, notif)
		}
	}
	return result
}

func (s *Store) ListAllNotifications() []*protocol.Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*protocol.Notification, 0, len(s.notifications))
	for _, notif := range s.notifications {
		result = append(result, notif)
	}
	return result
}

func (s *Store) MarkNotificationRead(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if notif, exists := s.notifications[id]; exists {
		notif.Read = true
	}
}

func (s *Store) CreateReport(report *protocol.MonthlyReport) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reports[report.Month] = report
}

func (s *Store) GetReport(month string) *protocol.MonthlyReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.reports[month]
}

func (s *Store) ListAllReports() []*protocol.MonthlyReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*protocol.MonthlyReport, 0, len(s.reports))
	for _, report := range s.reports {
		result = append(result, report)
	}
	return result
}

func (s *Store) AddComment(fbID string, comment *protocol.Comment) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	fb, exists := s.feedbacks[fbID]
	if !exists {
		return false
	}
	fb.Comments = append(fb.Comments, comment)
	fb.LastUpdatedAt = time.Now()
	return true
}

func (s *Store) AddTagToFeedback(fbID, tagID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	fb, exists := s.feedbacks[fbID]
	if !exists {
		return false
	}
	for _, id := range fb.TagIDs {
		if id == tagID {
			return true
		}
	}
	fb.TagIDs = append(fb.TagIDs, tagID)
	return true
}

func (s *Store) RemoveTagFromFeedback(fbID, tagID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	fb, exists := s.feedbacks[fbID]
	if !exists {
		return false
	}
	newTagIDs := make([]string, 0, len(fb.TagIDs))
	for _, id := range fb.TagIDs {
		if id != tagID {
			newTagIDs = append(newTagIDs, id)
		}
	}
	fb.TagIDs = newTagIDs
	return true
}

func (s *Store) GetFeedbacksForMergeCheck(userID string, beforeTime time.Time) []*protocol.Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*protocol.Feedback
	for _, fb := range s.feedbacks {
		if fb.UserID == userID && fb.CreatedAt.Before(beforeTime) && fb.MergedInto == "" {
			result = append(result, fb)
		}
	}
	return result
}

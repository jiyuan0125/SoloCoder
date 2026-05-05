package main

import (
	"notification-center/common"
	"sync"
	"time"
)

type Store struct {
	notifications      map[string]*common.Notification
	notificationsByUser map[string][]*common.Notification
	archivedNotifications map[string]*common.Notification
	templates          map[string]*common.NotificationTemplate
	users              map[string]*common.UserInfo
	failedLogs         map[string]*common.FailedLog
	dedupCache         map[string]time.Time
	mu                 sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		notifications:         make(map[string]*common.Notification),
		notificationsByUser:   make(map[string][]*common.Notification),
		archivedNotifications: make(map[string]*common.Notification),
		templates:             make(map[string]*common.NotificationTemplate),
		users:                 make(map[string]*common.UserInfo),
		failedLogs:            make(map[string]*common.FailedLog),
		dedupCache:            make(map[string]time.Time),
	}
}

func (s *Store) GetOrCreateUser(userID string) *common.UserInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	user, exists := s.users[userID]
	if !exists {
		user = &common.UserInfo{
			UserID:       userID,
			Status:       common.UserStatusActive,
			LastActiveAt: time.Now(),
		}
		s.users[userID] = user
	}
	return user
}

func (s *Store) UpdateUserActivity(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	user, exists := s.users[userID]
	if !exists {
		user = &common.UserInfo{
			UserID:       userID,
			Status:       common.UserStatusActive,
			LastActiveAt: time.Now(),
		}
		s.users[userID] = user
		return
	}
	user.LastActiveAt = time.Now()
	user.Status = common.UserStatusActive
}

func (s *Store) CheckDuplicate(receiver, content string, notificationType common.NotificationType) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if notificationType == common.TypeSecurity {
		return false
	}
	
	key := receiver + ":" + content
	lastSent, exists := s.dedupCache[key]
	if !exists {
		return false
	}
	
	return time.Since(lastSent) < time.Duration(common.DedupWindowMinutes)*time.Minute
}

func (s *Store) UpdateDedupCache(receiver, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	key := receiver + ":" + content
	s.dedupCache[key] = time.Now()
}

func (s *Store) SaveNotification(notification *common.Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.notifications[notification.ID] = notification
	
	if s.notificationsByUser[notification.Receiver] == nil {
		s.notificationsByUser[notification.Receiver] = []*common.Notification{}
	}
	s.notificationsByUser[notification.Receiver] = append(
		s.notificationsByUser[notification.Receiver],
		notification,
	)
}

func (s *Store) GetNotification(id string) (*common.Notification, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	notif, exists := s.notifications[id]
	if exists {
		return notif, true
	}
	notif, exists = s.archivedNotifications[id]
	return notif, exists
}

func (s *Store) ListNotifications(req common.ListRequest) ([]common.Notification, int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	userNotifications := s.notificationsByUser[req.Receiver]
	if userNotifications == nil {
		return []common.Notification{}, 0, 0
	}
	
	totalUnreadCount := 0
	for _, notif := range userNotifications {
		if notif.Status == common.StatusUnread && !notif.IsArchived {
			totalUnreadCount++
		}
	}
	
	var filtered []*common.Notification
	
	for _, notif := range userNotifications {
		if notif.IsArchived && !req.IncludeArchived {
			continue
		}
		
		if req.Type != nil && notif.Type != *req.Type {
			continue
		}
		
		if req.StartTime != nil && notif.CreatedAt.Before(*req.StartTime) {
			continue
		}
		
		if req.EndTime != nil && notif.CreatedAt.After(*req.EndTime) {
			continue
		}
		
		if req.Status != nil && notif.Status != *req.Status {
			continue
		}
		
		filtered = append(filtered, notif)
	}
	
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if shouldSwap(filtered[i], filtered[j]) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}
	
	total := len(filtered)
	
	if req.PageSize > 0 && req.Page > 0 {
		start := (req.Page - 1) * req.PageSize
		end := start + req.PageSize
		if start >= total {
			filtered = []*common.Notification{}
		} else if end > total {
			filtered = filtered[start:]
		} else {
			filtered = filtered[start:end]
		}
	}
	
	result := make([]common.Notification, len(filtered))
	for i, notif := range filtered {
		result[i] = *notif
	}
	
	return result, total, totalUnreadCount
}

func shouldSwap(a, b *common.Notification) bool {
	if a.Priority == common.PriorityHigh && b.Priority != common.PriorityHigh {
		return false
	}
	if b.Priority == common.PriorityHigh && a.Priority != common.PriorityHigh {
		return true
	}
	return a.CreatedAt.Before(b.CreatedAt)
}

func (s *Store) MarkAsRead(receiver string, notificationIDs []string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	count := 0
	now := time.Now()
	
	for _, id := range notificationIDs {
		notif, exists := s.notifications[id]
		if !exists {
			continue
		}
		if notif.Receiver != receiver {
			continue
		}
		if notif.Status == common.StatusRead {
			continue
		}
		
		notif.Status = common.StatusRead
		notif.ReadAt = &now
		count++
	}
	
	return count
}

func (s *Store) GetUnreadCount(receiver string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	userNotifications := s.notificationsByUser[receiver]
	if userNotifications == nil {
		return 0
	}
	
	count := 0
	for _, notif := range userNotifications {
		if notif.Status == common.StatusUnread && !notif.IsArchived {
			count++
		}
	}
	return count
}

func (s *Store) ArchiveOldNotifications() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	cutoff := time.Now().AddDate(0, 0, -common.ArchiveDays)
	count := 0
	
	for id, notif := range s.notifications {
		if notif.CreatedAt.Before(cutoff) && !notif.IsArchived {
			notif.IsArchived = true
			s.archivedNotifications[id] = notif
			count++
		}
	}
	
	return count
}

func (s *Store) CheckInactiveUsers() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	cutoff := time.Now().AddDate(0, 0, -common.InactiveDays)
	
	for _, user := range s.users {
		if user.LastActiveAt.Before(cutoff) {
			user.Status = common.UserStatusInactive
		}
	}
}

func (s *Store) GetInactiveUsers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var inactive []string
	for id, user := range s.users {
		if user.Status == common.UserStatusInactive {
			inactive = append(inactive, id)
		}
	}
	return inactive
}

func (s *Store) GetAllUsers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var users []string
	for id := range s.users {
		users = append(users, id)
	}
	return users
}

func (s *Store) GetUserTotalUnreadCount(receiver string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	userNotifications := s.notificationsByUser[receiver]
	if userNotifications == nil {
		return 0
	}
	
	count := 0
	for _, notif := range userNotifications {
		if notif.Status == common.StatusUnread && !notif.IsArchived {
			count++
		}
	}
	return count
}

func (s *Store) GetHighPriorityUnreadNotifications(receiver string) []*common.Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	userNotifications := s.notificationsByUser[receiver]
	if userNotifications == nil {
		return []*common.Notification{}
	}
	
	var result []*common.Notification
	for _, notif := range userNotifications {
		if notif.Status == common.StatusUnread && 
		   notif.Priority == common.PriorityHigh && 
		   !notif.IsArchived {
			result = append(result, notif)
		}
	}
	return result
}

func (s *Store) SaveTemplate(template *common.NotificationTemplate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.templates[template.ID] = template
}

func (s *Store) GetTemplate(id string) (*common.NotificationTemplate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	template, exists := s.templates[id]
	return template, exists
}

func (s *Store) ListTemplates() []common.NotificationTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	templates := make([]common.NotificationTemplate, 0, len(s.templates))
	for _, t := range s.templates {
		templates = append(templates, *t)
	}
	return templates
}

func (s *Store) DeleteTemplate(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	_, exists := s.templates[id]
	if exists {
		delete(s.templates, id)
	}
	return exists
}

func (s *Store) SaveFailedLog(log *common.FailedLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.failedLogs[log.ID] = log
}

func (s *Store) ListFailedLogs() []common.FailedLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	logs := make([]common.FailedLog, 0, len(s.failedLogs))
	for _, l := range s.failedLogs {
		logs = append(logs, *l)
	}
	return logs
}

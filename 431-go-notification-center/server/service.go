package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"notification-center/common"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Service struct {
	store *Store
	mu    sync.Mutex
}

func NewService(store *Store) *Service {
	return &Service{
		store: store,
	}
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *Service) SendNotification(req common.SendRequest) (*common.SendResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if req.Receiver == "" {
		return nil, common.ErrInvalidRequest("receiver is required")
	}
	if req.Sender == "" {
		return nil, common.ErrInvalidRequest("sender is required")
	}
	if req.Title == "" {
		return nil, common.ErrInvalidRequest("title is required")
	}
	
	title := req.Title
	content := req.Content
	
	if req.TemplateID != "" {
		template, exists := s.store.GetTemplate(req.TemplateID)
		if !exists {
			return nil, common.ErrNotFound(fmt.Sprintf("template %s not found", req.TemplateID))
		}
		
		var err error
		title, content, err = s.renderTemplate(template, req.Variables)
		if err != nil {
			return nil, err
		}
	}
	
	if content == "" {
		return nil, common.ErrInvalidRequest("content is required")
	}
	
	isEmergency := req.Type == common.TypeSecurity
	
	if !isEmergency {
		if s.store.CheckDuplicate(req.Receiver, content, req.Type) {
			return nil, common.ErrDuplicate("duplicate notification within 5 minutes")
		}
	}
	
	user := s.store.GetOrCreateUser(req.Receiver)
	
	priority := req.Priority
	if priority == "" {
		priority = common.PriorityNormal
	}
	
	now := time.Now()
	notification := &common.Notification{
		ID:          generateID(),
		Type:        req.Type,
		Priority:    priority,
		Sender:      req.Sender,
		Receiver:    req.Receiver,
		Title:       title,
		Content:     content,
		Status:      common.StatusUnread,
		CreatedAt:   now,
		IsArchived:  false,
		IsEmergency: isEmergency,
		RetryCount:  0,
		SendFailed:  false,
	}
	
	if user.Status == common.UserStatusInactive && !isEmergency {
		err := s.addToDailySummary(user, notification)
		if err != nil {
			return s.handleSendFailure(notification, err)
		}
	} else {
		err := s.deliverNotification(notification)
		if err != nil {
			return s.handleSendFailure(notification, err)
		}
	}
	
	s.store.SaveNotification(notification)
	if !isEmergency {
		s.store.UpdateDedupCache(req.Receiver, content)
	}
	
	return &common.SendResponse{
		Success:        true,
		NotificationID: notification.ID,
	}, nil
}

func (s *Service) renderTemplate(template *common.NotificationTemplate, vars map[string]string) (string, string, error) {
	title := template.Title
	content := template.Content
	
	for _, varName := range template.Variables {
		placeholder := fmt.Sprintf("{{%s}}", varName)
		value, exists := vars[varName]
		if !exists {
			return "", "", common.ErrInvalidTemplate(fmt.Sprintf("missing variable: %s", varName))
		}
		title = strings.ReplaceAll(title, placeholder, value)
		content = strings.ReplaceAll(content, placeholder, value)
	}
	
	re := regexp.MustCompile(`\{\{[^}]+\}\}`)
	remainingVars := re.FindAllString(content, -1)
	if len(remainingVars) > 0 {
		return "", "", common.ErrInvalidTemplate(fmt.Sprintf("unexpected variables in content: %v", remainingVars))
	}
	
	return title, content, nil
}

func (s *Service) deliverNotification(notification *common.Notification) error {
	fmt.Printf("[DELIVER] Sending notification to %s: %s\n", notification.Receiver, notification.Title)
	return nil
}

func (s *Service) addToDailySummary(user *common.UserInfo, notification *common.Notification) error {
	fmt.Printf("[SUMMARY] Adding notification for inactive user %s to daily summary: %s\n", user.UserID, notification.Title)
	return nil
}

func (s *Service) handleSendFailure(notification *common.Notification, originalErr error) (*common.SendResponse, error) {
	var errorMessages []string
	lastErr := originalErr
	
	for notification.RetryCount < common.MaxRetryCount {
		notification.RetryCount++
		errorMessages = append(errorMessages, lastErr.Error())
		
		fmt.Printf("[RETRY] Retry %d/%d for notification %s\n", 
			notification.RetryCount, common.MaxRetryCount, notification.ID)
		
		err := s.deliverNotification(notification)
		if err == nil {
			notification.SendFailed = false
			s.store.SaveNotification(notification)
			return &common.SendResponse{
				Success:        true,
				NotificationID: notification.ID,
			}, nil
		}
		lastErr = err
	}
	
	notification.SendFailed = true
	s.store.SaveNotification(notification)
	
	failedLog := &common.FailedLog{
		ID:              generateID(),
		NotificationID:  notification.ID,
		Receiver:        notification.Receiver,
		Title:           notification.Title,
		Content:         notification.Content,
		RetryCount:      notification.RetryCount,
		ErrorMessages:   append(errorMessages, lastErr.Error()),
		FailedAt:        time.Now(),
	}
	s.store.SaveFailedLog(failedLog)
	
	return nil, common.ErrSendFailed(fmt.Sprintf("all retries failed: %s", lastErr.Error()))
}

func (s *Service) ListNotifications(req common.ListRequest) (*common.ListResponse, error) {
	if req.Receiver == "" {
		return nil, common.ErrInvalidRequest("receiver is required")
	}
	
	s.store.UpdateUserActivity(req.Receiver)
	
	notifications, total, unreadCount := s.store.ListNotifications(req)
	
	return &common.ListResponse{
		Notifications: notifications,
		Total:         total,
		UnreadCount:   unreadCount,
	}, nil
}

func (s *Service) MarkAsRead(req common.MarkReadRequest) (*common.MarkReadResponse, error) {
	if req.Receiver == "" {
		return nil, common.ErrInvalidRequest("receiver is required")
	}
	if len(req.NotificationIDs) == 0 {
		return nil, common.ErrInvalidRequest("notification_ids is required")
	}
	if len(req.NotificationIDs) > common.MaxBatchReadCount {
		return nil, common.ErrExceedLimit(
			fmt.Sprintf("max batch read count is %d, got %d", 
				common.MaxBatchReadCount, len(req.NotificationIDs)),
		)
	}
	
	s.store.UpdateUserActivity(req.Receiver)
	
	count := s.store.MarkAsRead(req.Receiver, req.NotificationIDs)
	
	return &common.MarkReadResponse{
		Success:     true,
		MarkedCount: count,
	}, nil
}

func (s *Service) GetUnreadCount(receiver string) (*common.UnreadCountResponse, error) {
	if receiver == "" {
		return nil, common.ErrInvalidRequest("receiver is required")
	}
	
	count := s.store.GetUnreadCount(receiver)
	
	return &common.UnreadCountResponse{
		Receiver:    receiver,
		UnreadCount: count,
	}, nil
}

func (s *Service) CreateTemplate(req common.TemplateCreateRequest) (*common.NotificationTemplate, error) {
	if req.Name == "" {
		return nil, common.ErrInvalidRequest("name is required")
	}
	if req.Title == "" {
		return nil, common.ErrInvalidRequest("title is required")
	}
	if req.Content == "" {
		return nil, common.ErrInvalidRequest("content is required")
	}
	
	now := time.Now()
	template := &common.NotificationTemplate{
		ID:        generateID(),
		Name:      req.Name,
		Type:      req.Type,
		Title:     req.Title,
		Content:   req.Content,
		Variables: req.Variables,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	s.store.SaveTemplate(template)
	return template, nil
}

func (s *Service) UpdateTemplate(req common.TemplateUpdateRequest) (*common.NotificationTemplate, error) {
	if req.ID == "" {
		return nil, common.ErrInvalidRequest("id is required")
	}
	
	template, exists := s.store.GetTemplate(req.ID)
	if !exists {
		return nil, common.ErrNotFound(fmt.Sprintf("template %s not found", req.ID))
	}
	
	if req.Name != nil {
		template.Name = *req.Name
	}
	if req.Type != nil {
		template.Type = *req.Type
	}
	if req.Title != nil {
		template.Title = *req.Title
	}
	if req.Content != nil {
		template.Content = *req.Content
	}
	if req.Variables != nil {
		template.Variables = *req.Variables
	}
	
	template.UpdatedAt = time.Now()
	s.store.SaveTemplate(template)
	
	return template, nil
}

func (s *Service) ListTemplates() []common.NotificationTemplate {
	return s.store.ListTemplates()
}

func (s *Service) DeleteTemplate(id string) error {
	if id == "" {
		return common.ErrInvalidRequest("id is required")
	}
	
	exists := s.store.DeleteTemplate(id)
	if !exists {
		return common.ErrNotFound(fmt.Sprintf("template %s not found", id))
	}
	
	return nil
}

func (s *Service) RecordActivity(userID string) {
	s.store.UpdateUserActivity(userID)
}

func (s *Service) RunBackgroundTasks() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		s.store.CheckInactiveUsers()
		s.store.ArchiveOldNotifications()
		s.checkHighPriorityReminders()
	}
}

func (s *Service) checkHighPriorityReminders() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	
	allUsers := s.store.GetAllUsers()
	
	for _, userID := range allUsers {
		user := s.store.GetOrCreateUser(userID)
		
		if user.Status == common.UserStatusInactive {
			if user.LastSummaryAt != nil {
				lastSummary := *user.LastSummaryAt
				if now.Sub(lastSummary).Hours() < 24 {
					continue
				}
			}
			
			req := common.ListRequest{
				Receiver:        userID,
				Status:          func() *common.ReadStatus { s := common.StatusUnread; return &s }(),
				IncludeArchived: false,
			}
			notifications, _, _ := s.store.ListNotifications(req)
			
			if len(notifications) > 0 {
				fmt.Printf("[DAILY-SUMMARY] Sending daily summary to inactive user %s: %d unread notifications\n", 
					userID, len(notifications))
				user.LastSummaryAt = &now
			}
		} else {
			highPriorityNotifs := s.store.GetHighPriorityUnreadNotifications(userID)
			
			for _, notif := range highPriorityNotifs {
				if notif.LastReminderAt != nil {
					lastReminder := *notif.LastReminderAt
					if now.Sub(lastReminder).Hours() < 24 {
						continue
					}
				}
				
				fmt.Printf("[HIGH-PRIORITY-REMINDER] Reminding active user %s about high priority notification: %s\n", 
					userID, notif.Title)
				
				notif.LastReminderAt = &now
				s.store.SaveNotification(notif)
			}
		}
	}
}

func (s *Service) ListFailedLogs() []common.FailedLog {
	return s.store.ListFailedLogs()
}

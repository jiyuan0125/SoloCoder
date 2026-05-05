package server

import (
	"crypto/rand"
	"encoding/hex"
	"go-live-chat/common"
	"time"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{
		store: store,
	}
}

func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *Service) CreateSession(userID, username string) (*common.CreateSessionResponse, error) {
	if s.store.IsBlacklisted(userID) {
		return nil, &ServiceError{Message: "user is blacklisted"}
	}
	sessionID := generateID()
	agent, hasAgent := s.store.GetLeastBusyAgent()
	if !hasAgent {
		now := time.Now()
		session := &common.Session{
			ID:            sessionID,
			UserID:        userID,
			Status:        common.SessionStatusQueued,
			CreatedAt:     now,
			LastMessageAt: now,
			Messages:      make([]common.Message, 0),
		}
		queuePosition := s.store.AddToQueue(session)
		return &common.CreateSessionResponse{
			SessionID:     sessionID,
			Status:        common.SessionStatusQueued,
			QueuePosition: queuePosition,
		}, nil
	}
	s.store.CreateSession(sessionID, userID, agent.ID)
	return &common.CreateSessionResponse{
		SessionID: sessionID,
		Status:    common.SessionStatusActive,
		AgentID:   agent.ID,
	}, nil
}

func (s *Service) SendMessage(request common.SendMessageRequest) (*common.SendMessageResponse, error) {
	session, exists := s.store.GetSession(request.SessionID)
	if !exists {
		return nil, &ServiceError{Message: "session not found"}
	}
	if session.Status == common.SessionStatusClosed {
		return nil, &ServiceError{Message: "session is already closed"}
	}
	if session.Status == common.SessionStatusQueued {
		return nil, &ServiceError{Message: "session is still in queue"}
	}
	messageID := generateID()
	message := common.Message{
		ID:         messageID,
		SessionID:  request.SessionID,
		SenderID:   request.SenderID,
		SenderType: request.SenderType,
		Content:    request.Content,
		Timestamp:  time.Now(),
	}
	s.store.AddMessage(request.SessionID, message)
	return &common.SendMessageResponse{
		MessageID: messageID,
		Timestamp: message.Timestamp,
	}, nil
}

func (s *Service) CloseSession(request common.CloseSessionRequest) (*common.CloseSessionResponse, error) {
	session, exists := s.store.GetSession(request.SessionID)
	if !exists {
		return nil, &ServiceError{Message: "session not found"}
	}
	if session.Status == common.SessionStatusClosed {
		return &common.CloseSessionResponse{
			SessionID: request.SessionID,
			Closed:    true,
		}, nil
	}
	waitTime := time.Since(session.CreatedAt).Seconds()
	sessionTime := time.Since(session.CreatedAt).Seconds()
	s.store.UpdateStatistics(waitTime, sessionTime)
	success := s.store.CloseSession(request.SessionID)
	if !success {
		return nil, &ServiceError{Message: "failed to close session"}
	}
	s.processQueue()
	return &common.CloseSessionResponse{
		SessionID: request.SessionID,
		Closed:    true,
	}, nil
}

func (s *Service) SubmitSatisfaction(request common.SubmitSatisfactionRequest) (*common.SubmitSatisfactionResponse, error) {
	session, exists := s.store.GetSession(request.SessionID)
	if !exists {
		return nil, &ServiceError{Message: "session not found"}
	}
	if session.UserID != request.UserID {
		return nil, &ServiceError{Message: "unauthorized to submit satisfaction for this session"}
	}
	if session.Satisfaction != nil {
		return nil, &ServiceError{Message: "satisfaction already submitted and cannot be modified"}
	}
	if request.Rating < 1 || request.Rating > 5 {
		return nil, &ServiceError{Message: "rating must be between 1 and 5"}
	}
	satisfaction := common.Satisfaction{
		Rating:    request.Rating,
		Comment:   request.Comment,
		CreatedAt: time.Now(),
	}
	success := s.store.SubmitSatisfaction(request.SessionID, satisfaction)
	if !success {
		return nil, &ServiceError{Message: "failed to submit satisfaction"}
	}
	return &common.SubmitSatisfactionResponse{
		Success: true,
	}, nil
}

func (s *Service) TransferSession(request common.TransferSessionRequest) (*common.TransferSessionResponse, error) {
	session, exists := s.store.GetSession(request.SessionID)
	if !exists {
		return nil, &ServiceError{Message: "session not found"}
	}
	if session.Status != common.SessionStatusActive {
		return nil, &ServiceError{Message: "only active sessions can be transferred"}
	}
	if session.AgentID != request.FromAgentID {
		return nil, &ServiceError{Message: "only the current agent can transfer this session"}
	}
	toAgent, exists := s.store.GetAgent(request.ToAgentID)
	if !exists {
		return nil, &ServiceError{Message: "target agent not found"}
	}
	if toAgent.Status != common.AgentStatusOnline {
		return nil, &ServiceError{Message: "target agent is not online"}
	}
	if toAgent.CurrentSessions >= common.MaxAgentSessions {
		return nil, &ServiceError{Message: "target agent has reached maximum session limit"}
	}
	success := s.store.TransferSession(request.SessionID, request.FromAgentID, request.ToAgentID)
	if !success {
		return nil, &ServiceError{Message: "failed to transfer session"}
	}
	return &common.TransferSessionResponse{
		Success:    true,
		NewAgentID: request.ToAgentID,
	}, nil
}

func (s *Service) CreateQuickReply(request common.CreateQuickReplyRequest) (*common.CreateQuickReplyResponse, error) {
	templateID := generateID()
	template := &common.QuickReplyTemplate{
		ID:        templateID,
		AgentID:   request.AgentID,
		Title:     request.Title,
		Content:   request.Content,
		CreatedAt: time.Now(),
	}
	s.store.AddQuickReply(request.AgentID, template)
	return &common.CreateQuickReplyResponse{
		TemplateID: templateID,
	}, nil
}

func (s *Service) GetQuickReplies(agentID string) (*common.GetQuickRepliesResponse, error) {
	templates := s.store.GetQuickReplies(agentID)
	templateList := make([]common.QuickReplyTemplate, len(templates))
	for i, t := range templates {
		templateList[i] = *t
	}
	return &common.GetQuickRepliesResponse{
		Templates: templateList,
	}, nil
}

func (s *Service) AddToBlacklist(request common.AddToBlacklistRequest) (*common.AddToBlacklistResponse, error) {
	entry := &common.BlacklistEntry{
		UserID:    request.UserID,
		Reason:    request.Reason,
		CreatedAt: time.Now(),
		ExpiresAt: request.ExpiresAt,
	}
	s.store.AddToBlacklist(entry)
	return &common.AddToBlacklistResponse{
		Success: true,
	}, nil
}

func (s *Service) RemoveFromBlacklist(userID string) (*common.RemoveFromBlacklistResponse, error) {
	s.store.RemoveFromBlacklist(userID)
	return &common.RemoveFromBlacklistResponse{
		Success: true,
	}, nil
}

func (s *Service) AgentLogin(request common.AgentLoginRequest) (*common.AgentLoginResponse, error) {
	_, exists := s.store.GetAgent(request.AgentID)
	if !exists {
		s.store.CreateAgent(request.AgentID, request.Username)
	} else {
		s.store.UpdateAgentStatus(request.AgentID, common.AgentStatusOnline)
	}
	s.processQueue()
	return &common.AgentLoginResponse{
		Success: true,
	}, nil
}

func (s *Service) AgentLogout(agentID string) (*common.AgentLogoutResponse, error) {
	agentSessions := s.store.GetAgentSessions(agentID)
	for _, session := range agentSessions {
		if session.Status == common.SessionStatusActive {
			s.transferToAvailableAgent(session, agentID)
		}
	}
	s.store.UpdateAgentStatus(agentID, common.AgentStatusOffline)
	return &common.AgentLogoutResponse{
		Success: true,
	}, nil
}

func (s *Service) transferToAvailableAgent(session *common.Session, fromAgentID string) {
	agent, hasAgent := s.store.GetLeastBusyAgent()
	if !hasAgent {
		return
	}
	s.store.TransferSession(session.ID, fromAgentID, agent.ID)
}

func (s *Service) GetAgentSessions(agentID string) (*common.GetAgentSessionsResponse, error) {
	sessions := s.store.GetAgentSessions(agentID)
	sessionList := make([]common.Session, 0)
	for _, sess := range sessions {
		if sess.Status == common.SessionStatusActive {
			sessionList = append(sessionList, *sess)
		}
	}
	return &common.GetAgentSessionsResponse{
		Sessions: sessionList,
	}, nil
}

func (s *Service) GetSessionHistory(sessionID string) (*common.GetSessionHistoryResponse, error) {
	session, exists := s.store.GetSession(sessionID)
	if !exists {
		return nil, &ServiceError{Message: "session not found"}
	}
	return &common.GetSessionHistoryResponse{
		Session:  *session,
		Messages: session.Messages,
	}, nil
}

func (s *Service) GetStatistics() (*common.GetStatisticsResponse, error) {
	stats := s.store.GetStatistics()
	return &common.GetStatisticsResponse{
		Statistics: stats,
	}, nil
}

func (s *Service) AddSchedule(request common.AddScheduleRequest) (*common.AddScheduleResponse, error) {
	_, exists := s.store.GetAgent(request.AgentID)
	if !exists {
		return nil, &ServiceError{Message: "agent not found"}
	}
	scheduleID := generateID()
	schedule := &common.Schedule{
		ID:        scheduleID,
		AgentID:   request.AgentID,
		StartTime: request.StartTime,
		EndTime:   request.EndTime,
	}
	s.store.AddSchedule(request.AgentID, schedule)
	return &common.AddScheduleResponse{
		ScheduleID: scheduleID,
		Success:    true,
	}, nil
}

func (s *Service) GetSchedules(agentID string) (*common.GetSchedulesResponse, error) {
	schedules := s.store.GetSchedules(agentID)
	scheduleList := make([]common.Schedule, len(schedules))
	for i, sched := range schedules {
		scheduleList[i] = *sched
	}
	return &common.GetSchedulesResponse{
		Schedules: scheduleList,
	}, nil
}

func (s *Service) processQueue() {
	queue := s.store.GetQueue()
	for _, session := range queue {
		agent, hasAgent := s.store.GetLeastBusyAgent()
		if !hasAgent {
			break
		}
		success := s.store.AssignQueuedSessionToAgent(session.ID, agent.ID)
		if !success {
			break
		}
	}
}

func (s *Service) CheckQueueReminders() {
	queue := s.store.GetQueue()
	now := time.Now()
	for _, session := range queue {
		waitTime := now.Sub(session.CreatedAt)
		if waitTime.Minutes() >= float64(common.QueueReminderMinutes) {
			if s.store.HasQueueReminderBeenSent(session.ID) {
				continue
			}
			messageID := generateID()
			message := common.Message{
				ID:         messageID,
				SessionID:  session.ID,
				SenderID:   "system",
				SenderType: common.SenderTypeSystem,
				Content:    "感谢您的耐心等待，我们正在为您安排客服，请稍候...",
				Timestamp:  now,
			}
			s.store.AddMessage(session.ID, message)
			s.store.MarkQueueReminderSent(session.ID)
		}
	}
}

func (s *Service) CheckSessionTimeouts() {
	allSessions := s.store.GetAllSessions()
	now := time.Now()
	for _, session := range allSessions {
		if session.Status == common.SessionStatusActive {
			idleTime := now.Sub(session.LastMessageAt)
			if idleTime.Minutes() >= float64(common.SessionTimeoutMinutes) {
				messageID := generateID()
				message := common.Message{
					ID:         messageID,
					SessionID:  session.ID,
					SenderID:   "system",
					SenderType: common.SenderTypeSystem,
					Content:    "由于您长时间未回复，会话已自动结束。如有其他问题，请重新发起会话。",
					Timestamp:  now,
				}
				s.store.AddMessage(session.ID, message)
				s.CloseSession(common.CloseSessionRequest{
					SessionID: session.ID,
				})
			}
		}
	}
}

func (s *Service) CleanOldChats() {
	s.store.CleanOldChats()
}

type ServiceError struct {
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

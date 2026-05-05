package server

import (
	"go-live-chat/common"
	"sync"
	"time"
)

type Store struct {
	mu               sync.RWMutex
	agents           map[string]*common.Agent
	sessions         map[string]*common.Session
	userSessions     map[string][]*common.Session
	agentSessions    map[string][]*common.Session
	queue            []*common.Session
	quickReplies     map[string][]*common.QuickReplyTemplate
	blacklist        map[string]*common.BlacklistEntry
	schedules        map[string][]*common.Schedule
	queueReminded    map[string]bool
	statistics       common.Statistics
	totalSessions    int
	totalWaitTime    float64
	totalSessionTime float64
	totalSatisfaction float64
	satisfactionCount int
}

func NewStore() *Store {
	return &Store{
		agents:        make(map[string]*common.Agent),
		sessions:      make(map[string]*common.Session),
		userSessions:  make(map[string][]*common.Session),
		agentSessions: make(map[string][]*common.Session),
		queue:         make([]*common.Session, 0),
		quickReplies:  make(map[string][]*common.QuickReplyTemplate),
		blacklist:     make(map[string]*common.BlacklistEntry),
		schedules:     make(map[string][]*common.Schedule),
		queueReminded: make(map[string]bool),
	}
}

func (s *Store) GetAgent(agentID string) (*common.Agent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	agent, exists := s.agents[agentID]
	return agent, exists
}

func (s *Store) CreateAgent(agentID, username string) *common.Agent {
	s.mu.Lock()
	defer s.mu.Unlock()
	agent := &common.Agent{
		ID:              agentID,
		Username:        username,
		Status:          common.AgentStatusOnline,
		CurrentSessions: 0,
	}
	s.agents[agentID] = agent
	s.agentSessions[agentID] = make([]*common.Session, 0)
	return agent
}

func (s *Store) UpdateAgentStatus(agentID string, status common.AgentStatus) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, exists := s.agents[agentID]
	if !exists {
		return false
	}
	agent.Status = status
	return true
}

func (s *Store) GetOnlineAgents() []*common.Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var onlineAgents []*common.Agent
	for _, agent := range s.agents {
		if agent.Status == common.AgentStatusOnline {
			onlineAgents = append(onlineAgents, agent)
		}
	}
	return onlineAgents
}

func (s *Store) GetLeastBusyAgent() (*common.Agent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var leastBusy *common.Agent
	for _, agent := range s.agents {
		if agent.Status == common.AgentStatusOnline && agent.CurrentSessions < common.MaxAgentSessions {
			if !s.isAgentOnShiftInternal(agent.ID) {
				continue
			}
			if leastBusy == nil || agent.CurrentSessions < leastBusy.CurrentSessions {
				leastBusy = agent
			}
		}
	}
	return leastBusy, leastBusy != nil
}

func (s *Store) IncrementAgentSessions(agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, exists := s.agents[agentID]
	if exists {
		agent.CurrentSessions++
		if agent.CurrentSessions >= common.MaxAgentSessions {
			agent.Status = common.AgentStatusBusy
		}
	}
}

func (s *Store) DecrementAgentSessions(agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, exists := s.agents[agentID]
	if exists {
		agent.CurrentSessions--
		if agent.CurrentSessions < common.MaxAgentSessions {
			agent.Status = common.AgentStatusOnline
		}
	}
}

func (s *Store) CreateSession(sessionID, userID, agentID string) *common.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createSessionInternal(sessionID, userID, agentID)
}

func (s *Store) createSessionInternal(sessionID, userID, agentID string) *common.Session {
	now := time.Now()
	session := &common.Session{
		ID:            sessionID,
		UserID:        userID,
		AgentID:       agentID,
		Status:        common.SessionStatusActive,
		CreatedAt:     now,
		LastMessageAt: now,
		Messages:      make([]common.Message, 0),
	}
	s.sessions[sessionID] = session
	s.userSessions[userID] = append(s.userSessions[userID], session)
	if agentID != "" {
		s.agentSessions[agentID] = append(s.agentSessions[agentID], session)
		agent, exists := s.agents[agentID]
		if exists {
			agent.CurrentSessions++
			if agent.CurrentSessions >= common.MaxAgentSessions {
				agent.Status = common.AgentStatusBusy
			}
		}
	}
	return session
}

func (s *Store) GetSession(sessionID string) (*common.Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[sessionID]
	return session, exists
}

func (s *Store) AddToQueue(session *common.Session) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	session.Status = common.SessionStatusQueued
	session.QueuePosition = len(s.queue) + 1
	s.queue = append(s.queue, session)
	s.sessions[session.ID] = session
	s.userSessions[session.UserID] = append(s.userSessions[session.UserID], session)
	return session.QueuePosition
}

func (s *Store) RemoveFromQueue(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sess := range s.queue {
		if sess.ID == sessionID {
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			for j := i; j < len(s.queue); j++ {
				s.queue[j].QueuePosition--
			}
			break
		}
	}
}

func (s *Store) GetQueue() []*common.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*common.Session, len(s.queue))
	copy(result, s.queue)
	return result
}

func (s *Store) AssignQueuedSessionToAgent(sessionID, agentID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	agent, exists := s.agents[agentID]
	if !exists || agent.Status != common.AgentStatusOnline || agent.CurrentSessions >= common.MaxAgentSessions {
		return false
	}
	var session *common.Session
	sessionIndex := -1
	for i, sess := range s.queue {
		if sess.ID == sessionID {
			session = sess
			sessionIndex = i
			break
		}
	}
	if session == nil {
		return false
	}
	s.queue = append(s.queue[:sessionIndex], s.queue[sessionIndex+1:]...)
	for j := sessionIndex; j < len(s.queue); j++ {
		s.queue[j].QueuePosition--
	}
	session.Status = common.SessionStatusActive
	session.AgentID = agentID
	session.QueuePosition = 0
	s.agentSessions[agentID] = append(s.agentSessions[agentID], session)
	agent.CurrentSessions++
	if agent.CurrentSessions >= common.MaxAgentSessions {
		agent.Status = common.AgentStatusBusy
	}
	delete(s.queueReminded, sessionID)
	return true
}

func (s *Store) AddMessage(sessionID string, message common.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, exists := s.sessions[sessionID]
	if exists {
		session.Messages = append(session.Messages, message)
		session.LastMessageAt = message.Timestamp
	}
}

func (s *Store) CloseSession(sessionID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, exists := s.sessions[sessionID]
	if !exists || session.Status == common.SessionStatusClosed {
		return false
	}
	session.Status = common.SessionStatusClosed
	if session.AgentID != "" {
		agent, agentExists := s.agents[session.AgentID]
		if agentExists {
			agent.CurrentSessions--
			if agent.CurrentSessions < common.MaxAgentSessions {
				agent.Status = common.AgentStatusOnline
			}
		}
	}
	return true
}

func (s *Store) TransferSession(sessionID, fromAgentID, toAgentID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, exists := s.sessions[sessionID]
	if !exists || session.Status != common.SessionStatusActive {
		return false
	}
	if session.AgentID != fromAgentID {
		return false
	}
	oldAgent, oldExists := s.agents[fromAgentID]
	newAgent, newExists := s.agents[toAgentID]
	if !oldExists || !newExists {
		return false
	}
	if newAgent.Status != common.AgentStatusOnline || newAgent.CurrentSessions >= common.MaxAgentSessions {
		return false
	}
	if !s.isAgentOnShiftInternal(toAgentID) {
		return false
	}
	oldAgent.CurrentSessions--
	if oldAgent.CurrentSessions < common.MaxAgentSessions && oldAgent.Status == common.AgentStatusBusy {
		oldAgent.Status = common.AgentStatusOnline
	}
	newAgent.CurrentSessions++
	if newAgent.CurrentSessions >= common.MaxAgentSessions {
		newAgent.Status = common.AgentStatusBusy
	}
	session.AgentID = toAgentID
	for i, sess := range s.agentSessions[fromAgentID] {
		if sess.ID == sessionID {
			s.agentSessions[fromAgentID] = append(s.agentSessions[fromAgentID][:i], s.agentSessions[fromAgentID][i+1:]...)
			break
		}
	}
	s.agentSessions[toAgentID] = append(s.agentSessions[toAgentID], session)
	return true
}

func (s *Store) GetAgentSessions(agentID string) []*common.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sessions := s.agentSessions[agentID]
	result := make([]*common.Session, len(sessions))
	copy(result, sessions)
	return result
}

func (s *Store) GetUserSessions(userID string) []*common.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userSessions[userID]
}

func (s *Store) SubmitSatisfaction(sessionID string, satisfaction common.Satisfaction) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, exists := s.sessions[sessionID]
	if !exists || session.Satisfaction != nil {
		return false
	}
	session.Satisfaction = &satisfaction
	s.satisfactionCount++
	s.totalSatisfaction += float64(satisfaction.Rating)
	return true
}

func (s *Store) AddQuickReply(agentID string, template *common.QuickReplyTemplate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.quickReplies[agentID] = append(s.quickReplies[agentID], template)
}

func (s *Store) GetQuickReplies(agentID string) []*common.QuickReplyTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.quickReplies[agentID]
}

func (s *Store) AddToBlacklist(entry *common.BlacklistEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blacklist[entry.UserID] = entry
}

func (s *Store) RemoveFromBlacklist(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.blacklist, userID)
}

func (s *Store) IsBlacklisted(userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, exists := s.blacklist[userID]
	if !exists {
		return false
	}
	if entry.ExpiresAt.IsZero() {
		return true
	}
	return time.Now().Before(entry.ExpiresAt)
}

func (s *Store) AddSchedule(agentID string, schedule *common.Schedule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedules[agentID] = append(s.schedules[agentID], schedule)
}

func (s *Store) IsAgentOnShift(agentID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isAgentOnShiftInternal(agentID)
}

func (s *Store) isAgentOnShiftInternal(agentID string) bool {
	schedules, exists := s.schedules[agentID]
	if !exists || len(schedules) == 0 {
		return true
	}
	now := time.Now()
	for _, schedule := range schedules {
		if now.After(schedule.StartTime) && now.Before(schedule.EndTime) {
			return true
		}
	}
	return false
}

func (s *Store) GetSchedules(agentID string) []*common.Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.schedules[agentID]
}

func (s *Store) MarkQueueReminderSent(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queueReminded[sessionID] = true
}

func (s *Store) HasQueueReminderBeenSent(sessionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.queueReminded[sessionID]
}

func (s *Store) UpdateStatistics(waitTime, sessionTime float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalSessions++
	s.totalWaitTime += waitTime
	s.totalSessionTime += sessionTime
	s.statistics.TotalSessions = s.totalSessions
	if s.totalSessions > 0 {
		s.statistics.AverageWaitTime = s.totalWaitTime / float64(s.totalSessions)
		s.statistics.AverageSessionTime = s.totalSessionTime / float64(s.totalSessions)
	}
	if s.satisfactionCount > 0 {
		s.statistics.AverageSatisfaction = s.totalSatisfaction / float64(s.satisfactionCount)
	}
}

func (s *Store) GetStatistics() common.Statistics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.statistics
}

func (s *Store) GetAllSessions() []*common.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sessions := make([]*common.Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		sessions = append(sessions, sess)
	}
	return sessions
}

func (s *Store) CleanOldChats() {
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := time.Now().AddDate(0, 0, -common.ChatRetentionDays)
	for sessionID, session := range s.sessions {
		if session.Status == common.SessionStatusClosed && session.LastMessageAt.Before(cutoff) {
			delete(s.sessions, sessionID)
			for userID, sessions := range s.userSessions {
				for i, sess := range sessions {
					if sess.ID == sessionID {
						s.userSessions[userID] = append(s.userSessions[userID][:i], s.userSessions[userID][i+1:]...)
						break
					}
				}
			}
			for agentID, sessions := range s.agentSessions {
				for i, sess := range sessions {
					if sess.ID == sessionID {
						s.agentSessions[agentID] = append(s.agentSessions[agentID][:i], s.agentSessions[agentID][i+1:]...)
						break
					}
				}
			}
			delete(s.queueReminded, sessionID)
		}
	}
}

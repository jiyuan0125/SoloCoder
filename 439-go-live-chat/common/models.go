package common

import (
	"time"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Agent struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	Status         AgentStatus `json:"status"`
	CurrentSessions int `json:"current_sessions"`
}

type AgentStatus string

const (
	AgentStatusOnline  AgentStatus = "online"
	AgentStatusOffline AgentStatus = "offline"
	AgentStatusBusy    AgentStatus = "busy"
)

type SessionStatus string

const (
	SessionStatusPending  SessionStatus = "pending"
	SessionStatusActive   SessionStatus = "active"
	SessionStatusClosed   SessionStatus = "closed"
	SessionStatusQueued   SessionStatus = "queued"
)

type MessageSenderType string

const (
	SenderTypeUser  MessageSenderType = "user"
	SenderTypeAgent MessageSenderType = "agent"
	SenderTypeSystem MessageSenderType = "system"
)

type Message struct {
	ID        string            `json:"id"`
	SessionID string            `json:"session_id"`
	SenderID  string            `json:"sender_id"`
	SenderType MessageSenderType `json:"sender_type"`
	Content   string            `json:"content"`
	Timestamp time.Time         `json:"timestamp"`
}

type Session struct {
	ID               string         `json:"id"`
	UserID           string         `json:"user_id"`
	AgentID          string         `json:"agent_id,omitempty"`
	Status           SessionStatus  `json:"status"`
	CreatedAt        time.Time      `json:"created_at"`
	LastMessageAt    time.Time      `json:"last_message_at"`
	Messages         []Message      `json:"messages"`
	Satisfaction     *Satisfaction  `json:"satisfaction,omitempty"`
	QueuePosition    int            `json:"queue_position,omitempty"`
}

type Satisfaction struct {
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type QuickReplyTemplate struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type BlacklistEntry struct {
	UserID    string    `json:"user_id"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

type Schedule struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type Statistics struct {
	AverageWaitTime    float64 `json:"average_wait_time_seconds"`
	AverageSessionTime float64 `json:"average_session_time_seconds"`
	AverageSatisfaction float64 `json:"average_satisfaction"`
	TotalSessions      int     `json:"total_sessions"`
}

type CreateSessionRequest struct {
	UserID   string `json:"user_id"`
	Username string `json:"username,omitempty"`
}

type CreateSessionResponse struct {
	SessionID   string         `json:"session_id"`
	Status      SessionStatus  `json:"status"`
	AgentID     string         `json:"agent_id,omitempty"`
	QueuePosition int          `json:"queue_position,omitempty"`
}

type SendMessageRequest struct {
	SessionID  string            `json:"session_id"`
	SenderID   string            `json:"sender_id"`
	SenderType MessageSenderType `json:"sender_type"`
	Content    string            `json:"content"`
}

type SendMessageResponse struct {
	MessageID string    `json:"message_id"`
	Timestamp time.Time `json:"timestamp"`
}

type CloseSessionRequest struct {
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id,omitempty"`
}

type CloseSessionResponse struct {
	SessionID string `json:"session_id"`
	Closed    bool   `json:"closed"`
}

type SubmitSatisfactionRequest struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment,omitempty"`
}

type SubmitSatisfactionResponse struct {
	Success bool `json:"success"`
}

type TransferSessionRequest struct {
	SessionID    string `json:"session_id"`
	FromAgentID  string `json:"from_agent_id"`
	ToAgentID    string `json:"to_agent_id"`
	Reason       string `json:"reason,omitempty"`
}

type TransferSessionResponse struct {
	Success   bool   `json:"success"`
	NewAgentID string `json:"new_agent_id"`
}

type CreateQuickReplyRequest struct {
	AgentID string `json:"agent_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreateQuickReplyResponse struct {
	TemplateID string `json:"template_id"`
}

type GetQuickRepliesRequest struct {
	AgentID string `json:"agent_id"`
}

type GetQuickRepliesResponse struct {
	Templates []QuickReplyTemplate `json:"templates"`
}

type AddToBlacklistRequest struct {
	UserID    string    `json:"user_id"`
	Reason    string    `json:"reason,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

type AddToBlacklistResponse struct {
	Success bool `json:"success"`
}

type RemoveFromBlacklistRequest struct {
	UserID string `json:"user_id"`
}

type RemoveFromBlacklistResponse struct {
	Success bool `json:"success"`
}

type AgentLoginRequest struct {
	AgentID  string `json:"agent_id"`
	Username string `json:"username"`
}

type AgentLoginResponse struct {
	Success bool `json:"success"`
}

type AgentLogoutRequest struct {
	AgentID string `json:"agent_id"`
}

type AgentLogoutResponse struct {
	Success bool `json:"success"`
}

type GetAgentSessionsRequest struct {
	AgentID string `json:"agent_id"`
}

type GetAgentSessionsResponse struct {
	Sessions []Session `json:"sessions"`
}

type GetSessionHistoryRequest struct {
	SessionID string `json:"session_id"`
}

type GetSessionHistoryResponse struct {
	Session  Session `json:"session"`
	Messages []Message `json:"messages"`
}

type AddScheduleRequest struct {
	AgentID   string    `json:"agent_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type AddScheduleResponse struct {
	ScheduleID string `json:"schedule_id"`
	Success    bool   `json:"success"`
}

type GetSchedulesRequest struct {
	AgentID string `json:"agent_id"`
}

type GetSchedulesResponse struct {
	Schedules []Schedule `json:"schedules"`
}

type GetStatisticsResponse struct {
	Statistics Statistics `json:"statistics"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

const (
	MaxAgentSessions    = 3
	QueueReminderMinutes = 2
	SessionTimeoutMinutes = 30
	ChatRetentionDays    = 90
)

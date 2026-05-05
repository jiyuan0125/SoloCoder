package common

import "time"

type Notification struct {
	ID          string           `json:"id"`
	Type        NotificationType `json:"type"`
	Priority    Priority         `json:"priority"`
	Sender      string           `json:"sender"`
	Receiver    string           `json:"receiver"`
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	Status      ReadStatus       `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	ReadAt      *time.Time       `json:"read_at,omitempty"`
	IsArchived  bool             `json:"is_archived"`
	IsEmergency bool             `json:"is_emergency"`
	RetryCount  int              `json:"-"`
	SendFailed  bool             `json:"-"`
	LastReminderAt *time.Time    `json:"-"`
}

type NotificationTemplate struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        NotificationType `json:"type"`
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	Variables   []string         `json:"variables"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type UserInfo struct {
	UserID         string     `json:"user_id"`
	Status         UserStatus `json:"status"`
	LastActiveAt   time.Time  `json:"last_active_at"`
	LastSummaryAt  *time.Time `json:"last_summary_at,omitempty"`
}

type FailedLog struct {
	ID              string           `json:"id"`
	NotificationID  string           `json:"notification_id"`
	Receiver        string           `json:"receiver"`
	Title           string           `json:"title"`
	Content         string           `json:"content"`
	RetryCount      int              `json:"retry_count"`
	ErrorMessages   []string         `json:"error_messages"`
	FailedAt        time.Time        `json:"failed_at"`
}

type SendRequest struct {
	Type        NotificationType `json:"type"`
	Priority    Priority         `json:"priority,omitempty"`
	Sender      string           `json:"sender"`
	Receiver    string           `json:"receiver"`
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	TemplateID  string           `json:"template_id,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
}

type SendResponse struct {
	Success       bool   `json:"success"`
	NotificationID string `json:"notification_id,omitempty"`
	Error         string `json:"error,omitempty"`
}

type ListRequest struct {
	Receiver    string           `json:"receiver"`
	Type        *NotificationType `json:"type,omitempty"`
	StartTime   *time.Time       `json:"start_time,omitempty"`
	EndTime     *time.Time       `json:"end_time,omitempty"`
	Status      *ReadStatus      `json:"status,omitempty"`
	IncludeArchived bool         `json:"include_archived,omitempty"`
	Page        int              `json:"page,omitempty"`
	PageSize    int              `json:"page_size,omitempty"`
}

type ListResponse struct {
	Notifications []Notification `json:"notifications"`
	Total         int            `json:"total"`
	UnreadCount   int            `json:"unread_count"`
}

type MarkReadRequest struct {
	Receiver        string   `json:"receiver"`
	NotificationIDs []string `json:"notification_ids"`
}

type MarkReadResponse struct {
	Success     bool   `json:"success"`
	MarkedCount int    `json:"marked_count"`
	Error       string `json:"error,omitempty"`
}

type UnreadCountResponse struct {
	Receiver    string `json:"receiver"`
	UnreadCount int    `json:"unread_count"`
}

type TemplateCreateRequest struct {
	Name        string           `json:"name"`
	Type        NotificationType `json:"type"`
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	Variables   []string         `json:"variables"`
}

type TemplateUpdateRequest struct {
	ID          string            `json:"id"`
	Name        *string           `json:"name,omitempty"`
	Type        *NotificationType `json:"type,omitempty"`
	Title       *string           `json:"title,omitempty"`
	Content     *string           `json:"content,omitempty"`
	Variables   *[]string         `json:"variables,omitempty"`
}

type UserActivityRequest struct {
	UserID string `json:"user_id"`
}

type APIResponse struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

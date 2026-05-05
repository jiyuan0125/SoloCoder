package protocol

import "time"

const (
	MaxRecipientsPerBatch = 100000
	BatchIntervalSeconds  = 30
	MaxInvalidAttempts    = 3
	BounceRateThreshold   = 0.10
	ResendWindowHours     = 24
)

type SubscriptionStatus string

const (
	StatusSubscribed   SubscriptionStatus = "subscribed"
	StatusUnsubscribed SubscriptionStatus = "unsubscribed"
)

type SendStatus string

const (
	SendStatusSuccess SendStatus = "success"
	SendStatusFailed  SendStatus = "failed"
	SendStatusBounced SendStatus = "bounced"
	SendStatusInvalid SendStatus = "invalid"
	SendStatusPending SendStatus = "pending"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusScheduled  TaskStatus = "scheduled"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusPaused     TaskStatus = "paused"
	TaskStatusFailed     TaskStatus = "failed"
)

type AlertLevel string

const (
	AlertLevelInfo    AlertLevel = "info"
	AlertLevelWarning AlertLevel = "warning"
	AlertLevelError   AlertLevel = "error"
)

type MailingList struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPaused    bool      `json:"is_paused"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Subscriber struct {
	Email              string             `json:"email"`
	Name               string             `json:"name"`
	Status             SubscriptionStatus `json:"status"`
	RegisteredAt       time.Time          `json:"registered_at"`
	InvalidAttempts    int                `json:"invalid_attempts"`
	LastInvalidAt      *time.Time         `json:"last_invalid_at,omitempty"`
	CustomFields       map[string]string  `json:"custom_fields,omitempty"`
}

type SubscriberListEntry struct {
	Subscriber
	ListID    string    `json:"list_id"`
	SubscribedAt time.Time `json:"subscribed_at"`
}

type EmailTemplate struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Subject        string            `json:"subject"`
	HTMLBody       string            `json:"html_body"`
	TextBody       string            `json:"text_body"`
	TrackOpens     bool              `json:"track_opens"`
	TrackClicks    bool              `json:"track_clicks"`
	Variables      map[string]string `json:"variables,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type ABTest struct {
	ID              string    `json:"id"`
	TaskID          string    `json:"task_id"`
	SubjectA        string    `json:"subject_a"`
	SubjectB        string    `json:"subject_b"`
	TotalRecipients int       `json:"total_recipients"`
	GroupAEmails    []string  `json:"group_a_emails,omitempty"`
	GroupBEmails    []string  `json:"group_b_emails,omitempty"`
	OpensA          int       `json:"opens_a"`
	OpensB          int       `json:"opens_b"`
	ClicksA         int       `json:"clicks_a"`
	ClicksB         int       `json:"clicks_b"`
	Winner          string    `json:"winner,omitempty"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type SendTask struct {
	ID              string     `json:"id"`
	ListID          string     `json:"list_id"`
	TemplateID      string     `json:"template_id"`
	Subject         string     `json:"subject"`
	HTMLBody        string     `json:"html_body"`
	TextBody        string     `json:"text_body"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	Status          TaskStatus `json:"status"`
	TotalRecipients int        `json:"total_recipients"`
	SentCount       int        `json:"sent_count"`
	FailedCount     int        `json:"failed_count"`
	BouncedCount    int        `json:"bounced_count"`
	IsABTest        bool       `json:"is_ab_test"`
	ABTestID        string     `json:"ab_test_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type SendRecord struct {
	ID        string     `json:"id"`
	TaskID    string     `json:"task_id"`
	Email     string     `json:"email"`
	Status    SendStatus `json:"status"`
	ErrorMsg  string     `json:"error_msg,omitempty"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	OpenedAt  *time.Time `json:"opened_at,omitempty"`
	ClickedAt *time.Time `json:"clicked_at,omitempty"`
	ABGroup   string     `json:"ab_group,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type Alert struct {
	ID        string     `json:"id"`
	ListID    string     `json:"list_id,omitempty"`
	TaskID    string     `json:"task_id,omitempty"`
	Level     AlertLevel `json:"level"`
	Message   string     `json:"message"`
	Resolved  bool       `json:"resolved"`
	CreatedAt time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

type CreateMailingListRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateMailingListResponse struct {
	Success    bool        `json:"success"`
	MailingList *MailingList `json:"mailing_list,omitempty"`
	Error      string      `json:"error,omitempty"`
}

type GetMailingListResponse struct {
	Success    bool        `json:"success"`
	MailingList *MailingList `json:"mailing_list,omitempty"`
	Error      string      `json:"error,omitempty"`
}

type ListMailingListsResponse struct {
	Success       bool           `json:"success"`
	MailingLists  []*MailingList `json:"mailing_lists"`
	Error         string         `json:"error,omitempty"`
}

type SubscribeRequest struct {
	Email        string            `json:"email"`
	Name         string            `json:"name"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
}

type SubscribeResponse struct {
	Success    bool              `json:"success"`
	Subscriber *SubscriberListEntry `json:"subscriber,omitempty"`
	Error      string            `json:"error,omitempty"`
}

type UnsubscribeRequest struct {
	Email string `json:"email"`
}

type UnsubscribeResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ListSubscribersResponse struct {
	Success     bool                   `json:"success"`
	Subscribers []*SubscriberListEntry `json:"subscribers"`
	Error       string                 `json:"error,omitempty"`
}

type CreateTemplateRequest struct {
	Name        string            `json:"name"`
	Subject     string            `json:"subject"`
	HTMLBody    string            `json:"html_body"`
	TextBody    string            `json:"text_body"`
	TrackOpens  bool              `json:"track_opens"`
	TrackClicks bool              `json:"track_clicks"`
	Variables   map[string]string `json:"variables,omitempty"`
}

type CreateTemplateResponse struct {
	Success  bool           `json:"success"`
	Template *EmailTemplate `json:"template,omitempty"`
	Error    string         `json:"error,omitempty"`
}

type ListTemplatesResponse struct {
	Success   bool              `json:"success"`
	Templates []*EmailTemplate  `json:"templates"`
	Error     string            `json:"error,omitempty"`
}

type ImportSubscribersRequest struct {
	FilePath string `json:"file_path"`
}

type ImportSubscribersResponse struct {
	Success     bool   `json:"success"`
	Total       int    `json:"total"`
	Imported    int    `json:"imported"`
	Skipped     int    `json:"skipped"`
	Error       string `json:"error,omitempty"`
}

type SendCampaignRequest struct {
	ListID         string     `json:"list_id"`
	TemplateID     string     `json:"template_id,omitempty"`
	Subject        string     `json:"subject,omitempty"`
	HTMLBody       string     `json:"html_body,omitempty"`
	TextBody       string     `json:"text_body,omitempty"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	IsABTest       bool       `json:"is_ab_test"`
	SubjectA       string     `json:"subject_a,omitempty"`
	SubjectB       string     `json:"subject_b,omitempty"`
}

type SendCampaignResponse struct {
	Success bool       `json:"success"`
	Task    *SendTask  `json:"task,omitempty"`
	Error   string     `json:"error,omitempty"`
}

type GetTaskResponse struct {
	Success bool       `json:"success"`
	Task    *SendTask  `json:"task,omitempty"`
	Error   string     `json:"error,omitempty"`
}

type ListTasksResponse struct {
	Success bool       `json:"success"`
	Tasks   []*SendTask `json:"tasks"`
	Error   string     `json:"error,omitempty"`
}

type GetSendRecordsResponse struct {
	Success     bool          `json:"success"`
	SendRecords []*SendRecord `json:"send_records"`
	Error       string        `json:"error,omitempty"`
}

type TrackOpenRequest struct {
	TaskID string `json:"task_id"`
	Email  string `json:"email"`
}

type TrackClickRequest struct {
	TaskID string `json:"task_id"`
	Email  string `json:"email"`
	URL    string `json:"url"`
}

type TrackResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetAlertsResponse struct {
	Success bool      `json:"success"`
	Alerts  []*Alert  `json:"alerts"`
	Error   string    `json:"error,omitempty"`
}

type ResolveAlertRequest struct {
	AlertID string `json:"alert_id"`
}

type ResolveAlertResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type PauseListRequest struct {
	ListID string `json:"list_id"`
}

type PauseListResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ResumeListRequest struct {
	ListID string `json:"list_id"`
}

type ResumeListResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetStatsResponse struct {
	Success bool `json:"success"`
	Stats   struct {
		TotalLists        int `json:"total_lists"`
		TotalSubscribers  int `json:"total_subscribers"`
		TotalTasks        int `json:"total_tasks"`
		TotalSent         int `json:"total_sent"`
		TotalBounced      int `json:"total_bounced"`
		ActiveAlerts      int `json:"active_alerts"`
	} `json:"stats"`
	Error string `json:"error,omitempty"`
}

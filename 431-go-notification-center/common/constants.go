package common

const (
	DefaultPort        = 8080
	MaxBatchReadCount  = 100
	DedupWindowMinutes = 5
	ArchiveDays        = 30
	InactiveDays       = 3
	MaxRetryCount      = 3
	DailySummaryHour   = 8
	HighPriorityReminderIntervalHours = 24
)

type NotificationType string

const (
	TypeSystem    NotificationType = "system"
	TypeApproval  NotificationType = "approval"
	TypeTask      NotificationType = "task"
	TypeSecurity  NotificationType = "security"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

type ReadStatus string

const (
	StatusUnread ReadStatus = "unread"
	StatusRead   ReadStatus = "read"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)

package common

import "time"

type CommentStatus string

const (
	StatusPending       CommentStatus = "pending"
	StatusAutoApproved  CommentStatus = "auto_approved"
	StatusApproved      CommentStatus = "approved"
	StatusRejected      CommentStatus = "rejected"
	StatusManualReview  CommentStatus = "manual_review"
	StatusPublished     CommentStatus = "published"
)

type SensitiveWordLevel string

const (
	LevelSevere  SensitiveWordLevel = "severe"
	LevelMedium  SensitiveWordLevel = "medium"
	LevelMild    SensitiveWordLevel = "mild"
)

type Comment struct {
	ID             string
	UserID         string
	Content        string
	Status         CommentStatus
	RejectReason   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	EditDeadline   time.Time
	EditCount      int
	AssignedTo     string
	ReportedBy     []string
	ReportCount    int
	ViolationLevel SensitiveWordLevel
}

type User struct {
	ID                  string
	ConsecutiveRejects  int
	LastRejectTime      time.Time
	Comments24h         int
	Last24hResetTime    time.Time
	UnderManualReview   bool
	ManualReviewEndTime time.Time
}

type Moderator struct {
	ID           string
	Name         string
	DailyCount   map[string]int
	TotalCount   int
	AssignedIDs  []string
}

type SensitiveWord struct {
	Word  string
	Level SensitiveWordLevel
}

type AuditLog struct {
	ID        string
	ModeratorID string
	Action    string
	CommentID string
	Timestamp time.Time
	Details   string
}

type RejectReasonTemplate struct {
	Level   SensitiveWordLevel
	Message string
}

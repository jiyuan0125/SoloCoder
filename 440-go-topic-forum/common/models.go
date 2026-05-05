package common

import "time"

type User struct {
	ID          string
	Username    string
	IsAdmin     bool
	PostCount   int
	ReplyCount  int
	LikeCount   int
	CreatedAt   time.Time
}

type Post struct {
	ID              string
	UserID          string
	Username        string
	Title           string
	Content         string
	Category        string
	Tags            []string
	ViewCount       int
	ReplyCount      int
	LikeCount       int
	IsTop           bool
	IsEssence       bool
	BestReplyID     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PostHistory struct {
	ID        string
	PostID    string
	Title     string
	Content   string
	Tags      []string
	EditedAt  time.Time
	EditorID  string
}

type Reply struct {
	ID        string
	PostID    string
	UserID    string
	Username  string
	Content   string
	ParentID  string
	Floor     int
	LikeCount int
	CreatedAt time.Time
}

type Like struct {
	ID         string
	UserID     string
	TargetID   string
	TargetType string
	CreatedAt  time.Time
}

type Report struct {
	ID          string
	PostID      string
	ReporterID  string
	Reason      string
	Status      string
	CreatedAt   time.Time
	ReviewedAt  time.Time
	ReviewerID  string
}

type TagUsage struct {
	Name      string
	PostCount int
}

type HotPost struct {
	PostID string
	Score  float64
	Date   string
}

type ViewRecord struct {
	PostID    string
	UserID    string
	ViewedAt  time.Time
}

const (
	MaxPostLength    = 5000
	MaxPostsPerDay   = 10
	HotPostCount     = 10
	ViewWeight       = 0.3
	ReplyWeight      = 0.7
)

const (
	ReportStatusPending   = "pending"
	ReportStatusResolved  = "resolved"
	ReportStatusRejected  = "rejected"
)

const (
	TargetTypePost  = "post"
	TargetTypeReply = "reply"
)

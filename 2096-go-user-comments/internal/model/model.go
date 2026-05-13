package model

import "time"

type Comment struct {
	ID           int64     `json:"id"`
	ArticleID    int64     `json:"article_id"`
	UserID       int64     `json:"user_id"`
	Nickname     string    `json:"nickname"`
	Avatar       string    `json:"avatar"`
	Content      string    `json:"content"`
	ContentType  string    `json:"content_type"`
	ParentID     *int64    `json:"parent_id,omitempty"`
	ReplyLevel   int       `json:"reply_level"`
	Likes        int       `json:"likes"`
	IsDeleted    bool      `json:"is_deleted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	RootParentID *int64    `json:"-"`
}

type Article struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type RateLimiter struct {
	LastRequestTime time.Time
}

const (
	ContentTypePlain = "plain"
	ContentTypeRich  = "rich"
	MaxReplyLevel    = 3
	MaxContentLength = 2000
	RateLimitSeconds = 5
)

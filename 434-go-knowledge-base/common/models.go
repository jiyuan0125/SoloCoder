package common

import (
	"time"
)

type ArticleStatus string

const (
	StatusDraft    ArticleStatus = "draft"
	StatusPublished ArticleStatus = "published"
	StatusArchived ArticleStatus = "archived"
)

type AccessLevel string

const (
	AccessPublic     AccessLevel = "public"
	AccessLoggedIn   AccessLevel = "logged_in"
	AccessDepartment AccessLevel = "department"
)

type User struct {
	ID         string
	Name       string
	Department string
	IsLoggedIn bool
}

type Article struct {
	ID             string
	Title          string
	Content        string
	Category       string
	Tags           []string
	Status         ArticleStatus
	AuthorID       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CurrentVersion int
	AccessLevel    AccessLevel
	DepartmentIDs  []string
	ViewCount      int
	ReferenceCount int
	RelatedArticleIDs []string
}

type ArticleVersion struct {
	ID         string
	ArticleID  string
	VersionNum int
	Title      string
	Content    string
	Category   string
	Tags       []string
	CreatedBy  string
	CreatedAt  time.Time
}

type Favorite struct {
	ID        string
	UserID    string
	ArticleID string
	CreatedAt time.Time
}

type Reference struct {
	ID             string
	FromArticleID  string
	ToArticleID    string
	CreatedAt      time.Time
}

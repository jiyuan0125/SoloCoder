package model

import "time"

type RouteMatchType string

const (
	MatchTypeExact  RouteMatchType = "exact"
	MatchTypePrefix RouteMatchType = "prefix"
)

type AuthType string

const (
	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeJWT    AuthType = "jwt"
	AuthTypeNone   AuthType = "none"
)

type FlowStatus string

const (
	StatusDraft      FlowStatus = "draft"
	StatusPending    FlowStatus = "pending"
	StatusApproval1  FlowStatus = "approval1"
	StatusApproval2  FlowStatus = "approval2"
	StatusApproved   FlowStatus = "approved"
	StatusCompleted  FlowStatus = "completed"
	StatusRejected   FlowStatus = "rejected"
)

type RouteRule struct {
	ID           int64
	Name         string
	Path         string
	MatchType    RouteMatchType
	Headers      map[string]string
	TargetURL    string
	Timeout      time.Duration
	AuthType     AuthType
	RateLimitIP  int
	RateLimitKey int
	Healthy      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RouteChange struct {
	ID            int64
	RouteID       int64
	Operation     string
	OldConfig     string
	NewConfig     string
	Status        FlowStatus
	Applicant     string
	Approver1     string
	Approver2     string
	FinalApprover string
	Reason        string
	AppliedAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type APIKey struct {
	ID        int64
	Key       string
	UserID    string
	ExpiresAt *time.Time
	CreatedAt time.Time
}

type JWTSecret struct {
	ID        int64
	Secret    string
	Issuer    string
	CreatedAt time.Time
}

type AccessLog struct {
	ID          int64
	RequestPath string
	StatusCode  int
	Duration    time.Duration
	ClientID    string
	ClientIP    string
	Timestamp   time.Time
}

type RateLimitKey struct {
	Key       string
	Count     int
	Window    time.Time
	CreatedAt time.Time
}

package shared

import (
	"time"
)

type BehaviorType string

const (
	BehaviorTypePageView   BehaviorType = "page_view"
	BehaviorTypeButtonClick BehaviorType = "button_click"
	BehaviorTypeFeatureUse  BehaviorType = "feature_use"
)

type Behavior struct {
	ID           string       `json:"id"`
	UserID       string       `json:"user_id"`
	SessionID    string       `json:"session_id"`
	Type         BehaviorType `json:"type"`
	Timestamp    time.Time    `json:"timestamp"`
	PageView     *PageViewData    `json:"page_view,omitempty"`
	ButtonClick  *ButtonClickData `json:"button_click,omitempty"`
	FeatureUse   *FeatureUseData  `json:"feature_use,omitempty"`
	IsFiltered   bool         `json:"is_filtered"`
	FilterReason string       `json:"filter_reason,omitempty"`
}

type PageViewData struct {
	URL           string        `json:"url"`
	Duration      time.Duration `json:"duration"`
	PageKey       string        `json:"page_key"`
}

type ButtonClickData struct {
	ButtonID string `json:"button_id"`
}

type FeatureUseData struct {
	FeatureName string        `json:"feature_name"`
	Duration    time.Duration `json:"duration"`
}

type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	StartTime    time.Time `json:"start_time"`
	LastActivity time.Time `json:"last_activity"`
	IsActive     bool      `json:"is_active"`
}

type FunnelStep struct {
	Name           string  `json:"name"`
	BehaviorType   BehaviorType `json:"behavior_type"`
	PageKey        string  `json:"page_key,omitempty"`
	ButtonID       string  `json:"button_id,omitempty"`
	FeatureName    string  `json:"feature_name,omitempty"`
}

type FunnelResult struct {
	Steps []FunnelStepResult `json:"steps"`
}

type FunnelStepResult struct {
	Name           string  `json:"name"`
	UserCount      int     `json:"user_count"`
	ConversionRate float64 `json:"conversion_rate"`
	ChurnRate      float64 `json:"churn_rate"`
	IsAvailable    bool    `json:"is_available"`
}

type RetentionRequest struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

type RetentionResult struct {
	Day1Retention  float64 `json:"day_1_retention"`
	Day7Retention  float64 `json:"day_7_retention"`
	Day30Retention float64 `json:"day_30_retention"`
}

type PathAnalysisRequest struct {
	FromPage string `json:"from_page"`
	ToPage   string `json:"to_page"`
}

type PathResult struct {
	Paths []PathItem `json:"paths"`
}

type PathItem struct {
	Path       []string `json:"path"`
	UserCount  int      `json:"user_count"`
	Percentage float64  `json:"percentage"`
}

type UserProfile struct {
	UserID       string            `json:"user_id"`
	InterestTags map[string]int    `json:"interest_tags"`
	ActivityLevel string           `json:"activity_level"`
	LastActive   time.Time         `json:"last_active"`
	FirstActive  time.Time         `json:"first_active"`
	TotalSessions int              `json:"total_sessions"`
}

type RealtimeStats struct {
	OnlineUserCount int             `json:"online_user_count"`
	ActivePages     []PageRanking   `json:"active_pages"`
}

type PageRanking struct {
	PageKey string `json:"page_key"`
	UserCount int   `json:"user_count"`
	Rank     int    `json:"rank"`
}

const (
	SessionTimeout      = 30 * time.Minute
	MaxFunnelSteps      = 10
	DeduplicationWindow = 1 * time.Minute
)

type TrackRequest struct {
	UserID       string       `json:"user_id"`
	SessionID    string       `json:"session_id,omitempty"`
	Type         BehaviorType `json:"type"`
	URL          string       `json:"url,omitempty"`
	Duration     int64        `json:"duration,omitempty"`
	PageKey      string       `json:"page_key,omitempty"`
	ButtonID     string       `json:"button_id,omitempty"`
	FeatureName  string       `json:"feature_name,omitempty"`
}

type TrackResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id"`
	Message   string `json:"message,omitempty"`
}

type FunnelRequest struct {
	Steps []FunnelStep `json:"steps"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

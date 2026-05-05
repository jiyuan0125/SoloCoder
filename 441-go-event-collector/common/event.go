package common

type Event struct {
	Name       string                 `json:"name"`
	Timestamp  int64                  `json:"timestamp"`
	Properties map[string]interface{} `json:"properties"`
	UserID     string                 `json:"user_id"`
	DeviceID   string                 `json:"device_id"`
}

type TrackRequest struct {
	Event
}

type TrackResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type QueryRequest struct {
	EventName   string `json:"event_name,omitempty"`
	StartTime   int64  `json:"start_time,omitempty"`
	EndTime     int64  `json:"end_time,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
}

type QueryResponse struct {
	Success bool    `json:"success"`
	Data    []Event `json:"data,omitempty"`
	Total   int     `json:"total,omitempty"`
	Message string  `json:"message,omitempty"`
}

type AggregationRequest struct {
	EventNames []string `json:"event_names"`
	StartTime  int64    `json:"start_time"`
	EndTime    int64    `json:"end_time"`
}

type AggregationResponse struct {
	Success bool                `json:"success"`
	Data    AggregationResult   `json:"data,omitempty"`
	Message string              `json:"message,omitempty"`
}

type AggregationResult struct {
	IndividualStats []EventStat     `json:"individual_stats"`
	OverlapUsers    []string        `json:"overlap_users"`
	OverlapCount    int             `json:"overlap_count"`
}

type EventStat struct {
	EventName  string `json:"event_name"`
	TotalCount int    `json:"total_count"`
	UserCount  int    `json:"user_count"`
}

type DailyReport struct {
	Date         string               `json:"date"`
	EventSummary map[string]EventDaily `json:"event_summary"`
}

type EventDaily struct {
	TotalCount  int      `json:"total_count"`
	UniqueUsers []string `json:"unique_users"`
}

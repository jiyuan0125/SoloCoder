package api

import "time"

type VariableRule struct {
	Type       string `json:"type"`
	HeaderName string `json:"header_name,omitempty"`
	JSONPath   string `json:"json_path,omitempty"`
	Name       string `json:"name"`
}

type VariableValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ReplaceRule struct {
	OldPrefix string `json:"old_prefix,omitempty"`
	NewPrefix string `json:"new_prefix,omitempty"`
}

type CreateRecordRequest struct {
	TargetURL     string          `json:"target_url"`
	VariableRules []VariableRule  `json:"variable_rules,omitempty"`
}

type CreateRecordResponse struct {
	SessionID  string    `json:"session_id"`
	ProxyAddr  string    `json:"proxy_addr"`
	TargetURL  string    `json:"target_url"`
	StartTime  time.Time `json:"start_time"`
}

type StopRecordResponse struct {
	SessionID    string    `json:"session_id"`
	EndTime      time.Time `json:"end_time"`
	RequestCount int       `json:"request_count"`
	FilePath     string    `json:"file_path"`
}

type PlayRequest struct {
	SessionID      string          `json:"session_id"`
	TargetURL      string          `json:"target_url,omitempty"`
	Mode           string          `json:"mode"`
	Concurrency    int             `json:"concurrency,omitempty"`
	ReplaceRules   []ReplaceRule   `json:"replace_rules,omitempty"`
	VariableValues []VariableValue `json:"variable_values,omitempty"`
}

type PlayResponse struct {
	SessionID   string          `json:"session_id"`
	Total       int             `json:"total"`
	Success     int             `json:"success"`
	Failed      int             `json:"failed"`
	StartTime   time.Time       `json:"start_time"`
	EndTime     time.Time       `json:"end_time"`
	DurationMs  int64           `json:"duration_ms"`
	Results     []PlayResult    `json:"results,omitempty"`
}

type PlayResult struct {
	Index        int             `json:"index"`
	Method       string          `json:"method"`
	URL          string          `json:"url"`
	Success      bool            `json:"success"`
	StatusCode   int             `json:"status_code,omitempty"`
	Error        string          `json:"error,omitempty"`
	DurationMs   int64           `json:"duration_ms"`
}

type RecordSession struct {
	SessionID    string    `json:"session_id"`
	TargetURL    string    `json:"target_url"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time,omitempty"`
	RequestCount int       `json:"request_count"`
	Active       bool      `json:"active"`
	FilePath     string    `json:"file_path,omitempty"`
}

type RecordSessionList struct {
	Sessions []RecordSession `json:"sessions"`
}

type RecordedSessionData struct {
	Meta      SessionMeta     `json:"meta"`
	Requests  []RecordedReq   `json:"requests"`
}

type SessionMeta struct {
	SessionID    string      `json:"session_id"`
	TargetURL    string      `json:"target_url"`
	StartTime    time.Time   `json:"start_time"`
	EndTime      time.Time   `json:"end_time"`
	RequestCount int         `json:"request_count"`
}

type RecordedReq struct {
	Index            int               `json:"index"`
	Timestamp        time.Time         `json:"timestamp"`
	Method           string            `json:"method"`
	URL              string            `json:"url"`
	Path             string            `json:"path"`
	Query            string            `json:"query,omitempty"`
	Headers          map[string]string `json:"headers,omitempty"`
	Body             string            `json:"body,omitempty"`
	ResponseStatus   int               `json:"response_status"`
	ResponseHeaders  map[string]string `json:"response_headers,omitempty"`
	ResponseBody     string            `json:"response_body,omitempty"`
	RequestDuration  int64             `json:"request_duration_ms"`
	DelayFromPrevMs  int64             `json:"delay_from_prev_ms"`
	ExtractedVars    map[string]string `json:"extracted_vars,omitempty"`
}

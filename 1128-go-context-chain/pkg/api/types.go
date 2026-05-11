package api

import "time"

type RequestCreate struct {
	Timeout    time.Duration `json:"timeout"`
	Values     []ValueItem   `json:"values"`
	ParentID   string        `json:"parent_id"`
}

type ValueItem struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type ResponseCreate struct {
	RequestID string      `json:"request_id"`
	CancelID  string      `json:"cancel_id"`
	Context   ContextInfo `json:"context"`
}

type ContextInfo struct {
	ID         string            `json:"id"`
	ParentID   string            `json:"parent_id"`
	Timeout    time.Duration     `json:"timeout,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	Deadline   time.Time         `json:"deadline,omitempty"`
	Values     map[string]string `json:"values"`
	Cancelled  bool              `json:"cancelled"`
	CanceledAt time.Time         `json:"canceled_at,omitempty"`
}

type RequestCancel struct {
	CancelID string `json:"cancel_id"`
}

type ResponseCancel struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ResponseChain struct {
	Triggers    []TriggerInfo `json:"triggers"`
	TotalCancel int           `json:"total_cancel"`
}

type TriggerInfo struct {
	TriggerID   string        `json:"trigger_id"`
	TriggerType string        `json:"trigger_type"`
	CanceledAt  time.Time     `json:"canceled_at"`
	Children    []CanceledCtx `json:"children"`
}

type CanceledCtx struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
}

type RequestPrecision struct {
	TargetTimeout time.Duration `json:"target_timeout"`
	Iterations    int           `json:"iterations"`
}

type ResponsePrecision struct {
	Iterations  int           `json:"iterations"`
	Target      time.Duration `json:"target"`
	Stats       PrecisionStats `json:"stats"`
	Results     []PrecisionResult `json:"results"`
}

type PrecisionStats struct {
	Mean   time.Duration `json:"mean"`
	Median time.Duration `json:"median"`
	Min    time.Duration `json:"min"`
	Max    time.Duration `json:"max"`
	StdDev time.Duration `json:"std_dev"`
}

type PrecisionResult struct {
	Index        int           `json:"index"`
	Actual       time.Duration `json:"actual"`
	Deviation    time.Duration `json:"deviation"`
	PercentError float64       `json:"percent_error"`
}

type RequestLeak struct {
	RequestID string `json:"request_id"`
}

type ResponseLeak struct {
	IsLeaking bool   `json:"is_leaking"`
	GoroutineCount int `json:"goroutine_count"`
	Message   string `json:"message"`
}

type ContextSummary struct {
	Total    int    `json:"total"`
	Active   int    `json:"active"`
	Canceled int    `json:"canceled"`
}

type ResponseStatus struct {
	Contexts ContextSummary `json:"contexts"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

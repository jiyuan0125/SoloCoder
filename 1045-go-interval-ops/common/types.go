package common

import "time"

type Operation string

const (
	OpMerge      Operation = "merge"
	OpDifference Operation = "difference"
	OpIntersect  Operation = "intersect"
	OpQuery      Operation = "query"
)

type IntervalType string

const (
	TypeInteger IntervalType = "integer"
	TypeTime    IntervalType = "time"
)

type IntegerInterval struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type TimeInterval struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type OperationRequest struct {
	BatchID      string         `json:"batch_id"`
	Operation    Operation      `json:"operation"`
	IntervalType IntervalType   `json:"interval_type"`
	Timezone     string         `json:"timezone,omitempty"`
	IntegersA    []IntegerInterval `json:"integers_a,omitempty"`
	IntegersB    []IntegerInterval `json:"integers_b,omitempty"`
	TimesA       []TimeInterval  `json:"times_a,omitempty"`
	TimesB       []TimeInterval  `json:"times_b,omitempty"`
	IntegerPoint *int           `json:"integer_point,omitempty"`
	TimePoint    *time.Time     `json:"time_point,omitempty"`
}

type OperationResult struct {
	Integers []IntegerInterval `json:"integers,omitempty"`
	Times    []TimeInterval  `json:"times,omitempty"`
}

type OperationResponse struct {
	BatchID string          `json:"batch_id"`
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
	Result  OperationResult `json:"result,omitempty"`
}

type BatchResult struct {
	BatchID    string          `json:"batch_id"`
	Request    OperationRequest `json:"request"`
	Result     OperationResult  `json:"result"`
	CreatedAt  time.Time       `json:"created_at"`
}

type QueryBatchRequest struct {
	BatchID string `json:"batch_id"`
}

type QueryBatchResponse struct {
	Found bool          `json:"found"`
	Error string        `json:"error,omitempty"`
	Data  *BatchResult  `json:"data,omitempty"`
}

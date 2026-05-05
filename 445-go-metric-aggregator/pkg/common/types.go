package common

type Granularity string

const (
	Granularity1Min   Granularity = "1m"
	Granularity5Min   Granularity = "5m"
	Granularity1Hour  Granularity = "1h"
)

type AggregationType string

const (
	AggregationAvg   AggregationType = "avg"
	AggregationMax   AggregationType = "max"
	AggregationMin   AggregationType = "min"
	AggregationSum   AggregationType = "sum"
)

const (
	MinValue = -1000000.0
	MaxValue = 1000000.0
)

type DataPoint struct {
	Metric    string  `json:"metric"`
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type AggregationResult struct {
	Metric     string          `json:"metric"`
	Start      int64           `json:"start"`
	End        int64           `json:"end"`
	Granularity Granularity    `json:"granularity"`
	Points     []AggregatedPoint `json:"points"`
}

type AggregatedPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
	Count     int     `json:"count,omitempty"`
	IsFilled  bool    `json:"is_filled,omitempty"`
}

type CreateMetricRequest struct {
	Metric string `json:"metric"`
}

type QueryRequest struct {
	Metrics     []string          `json:"metrics"`
	Start       int64             `json:"start"`
	End         int64             `json:"end"`
	Granularity Granularity       `json:"granularity"`
	Aggregation AggregationType   `json:"aggregation"`
}

type QueryResponse struct {
	Results []AggregationResult `json:"results"`
}

type ReportRequest struct {
	Points []DataPoint `json:"points"`
}

type ReportResponse struct {
	Success    int   `json:"success"`
	Outliers   int   `json:"outliers"`
	Timestamps []int64 `json:"timestamp,omitempty"`
}

type DeleteMetricRequest struct {
	Metric string `json:"metric"`
}

type ErrorCode string

const (
	ErrMetricExists      ErrorCode = "METRIC_EXISTS"
	ErrMetricNotFound    ErrorCode = "METRIC_NOT_FOUND"
	ErrInvalidGranularity ErrorCode = "INVALID_GRANULARITY"
	ErrInvalidAggregation ErrorCode = "INVALID_AGGREGATION"
	ErrInvalidTimestamp  ErrorCode = "INVALID_TIMESTAMP"
	ErrInvalidValue      ErrorCode = "INVALID_VALUE"
	ErrEmptyMetrics      ErrorCode = "EMPTY_METRICS"
	ErrInvalidTimeRange  ErrorCode = "INVALID_TIME_RANGE"
	ErrInternal          ErrorCode = "INTERNAL_ERROR"
)

type ErrorResponse struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

var errorMessages = map[ErrorCode]string{
	ErrMetricExists:      "指标名称已存在",
	ErrMetricNotFound:    "指标不存在",
	ErrInvalidGranularity: "无效的聚合粒度",
	ErrInvalidAggregation: "无效的聚合类型",
	ErrInvalidTimestamp:  "无效的时间戳",
	ErrInvalidValue:      "无效的数值",
	ErrEmptyMetrics:      "指标名称列表为空",
	ErrInvalidTimeRange:  "无效的时间范围",
	ErrInternal:          "内部错误",
}

func (e ErrorCode) Message() string {
	if msg, ok := errorMessages[e]; ok {
		return msg
	}
	return "未知错误"
}

func ValidGranularity(g Granularity) bool {
	switch g {
	case Granularity1Min, Granularity5Min, Granularity1Hour:
		return true
	}
	return false
}

func ValidAggregation(a AggregationType) bool {
	switch a {
	case AggregationAvg, AggregationMax, AggregationMin, AggregationSum:
		return true
	}
	return false
}

func ValidValue(v float64) bool {
	return v >= MinValue && v <= MaxValue
}

func TruncateTimestamp(ts int64) int64 {
	if ts > 10000000000 {
		return ts / 1000
	}
	return ts
}

func AlignToGranularity(ts int64, g Granularity) int64 {
	ts = TruncateTimestamp(ts)
	var interval int64
	switch g {
	case Granularity1Min:
		interval = 60
	case Granularity5Min:
		interval = 300
	case Granularity1Hour:
		interval = 3600
	default:
		return ts
	}
	return (ts / interval) * interval
}

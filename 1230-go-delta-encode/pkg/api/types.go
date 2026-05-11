package api

type DataType string

const (
	DataTypeInt   DataType = "int"
	DataTypeFloat DataType = "float"
)

type EncodeRequest struct {
	DataType DataType    `json:"data_type"`
	Data     interface{} `json:"data"`
}

type IntEncodeResponse struct {
	Base   int64   `json:"base"`
	Deltas []int64 `json:"deltas"`
}

type FloatEncodeResponse struct {
	Base   float64   `json:"base"`
	Deltas []float64 `json:"deltas"`
}

type DecodeRequest struct {
	DataType DataType    `json:"data_type"`
	Encoded  interface{} `json:"encoded"`
}

type StatsRequest struct {
	DataType DataType    `json:"data_type"`
	Data     interface{} `json:"data"`
}

type IntStatsResponse struct {
	AbsoluteSum int64   `json:"absolute_sum"`
	MaxAbsDelta int64   `json:"max_abs_delta"`
	ZeroRatio   float64 `json:"zero_ratio"`
	TotalCount  int     `json:"total_count"`
	ZeroCount   int     `json:"zero_count"`
}

type FloatStatsResponse struct {
	AbsoluteSum float64 `json:"absolute_sum"`
	MaxAbsDelta float64 `json:"max_abs_delta"`
	ZeroRatio   float64 `json:"zero_ratio"`
	TotalCount  int     `json:"total_count"`
	ZeroCount   int     `json:"zero_count"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

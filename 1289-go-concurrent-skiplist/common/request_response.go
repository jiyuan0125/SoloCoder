package common

type InsertRequest struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type InsertResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type GetRequest struct {
	Key string `json:"key"`
}

type GetResponse struct {
	Success bool        `json:"success"`
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"message,omitempty"`
}

type DeleteRequest struct {
	Key string `json:"key"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type RangeQueryRequest struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type RangeQueryResponse struct {
	Success bool        `json:"success"`
	Data    []KeyValue  `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type KeyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type SwitchStrategyRequest struct {
	Strategy string `json:"strategy"`
}

type SwitchStrategyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type GetStatsResponse struct {
	Success bool            `json:"success"`
	Stats   PerformanceStats `json:"stats,omitempty"`
	Message string          `json:"message,omitempty"`
}

type PerformanceStats struct {
	TotalOperations   int64   `json:"total_operations"`
	SuccessOperations int64   `json:"success_operations"`
	FailedOperations  int64   `json:"failed_operations"`
	ConflictCount     int64   `json:"conflict_count"`
	TotalWaitTime     int64   `json:"total_wait_time_ns"`
	AvgWaitTime       float64 `json:"avg_wait_time_ms"`
	InsertCount       int64   `json:"insert_count"`
	DeleteCount       int64   `json:"delete_count"`
	GetCount          int64   `json:"get_count"`
	RangeCount        int64   `json:"range_count"`
	CurrentStrategy   string  `json:"current_strategy"`
}

type StressTestRequest struct {
	Operations int `json:"operations"`
	Concurrency int `json:"concurrency"`
}

type StressTestResponse struct {
	Success       bool   `json:"success"`
	DurationMs    int64  `json:"duration_ms"`
	Operations    int    `json:"operations"`
	Message       string `json:"message,omitempty"`
	Throughput    float64 `json:"throughput_ops_sec"`
}

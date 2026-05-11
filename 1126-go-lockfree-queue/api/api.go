package api

type QueueRequest struct {
	ID      int64  `json:"id"`
	Payload string `json:"payload"`
}

type QueueResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type StatsResponse struct {
	Depth             int64 `json:"depth"`
	CASSuccess        int64 `json:"cas_success"`
	CASFailure        int64 `json:"cas_failure"`
	AllocsTotal       int64 `json:"allocs_total"`
	PoolHits          int64 `json:"pool_hits"`
	PoolMisses        int64 `json:"pool_misses"`
	ProcessedTotal    int64 `json:"processed_total"`
	EnqueueTotal      int64 `json:"enqueue_total"`
}

type VerifyRequest struct {
	Messages []string `json:"messages"`
	UseLock  bool     `json:"use_lock"`
}

type VerifyResponse struct {
	Success bool     `json:"success"`
	Expected []string `json:"expected"`
	Actual   []string `json:"actual"`
	Message  string   `json:"message,omitempty"`
}

type BenchResult struct {
	TotalMessages int64   `json:"total_messages"`
	Producers     int     `json:"producers"`
	DurationMs    int64   `json:"duration_ms"`
	Throughput    float64 `json:"throughput"`
	LatencyMin    float64 `json:"latency_min_ms"`
	LatencyMax    float64 `json:"latency_max_ms"`
	LatencyAvg    float64 `json:"latency_avg_ms"`
	LatencyP50    float64 `json:"latency_p50_ms"`
	LatencyP95    float64 `json:"latency_p95_ms"`
	LatencyP99    float64 `json:"latency_p99_ms"`
}

type BenchRequest struct {
	Producers  int `json:"producers"`
	MessagesPerProducer int `json:"messages_per_producer"`
}

type BenchResponse struct {
	Success bool        `json:"success"`
	Result  *BenchResult `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
}

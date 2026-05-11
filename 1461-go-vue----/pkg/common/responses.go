package common

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type BatchMetricsResponse struct {
	OneTimePassRate float64          `json:"one_time_pass_rate"`
	PeriodStart     string           `json:"period_start"`
	PeriodEnd       string           `json:"period_end"`
	TotalBatches    int              `json:"total_batches"`
	OneTimePassBatches int           `json:"one_time_pass_batches"`
}

type ProcessMetricsResponse struct {
	ProcessName string  `json:"process_name"`
	TotalCount  int     `json:"total_count"`
	FailCount   int     `json:"fail_count"`
	DefectRate  float64 `json:"defect_rate"`
}

type MetricsSummaryResponse struct {
	BatchMetrics   BatchMetricsResponse    `json:"batch_metrics"`
	ProcessMetrics []ProcessMetricsResponse `json:"process_metrics"`
}

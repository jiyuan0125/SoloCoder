package common

import "time"

type AuditLog struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	UserName        string                 `json:"user_name"`
	IPAddress       string                 `json:"ip_address"`
	OperationTime   time.Time              `json:"operation_time"`
	OperationType   string                 `json:"operation_type"`
	Description     string                 `json:"description"`
	Result          string                 `json:"result"`
	RecordIDs       []string               `json:"record_ids"`
	BeforeValue     map[string]interface{} `json:"before_value"`
	AfterValue      map[string]interface{} `json:"after_value"`
	Snapshot        map[string]interface{} `json:"snapshot"`
	ExportRange     *ExportRange           `json:"export_range"`
	Approver        string                 `json:"approver"`
	IsAbnormal      bool                   `json:"is_abnormal"`
	AbnormalReason  string                 `json:"abnormal_reason"`
	MonthPartition  string                 `json:"month_partition"`
}

type ExportRange struct {
	TableName    string `json:"table_name"`
	QueryFilter  string `json:"query_filter"`
	RecordCount  int    `json:"record_count"`
}

type CreateLogRequest struct {
	UserID         string                 `json:"user_id"`
	UserName       string                 `json:"user_name"`
	IPAddress      string                 `json:"ip_address"`
	OperationType  string                 `json:"operation_type"`
	Description    string                 `json:"description"`
	Result         string                 `json:"result"`
	RecordIDs      []string               `json:"record_ids"`
	BeforeValue    map[string]interface{} `json:"before_value"`
	AfterValue     map[string]interface{} `json:"after_value"`
	Snapshot       map[string]interface{} `json:"snapshot"`
	ExportRange    *ExportRange           `json:"export_range"`
	Approver       string                 `json:"approver"`
}

type QueryLogsRequest struct {
	UserID        string    `json:"user_id"`
	OperationType string    `json:"operation_type"`
	IPAddress     string    `json:"ip_address"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	IsAbnormal    bool      `json:"is_abnormal"`
	Page          int       `json:"page"`
	PageSize      int       `json:"page_size"`
}

type QueryLogsResponse struct {
	Logs      []*AuditLog `json:"logs"`
	Total     int         `json:"total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
}

type StatisticsRequest struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type OperationTrendItem struct {
	TimePoint   string `json:"time_point"`
	Count       int    `json:"count"`
	SuccessCount int   `json:"success_count"`
	FailedCount  int   `json:"failed_count"`
}

type OperationFrequencyTrend struct {
	Items []*OperationTrendItem `json:"items"`
}

type AbnormalBehaviorItem struct {
	UserID         string    `json:"user_id"`
	UserName       string    `json:"user_name"`
	OperationTime  time.Time `json:"operation_time"`
	Reason         string    `json:"reason"`
	LogID          string    `json:"log_id"`
}

type AbnormalBehaviorList struct {
	Items []*AbnormalBehaviorItem `json:"items"`
}

type SensitiveOperationItem struct {
	OperationType string `json:"operation_type"`
	Count         int    `json:"count"`
	FailedCount   int    `json:"failed_count"`
}

type SensitiveOperationRanking struct {
	Items []*SensitiveOperationItem `json:"items"`
}

type StatisticsResponse struct {
	OperationFrequencyTrend *OperationFrequencyTrend   `json:"operation_frequency_trend"`
	AbnormalBehaviorList    *AbnormalBehaviorList      `json:"abnormal_behavior_list"`
	SensitiveOperationRanking *SensitiveOperationRanking `json:"sensitive_operation_ranking"`
}

type LoginRequest struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	IPAddress string `json:"ip_address"`
	Success   bool   `json:"success"`
}

type LoginResponse struct {
	IsLocked     bool   `json:"is_locked"`
	LockUntil    string `json:"lock_until"`
	FailureCount int    `json:"failure_count"`
}

type ExportApprovalRequest struct {
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	ApproverID string `json:"approver_id"`
	Approval   bool   `json:"approval"`
}

type ExportApprovalResponse struct {
	Approved bool `json:"approved"`
}

type StorageStatusRequest struct{}

type StorageStatusResponse struct {
	TotalCapacity   int64   `json:"total_capacity"`
	UsedCapacity    int64   `json:"used_capacity"`
	UsagePercent    float64 `json:"usage_percent"`
	NeedAlert       bool    `json:"need_alert"`
	LogCount        int64   `json:"log_count"`
	ArchivedCount   int64   `json:"archived_count"`
}

type ArchiveInfo struct {
	MonthPartition string    `json:"month_partition"`
	LogCount       int       `json:"log_count"`
	CreatedAt      time.Time `json:"created_at"`
	IsArchived     bool      `json:"is_archived"`
}

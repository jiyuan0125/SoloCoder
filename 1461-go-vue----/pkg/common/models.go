package common

import "time"

type Batch struct {
	BatchID          string      `json:"batch_id"`
	ProductName      string      `json:"product_name"`
	FactoryCode      string      `json:"factory_code"`
	PlanQuantity     int         `json:"plan_quantity"`
	ActualQuantity   int         `json:"actual_quantity"`
	StartTime        time.Time   `json:"start_time"`
	EndTime          *time.Time  `json:"end_time,omitempty"`
	Status           BatchStatus `json:"status"`
	IsAbnormal       bool        `json:"is_abnormal"`
	IsOneTimePass    bool        `json:"is_one_time_pass"`
	ReworkCount      int         `json:"rework_count"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

type ProcessFlow struct {
	ProductName string      `json:"product_name"`
	Processes   []ProcessDef `json:"processes"`
}

type ProcessDef struct {
	Sequence    int    `json:"sequence"`
	ProcessName string `json:"process_name"`
}

type ProcessRecord struct {
	ID             int64         `json:"id"`
	BatchID        string        `json:"batch_id"`
	Sequence       int           `json:"sequence"`
	ProcessName    string        `json:"process_name"`
	Operator       string        `json:"operator"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        *time.Time    `json:"end_time,omitempty"`
	Result         ProcessResult `json:"result"`
	IsRework       bool          `json:"is_rework"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type InspectionSpec struct {
	ID            int64          `json:"id"`
	ProductName   string         `json:"product_name"`
	InspectionType InspectionType `json:"inspection_type"`
	MetricName    string         `json:"metric_name"`
	LowerLimit    float64        `json:"lower_limit"`
	UpperLimit    float64        `json:"upper_limit"`
}

type InspectionRecord struct {
	ID             int64           `json:"id"`
	BatchID        string          `json:"batch_id"`
	InspectionType InspectionType  `json:"inspection_type"`
	MetricName     string          `json:"metric_name"`
	ActualValue    float64         `json:"actual_value"`
	LowerLimit     float64         `json:"lower_limit"`
	UpperLimit     float64         `json:"upper_limit"`
	Result         InspectionResult `json:"result"`
	InspectedBy    string          `json:"inspected_by"`
	InspectedAt    time.Time       `json:"inspected_at"`
	CreatedAt      time.Time       `json:"created_at"`
}

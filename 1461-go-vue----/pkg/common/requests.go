package common

import "time"

type CreateBatchRequest struct {
	ProductName  string    `json:"product_name"`
	FactoryCode  string    `json:"factory_code"`
	PlanQuantity int       `json:"plan_quantity"`
	StartTime    time.Time `json:"start_time"`
}

type CompleteBatchRequest struct {
	BatchID        string    `json:"batch_id"`
	ActualQuantity int       `json:"actual_quantity"`
	EndTime        time.Time `json:"end_time"`
}

type StartProcessRequest struct {
	BatchID   string    `json:"batch_id"`
	Operator  string    `json:"operator"`
	StartTime time.Time `json:"start_time"`
}

type CompleteProcessRequest struct {
	BatchID   string        `json:"batch_id"`
	Result    ProcessResult `json:"result"`
	EndTime   time.Time     `json:"end_time"`
}

type ReworkDecisionRequest struct {
	BatchID  string `json:"batch_id"`
	Decision string `json:"decision"`
}

type AddInspectionRequest struct {
	BatchID        string         `json:"batch_id"`
	InspectionType InspectionType `json:"inspection_type"`
	MetricName     string         `json:"metric_name"`
	ActualValue    float64        `json:"actual_value"`
	InspectedBy    string         `json:"inspected_by"`
	InspectedAt    time.Time      `json:"inspected_at"`
}

type AddProcessFlowRequest struct {
	ProductName string       `json:"product_name"`
	Processes   []ProcessDef `json:"processes"`
}

type AddInspectionSpecRequest struct {
	ProductName   string         `json:"product_name"`
	InspectionType InspectionType `json:"inspection_type"`
	MetricName    string         `json:"metric_name"`
	LowerLimit    float64        `json:"lower_limit"`
	UpperLimit    float64        `json:"upper_limit"`
}

type ListBatchesRequest struct {
	Status     BatchStatus `json:"status,omitempty"`
	ProductName string      `json:"product_name,omitempty"`
	StartTime  *time.Time  `json:"start_time,omitempty"`
	EndTime    *time.Time  `json:"end_time,omitempty"`
}

type ListProcessRecordsRequest struct {
	BatchID string `json:"batch_id"`
}

type ListInspectionRecordsRequest struct {
	BatchID        string         `json:"batch_id,omitempty"`
	InspectionType InspectionType `json:"inspection_type,omitempty"`
}

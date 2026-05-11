package common

import "time"

type SampleStatus string

const (
	SampleStatusPending   SampleStatus = "待检测"
	SampleStatusTesting   SampleStatus = "检测中"
	SampleStatusCompleted SampleStatus = "已检毕"
	SampleStatusReturned  SampleStatus = "已退回"
	SampleStatusDestroyed SampleStatus = "已销毁"
)

type ReportStatus string

const (
	ReportStatusDraft   ReportStatus = "草稿"
	ReportStatusIssued  ReportStatus = "已签发"
	ReportStatusVoid    ReportStatus = "已作废"
)

type TestType string

const (
	TestTypeNumeric  TestType = "数值型"
	TestTypeJudgment TestType = "判定型"
)

type TestResultStatus string

const (
	TestResultPass    TestResultStatus = "合格"
	TestResultFail    TestResultStatus = "不合格"
	TestResultPending TestResultStatus = "待检测"
)

type RetestRequestStatus string

const (
	RetestStatusPending   RetestRequestStatus = "待审批"
	RetestStatusApproved  RetestRequestStatus = "已批准"
	RetestStatusRejected  RetestRequestStatus = "已拒绝"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateSampleRequest struct {
	Name        string    `json:"name"`
	Quantity    int       `json:"quantity"`
	Customer    string    `json:"customer"`
	DeliveryDate time.Time `json:"delivery_date"`
	Requirements string   `json:"requirements"`
	Volume      float64   `json:"volume"`
}

type CreateSampleResponse struct {
	SampleID     string `json:"sample_id"`
	IsLargeSample bool  `json:"is_large_sample"`
}

type ListSamplesRequest struct {
	Status SampleStatus `json:"status,omitempty"`
}

type AssignCabinetRequest struct {
	SampleID  string `json:"sample_id"`
	CabinetID string `json:"cabinet_id"`
}

type UpdateSampleStatusRequest struct {
	SampleID   string       `json:"sample_id"`
	NewStatus  SampleStatus `json:"new_status"`
	Operator   string       `json:"operator"`
	Remark     string       `json:"remark,omitempty"`
}

type CreateTestItemRequest struct {
	MethodID      string `json:"method_id"`
	Name          string `json:"name"`
	Steps         string `json:"steps"`
	JudgmentStd   string `json:"judgment_std"`
	Type          TestType `json:"type"`
	Unit          string `json:"unit,omitempty"`
	PassThreshold float64 `json:"pass_threshold,omitempty"`
	AllowRetest   bool   `json:"allow_retest"`
}

type AssignTestItemsRequest struct {
	SampleID  string   `json:"sample_id"`
	ItemIDs   []string `json:"item_ids"`
}

type ClaimSampleRequest struct {
	SampleID   string `json:"sample_id"`
	TesterName string `json:"tester_name"`
}

type RecordTestResultRequest struct {
	SampleID    string             `json:"sample_id"`
	ItemID      string             `json:"item_id"`
	IsRetest    bool               `json:"is_retest"`
	NumericValue float64           `json:"numeric_value,omitempty"`
	JudgmentResult TestResultStatus `json:"judgment_result,omitempty"`
	Unit        string             `json:"unit,omitempty"`
	Operator    string             `json:"operator"`
	Remark      string             `json:"remark,omitempty"`
}

type ApplyRetestRequest struct {
	SampleID  string `json:"sample_id"`
	ItemID    string `json:"item_id"`
	Applicant string `json:"applicant"`
	Reason    string `json:"reason"`
}

type ApproveRetestRequest struct {
	RequestID string `json:"request_id"`
	Approved  bool   `json:"approved"`
	Approver  string `json:"approver"`
	Remark    string `json:"remark,omitempty"`
}

type GenerateReportRequest struct {
	SampleID string `json:"sample_id"`
}

type IssueReportRequest struct {
	ReportID string `json:"report_id"`
	Issuer   string `json:"issuer"`
}

type VoidReportRequest struct {
	ReportID string `json:"report_id"`
	Operator string `json:"operator"`
	Reason   string `json:"reason"`
}

type CreateCabinetRequest struct {
	ID       string `json:"id"`
	Capacity int    `json:"capacity"`
	IsSpecial bool  `json:"is_special"`
}

type StatusLogEntry struct {
	Time     time.Time    `json:"time"`
	Status   SampleStatus `json:"status"`
	Operator string       `json:"operator"`
	Remark   string       `json:"remark,omitempty"`
}

type SampleInfo struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Quantity     int          `json:"quantity"`
	Customer     string       `json:"customer"`
	DeliveryDate time.Time    `json:"delivery_date"`
	Requirements string       `json:"requirements"`
	Volume       float64      `json:"volume"`
	IsLargeSample bool        `json:"is_large_sample"`
	Status       SampleStatus `json:"status"`
	CabinetID    string       `json:"cabinet_id,omitempty"`
	StatusLogs   []StatusLogEntry `json:"status_logs"`
}

type TestItemInfo struct {
	ID            string   `json:"id"`
	MethodID      string   `json:"method_id"`
	Name          string   `json:"name"`
	Steps         string   `json:"steps"`
	JudgmentStd   string   `json:"judgment_std"`
	Type          TestType `json:"type"`
	Unit          string   `json:"unit,omitempty"`
	PassThreshold float64  `json:"pass_threshold,omitempty"`
	AllowRetest   bool     `json:"allow_retest"`
}

type TestRecordInfo struct {
	ItemID      string             `json:"item_id"`
	ItemName    string             `json:"item_name"`
	Status      TestResultStatus   `json:"status"`
	NumericValue float64           `json:"numeric_value,omitempty"`
	Unit        string             `json:"unit,omitempty"`
	JudgmentResult TestResultStatus `json:"judgment_result,omitempty"`
	IsRetest    bool               `json:"is_retest"`
	Operator    string             `json:"operator,omitempty"`
	TestTime    time.Time          `json:"test_time,omitempty"`
}

type ReportInfo struct {
	ID         string         `json:"id"`
	SampleID   string         `json:"sample_id"`
	CreateTime time.Time      `json:"create_time"`
	Status     ReportStatus   `json:"status"`
	Conclusion TestResultStatus `json:"conclusion"`
	Issuer     string         `json:"issuer,omitempty"`
	IssueTime  time.Time      `json:"issue_time,omitempty"`
	Tests      []TestRecordInfo `json:"tests"`
}

type CabinetInfo struct {
	ID       string `json:"id"`
	Capacity int    `json:"capacity"`
	Used     int    `json:"used"`
	IsSpecial bool  `json:"is_special"`
}

type RetestRequestInfo struct {
	ID         string               `json:"id"`
	SampleID   string               `json:"sample_id"`
	ItemID     string               `json:"item_id"`
	Applicant  string               `json:"applicant"`
	Reason     string               `json:"reason"`
	Status     RetestRequestStatus  `json:"status"`
	ApplyTime  time.Time            `json:"apply_time"`
	Approver   string               `json:"approver,omitempty"`
	Remark     string               `json:"remark,omitempty"`
}

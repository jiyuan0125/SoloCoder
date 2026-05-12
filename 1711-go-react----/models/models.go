package models

import "time"

type Role string

const (
	RoleSubmitter    Role = "submitter"
	RoleTechnician   Role = "technician"
	RoleReviewer     Role = "reviewer"
	RoleAdmin        Role = "admin"
)

type SampleStatus string

const (
	SampleStatusReceived      SampleStatus = "received"
	SampleStatusTesting       SampleStatus = "testing"
	SampleStatusTestComplete  SampleStatus = "test_complete"
	SampleStatusReportGenerating SampleStatus = "report_generating"
	SampleStatusReported      SampleStatus = "reported"
)

type SampleType string

const (
	SampleTypeSaliva     SampleType = "saliva"
	SampleTypeBlood      SampleType = "blood"
	SampleTypeSwab       SampleType = "swab"
	SampleTypeTissue     SampleType = "tissue"
)

type ReportStatus string

const (
	ReportStatusDraft     ReportStatus = "draft"
	ReportStatusPending   ReportStatus = "pending"
	ReportStatusPublished ReportStatus = "published"
)

type ApprovalStatus string

const (
	ApprovalSubmitted ApprovalStatus = "submitted"
	ApprovalFirst     ApprovalStatus = "first_review"
	ApprovalSecond    ApprovalStatus = "second_review"
	ApprovalFinal     ApprovalStatus = "final_review"
	ApprovalApproved  ApprovalStatus = "approved"
	ApprovalRejected  ApprovalStatus = "rejected"
)

type TodoType string

const (
	TodoArrangeTest  TodoType = "arrange_test"
	TodoReviewReport TodoType = "review_report"
	TodoNotifyClient TodoType = "notify_client"
)

type Unit struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Contact string `json:"contact,omitempty"`
	Phone   string `json:"phone,omitempty"`
}

type Patient struct {
	Name        string `json:"name"`
	IDCard      string `json:"id_card"`
	Phone       string `json:"phone"`
	Gender      string `json:"gender,omitempty"`
	Age         int    `json:"age,omitempty"`
}

type TestItem struct {
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Price          float64 `json:"price"`
	ReportDays     int     `json:"report_days"`
	RawData        string  `json:"raw_data,omitempty"`
	ResultSummary  string  `json:"result_summary,omitempty"`
	RiskLevel      string  `json:"risk_level,omitempty"`
	RiskConclusion string  `json:"risk_conclusion,omitempty"`
	Completed      bool    `json:"completed"`
	CompletedAt    string  `json:"completed_at,omitempty"`
}

type Sample struct {
	ID              string       `json:"id"`
	SampleCode      string       `json:"sample_code"`
	SampleType      SampleType   `json:"sample_type"`
	ReceivedDate    string       `json:"received_date"`
	CollectionDate  string       `json:"collection_date"`
	UnitID          string       `json:"unit_id"`
	UnitName        string       `json:"unit_name,omitempty"`
	Submitter       string       `json:"submitter"`
	Status          SampleStatus `json:"status"`
	Patient         Patient      `json:"patient"`
	TestItems       []TestItem   `json:"test_items"`
	TotalPrice      float64      `json:"total_price"`
	Discount        float64      `json:"discount"`
	FinalPrice      float64      `json:"final_price"`
	Archived        bool         `json:"archived"`
	CreatedAt       time.Time    `json:"created_at"`
}

type Report struct {
	ID              string       `json:"id"`
	SampleID        string       `json:"sample_id"`
	SampleCode      string       `json:"sample_code"`
	Version         int          `json:"version"`
	Status          ReportStatus `json:"status"`
	ApprovalStatus  ApprovalStatus `json:"approval_status"`
	Content         ReportContent `json:"content"`
	ReviewComments  []ReviewComment `json:"review_comments,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	PublishedAt     *time.Time   `json:"published_at,omitempty"`
	IsSupplementary bool         `json:"is_supplementary"`
	ParentReportID  string       `json:"parent_report_id,omitempty"`
	Amount          float64      `json:"amount,omitempty"`
}

type ReportContent struct {
	SampleInfo     SampleSummary   `json:"sample_info"`
	TestItems      []TestItemSummary `json:"test_items"`
	RiskAssessment RiskAssessment  `json:"risk_assessment"`
	Recommendations string         `json:"recommendations"`
	GeneratedAt    string         `json:"generated_at"`
}

type SampleSummary struct {
	SampleCode      string `json:"sample_code"`
	SampleType      string `json:"sample_type"`
	ReceivedDate    string `json:"received_date"`
	CollectionDate  string `json:"collection_date"`
	UnitName        string `json:"unit_name"`
	Submitter       string `json:"submitter"`
	PatientName     string `json:"patient_name,omitempty"`
}

type TestItemSummary struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	ResultSummary string `json:"result_summary"`
	RiskLevel     string `json:"risk_level"`
	RawData       string `json:"raw_data,omitempty"`
}

type RiskAssessment struct {
	OverallLevel  string `json:"overall_level"`
	HighRisks     []string `json:"high_risks"`
	LowRisks      []string `json:"low_risks"`
}

type ReviewComment struct {
	Reviewer string `json:"reviewer"`
	Comment  string `json:"comment"`
	Stage    string `json:"stage"`
	Time     string `json:"time"`
}

type Todo struct {
	ID        string    `json:"id"`
	Type      TodoType  `json:"type"`
	Title     string    `json:"title"`
	SampleID  string    `json:"sample_id"`
	SampleCode string    `json:"sample_code"`
	Assignee  string    `json:"assignee"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

package core

import (
	"laboratory/common"
	"sync"
	"time"
)

type Sample struct {
	ID            string
	Name          string
	Quantity      int
	Customer      string
	DeliveryDate  time.Time
	Requirements  string
	Volume        float64
	IsLargeSample bool
	Status        common.SampleStatus
	CabinetID     string
	StatusLogs    []common.StatusLogEntry
	TesterName    string
}

type Cabinet struct {
	ID        string
	Capacity  int
	Used      int
	IsSpecial bool
	mu        sync.Mutex
}

type TestItem struct {
	ID            string
	MethodID      string
	Name          string
	Steps         string
	JudgmentStd   string
	Type          common.TestType
	Unit          string
	PassThreshold float64
	AllowRetest   bool
}

type TestRecord struct {
	ItemID         string
	ItemName       string
	Status         common.TestResultStatus
	NumericValue   float64
	Unit           string
	JudgmentResult common.TestResultStatus
	IsRetest       bool
	Operator       string
	TestTime       time.Time
}

type SampleTests struct {
	SampleID string
	ItemIDs  []string
	Records  map[string]*TestRecord
}

type Report struct {
	ID         string
	SampleID   string
	CreateTime time.Time
	Status     common.ReportStatus
	Conclusion common.TestResultStatus
	Issuer     string
	IssueTime  time.Time
	Tests      []common.TestRecordInfo
}

type RetestRequest struct {
	ID        string
	SampleID  string
	ItemID    string
	Applicant string
	Reason    string
	Status    common.RetestRequestStatus
	ApplyTime time.Time
	Approver  string
	Remark    string
}

type Store struct {
	mu          sync.RWMutex
	samples     map[string]*Sample
	cabinets    map[string]*Cabinet
	testItems   map[string]*TestItem
	sampleTests map[string]*SampleTests
	reports     map[string]*Report
	retestReqs  map[string]*RetestRequest

	sampleSeq   int
	reportSeq   int
	retestSeq   int
}

func NewStore() *Store {
	return &Store{
		samples:     make(map[string]*Sample),
		cabinets:    make(map[string]*Cabinet),
		testItems:   make(map[string]*TestItem),
		sampleTests: make(map[string]*SampleTests),
		reports:     make(map[string]*Report),
		retestReqs:  make(map[string]*RetestRequest),
		sampleSeq:   1,
		reportSeq:   1,
		retestSeq:   1,
	}
}

const LargeSampleVolume = 10.0

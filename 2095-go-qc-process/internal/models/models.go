package models

type QCStatus string

const (
	StatusPendingSampling  QCStatus = "待抽样"
	StatusSampling      QCStatus = "抽样中"
	StatusTesting      QCStatus = "检测中"
	StatusTestComplete QCStatus = "检测完成"
	StatusQualified     QCStatus = "合格"
	StatusUnqualified QCStatus = "不合格"
	StatusRecheck     QCStatus = "复检"
	StatusFinalCheck   QCStatus = "终检"
	StatusArchived    QCStatus = "已归档"
)

var StatusOrder = []QCStatus{
	StatusPendingSampling,
	StatusSampling,
	StatusTesting,
	StatusTestComplete,
	StatusQualified,
	StatusUnqualified,
	StatusRecheck,
	StatusFinalCheck,
	StatusArchived,
}

type Severity string

const (
	SeverityMinor   Severity = "轻微"
	SeverityGeneral Severity = "一般"
	SeverityMajor   Severity = "严重"
)

type InspectionItem struct {
	Name          string  `json:"name"`
	StandardValue  float64 `json:"standard_value"`
	ToleranceMin  float64 `json:"tolerance_min"`
	ToleranceMax  float64 `json:"tolerance_max"`
	ActualValue   *float64 `json:"actual_value,omitempty"`
	IsQualified   *bool   `json:"is_qualified,omitempty"`
}

type UnqualifiedRecord struct {
	ItemName  string   `json:"item_name"`
	Reason     string   `json:"reason"`
	Severity   Severity `json:"severity"`
}

type RecheckRecord struct {
	Attempt     bool    `json:"attempt"`
	Items       bool      `json:"items"`
}

type ReworkOrder struct {
	ID          string `json:"id"`
	Reason      string `json:"reason"`
	Completed   bool   `json:"completed"`
}

type QCReport struct {
	ID             string          `json:"id"`
	BatchNo        string        `json:"batch_no"`
	BatchSize      int           `json:"batch_size"`
	InspectionLevel string       `json:"inspection_level"`
	SampleSize      int           `json:"sample_size"`
	Status          QCStatus      `json:"status"`
	Items           []InspectionItem `json:"items"`
	UnqualifiedItems []UnqualifiedRecord `json:"unqualified_items,omitempty"`
	Recheck         *RecheckRecord `json:"recheck,omitempty"`
	ReworkOrder    *ReworkOrder `json:"rework_order,omitempty"`
	FinalResult     string        `json:"final_result,omitempty"`
	CreatedAt       string        `json:"created_at"`
	UpdatedAt       string        `json:"updated_at"`
}

func (i *InspectionItem) Evaluate() {
	if i.ActualValue == nil {
		return
	}
	actual := *i.ActualValue
	isQualified := actual >= i.ToleranceMin && actual <= i.ToleranceMax
	i.IsQualified = &isQualified
}

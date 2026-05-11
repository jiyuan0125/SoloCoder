package common

import (
	"encoding/json"
	"time"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(data interface{}) *Response {
	return &Response{Code: 200, Message: "success", Data: data}
}

func Error(code int, message string) *Response {
	return &Response{Code: code, Message: message}
}

func (r *Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data,omitempty"`
	}{
		Code:    r.Code,
		Message: r.Message,
		Data:    r.Data,
	})
}

func (r *Response) UnmarshalJSON(b []byte) error {
	aux := struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data,omitempty"`
	}{}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	r.Code = aux.Code
	r.Message = aux.Message
	r.Data = aux.Data
	return nil
}

type CreateProjectRequest struct {
	Name         string `json:"name"`
	Client       string `json:"client"`
	Contractor   string `json:"contractor"`
	ContractFee  int64  `json:"contract_fee"`
	PlanStart    string `json:"plan_start"`
	PlanEnd      string `json:"plan_end"`
	ProjectManager string `json:"project_manager"`
}

type UpdateProgressRequest struct {
	MilestoneName     string `json:"milestone_name"`
	Progress          int    `json:"progress"`
	ActualFinishDate  string `json:"actual_finish_date,omitempty"`
}

type AddExpenseRequest struct {
	Category    string `json:"category"`
	Amount      int64  `json:"amount"`
	Date        string `json:"date"`
	VoucherNo   string `json:"voucher_no"`
}

type AddQualityCheckRequest struct {
	CheckItem   string `json:"check_item"`
	Result      string `json:"result"`
	Inspector   string `json:"inspector"`
	CheckDate   string `json:"check_date"`
}

type RectificationRequest struct {
	CheckID   string `json:"check_id"`
	Result    string `json:"result"`
	Inspector string `json:"inspector"`
	Date      string `json:"date"`
}

type ProjectListResponse struct {
	Projects []ProjectSummary `json:"projects"`
}

type ProjectSummary struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	ProjectManager string  `json:"project_manager"`
	Progress       float64 `json:"progress"`
	Status         string  `json:"status"`
	IsRiskProject  bool    `json:"is_risk_project"`
}

type ProjectDetailResponse struct {
	*Project
	OverallProgress float64           `json:"overall_progress"`
	IsRiskProject   bool              `json:"is_risk_project"`
	BudgetSummary   []BudgetCategory  `json:"budget_summary"`
}

type BudgetCategory struct {
	Name       string `json:"name"`
	Budget     int64  `json:"budget"`
	Actual     int64  `json:"actual"`
	OverBudget bool   `json:"over_budget"`
}

type Project struct {
	ID             string
	Name           string
	Client         string
	Contractor     string
	ContractFee    int64
	PlanStart      time.Time
	PlanEnd        time.Time
	ProjectManager string
	Milestones     []Milestone
	Expenses       []Expense
	QualityChecks  []QualityCheck
	CreatedAt      time.Time
}

type Milestone struct {
	Name              string
	PlanFinishDate    time.Time
	Progress          int
	ActualFinishDate  *time.Time
	IsDelayed         bool
}

type Expense struct {
	ID        string
	Category  string
	Amount    int64
	Date      time.Time
	VoucherNo string
}

type QualityCheck struct {
	ID            string
	CheckItem     string
	Result        string
	IsPass        bool
	Inspector     string
	CheckDate     time.Time
	Rectification *Rectification
	Closed        bool
}

type Rectification struct {
	ID          string
	Result      string
	Passed      bool
	Inspector   string
	Date        time.Time
}

const (
	ExpenseCategoryLabor     = "人工费"
	ExpenseCategoryMaterial  = "材料费"
	ExpenseCategoryMachine   = "机械费"
	ExpenseCategoryAdmin     = "管理费"
)

const (
	BudgetRatioLabor    = 0.30
	BudgetRatioMaterial = 0.45
	BudgetRatioMachine  = 0.15
	BudgetRatioAdmin    = 0.10
)

var ExpenseCategories = []string{
	ExpenseCategoryLabor,
	ExpenseCategoryMaterial,
	ExpenseCategoryMachine,
	ExpenseCategoryAdmin,
}

var BudgetRatios = map[string]float64{
	ExpenseCategoryLabor:    BudgetRatioLabor,
	ExpenseCategoryMaterial: BudgetRatioMaterial,
	ExpenseCategoryMachine:  BudgetRatioMachine,
	ExpenseCategoryAdmin:    BudgetRatioAdmin,
}

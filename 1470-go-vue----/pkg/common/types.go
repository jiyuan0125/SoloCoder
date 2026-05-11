package common

import "time"

type ProjectStatus string

const (
	ProjectStatusDraft     ProjectStatus = "draft"
	ProjectStatusAuditing  ProjectStatus = "auditing"
	ProjectStatusLocked    ProjectStatus = "locked"
)

type VisaStatus string

const (
	VisaStatusPending   VisaStatus = "pending"
	VisaStatusConfirmed VisaStatus = "confirmed"
	VisaStatusRejected  VisaStatus = "rejected"
)

type AuditAction string

const (
	AuditActionCreate           AuditAction = "create"
	AuditActionUpdate           AuditAction = "update"
	AuditActionDelete           AuditAction = "delete"
	AuditActionVisaSubmit       AuditAction = "visa_submit"
	AuditActionVisaConfirm      AuditAction = "visa_confirm"
	AuditActionVisaReject       AuditAction = "visa_reject"
	AuditActionAuditPass        AuditAction = "audit_pass"
	AuditActionAuditComment     AuditAction = "audit_comment"
)

type Project struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	ContractAmount float64    `json:"contract_amount"`
	Status      ProjectStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type BillOfQuantity struct {
	ID             string  `json:"id"`
	ProjectID      string  `json:"project_id"`
	ItemCode       string  `json:"item_code"`
	ItemName       string  `json:"item_name"`
	Unit           string  `json:"unit"`
	ContractQty    float64 `json:"contract_qty"`
	ActualQty      float64 `json:"actual_qty"`
	DeviationRate  float64 `json:"deviation_rate"`
	IsAbnormal     bool    `json:"is_abnormal"`
	UnitPrice      float64 `json:"unit_price"`
	ContractAmount float64 `json:"contract_amount"`
	Adjustment     float64 `json:"adjustment"`
}

type Visa struct {
	ID               string     `json:"id"`
	ProjectID        string     `json:"project_id"`
	VisaNumber       string     `json:"visa_number"`
	Reason           string     `json:"reason"`
	ChangeContent    string     `json:"change_content"`
	IncreaseQty      float64    `json:"increase_qty"`
	DecreaseQty      float64    `json:"decrease_qty"`
	UnitPrice        float64    `json:"unit_price"`
	Amount           float64    `json:"amount"`
	Status           VisaStatus `json:"status"`
	SupervisorConfirm bool     `json:"supervisor_confirm"`
	OwnerConfirm     bool       `json:"owner_confirm"`
	Creator          string     `json:"creator"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Settlement struct {
	ProjectID            string  `json:"project_id"`
	ContractAmount       float64 `json:"contract_amount"`
	VisaIncreaseAmount   float64 `json:"visa_increase_amount"`
	VisaDecreaseAmount   float64 `json:"visa_decrease_amount"`
	QuantityAdjustment   float64 `json:"quantity_adjustment"`
	TotalAmount          float64 `json:"total_amount"`
	AuditOpinion         string  `json:"audit_opinion"`
	AuditedAt            *time.Time `json:"audited_at,omitempty"`
	Auditor              string  `json:"auditor,omitempty"`
}

type AuditLog struct {
	ID          string       `json:"id"`
	ProjectID   string       `json:"project_id"`
	Action      AuditAction  `json:"action"`
	Operator    string       `json:"operator"`
	Description string       `json:"description"`
	CreatedAt   time.Time    `json:"created_at"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateProjectRequest struct {
	Name           string  `json:"name"`
	ContractAmount float64 `json:"contract_amount"`
}

type CreateBOQRequest struct {
	ProjectID   string  `json:"project_id"`
	ItemCode    string  `json:"item_code"`
	ItemName    string  `json:"item_name"`
	Unit        string  `json:"unit"`
	ContractQty float64 `json:"contract_qty"`
	ActualQty   float64 `json:"actual_qty"`
	UnitPrice   float64 `json:"unit_price"`
}

type UpdateBOQRequest struct {
	ItemName    string  `json:"item_name,omitempty"`
	Unit        string  `json:"unit,omitempty"`
	ContractQty float64 `json:"contract_qty,omitempty"`
	ActualQty   float64 `json:"actual_qty,omitempty"`
	UnitPrice   float64 `json:"unit_price,omitempty"`
}

type CreateVisaRequest struct {
	ProjectID     string  `json:"project_id"`
	Reason        string  `json:"reason"`
	ChangeContent string  `json:"change_content"`
	IncreaseQty   float64 `json:"increase_qty"`
	DecreaseQty   float64 `json:"decrease_qty"`
	UnitPrice     float64 `json:"unit_price"`
	Creator       string  `json:"creator"`
}

type UpdateVisaRequest struct {
	Reason        string  `json:"reason,omitempty"`
	ChangeContent string  `json:"change_content,omitempty"`
	IncreaseQty   float64 `json:"increase_qty,omitempty"`
	DecreaseQty   float64 `json:"decrease_qty,omitempty"`
	UnitPrice     float64 `json:"unit_price,omitempty"`
}

type AuditRequest struct {
	ProjectID string `json:"project_id"`
	Auditor   string `json:"auditor"`
	Opinion   string `json:"opinion"`
}

type ConfirmVisaRequest struct {
	VisaID   string `json:"visa_id"`
	Operator string `json:"operator"`
	Role     string `json:"role"`
}

type RejectVisaRequest struct {
	VisaID   string `json:"visa_id"`
	Operator string `json:"operator"`
	Role     string `json:"role"`
	Reason   string `json:"reason"`
}

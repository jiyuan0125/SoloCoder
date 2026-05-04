package api

import "time"

type Plan struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	MonthlyFee          float64 `json:"monthly_fee"`
	SmsQuota            int     `json:"sms_quota"`
	StorageQuotaGB      float64 `json:"storage_quota_gb"`
	Description         string  `json:"description,omitempty"`
}

type PlanChange struct {
	ID           string    `json:"id"`
	CustomerID   string    `json:"customer_id"`
	FromPlanID   string    `json:"from_plan_id"`
	ToPlanID     string    `json:"to_plan_id"`
	EffectiveDate time.Time `json:"effective_date"`
	CreatedAt    time.Time `json:"created_at"`
}

type Customer struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	CurrentPlanID    string    `json:"current_plan_id"`
	RegisteredAt     time.Time `json:"registered_at"`
	IsActive         bool      `json:"is_active"`
}

type UsageRecord struct {
	ID              string    `json:"id"`
	CustomerID      string    `json:"customer_id"`
	Year            int       `json:"year"`
	Month           int       `json:"month"`
	SmsCount        int       `json:"sms_count"`
	StorageUsageGB  float64   `json:"storage_usage_gb"`
	LastUpdated     time.Time `json:"last_updated"`
}

type Bill struct {
	ID                    string    `json:"id"`
	CustomerID            string    `json:"customer_id"`
	Year                  int       `json:"year"`
	Month                 int       `json:"month"`
	PlanFee               float64   `json:"plan_fee"`
	SmsOverageFee         float64   `json:"sms_overage_fee"`
	StorageOverageFee     float64   `json:"storage_overage_fee"`
	TotalAmount           float64   `json:"total_amount"`
	Status                string    `json:"status"`
	GeneratedAt           time.Time `json:"generated_at"`
	DueDate               time.Time `json:"due_date"`
	PaymentRecord         *PaymentRecord `json:"payment_record,omitempty"`
	PlanBreakdown         []PlanProration `json:"plan_breakdown,omitempty"`
}

type PlanProration struct {
	PlanID    string    `json:"plan_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Days      int       `json:"days"`
	Fee       float64   `json:"fee"`
}

type PaymentRecord struct {
	ID          string    `json:"id"`
	BillID      string    `json:"bill_id"`
	Operator    string    `json:"operator"`
	MarkedAt    time.Time `json:"marked_at"`
	Remark      string    `json:"remark"`
	Amount      float64   `json:"amount"`
}

type PricingConfig struct {
	SmsPricePerUnit     float64 `json:"sms_price_per_unit"`
	StoragePricePerGB   float64 `json:"storage_price_per_gb"`
	PaymentDueDays      int     `json:"payment_due_days"`
	SeriousOverdueDays  int     `json:"serious_overdue_days"`
}

type CreateCustomerRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	PlanID    string `json:"plan_id"`
}

type CreateCustomerResponse struct {
	Customer *Customer `json:"customer"`
}

type GetCustomerResponse struct {
	Customer *Customer `json:"customer"`
}

type ListCustomersResponse struct {
	Customers []*Customer `json:"customers"`
}

type ChangePlanRequest struct {
	CustomerID    string    `json:"customer_id"`
	NewPlanID     string    `json:"new_plan_id"`
	EffectiveDate time.Time `json:"effective_date"`
}

type ChangePlanResponse struct {
	PlanChange *PlanChange `json:"plan_change"`
}

type RecordUsageRequest struct {
	CustomerID     string  `json:"customer_id"`
	SmsIncrement   int     `json:"sms_increment,omitempty"`
	StorageUpdate  float64 `json:"storage_update,omitempty"`
}

type RecordUsageResponse struct {
	Usage *UsageRecord `json:"usage"`
}

type GetUsageResponse struct {
	Usage *UsageRecord `json:"usage"`
}

type GenerateBillRequest struct {
	CustomerID string `json:"customer_id,omitempty"`
	Year       int    `json:"year"`
	Month      int    `json:"month"`
}

type GenerateBillResponse struct {
	Bill *Bill `json:"bill"`
}

type GetBillResponse struct {
	Bill *Bill `json:"bill"`
}

type ListCustomerBillsResponse struct {
	Bills []*Bill `json:"bills"`
}

type ListAllBillsResponse struct {
	Bills []*Bill `json:"bills"`
}

type MarkPaidRequest struct {
	BillID   string  `json:"bill_id"`
	Operator string  `json:"operator"`
	Amount   float64 `json:"amount"`
	Remark   string  `json:"remark,omitempty"`
}

type MarkPaidResponse struct {
	Bill *Bill `json:"bill"`
}

type GetPlanResponse struct {
	Plan *Plan `json:"plan"`
}

type ListPlansResponse struct {
	Plans []*Plan `json:"plans"`
}

type CreatePlanRequest struct {
	Name            string  `json:"name"`
	MonthlyFee      float64 `json:"monthly_fee"`
	SmsQuota        int     `json:"sms_quota"`
	StorageQuotaGB  float64 `json:"storage_quota_gb"`
	Description     string  `json:"description,omitempty"`
}

type CreatePlanResponse struct {
	Plan *Plan `json:"plan"`
}

type UpdatePricingRequest struct {
	SmsPricePerUnit    *float64 `json:"sms_price_per_unit,omitempty"`
	StoragePricePerGB  *float64 `json:"storage_price_per_gb,omitempty"`
	PaymentDueDays     *int     `json:"payment_due_days,omitempty"`
	SeriousOverdueDays *int     `json:"serious_overdue_days,omitempty"`
}

type GetPricingResponse struct {
	Pricing *PricingConfig `json:"pricing"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

const (
	BillStatusUnpaid       = "unpaid"
	BillStatusPaid         = "paid"
	BillStatusOverdue      = "overdue"
	BillStatusSeriousOverdue = "serious_overdue"
)

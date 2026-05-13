package models

import (
	"time"
)

const (
	OrderStatusPendingPayment = "pending_payment"
	OrderStatusPaid           = "paid"
	OrderStatusPreparing      = "preparing"
	OrderStatusShipped        = "shipped"
	OrderStatusSigned         = "signed"
	OrderStatusCompleted      = "completed"
	OrderStatusCancelled      = "cancelled"

	CancellationTypeDirect    = "direct"
	CancellationTypeRefund    = "refund_approval"
	CancellationTypeReturn    = "return"
	CancellationTypeAftersale = "aftersale"

	AfterSalesTypeRepair   = "repair"
	AfterSalesTypeExchange = "exchange"
	AfterSalesTypeRefund   = "refund"

	RepairStatusApplied   = "applied"
	RepairStatusChecking  = "checking"
	RepairStatusRepairing = "repairing"
	RepairStatusDone      = "done"

	ExchangeStatusApplied    = "applied"
	ExchangeStatusReviewing  = "reviewing"
	ExchangeStatusShipping   = "shipping"
	ExchangeStatusSigned     = "signed"

	RefundStatusApplied   = "applied"
	RefundStatusReviewing = "reviewing"
	RefundStatusRefunding = "refunding"
	RefundStatusDone      = "done"
)

var OrderStatusFlow = []string{
	OrderStatusPendingPayment,
	OrderStatusPaid,
	OrderStatusPreparing,
	OrderStatusShipped,
	OrderStatusSigned,
	OrderStatusCompleted,
}

var RepairStatusFlow = []string{
	RepairStatusApplied,
	RepairStatusChecking,
	RepairStatusRepairing,
	RepairStatusDone,
}

var ExchangeStatusFlow = []string{
	ExchangeStatusApplied,
	ExchangeStatusReviewing,
	ExchangeStatusShipping,
	ExchangeStatusSigned,
}

var RefundStatusFlow = []string{
	RefundStatusApplied,
	RefundStatusReviewing,
	RefundStatusRefunding,
	RefundStatusDone,
}

var OrderStatusNames = map[string]string{
	OrderStatusPendingPayment: "待支付",
	OrderStatusPaid:           "已支付",
	OrderStatusPreparing:      "备货中",
	OrderStatusShipped:        "已发货",
	OrderStatusSigned:         "已签收",
	OrderStatusCompleted:      "已完成",
	OrderStatusCancelled:      "已取消",
}

var RepairStatusNames = map[string]string{
	RepairStatusApplied:   "申请中",
	RepairStatusChecking:  "检测中",
	RepairStatusRepairing: "维修中",
	RepairStatusDone:      "已完成",
}

var ExchangeStatusNames = map[string]string{
	ExchangeStatusApplied:   "申请中",
	ExchangeStatusReviewing: "审核中",
	ExchangeStatusShipping:  "已发货",
	ExchangeStatusSigned:    "已签收",
}

var RefundStatusNames = map[string]string{
	RefundStatusApplied:   "申请中",
	RefundStatusReviewing: "审核中",
	RefundStatusRefunding: "退款中",
	RefundStatusDone:      "已退款",
}

func GetStatusName(status string) string {
	if name, ok := OrderStatusNames[status]; ok {
		return name
	}
	if name, ok := RepairStatusNames[status]; ok {
		return name
	}
	if name, ok := ExchangeStatusNames[status]; ok {
		return name
	}
	if name, ok := RefundStatusNames[status]; ok {
		return name
	}
	return status
}

type Order struct {
	ID              int       `json:"id"`
	OrderNo         string    `json:"order_no"`
	ResourceID      *int      `json:"resource_id,omitempty"`
	Amount          float64   `json:"amount"`
	CurrentStatus   string    `json:"current_status"`
	StatusName      string    `json:"status_name"`
	IsCancelled     bool      `json:"is_cancelled"`
	CancellationType *string  `json:"cancellation_type,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type OrderHistory struct {
	ID         int       `json:"id"`
	OrderID    int       `json:"order_id"`
	FromStatus string    `json:"from_status"`
	FromName   string    `json:"from_name"`
	ToStatus   string    `json:"to_status"`
	ToName     string    `json:"to_name"`
	Operator   string    `json:"operator"`
	Reason     string    `json:"reason"`
	ResourceID *int      `json:"resource_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type AfterSales struct {
	ID            int       `json:"id"`
	OrderID       int       `json:"order_id"`
	ResourceID    *int      `json:"resource_id,omitempty"`
	Type          string    `json:"type"`
	CurrentStatus string    `json:"current_status"`
	StatusName    string    `json:"status_name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AfterSalesHistory struct {
	ID           int       `json:"id"`
	AfterSalesID int       `json:"aftersales_id"`
	FromStatus   string    `json:"from_status"`
	FromName     string    `json:"from_name"`
	ToStatus     string    `json:"to_status"`
	ToName       string    `json:"to_name"`
	Operator     string    `json:"operator"`
	Reason       string    `json:"reason"`
	ResourceID   *int      `json:"resource_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Resource struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type ResourceSummary struct {
	ResourceID   int    `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	OrderCount   int    `json:"order_count"`
	TotalAmount  float64 `json:"total_amount"`
	OperationCount int  `json:"operation_count"`
}

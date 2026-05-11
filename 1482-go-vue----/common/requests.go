package common

type CreateOrderRequest struct {
	PlateNumber  string    `json:"plate_number"`
	CustomerName string    `json:"customer_name"`
	Phone        string    `json:"phone"`
	Description  string    `json:"description"`
	FaultType    FaultType `json:"fault_type"`
}

type AssignOrderRequest struct {
	OrderID int64 `json:"order_id"`
}

type CreateRepairItemRequest struct {
	OrderID         int64           `json:"order_id"`
	Name            string          `json:"name"`
	TechnicianLevel TechnicianLevel `json:"technician_level"`
	EstimatedHours  float64         `json:"estimated_hours"`
}

type CompleteRepairItemRequest struct {
	ItemID         int64   `json:"item_id"`
	ActualHours    float64 `json:"actual_hours"`
	OvertimeReason string  `json:"overtime_reason"`
}

type AddPartToItemRequest struct {
	ItemID   int64 `json:"item_id"`
	PartID   int64 `json:"part_id"`
	Quantity int64 `json:"quantity"`
}

type ReturnPartRequest struct {
	UsedPartID int64 `json:"used_part_id"`
	Quantity   int64 `json:"quantity"`
}

type CompleteOrderRequest struct {
	OrderID int64 `json:"order_id"`
}

type CancelOrderRequest struct {
	OrderID int64 `json:"order_id"`
}

type CreatePartRequest struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	Spec         string `json:"spec"`
	UnitPrice    int64  `json:"unit_price"`
	StockQty     int64  `json:"stock_qty"`
	WarningLevel int64  `json:"warning_level"`
}

type UpdatePartPriceRequest struct {
	PartID    int64 `json:"part_id"`
	UnitPrice int64 `json:"unit_price"`
}

type UpdatePartStockRequest struct {
	PartID    int64 `json:"part_id"`
	StockQty  int64 `json:"stock_qty"`
}

type ResolveReplenishmentRequest struct {
	TodoID int64 `json:"todo_id"`
}

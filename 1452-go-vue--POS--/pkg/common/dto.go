package common

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateProductReq struct {
	Name     string `json:"name"`
	Barcode  string `json:"barcode"`
	Price    int64  `json:"price"`
	StockQty int    `json:"stock_qty"`
}

type CreateProductResp struct {
	ID string `json:"id"`
}

type GetProductByBarcodeReq struct {
	Barcode string `json:"barcode"`
}

type GetProductResp struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Barcode  string `json:"barcode"`
	Price    int64  `json:"price"`
	StockQty int    `json:"stock_qty"`
}

type CreateMemberLevelReq struct {
	Name      string  `json:"name"`
	DiscountRate float64 `json:"discount_rate"`
}

type CreateMemberLevelResp struct {
	ID string `json:"id"`
}

type CreateMemberReq struct {
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	MemberLevelID  string `json:"member_level_id"`
}

type CreateMemberResp struct {
	ID string `json:"id"`
}

type AddOrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CreateOrderReq struct {
	CashierID string          `json:"cashier_id"`
	MemberID  string          `json:"member_id,omitempty"`
	Items     []AddOrderItem  `json:"items"`
	PaymentMethod string      `json:"payment_method"`
}

type CreateOrderResp struct {
	OrderID string `json:"order_id"`
	TotalAmount int64 `json:"total_amount"`
	DiscountAmount int64 `json:"discount_amount"`
	PayableAmount int64 `json:"payable_amount"`
	StockStatus string `json:"stock_status"`
}

type GetOrderReq struct {
	OrderID string `json:"order_id"`
}

type GetOrderResp struct {
	OrderID        string `json:"order_id"`
	CashierID      string `json:"cashier_id"`
	MemberID       string `json:"member_id,omitempty"`
	TotalAmount    int64  `json:"total_amount"`
	DiscountAmount int64  `json:"discount_amount"`
	PayableAmount  int64  `json:"payable_amount"`
	PaymentMethod  string `json:"payment_method"`
	PaymentTime    string `json:"payment_time"`
	StockStatus    string `json:"stock_status"`
	Items          []OrderItemResp `json:"items"`
}

type OrderItemResp struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Price       int64  `json:"price"`
	Quantity    int    `json:"quantity"`
	SubTotal    int64  `json:"sub_total"`
}

type DailyClosingReq struct {
	Date string `json:"date"`
}

type DailyClosingResp struct {
	ClosingID     string           `json:"closing_id"`
	Date          string           `json:"date"`
	TotalAmount   int64            `json:"total_amount"`
	PaymentStats  []PaymentStatResp `json:"payment_stats"`
	OrderCount    int              `json:"order_count"`
	ClosedAt      string           `json:"closed_at"`
}

type PaymentStatResp struct {
	PaymentMethod string `json:"payment_method"`
	TotalAmount   int64  `json:"total_amount"`
	OrderCount    int    `json:"order_count"`
}

type CreateStocktakingReq struct {
	OperatorID string `json:"operator_id"`
}

type CreateStocktakingResp struct {
	StocktakingID string `json:"stocktaking_id"`
	Status        string `json:"status"`
}

type GetStocktakingListReq struct {
}

type GetStocktakingListResp struct {
	Stocktakings []StocktakingSummaryResp `json:"stocktakings"`
}

type StocktakingSummaryResp struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	OperatorID string `json:"operator_id"`
	CreatedAt  string `json:"created_at"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type GetStocktakingDetailReq struct {
	StocktakingID string `json:"stocktaking_id"`
}

type GetStocktakingDetailResp struct {
	ID          string                     `json:"id"`
	Status      string                     `json:"status"`
	OperatorID  string                     `json:"operator_id"`
	CreatedAt   string                     `json:"created_at"`
	CompletedAt string                     `json:"completed_at,omitempty"`
	Items       []StocktakingItemResp      `json:"items"`
	PendingOps  []StocktakingPendingOpResp `json:"pending_ops,omitempty"`
}

type StocktakingItemResp struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Barcode     string `json:"barcode"`
	Price       int64  `json:"price"`
	SnapshotQty int    `json:"snapshot_qty"`
	ActualQty   int    `json:"actual_qty,omitempty"`
	DiffQty     int    `json:"diff_qty,omitempty"`
	DiffAmount  int64  `json:"diff_amount,omitempty"`
}

type StocktakingPendingOpResp struct {
	OperationType string `json:"operation_type"`
	ProductID     string `json:"product_id"`
	Quantity      int    `json:"quantity"`
	OperationTime string `json:"operation_time"`
}

type SubmitStocktakingReq struct {
	StocktakingID string                   `json:"stocktaking_id"`
	Items         []SubmitStocktakingItem  `json:"items"`
}

type SubmitStocktakingItem struct {
	ProductID string `json:"product_id"`
	ActualQty int    `json:"actual_qty"`
}

type CompleteStocktakingReq struct {
	StocktakingID string `json:"stocktaking_id"`
}

type CompleteStocktakingResp struct {
	StocktakingID string `json:"stocktaking_id"`
	Status        string `json:"status"`
}

type StockInReq struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	OperatorID string `json:"operator_id"`
}

type StockInResp struct {
	ID string `json:"id"`
}

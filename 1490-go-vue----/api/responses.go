package api

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateRepairOrderResponse struct {
	OrderID    string `json:"order_id"`
	MasterName string `json:"master_name,omitempty"`
	SiteName   string `json:"site_name,omitempty"`
	Status     string `json:"status"`
}

type GetOrderResponse struct {
	Order         *RepairOrder  `json:"order"`
	RepairRecord  *RepairRecord `json:"repair_record,omitempty"`
	MasterName    string        `json:"master_name,omitempty"`
	SiteName      string        `json:"site_name,omitempty"`
}

type ListOrdersResponse struct {
	Orders []*RepairOrder `json:"orders"`
}

type MasterInfoResponse struct {
	Master      *Master          `json:"master"`
	CurrentOrder *RepairOrder    `json:"current_order,omitempty"`
	TodayOrders  int             `json:"today_orders"`
}

type CostDetailResponse struct {
	LaborCost  float64 `json:"labor_cost"`
	PartCost   float64 `json:"part_cost,omitempty"`
	PartName   string  `json:"part_name,omitempty"`
	TotalCost  float64 `json:"total_cost"`
}

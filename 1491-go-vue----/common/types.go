package common

import "time"

type ApplianceCategory string

const (
	AirConditioner ApplianceCategory = "空调"
	Refrigerator   ApplianceCategory = "冰箱"
	WashingMachine ApplianceCategory = "洗衣机"
	TV             ApplianceCategory = "电视"
	WaterHeater    ApplianceCategory = "热水器"
	RangeHood      ApplianceCategory = "油烟机"
	GasStove       ApplianceCategory = "燃气灶"
)

var AllCategories = []ApplianceCategory{
	AirConditioner, Refrigerator, WashingMachine, TV,
	WaterHeater, RangeHood, GasStove,
}

var LaborFeeMap = map[ApplianceCategory]int{
	AirConditioner: 100,
	Refrigerator:   80,
	WashingMachine: 80,
	TV:             60,
	WaterHeater:    100,
	RangeHood:      60,
	GasStove:       60,
}

const ServiceVisitFee = 50

type RepairStatus string

const (
	StatusPending      RepairStatus = "待派单"
	StatusAssigned     RepairStatus = "已派单"
	StatusSpareApplied RepairStatus = "备件已申请"
	StatusRepairing    RepairStatus = "维修中"
	StatusCompleted    RepairStatus = "已完成"
)

type ApplianceInfo struct {
	Category ApplianceCategory `json:"category"`
	Brand    string            `json:"brand"`
	Model    string            `json:"model"`
}

type FaultDiagnosis struct {
	PossibleCauses []PossibleCause `json:"possible_causes"`
}

type PossibleCause struct {
	Cause           string `json:"cause"`
	MinEstimateFee  int    `json:"min_estimate_fee"`
	MaxEstimateFee  int    `json:"max_estimate_fee"`
}

type SparePart struct {
	Code          string            `json:"code"`
	Name          string            `json:"name"`
	ApplicableCat ApplianceCategory `json:"applicable_category"`
	ApplicableBrand string          `json:"applicable_brand"`
	UnitPrice     float64           `json:"unit_price"`
	StockQuantity int               `json:"stock_quantity"`
}

type SparePartOutboundRecord struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	OutboundAt  time.Time `json:"outbound_at"`
}

type RepairRequest struct {
	ID           string            `json:"id"`
	UserID       string            `json:"user_id"`
	Address      string            `json:"address"`
	Area         string            `json:"area"`
	Appliance    ApplianceInfo     `json:"appliance"`
	FaultDesc    string            `json:"fault_desc"`
	Diagnosis    *FaultDiagnosis   `json:"diagnosis,omitempty"`
	AssignedTechID string          `json:"assigned_tech_id,omitempty"`
	Status       RepairStatus      `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
}

type Technician struct {
	ID                string              `json:"id"`
	Name              string              `json:"name"`
	Specialties       []ApplianceCategory `json:"specialties"`
	ServiceAreas      []string            `json:"service_areas"`
	DailyOrderCount   int                 `json:"daily_order_count"`
	LastOrderTime     *time.Time          `json:"last_order_time,omitempty"`
}

type PurchaseOrder struct {
	ID         string    `json:"id"`
	SpareCode  string    `json:"spare_code"`
	Quantity   int       `json:"quantity"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type RepairReport struct {
	RequestID     string                    `json:"request_id"`
	TechnicianID  string                    `json:"technician_id"`
	FaultCause    string                    `json:"fault_cause"`
	SparePartsUsed []SparePartUsage         `json:"spare_parts_used"`
	FeeBreakdown  FeeBreakdown              `json:"fee_breakdown"`
	CompletedAt   time.Time                 `json:"completed_at"`
}

type SparePartUsage struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type FeeBreakdown struct {
	VisitFee      float64 `json:"visit_fee"`
	LaborFee      float64 `json:"labor_fee"`
	SparePartsFee float64 `json:"spare_parts_fee"`
	Total         float64 `json:"total"`
}

type CreateRepairRequest struct {
	UserID    string           `json:"user_id"`
	Address   string           `json:"address"`
	Area      string           `json:"area"`
	Appliance ApplianceInfo    `json:"appliance"`
	FaultDesc string           `json:"fault_desc"`
}

type CreateRepairResponse struct {
	Request   RepairRequest    `json:"request"`
	Diagnosis *FaultDiagnosis  `json:"diagnosis,omitempty"`
}

type AssignTechnicianRequest struct {
	RequestID string `json:"request_id"`
}

type AssignTechnicianResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	TechID  string `json:"tech_id,omitempty"`
}

type ApplySpareRequest struct {
	RequestID  string          `json:"request_id"`
	TechID     string          `json:"tech_id"`
	SpareParts []ApplySpareItem `json:"spare_parts"`
}

type ApplySpareItem struct {
	Code     string `json:"code"`
	Quantity int    `json:"quantity"`
}

type ApplySpareResponse struct {
	Success        bool                    `json:"success"`
	Message        string                  `json:"message,omitempty"`
	AppliedParts   []SparePartOutboundRecord `json:"applied_parts,omitempty"`
	FailedItems    []string                `json:"failed_items,omitempty"`
	PurchaseOrders []PurchaseOrder         `json:"purchase_orders,omitempty"`
}

type CompleteRepairRequest struct {
	RequestID    string          `json:"request_id"`
	TechID       string          `json:"tech_id"`
	FaultCause   string          `json:"fault_cause"`
	PartsUsed    []UsedPartItem  `json:"parts_used"`
	PartsReturn  []ReturnPartItem `json:"parts_return"`
}

type UsedPartItem struct {
	Code     string `json:"code"`
	Quantity int    `json:"quantity"`
}

type ReturnPartItem struct {
	Code     string `json:"code"`
	Quantity int    `json:"quantity"`
}

type CompleteRepairResponse struct {
	Success  bool          `json:"success"`
	Message  string        `json:"message,omitempty"`
	Report   *RepairReport `json:"report,omitempty"`
}

type AddSparePartRequest struct {
	Part SparePart `json:"part"`
}

type AddTechnicianRequest struct {
	Tech Technician `json:"tech"`
}

type UpdateSparePriceRequest struct {
	Code      string  `json:"code"`
	NewPrice  float64 `json:"new_price"`
}

type ListRepairRequestsResponse struct {
	Requests []RepairRequest `json:"requests"`
}

type GetSparePartsResponse struct {
	Parts []SparePart `json:"parts"`
}

type GetTechniciansResponse struct {
	Techs []Technician `json:"techs"`
}

type GetPurchaseOrdersResponse struct {
	Orders []PurchaseOrder `json:"orders"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

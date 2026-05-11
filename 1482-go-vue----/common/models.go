package common

import "time"

type FaultType string

const (
	FaultTypeMaintenance FaultType = "maintenance"
	FaultTypeEngine      FaultType = "engine"
	FaultTypeTransmission FaultType = "transmission"
	FaultTypeChassis     FaultType = "chassis"
	FaultTypeElectrical  FaultType = "electrical"
	FaultTypeAC          FaultType = "ac"
	FaultTypeBody        FaultType = "body"
)

func ValidFaultTypes() []FaultType {
	return []FaultType{
		FaultTypeMaintenance,
		FaultTypeEngine,
		FaultTypeTransmission,
		FaultTypeChassis,
		FaultTypeElectrical,
		FaultTypeAC,
		FaultTypeBody,
	}
}

func (ft FaultType) String() string {
	switch ft {
	case FaultTypeMaintenance:
		return "保养"
	case FaultTypeEngine:
		return "发动机"
	case FaultTypeTransmission:
		return "变速箱"
	case FaultTypeChassis:
		return "底盘"
	case FaultTypeElectrical:
		return "电气"
	case FaultTypeAC:
		return "空调"
	case FaultTypeBody:
		return "钣金喷漆"
	default:
		return "未知"
	}
}

type TechnicianLevel string

const (
	TechnicianLevelApprentice TechnicianLevel = "apprentice"
	TechnicianLevelIntermediate TechnicianLevel = "intermediate"
	TechnicianLevelSenior TechnicianLevel = "senior"
)

func ValidTechnicianLevels() []TechnicianLevel {
	return []TechnicianLevel{
		TechnicianLevelApprentice,
		TechnicianLevelIntermediate,
		TechnicianLevelSenior,
	}
}

func (tl TechnicianLevel) String() string {
	switch tl {
	case TechnicianLevelApprentice:
		return "学徒工"
	case TechnicianLevelIntermediate:
		return "中级工"
	case TechnicianLevelSenior:
		return "高级工"
	default:
		return "未知"
	}
}

func (tl TechnicianLevel) HourlyRate() float64 {
	switch tl {
	case TechnicianLevelApprentice:
		return 60.0
	case TechnicianLevelIntermediate:
		return 100.0
	case TechnicianLevelSenior:
		return 150.0
	default:
		return 0
	}
}

type OrderStatus string

const (
	OrderStatusCreated OrderStatus = "created"
	OrderStatusAssigned OrderStatus = "assigned"
	OrderStatusInProgress OrderStatus = "in_progress"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

func ValidOrderStatuses() []OrderStatus {
	return []OrderStatus{
		OrderStatusCreated,
		OrderStatusAssigned,
		OrderStatusInProgress,
		OrderStatusCompleted,
		OrderStatusCancelled,
	}
}

func (os OrderStatus) String() string {
	switch os {
	case OrderStatusCreated:
		return "已创建"
	case OrderStatusAssigned:
		return "已接单"
	case OrderStatusInProgress:
		return "进行中"
	case OrderStatusCompleted:
		return "已完成"
	case OrderStatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

type Part struct {
	ID           int64   `json:"id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Spec         string  `json:"spec"`
	UnitPrice    int64   `json:"unit_price"`
	StockQty     int64   `json:"stock_qty"`
	WarningLevel int64   `json:"warning_level"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UsedPart struct {
	ID        int64   `json:"id"`
	PartID    int64   `json:"part_id"`
	PartCode  string  `json:"part_code"`
	PartName  string  `json:"part_name"`
	Spec      string  `json:"spec"`
	UnitPrice int64   `json:"unit_price"`
	Quantity  int64   `json:"quantity"`
	Returned  bool    `json:"returned"`
}

type RepairItem struct {
	ID                int64            `json:"id"`
	OrderID           int64            `json:"order_id"`
	Name              string           `json:"name"`
	TechnicianLevel   TechnicianLevel  `json:"technician_level"`
	EstimatedHours    float64          `json:"estimated_hours"`
	ActualHours       float64          `json:"actual_hours"`
	LaborCost         int64            `json:"labor_cost"`
	Parts             []UsedPart       `json:"parts"`
	IsAbnormalHours   bool             `json:"is_abnormal_hours"`
	OvertimeReason    string           `json:"overtime_reason"`
	Completed         bool             `json:"completed"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

type Order struct {
	ID           int64       `json:"id"`
	PlateNumber  string      `json:"plate_number"`
	CustomerName string      `json:"customer_name"`
	Phone        string      `json:"phone"`
	Description  string      `json:"description"`
	FaultType    FaultType   `json:"fault_type"`
	Status       OrderStatus `json:"status"`
	Items        []RepairItem `json:"items"`
	LaborTotal   int64       `json:"labor_total"`
	PartsTotal   int64       `json:"parts_total"`
	GrandTotal   int64       `json:"grand_total"`
	HasAbnormal  bool        `json:"has_abnormal"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type ReplenishmentTodo struct {
	ID        int64     `json:"id"`
	PartID    int64     `json:"part_id"`
	PartCode  string    `json:"part_code"`
	PartName  string    `json:"part_name"`
	CurrentQty int64    `json:"current_qty"`
	WarningLevel int64  `json:"warning_level"`
	CreatedAt time.Time `json:"created_at"`
	Resolved  bool      `json:"resolved"`
}

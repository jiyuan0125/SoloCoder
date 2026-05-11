package api

import "time"

type BlockageType string

const (
	BlockageTypeKitchenSink   BlockageType = "厨房下水道"
	BlockageTypeBathroomSink  BlockageType = "卫生间下水道"
	BlockageTypeToilet        BlockageType = "马桶"
	BlockageTypeFloorDrain    BlockageType = "地漏"
	BlockageTypeMainPipe      BlockageType = "主管道"
)

type BlockageSeverity string

const (
	SeverityMild    BlockageSeverity = "轻微流水慢"
	SeveritySevere  BlockageSeverity = "严重完全堵塞"
)

type SkillTag string

const (
	SkillGeneral   SkillTag = "普通疏通"
	SkillHighPressure SkillTag = "高压清洗"
	SkillInspection SkillTag = "管道检测"
	SkillReplacement SkillTag = "管道更换"
)

type UncloggingMethod string

const (
	MethodManual      UncloggingMethod = "手工疏通"
	MethodMachine     UncloggingMethod = "管道疏通机"
	MethodHighPressure UncloggingMethod = "高压清洗"
)

const (
	LaborCostManual     = 80.0
	LaborCostMachine    = 120.0
	LaborCostHighPressure = 200.0
)

type RepairStatus string

const (
	StatusPending       RepairStatus = "待派单"
	StatusDispatched    RepairStatus = "已派单"
	StatusInProgress    RepairStatus = "处理中"
	StatusCompleted     RepairStatus = "待验收"
	StatusAccepted      RepairStatus = "已验收"
	StatusRejected      RepairStatus = "验收驳回"
	StatusComplaint     RepairStatus = "投诉工单"
)

type RepairOrder struct {
	ID               string
	Address          string
	Contact          string
	ContactPhone     string
	BlockageType     BlockageType
	BlockageSeverity BlockageSeverity
	IsRecurring      bool
	AreaCode         string
	Status           RepairStatus
	AssignedMasterID string
	SiteID           string
	DispatchCount    int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type RepairRecord struct {
	OrderID            string
	MasterID           string
	UncloggingMethod   UncloggingMethod
	DurationMinutes    int
	ReplacedParts      bool
	PartName           string
	PartCost           float64
	LaborCost          float64
	TotalCost          float64
	CompletedAt        time.Time
}

type Master struct {
	ID         string
	Name       string
	SiteID     string
	Skills     []SkillTag
	Status     MasterStatus
	DailyOrders int
}

type MasterStatus string

const (
	MasterStatusIdle    MasterStatus = "空闲"
	MasterStatusBusy    MasterStatus = "忙碌"
)

type RepairSite struct {
	ID           string
	Name         string
	AreaCodes    []string
	NeighborSiteIDs []string
	MasterIDs    []string
}

type AcceptanceRecord struct {
	OrderID     string
	Accepted    bool
	Comment     string
	CompletedAt time.Time
}

package common

type TicketCategory string

const (
	CategoryNetwork TicketCategory = "network"
	CategoryHardware TicketCategory = "hardware"
	CategorySoftware TicketCategory = "software"
	CategoryAccount  TicketCategory = "account"
)

type TicketPriority string

const (
	PriorityNormal   TicketPriority = "normal"
	PriorityUrgent   TicketPriority = "urgent"
	PriorityCritical TicketPriority = "critical"
)

type TicketStatus string

const (
	StatusNew        TicketStatus = "new"
	StatusAssigned   TicketStatus = "assigned"
	StatusInProgress TicketStatus = "in_progress"
	StatusResolved   TicketStatus = "resolved"
	StatusClosed     TicketStatus = "closed"
)

const (
	SLAHoursNormal   = 24
	SLAHoursUrgent   = 8
	SLAHoursCritical = 4
	SLAHoursEscalateToManager = 7 * 24
)

type OperationType string

const (
	OpCreate     OperationType = "create"
	OpAssign     OperationType = "assign"
	OpReassign   OperationType = "reassign"
	OpStatusChange OperationType = "status_change"
	OpAddComment OperationType = "add_comment"
	OpLink       OperationType = "link"
	OpUnlink     OperationType = "unlink"
	OpRate       OperationType = "rate"
	OpEscalate   OperationType = "escalate"
	OpTransfer   OperationType = "transfer"
)

var CategoryKeywords = map[TicketCategory][]string{
	CategoryNetwork:  {"网络", "wifi", "WiFi", "连接", "访问", "IP", "DNS", "ping", "网卡", "network", "connect"},
	CategoryHardware: {"硬件", "电脑", "显示器", "键盘", "鼠标", "打印机", "硬件故障", "hardware", "device"},
	CategorySoftware: {"软件", "安装", "升级", "bug", "Bug", "崩溃", "报错", "程序", "software", "install"},
	CategoryAccount:  {"账号", "密码", "登录", "权限", "注册", "注销", "account", "login", "password"},
}

package common

import "time"

const (
	MaxActiveItemsPerUser = 20
	MaxNegotiationRounds  = 3
	NegotiationTimeout    = 24 * time.Hour
	PaymentTimeout        = 48 * time.Hour
	ConfirmReceiptTimeout = 7 * 24 * time.Hour
	MinFeeThreshold       = 1000
	ServiceFeeRate        = 0.02
)

type Category string

const (
	CategoryElectronics Category = "数码电子"
	CategoryHome        Category = "家居用品"
	CategoryBooks       Category = "图书文具"
	CategoryClothing    Category = "服饰鞋包"
	CategorySports      Category = "运动户外"
	CategoryOther       Category = "其他"
)

var ValidCategories = []Category{
	CategoryElectronics,
	CategoryHome,
	CategoryBooks,
	CategoryClothing,
	CategorySports,
	CategoryOther,
}

type Condition string

const (
	ConditionNew        Condition = "全新未拆"
	ConditionLikeNew    Condition = "几乎全新"
	ConditionLightUsed  Condition = "轻微使用痕迹"
	ConditionHeavyUsed  Condition = "明显使用痕迹"
)

var ValidConditions = []Condition{
	ConditionNew,
	ConditionLikeNew,
	ConditionLightUsed,
	ConditionHeavyUsed,
}

type ItemStatus string

const (
	ItemStatusPending    ItemStatus = "待审核"
	ItemStatusOnSale     ItemStatus = "在售"
	ItemStatusNegotiating ItemStatus = "议价中"
	ItemStatusSold       ItemStatus = "已售出"
	ItemStatusRejected   ItemStatus = "审核未通过"
	ItemStatusRemoved    ItemStatus = "已下架"
)

type NegotiationStatus string

const (
	NegotiationStatusActive    NegotiationStatus = "进行中"
	NegotiationStatusAccepted  NegotiationStatus = "已接受"
	NegotiationStatusRejected  NegotiationStatus = "已拒绝"
	NegotiationStatusClosed    NegotiationStatus = "已关闭"
	NegotiationStatusTimeout   NegotiationStatus = "超时"
)

type NegotiationRole string

const (
	RoleBuyer  NegotiationRole = "买家"
	RoleSeller NegotiationRole = "卖家"
)

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "待付款"
	OrderStatusPaid           OrderStatus = "待发货"
	OrderStatusShipped        OrderStatus = "待收货"
	OrderStatusCompleted      OrderStatus = "已完成"
	OrderStatusCancelled      OrderStatus = "已取消"
)

type UserRole string

const (
	UserRoleAdmin UserRole = "管理员"
	UserRoleUser  UserRole = "普通用户"
)

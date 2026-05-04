package common

import (
	"errors"
	"time"
)

var (
	ErrEventNameEmpty     = errors.New("活动名称不能为空")
	ErrEventNameTooLong   = errors.New("活动名称不能超过100字")
	ErrLocationEmpty      = errors.New("活动地点不能为空")
	ErrInvalidTime        = errors.New("活动时间必须在未来")
	ErrEventEnded         = errors.New("活动已结束，无法购票")
	ErrInvalidTier        = errors.New("无效的票档次")
	ErrInsufficientStock  = errors.New("库存不足")
	ErrPurchaseLimit      = errors.New("一次购票最多限购10张")
	ErrTicketNotFound     = errors.New("票号不存在")
	ErrTicketAlreadyUsed  = errors.New("该票已签到")
	ErrTicketRefunded     = errors.New("该票已退票")
	ErrEventNotFound      = errors.New("活动不存在")
	ErrNoTiers            = errors.New("至少需要一个票档次")
	ErrInvalidQuantity    = errors.New("购票数量必须大于0")
	ErrTierNameEmpty      = errors.New("票档次名称不能为空")
	ErrInvalidPrice       = errors.New("票价必须大于0")
	ErrInvalidCapacity    = errors.New("可售票数必须大于0")
)

type Tier struct {
	Name       string `json:"name"`
	Price      int    `json:"price"`
	Capacity   int    `json:"capacity"`
	Available  int    `json:"available"`
}

type Event struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Time      time.Time `json:"time"`
	Location  string    `json:"location"`
	Tiers     []Tier    `json:"tiers"`
	CreatedAt time.Time `json:"created_at"`
}

type Ticket struct {
	TicketNumber string    `json:"ticket_number"`
	EventID      string    `json:"event_id"`
	TierName     string    `json:"tier_name"`
	Price        int       `json:"price"`
	PurchasedAt  time.Time `json:"purchased_at"`
	CheckedIn    bool      `json:"checked_in"`
	CheckedInAt  time.Time `json:"checked_in_at,omitempty"`
	Refunded     bool      `json:"refunded"`
	RefundedAt   time.Time `json:"refunded_at,omitempty"`
}

type TierStats struct {
	Name       string `json:"name"`
	Price      int    `json:"price"`
	Capacity   int    `json:"capacity"`
	Available  int    `json:"available"`
	Sold       int    `json:"sold"`
	CheckedIn  int    `json:"checked_in"`
}

type EventStats struct {
	EventID    string      `json:"event_id"`
	EventName  string      `json:"event_name"`
	TierStats  []TierStats `json:"tier_stats"`
}

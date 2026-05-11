package common

import "encoding/json"

type CreateTopicReq struct {
	Topic string `json:"topic"`
}

type DeleteTopicReq struct {
	Topic string `json:"topic"`
}

type SubscribeReq struct {
	Topic     string `json:"topic"`
	SubID     string `json:"sub_id"`
	Priority  int    `json:"priority"`
	Endpoint  string `json:"endpoint"`
}

type UnsubscribeReq struct {
	Topic string `json:"topic"`
	SubID string `json:"sub_id"`
}

type PublishReq struct {
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
	Async   bool            `json:"async"`
}

type ListSubscribersReq struct {
	Topic string `json:"topic"`
}

type EventStatusReq struct {
	EventID string `json:"event_id"`
}

type SubscriberInfo struct {
	ID        string `json:"id"`
	Priority  int    `json:"priority"`
	Endpoint  string `json:"endpoint"`
	IsActive  bool   `json:"is_active"`
	LastSeen  int64  `json:"last_seen"`
}

type EventResult struct {
	EventID     string `json:"event_id"`
	Status      string `json:"status"`
	Total       int    `json:"total"`
	Success     int    `json:"success"`
	Failed      int    `json:"failed"`
	Intercepted bool   `json:"intercepted"`
}

const (
	StatusPending    = "等待中"
	StatusProcessing = "处理中"
	StatusDone       = "已完成"
	StatusPartial    = "部分完成"
	StatusBlocked    = "已拦截"
)

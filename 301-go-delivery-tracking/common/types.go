package common

import (
	"encoding/json"
	"fmt"
)

type StatusName string

const (
	StatusPickupConfirmed StatusName = "取餐确认"
	StatusDepartedStore    StatusName = "商家出发"
	StatusArrivedCommunity StatusName = "到达小区"
	StatusDelivered        StatusName = "已送达"
)

var validStatuses = map[StatusName]bool{
	StatusPickupConfirmed: true,
	StatusDepartedStore:    true,
	StatusArrivedCommunity: true,
	StatusDelivered:        true,
}

func IsValidStatusName(name string) bool {
	return validStatuses[StatusName(name)]
}

type DeliveryStatus struct {
	OrderID     string     `json:"order_id"`
	StatusName  StatusName `json:"status_name"`
	Longitude   float64    `json:"longitude"`
	Latitude    float64    `json:"latitude"`
	Timestamp   int64      `json:"timestamp"`
	ReceiveTime int64      `json:"-"`
}

func (ds *DeliveryStatus) Validate() error {
	if ds.OrderID == "" {
		return fmt.Errorf("订单号不能为空")
	}
	if !IsValidStatusName(string(ds.StatusName)) {
		return fmt.Errorf("无效的状态名称: %s", ds.StatusName)
	}
	if ds.Longitude < -180 || ds.Longitude > 180 {
		return fmt.Errorf("经度范围无效: %f", ds.Longitude)
	}
	if ds.Latitude < -90 || ds.Latitude > 90 {
		return fmt.Errorf("纬度范围无效: %f", ds.Latitude)
	}
	return nil
}

type MessageType string

const (
	MsgTypeReport    MessageType = "report"
	MsgTypeQuery     MessageType = "query"
	MsgTypeResponse  MessageType = "response"
)

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ReportRequest struct {
	Statuses []DeliveryStatus `json:"statuses"`
}

type QueryRequest struct {
	OrderID string `json:"order_id"`
}

type Response struct {
	Success bool             `json:"success"`
	Error   string           `json:"error,omitempty"`
	Data    *QueryResultData `json:"data,omitempty"`
}

type QueryResultData struct {
	OrderID  string           `json:"order_id"`
	Statuses []DeliveryStatus `json:"statuses"`
}

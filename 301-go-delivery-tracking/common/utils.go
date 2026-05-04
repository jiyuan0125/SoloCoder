package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
)

func ReadMessage(conn net.Conn) (*Message, error) {
	var msg Message
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&msg); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, fmt.Errorf("解码消息失败: %w", err)
	}
	return &msg, nil
}

func WriteMessage(conn net.Conn, msg *Message) error {
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		return fmt.Errorf("编码消息失败: %w", err)
	}
	return nil
}

func BuildReportMessage(statuses []DeliveryStatus) (*Message, error) {
	req := ReportRequest{Statuses: statuses}
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化上报请求失败: %w", err)
	}
	return &Message{
		Type:    MsgTypeReport,
		Payload: payload,
	}, nil
}

func BuildQueryMessage(orderID string) (*Message, error) {
	req := QueryRequest{OrderID: orderID}
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化查询请求失败: %w", err)
	}
	return &Message{
		Type:    MsgTypeQuery,
		Payload: payload,
	}, nil
}

func BuildResponseMessage(success bool, errorMsg string, data *QueryResultData) (*Message, error) {
	resp := Response{
		Success: success,
		Error:   errorMsg,
		Data:    data,
	}
	payload, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("序列化响应失败: %w", err)
	}
	return &Message{
		Type:    MsgTypeResponse,
		Payload: payload,
	}, nil
}

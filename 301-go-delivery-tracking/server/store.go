package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"time"

	"go-delivery-tracking/common"
)

type DeliveryStore struct {
	mu       sync.RWMutex
	orders   map[string][]*common.DeliveryStatus
}

func NewDeliveryStore() *DeliveryStore {
	return &DeliveryStore{
		orders: make(map[string][]*common.DeliveryStatus),
	}
}

func (s *DeliveryStore) AddStatus(status *common.DeliveryStatus) error {
	if err := status.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	orderID := status.OrderID
	existing, exists := s.orders[orderID]

	status.ReceiveTime = time.Now().UnixNano()

	if exists {
		for _, existingStatus := range existing {
			if existingStatus.StatusName == status.StatusName && existingStatus.Timestamp == status.Timestamp {
				log.Printf("重复状态，跳过: order_id=%s, status=%s, timestamp=%d",
					orderID, status.StatusName, status.Timestamp)
				return nil
			}
		}
	}

	s.orders[orderID] = append(existing, status)
	sort.Slice(s.orders[orderID], func(i, j int) bool {
		a := s.orders[orderID][i]
		b := s.orders[orderID][j]
		if a.Timestamp != b.Timestamp {
			return a.Timestamp < b.Timestamp
		}
		return a.ReceiveTime < b.ReceiveTime
	})

	log.Printf("状态已添加: order_id=%s, status=%s, timestamp=%d",
		orderID, status.StatusName, status.Timestamp)
	return nil
}

func (s *DeliveryStore) GetOrderTrajectory(orderID string) (*common.QueryResultData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statuses, exists := s.orders[orderID]
	if !exists {
		return nil, fmt.Errorf("订单不存在: %s", orderID)
	}

	resultStatuses := make([]common.DeliveryStatus, len(statuses))
	for i, status := range statuses {
		resultStatuses[i] = *status
	}

	return &common.QueryResultData{
		OrderID:  orderID,
		Statuses: resultStatuses,
	}, nil
}

func (s *Server) handleReport(conn net.Conn, msg *common.Message) {
	var req common.ReportRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		log.Printf("解析上报请求失败: %v", err)
		respMsg, _ := common.BuildResponseMessage(false, fmt.Sprintf("解析请求失败: %v", err), nil)
		common.WriteMessage(conn, respMsg)
		return
	}

	if len(req.Statuses) == 0 {
		log.Printf("上报状态列表为空")
		respMsg, _ := common.BuildResponseMessage(false, "状态列表不能为空", nil)
		common.WriteMessage(conn, respMsg)
		return
	}

	var successCount int
	var errorMsg string
	for i := range req.Statuses {
		err := s.store.AddStatus(&req.Statuses[i])
		if err != nil {
			log.Printf("添加状态失败: %v", err)
			if errorMsg == "" {
				errorMsg = err.Error()
			}
		} else {
			successCount++
		}
	}

	var respMsg *common.Message
	if successCount > 0 {
		if errorMsg != "" {
			log.Printf("部分状态添加成功，部分失败: 成功%d个", successCount)
			respMsg, _ = common.BuildResponseMessage(false, fmt.Sprintf("部分添加成功，失败原因: %s", errorMsg), nil)
		} else {
			respMsg, _ = common.BuildResponseMessage(true, "", nil)
		}
	} else {
		respMsg, _ = common.BuildResponseMessage(false, errorMsg, nil)
	}

	common.WriteMessage(conn, respMsg)
}

func (s *Server) handleQuery(conn net.Conn, msg *common.Message) {
	var req common.QueryRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		log.Printf("解析查询请求失败: %v", err)
		respMsg, _ := common.BuildResponseMessage(false, fmt.Sprintf("解析请求失败: %v", err), nil)
		common.WriteMessage(conn, respMsg)
		return
	}

	if req.OrderID == "" {
		log.Printf("查询订单号为空")
		respMsg, _ := common.BuildResponseMessage(false, "订单号不能为空", nil)
		common.WriteMessage(conn, respMsg)
		return
	}

	data, err := s.store.GetOrderTrajectory(req.OrderID)
	if err != nil {
		log.Printf("查询订单轨迹失败: %v", err)
		respMsg, _ := common.BuildResponseMessage(false, err.Error(), nil)
		common.WriteMessage(conn, respMsg)
		return
	}

	log.Printf("查询订单轨迹成功: order_id=%s, status_count=%d", req.OrderID, len(data.Statuses))
	respMsg, _ := common.BuildResponseMessage(true, "", data)
	common.WriteMessage(conn, respMsg)
}

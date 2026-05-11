package core

import (
	"errors"
	"sync"
	"time"
	"taxisystem/common"

	"github.com/google/uuid"
)

type RatingManager struct {
	dm         *DispatchManager
	complaints map[string]*common.Complaint
	mu         sync.RWMutex
}

func NewRatingManager(dm *DispatchManager) *RatingManager {
	return &RatingManager{
		dm:         dm,
		complaints: make(map[string]*common.Complaint),
	}
}

func (rm *RatingManager) SubmitRating(orderID string, stars int, content string) (*common.Rating, error) {
	if stars < 1 || stars > 5 {
		return nil, errors.New("评分必须在1-5星之间")
	}

	order, err := rm.dm.GetOrder(orderID)
	if err != nil {
		return nil, err
	}

	if order.Status != common.OrderStatusCompleted {
		return nil, errors.New("只能对已完成的订单进行评价")
	}

	if order.Rating != nil {
		return nil, errors.New("该订单已评价过")
	}

	rating := &common.Rating{
		OrderID:      orderID,
		Stars:        stars,
		Content:      content,
		CreateTime:   time.Now().Unix(),
		HasComplaint: stars <= 2,
	}

	err = rm.dm.UpdateOrderRating(orderID, rating)
	if err != nil {
		return nil, err
	}

	if stars <= 2 {
		complaint := &common.Complaint{
			ComplaintID: uuid.New().String(),
			OrderID:     orderID,
			Content:     content,
			CreateTime:  time.Now().Unix(),
			Handled:     false,
		}

		rm.mu.Lock()
		rm.complaints[complaint.ComplaintID] = complaint
		rm.mu.Unlock()
	}

	return rating, nil
}

func (rm *RatingManager) GetPendingComplaints() []*common.Complaint {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	list := make([]*common.Complaint, 0)
	for _, c := range rm.complaints {
		if !c.Handled {
			complaint := *c
			list = append(list, &complaint)
		}
	}
	return list
}

func (rm *RatingManager) GetAllComplaints() []*common.Complaint {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	list := make([]*common.Complaint, 0, len(rm.complaints))
	for _, c := range rm.complaints {
		complaint := *c
		list = append(list, &complaint)
	}
	return list
}

func (rm *RatingManager) HandleComplaint(complaintID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	complaint, exists := rm.complaints[complaintID]
	if !exists {
		return errors.New("投诉不存在")
	}

	complaint.Handled = true
	return nil
}

func (rm *RatingManager) GetComplaint(complaintID string) (*common.Complaint, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	complaint, exists := rm.complaints[complaintID]
	if !exists {
		return nil, errors.New("投诉不存在")
	}

	c := *complaint
	return &c, nil
}

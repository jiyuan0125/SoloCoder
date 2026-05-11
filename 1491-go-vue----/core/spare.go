package core

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"repair-platform/common"
)

type ApplyResult struct {
	AppliedRecords []*common.SparePartOutboundRecord
	FailedCodes    []string
	PurchaseOrders []common.PurchaseOrder
}

func (s *Service) ApplySpareParts(reqID string, techID string, items []common.ApplySpareItem) (*ApplyResult, error) {
	req, ok := s.storage.GetRepairRequest(reqID)
	if !ok {
		return nil, errors.New("维修请求不存在")
	}
	if req.Status != common.StatusAssigned && req.Status != common.StatusSpareApplied {
		return nil, errors.New("该订单状态不允许申请备件")
	}
	if req.AssignedTechID != techID {
		return nil, errors.New("只有接单师傅可以申请备件")
	}

	sortedItems := make([]common.ApplySpareItem, len(items))
	copy(sortedItems, items)
	for i := 0; i < len(sortedItems)-1; i++ {
		for j := i + 1; j < len(sortedItems); j++ {
			if sortedItems[i].Code > sortedItems[j].Code {
				sortedItems[i], sortedItems[j] = sortedItems[j], sortedItems[i]
			}
		}
	}

	result := &ApplyResult{}

	for _, item := range sortedItems {
		applyRes := s.applySingleSpare(reqID, item)
		if applyRes.success {
			result.AppliedRecords = append(result.AppliedRecords, applyRes.record)
		} else {
			result.FailedCodes = append(result.FailedCodes, item.Code)
			if applyRes.po != nil {
				result.PurchaseOrders = append(result.PurchaseOrders, *applyRes.po)
			}
		}
	}

	if len(result.AppliedRecords) > 0 {
		s.storage.UpdateRepairRequest(reqID, func(r *common.RepairRequest) {
			r.Status = common.StatusSpareApplied
		})
	}

	return result, nil
}

type singleApplyResult struct {
	success bool
	record  *common.SparePartOutboundRecord
	po      *common.PurchaseOrder
}

func (s *Service) applySingleSpare(reqID string, item common.ApplySpareItem) singleApplyResult {
	unlock := s.storage.LockSparePart(item.Code)
	defer unlock()

	part, ok := s.storage.GetSparePart(item.Code)
	if !ok {
		return singleApplyResult{success: false, po: s.createPO(item.Code, item.Quantity)}
	}

	if part.StockQuantity < item.Quantity {
		deficit := item.Quantity - part.StockQuantity
		po := s.createPO(item.Code, deficit)
		if part.StockQuantity > 0 {
			price, _ := s.storage.DeductSpareStock(item.Code, part.StockQuantity)
			record := &common.SparePartOutboundRecord{
				Code:       item.Code,
				Name:       part.Name,
				Quantity:   part.StockQuantity,
				UnitPrice:  price,
				OutboundAt: time.Now(),
			}
			s.storage.AddOutboundRecord(reqID, record)
			return singleApplyResult{success: true, record: record, po: po}
		}
		return singleApplyResult{success: false, po: po}
	}

	price, err := s.storage.DeductSpareStock(item.Code, item.Quantity)
	if err != nil {
		return singleApplyResult{success: false}
	}

	record := &common.SparePartOutboundRecord{
		Code:       item.Code,
		Name:       part.Name,
		Quantity:   item.Quantity,
		UnitPrice:  price,
		OutboundAt: time.Now(),
	}
	s.storage.AddOutboundRecord(reqID, record)
	return singleApplyResult{success: true, record: record}
}

func (s *Service) createPO(code string, qty int) *common.PurchaseOrder {
	po := &common.PurchaseOrder{
		ID:        s.storage.NextOrderID(),
		SpareCode: code,
		Quantity:  qty,
		Status:    "待采购",
		CreatedAt: time.Now(),
	}
	s.storage.AddPurchaseOrder(*po)
	return po
}

func (s *Service) ReturnSpareParts(reqID string, returns []common.ReturnPartItem) error {
	for _, ret := range returns {
		if ret.Quantity <= 0 {
			continue
		}

		unlock := s.storage.LockSparePart(ret.Code)
		_ = s.storage.ReturnSpareStock(ret.Code, ret.Quantity)
		unlock()
	}
	return nil
}

func (s *Service) ConcurrentApplyTest(techIDs []string, code string, qtyPerTech int) []error {
	var wg sync.WaitGroup
	errs := make([]error, len(techIDs))
	req := common.RepairRequest{
		ID:             "test-concurrent",
		Status:         common.StatusAssigned,
		AssignedTechID: techIDs[0],
	}
	_ = s.storage.CreateRepairRequest(req)

	for i, techID := range techIDs {
		wg.Add(1)
		go func(idx int, tid string) {
			defer wg.Done()
			_, err := s.ApplySpareParts("test-concurrent", tid, []common.ApplySpareItem{{Code: code, Quantity: qtyPerTech}})
			errs[idx] = err
		}(i, techID)
	}
	wg.Wait()
	return errs
}

func (s *Service) AddSparePart(part common.SparePart) error {
	return s.storage.AddSparePart(part)
}

func (s *Service) UpdateSparePrice(code string, newPrice float64) error {
	return s.storage.UpdateSparePrice(code, newPrice)
}

func (s *Service) ListSpareParts() []common.SparePart {
	return s.storage.ListSpareParts()
}

func (s *Service) ListPurchaseOrders() []common.PurchaseOrder {
	return s.storage.ListPurchaseOrders()
}

func (s *Service) GetOutboundPrice(reqID, code string) (float64, error) {
	records := s.storage.GetOutboundRecords(reqID)
	for _, r := range records {
		if r.Code == code {
			return r.UnitPrice, nil
		}
	}
	part, ok := s.storage.GetSparePart(code)
	if !ok {
		return 0, fmt.Errorf("备件不存在: %s", code)
	}
	return part.UnitPrice, nil
}

package core

import (
	"errors"
	"piperepair/api"
	"time"
)

type RepairService struct {
	store      *Store
	dispatcher *Dispatcher
}

func NewRepairService(store *Store, dispatcher *Dispatcher) *RepairService {
	return &RepairService{
		store:      store,
		dispatcher: dispatcher,
	}
}

func (s *RepairService) SubmitRepairRecord(req *api.SubmitRepairRecordRequest) (*api.RepairRecord, error) {
	order, ok := s.store.GetOrder(req.OrderID)
	if !ok {
		return nil, errors.New("工单不存在")
	}

	if order.Status != api.StatusInProgress {
		return nil, errors.New("当前工单状态不允许提交维修记录")
	}

	if order.AssignedMasterID != req.MasterID {
		return nil, errors.New("当前师傅不是该工单的指派师傅")
	}

	laborCost := s.calculateLaborCost(req.UncloggingMethod)
	partCost := 0.0
	partName := ""
	if req.ReplacedParts {
		partCost = req.PartCost
		partName = req.PartName
	}
	totalCost := laborCost + partCost

	record := &api.RepairRecord{
		OrderID:          req.OrderID,
		MasterID:         req.MasterID,
		UncloggingMethod: req.UncloggingMethod,
		DurationMinutes:  req.DurationMinutes,
		ReplacedParts:    req.ReplacedParts,
		PartName:         partName,
		PartCost:         partCost,
		LaborCost:        laborCost,
		TotalCost:        totalCost,
		CompletedAt:      time.Now(),
	}

	s.store.SaveRepairRecord(record)

	order.Status = api.StatusCompleted
	s.store.SaveOrder(order)

	return record, nil
}

func (s *RepairService) calculateLaborCost(method api.UncloggingMethod) float64 {
	switch method {
	case api.MethodManual:
		return api.LaborCostManual
	case api.MethodMachine:
		return api.LaborCostMachine
	case api.MethodHighPressure:
		return api.LaborCostHighPressure
	default:
		return api.LaborCostManual
	}
}

func (s *RepairService) AcceptOrder(req *api.AcceptOrderRequest) error {
	order, ok := s.store.GetOrder(req.OrderID)
	if !ok {
		return errors.New("工单不存在")
	}

	if order.Status != api.StatusCompleted {
		return errors.New("当前工单状态不允许验收")
	}

	acceptanceRecord := &api.AcceptanceRecord{
		OrderID:     req.OrderID,
		Accepted:    req.Accepted,
		Comment:     req.Comment,
		CompletedAt: time.Now(),
	}

	s.store.SaveAcceptanceRecord(acceptanceRecord)

	if req.Accepted {
		order.Status = api.StatusAccepted
		s.store.SaveOrder(order)
		err := s.dispatcher.ReleaseMaster(order.AssignedMasterID)
		if err != nil {
			return err
		}
	} else {
		order.Status = api.StatusRejected
		s.store.SaveOrder(order)
		err := s.dispatcher.ReleaseMaster(order.AssignedMasterID)
		if err != nil {
			return err
		}

		if order.DispatchCount >= 3 {
			order.Status = api.StatusComplaint
			s.store.SaveOrder(order)
		} else {
			_, err := s.dispatcher.RedispatchOrder(order)
			if err != nil {
				order.Status = api.StatusComplaint
				s.store.SaveOrder(order)
			}
		}
	}

	return nil
}

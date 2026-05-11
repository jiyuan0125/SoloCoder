package core

import (
	"errors"
	"fmt"
	"piperepair/api"
	"time"
)

type PipeRepairService struct {
	store      *Store
	dispatcher *Dispatcher
	repairSvc  *RepairService
}

func NewPipeRepairService() *PipeRepairService {
	store := NewStore()
	store.InitSampleData()
	dispatcher := NewDispatcher(store)
	repairSvc := NewRepairService(store, dispatcher)

	return &PipeRepairService{
		store:      store,
		dispatcher: dispatcher,
		repairSvc:  repairSvc,
	}
}

func (s *PipeRepairService) CreateRepairOrder(req *api.CreateRepairOrderRequest) (*api.RepairOrder, *api.Master, error) {
	site, ok := s.store.GetSiteByAreaCode(req.AreaCode)
	if !ok {
		return nil, nil, errors.New("该区域暂无服务站点")
	}

	orderID := generateOrderID()
	now := time.Now()

	order := &api.RepairOrder{
		ID:               orderID,
		Address:          req.Address,
		Contact:          req.Contact,
		ContactPhone:     req.ContactPhone,
		BlockageType:     req.BlockageType,
		BlockageSeverity: req.BlockageSeverity,
		IsRecurring:      req.IsRecurring,
		AreaCode:         req.AreaCode,
		Status:           api.StatusPending,
		SiteID:           site.ID,
		DispatchCount:    0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	s.store.SaveOrder(order)

	master, err := s.dispatcher.DispatchOrder(order)
	if err != nil {
		return order, nil, err
	}

	return order, master, nil
}

func (s *PipeRepairService) GetOrder(orderID string) (*api.RepairOrder, *api.Master, *api.RepairSite, *api.RepairRecord, error) {
	order, ok := s.store.GetOrder(orderID)
	if !ok {
		return nil, nil, nil, nil, errors.New("工单不存在")
	}

	var master *api.Master
	if order.AssignedMasterID != "" {
		m, ok := s.store.GetMaster(order.AssignedMasterID)
		if ok {
			master = m
		}
	}

	var site *api.RepairSite
	if order.SiteID != "" {
		st, ok := s.store.GetSite(order.SiteID)
		if ok {
			site = st
		}
	}

	var repairRecord *api.RepairRecord
	record, ok := s.store.GetRepairRecord(orderID)
	if ok {
		repairRecord = record
	}

	return order, master, site, repairRecord, nil
}

func (s *PipeRepairService) ListAllOrders() []*api.RepairOrder {
	return s.store.ListAllOrders()
}

func (s *PipeRepairService) StartProcessing(orderID string, masterID string) error {
	order, ok := s.store.GetOrder(orderID)
	if !ok {
		return errors.New("工单不存在")
	}

	if order.Status != api.StatusDispatched {
		return errors.New("当前工单状态不允许开始处理")
	}

	if order.AssignedMasterID != masterID {
		return errors.New("当前师傅不是该工单的指派师傅")
	}

	order.Status = api.StatusInProgress
	s.store.SaveOrder(order)

	return nil
}

func (s *PipeRepairService) SubmitRepairRecord(req *api.SubmitRepairRecordRequest) (*api.RepairRecord, error) {
	return s.repairSvc.SubmitRepairRecord(req)
}

func (s *PipeRepairService) AcceptOrder(req *api.AcceptOrderRequest) error {
	return s.repairSvc.AcceptOrder(req)
}

func (s *PipeRepairService) GetMasterInfo(masterID string) (*api.Master, *api.RepairOrder, error) {
	master, ok := s.store.GetMaster(masterID)
	if !ok {
		return nil, nil, errors.New("师傅不存在")
	}

	var currentOrder *api.RepairOrder
	orderID, ok := s.store.GetMasterCurrentOrder(masterID)
	if ok {
		order, ok := s.store.GetOrder(orderID)
		if ok {
			currentOrder = order
		}
	}

	return master, currentOrder, nil
}

func generateOrderID() string {
	now := time.Now()
	return fmt.Sprintf("ORD%s%06d", now.Format("20060102150405"), now.Nanosecond()%1000000)
}

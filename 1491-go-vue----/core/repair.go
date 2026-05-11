package core

import (
	"errors"
	"time"

	"repair-platform/common"
)

type CompleteResult struct {
	Report *common.RepairReport
}

func (s *Service) CompleteRepair(
	reqID string,
	techID string,
	faultCause string,
	partsUsed []common.UsedPartItem,
	partsReturn []common.ReturnPartItem,
) (*CompleteResult, error) {

	req, ok := s.storage.GetRepairRequest(reqID)
	if !ok {
		return nil, errors.New("维修请求不存在")
	}
	if req.AssignedTechID != techID {
		return nil, errors.New("只有接单师傅可以完成维修")
	}
	if req.Status == common.StatusCompleted {
		return nil, errors.New("订单已完成")
	}

	if len(partsReturn) > 0 {
		if err := s.ReturnSpareParts(reqID, partsReturn); err != nil {
			return nil, err
		}
	}

	partsUsage := make([]common.SparePartUsage, 0, len(partsUsed))
	var sparePartsTotal float64

	for _, pu := range partsUsed {
		if pu.Quantity <= 0 {
			continue
		}
		price, err := s.GetOutboundPrice(reqID, pu.Code)
		if err != nil {
			continue
		}
		part, ok := s.storage.GetSparePart(pu.Code)
		name := pu.Code
		if ok {
			name = part.Name
		}
		usage := common.SparePartUsage{
			Code:      pu.Code,
			Name:      name,
			Quantity:  pu.Quantity,
			UnitPrice: price,
		}
		partsUsage = append(partsUsage, usage)
		sparePartsTotal += price * float64(pu.Quantity)
	}

	visitFee := float64(common.ServiceVisitFee)
	laborFee := float64(common.LaborFeeMap[req.Appliance.Category])
	total := visitFee + laborFee + sparePartsTotal

	report := &common.RepairReport{
		RequestID:      reqID,
		TechnicianID:   techID,
		FaultCause:     faultCause,
		SparePartsUsed: partsUsage,
		FeeBreakdown: common.FeeBreakdown{
			VisitFee:      visitFee,
			LaborFee:      laborFee,
			SparePartsFee: sparePartsTotal,
			Total:         total,
		},
		CompletedAt: time.Now(),
	}

	s.storage.AddRepairReport(*report)
	s.storage.UpdateRepairRequest(reqID, func(r *common.RepairRequest) {
		r.Status = common.StatusCompleted
	})

	return &CompleteResult{Report: report}, nil
}

func (s *Service) CreateRepairRequest(userID, address, area string, appliance common.ApplianceInfo, faultDesc string) (*common.RepairRequest, *common.FaultDiagnosis, error) {
	if userID == "" {
		return nil, nil, errors.New("用户ID不能为空")
	}
	if area == "" {
		return nil, nil, errors.New("区域不能为空")
	}
	if appliance.Category == "" {
		return nil, nil, errors.New("品类不能为空")
	}

	diagnosis := s.DiagnoseFault(appliance.Category, faultDesc)

	req := common.RepairRequest{
		ID:           s.storage.NextReqID(),
		UserID:       userID,
		Address:      address,
		Area:         area,
		Appliance:    appliance,
		FaultDesc:    faultDesc,
		Diagnosis:    diagnosis,
		Status:       common.StatusPending,
		CreatedAt:    time.Now(),
	}

	if err := s.storage.CreateRepairRequest(req); err != nil {
		return nil, nil, err
	}

	s.ProcessWaitingQueue()

	return &req, diagnosis, nil
}

func (s *Service) AddTechnician(tech common.Technician) error {
	if tech.ID == "" {
		return errors.New("师傅ID不能为空")
	}
	return s.storage.AddTechnician(tech)
}

func (s *Service) ListTechnicians() []common.Technician {
	return s.storage.ListTechnicians()
}

func (s *Service) ListRepairRequests() []common.RepairRequest {
	return s.storage.ListRepairRequests()
}

func (s *Service) GetRepairRequest(id string) (*common.RepairRequest, bool) {
	return s.storage.GetRepairRequest(id)
}

func (s *Service) GetRepairReport(reqID string) (*common.RepairReport, bool) {
	return s.storage.GetRepairReport(reqID)
}

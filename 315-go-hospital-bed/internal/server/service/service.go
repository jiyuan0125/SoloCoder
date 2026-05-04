package service

import (
	"fmt"
	"hospital-bed/internal/server/model"
	"hospital-bed/internal/server/store"
	"hospital-bed/shared/protocol"
	"sort"
	"strconv"
	"strings"
	"time"
)

type BedService struct {
	store *store.Store
}

func NewBedService(s *store.Store) *BedService {
	return &BedService{store: s}
}

func (s *BedService) ConfigureDepartment(req protocol.DepartmentConfigRequest) error {
	if err := store.ValidateDepartmentName(req.DepartmentName); err != nil {
		return err
	}

	for _, wardCfg := range req.Wards {
		if err := store.ValidateWardName(wardCfg.WardName); err != nil {
			return fmt.Errorf("病区[%s]错误: %w", wardCfg.WardName, err)
		}
		if wardCfg.SingleBeds < 0 || wardCfg.DoubleBeds < 0 || wardCfg.TripleBeds < 0 || wardCfg.ExtraBedLimit < 0 {
			return fmt.Errorf("病区[%s]配置错误: 床位数量不能为负数", wardCfg.WardName)
		}
	}

	s.store.Lock()
	defer s.store.Unlock()

	dept := s.store.State().Departments[req.DepartmentName]
	if dept == nil {
		dept = &model.Department{
			Name:  req.DepartmentName,
			Wards: make(map[string]*model.Ward),
			Queue: []*model.QueueEntry{},
		}
		s.store.State().Departments[req.DepartmentName] = dept
	}

	existingWards := make(map[string]bool)
	for name := range dept.Wards {
		existingWards[name] = true
	}

	for _, wardCfg := range req.Wards {
		delete(existingWards, wardCfg.WardName)
		ward := dept.Wards[wardCfg.WardName]
		if ward == nil {
			ward = &model.Ward{
				Name:           wardCfg.WardName,
				DepartmentName: req.DepartmentName,
				Beds:           make(map[string]*model.Bed),
			}
			dept.Wards[wardCfg.WardName] = ward
		}

		if err := s.resizeWardBeds(ward, wardCfg); err != nil {
			return err
		}
	}

	for wardName := range existingWards {
		ward := dept.Wards[wardName]
		hasOccupied := false
		for _, bed := range ward.Beds {
			if bed.IsOccupied {
				hasOccupied = true
				break
			}
		}
		if hasOccupied {
			return fmt.Errorf("无法删除病区[%s]: 存在正在使用的床位", wardName)
		}
		delete(dept.Wards, wardName)
	}

	return s.store.SaveLocked()
}

func (s *BedService) resizeWardBeds(ward *model.Ward, cfg protocol.WardConfig) error {
	ward.SingleCount = cfg.SingleBeds
	ward.DoubleCount = cfg.DoubleBeds
	ward.TripleCount = cfg.TripleBeds
	ward.ExtraBedLimit = cfg.ExtraBedLimit

	s.resizeBedType(ward, protocol.BedTypeSingle, cfg.SingleBeds)
	s.resizeBedType(ward, protocol.BedTypeDouble, cfg.DoubleBeds)
	s.resizeBedType(ward, protocol.BedTypeTriple, cfg.TripleBeds)

	return nil
}

func (s *BedService) resizeBedType(ward *model.Ward, bedType protocol.BedType, targetCount int) {
	var typePrefix string
	switch bedType {
	case protocol.BedTypeSingle:
		typePrefix = "单"
	case protocol.BedTypeDouble:
		typePrefix = "双"
	case protocol.BedTypeTriple:
		typePrefix = "三"
	}

	currentBeds := s.getBedsByType(ward, bedType)
	currentCount := len(currentBeds)

	if currentCount < targetCount {
		for i := currentCount + 1; i <= targetCount; i++ {
			bedID := fmt.Sprintf("%s-%s-%s%d", ward.DepartmentName, ward.Name, typePrefix, i)
			bed := &model.Bed{
				ID:             bedID,
				Type:           bedType,
				DepartmentName: ward.DepartmentName,
				WardName:       ward.Name,
				IsOccupied:     false,
				IsExtra:        false,
			}
			ward.Beds[bedID] = bed
		}
	} else if currentCount > targetCount {
		sort.Slice(currentBeds, func(i, j int) bool {
			return getBedNumber(currentBeds[i].ID) > getBedNumber(currentBeds[j].ID)
		})

		toRemove := currentCount - targetCount
		removed := 0
		for _, bed := range currentBeds {
			if removed >= toRemove {
				break
			}
			if !bed.IsOccupied {
				delete(ward.Beds, bed.ID)
				removed++
			}
		}
	}
}

func (s *BedService) getBedsByType(ward *model.Ward, bedType protocol.BedType) []*model.Bed {
	var beds []*model.Bed
	for _, bed := range ward.Beds {
		if bed.Type == bedType && !bed.IsExtra {
			beds = append(beds, bed)
		}
	}
	return beds
}

func getBedNumber(bedID string) int {
	parts := strings.Split(bedID, "-")
	if len(parts) == 0 {
		return 0
	}
	lastPart := parts[len(parts)-1]
	numStr := ""
	for _, c := range lastPart {
		if c >= '0' && c <= '9' {
			numStr += string(c)
		}
	}
	if numStr == "" {
		return 0
	}
	num, _ := strconv.Atoi(numStr)
	return num
}

func (s *BedService) AdmitPatient(req protocol.AdmissionRequest) (*protocol.AdmissionResponse, error) {
	if req.PatientID == "" {
		return nil, fmt.Errorf("患者ID不能为空")
	}
	if req.PatientName == "" {
		return nil, fmt.Errorf("患者姓名不能为空")
	}

	s.store.Lock()
	defer s.store.Unlock()

	if existingPatient := s.store.State().Patients[req.PatientID]; existingPatient != nil {
		return nil, fmt.Errorf("患者[%s]已在系统中", req.PatientID)
	}

	dept := s.store.State().Departments[req.DepartmentName]
	if dept == nil {
		return nil, fmt.Errorf("科室[%s]不存在", req.DepartmentName)
	}

	allocatedBed, downgradeReason, err := s.tryAllocateBed(dept, req)
	if err == nil && allocatedBed != nil {
		allocatedBed.IsOccupied = true
		allocatedBed.PatientID = &req.PatientID
		allocatedBed.PatientName = &req.PatientName

		patient := &model.PatientRecord{
			PatientID:      req.PatientID,
			PatientName:    req.PatientName,
			BedID:          &allocatedBed.ID,
			DepartmentName: &dept.Name,
			WardName:       &allocatedBed.WardName,
			BedType:        &allocatedBed.Type,
			InQueue:        false,
		}
		s.store.State().Patients[req.PatientID] = patient

		if err := s.store.SaveLocked(); err != nil {
			return nil, err
		}

		resp := &protocol.AdmissionResponse{
			Response: protocol.Response{Success: true},
			Result:   protocol.AdmissionResultAllocated,
			BedID:    allocatedBed.ID,
			BedType:  allocatedBed.Type,
		}
		if downgradeReason != nil {
			resp.DowngradeReason = downgradeReason
		}
		return resp, nil
	}

	entry := &model.QueueEntry{
		PatientID:     req.PatientID,
		PatientName:   req.PatientName,
		WardName:      req.WardName,
		PreferredType: req.PreferredType,
		QueueTime:     time.Now(),
	}
	dept.Queue = append(dept.Queue, entry)

	patient := &model.PatientRecord{
		PatientID:      req.PatientID,
		PatientName:    req.PatientName,
		DepartmentName: &dept.Name,
		InQueue:        true,
		QueuePosition:  len(dept.Queue),
	}
	s.store.State().Patients[req.PatientID] = patient

	if err := s.store.SaveLocked(); err != nil {
		return nil, err
	}

	return &protocol.AdmissionResponse{
		Response:      protocol.Response{Success: true},
		Result:        protocol.AdmissionResultQueued,
		QueuePosition: len(dept.Queue),
	}, nil
}

func (s *BedService) tryAllocateBed(dept *model.Department, req protocol.AdmissionRequest) (*model.Bed, *protocol.DowngradeReason, error) {
	preferredOrder := s.getDowngradeOrder(req.PreferredType)

	var targetWards []*model.Ward
	if req.WardName != nil {
		ward := dept.Wards[*req.WardName]
		if ward == nil {
			return nil, nil, fmt.Errorf("病区[%s]不存在", *req.WardName)
		}
		targetWards = []*model.Ward{ward}
	} else {
		for _, ward := range dept.Wards {
			targetWards = append(targetWards, ward)
		}
	}

	for i, bedType := range preferredOrder {
		for _, ward := range targetWards {
			bed := s.findAvailableBed(ward, bedType)
			if bed != nil {
				var downgradeReason *protocol.DowngradeReason
				if i > 0 {
					downgradeReason = &protocol.DowngradeReason{
						PreferredType: req.PreferredType,
						ActualType:    bedType,
						Reason:        fmt.Sprintf("偏好的%s无空余，已分配%s", typeLabel(req.PreferredType), typeLabel(bedType)),
					}
				}
				return bed, downgradeReason, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("无可用床位")
}

func (s *BedService) getDowngradeOrder(preferred protocol.BedType) []protocol.BedType {
	allTypes := []protocol.BedType{
		protocol.BedTypeSingle,
		protocol.BedTypeDouble,
		protocol.BedTypeTriple,
		protocol.BedTypeExtra,
	}

	var order []protocol.BedType
	foundPreferred := false
	for _, t := range allTypes {
		if t == preferred {
			foundPreferred = true
		}
		if foundPreferred {
			order = append(order, t)
		}
	}

	for _, t := range allTypes {
		if t == preferred {
			break
		}
		order = append(order, t)
	}

	return order
}

func (s *BedService) findAvailableBed(ward *model.Ward, bedType protocol.BedType) *model.Bed {
	if bedType == protocol.BedTypeExtra {
		if ward.ExtraBedCount >= ward.ExtraBedLimit {
			return nil
		}
		for _, bed := range ward.Beds {
			if bed.Type == protocol.BedTypeExtra && bed.IsExtra && !bed.IsOccupied {
				return bed
			}
		}
		newExtraNum := ward.ExtraBedCount + 1
		bedID := fmt.Sprintf("%s-%s-加-%d", ward.DepartmentName, ward.Name, newExtraNum)
		bed := &model.Bed{
			ID:             bedID,
			Type:           protocol.BedTypeExtra,
			DepartmentName: ward.DepartmentName,
			WardName:       ward.Name,
			IsOccupied:     false,
			IsExtra:        true,
		}
		ward.Beds[bedID] = bed
		ward.ExtraBedCount++
		return bed
	}

	for _, bed := range ward.Beds {
		if bed.Type == bedType && !bed.IsExtra && !bed.IsOccupied {
			return bed
		}
	}
	return nil
}

func typeLabel(t protocol.BedType) string {
	switch t {
	case protocol.BedTypeSingle:
		return "单人间"
	case protocol.BedTypeDouble:
		return "双人间"
	case protocol.BedTypeTriple:
		return "三人间"
	case protocol.BedTypeExtra:
		return "加床"
	default:
		return string(t)
	}
}

func (s *BedService) DischargePatient(req protocol.DischargeRequest) (*protocol.DischargeResponse, error) {
	if req.PatientID == "" {
		return nil, fmt.Errorf("患者ID不能为空")
	}

	s.store.Lock()
	defer s.store.Unlock()

	patient := s.store.State().Patients[req.PatientID]
	if patient == nil {
		return nil, fmt.Errorf("患者[%s]不存在", req.PatientID)
	}

	if patient.InQueue {
		dept := s.store.State().Departments[*patient.DepartmentName]
		if dept != nil {
			newQueue := []*model.QueueEntry{}
			for _, entry := range dept.Queue {
				if entry.PatientID != req.PatientID {
					newQueue = append(newQueue, entry)
				}
			}
			dept.Queue = newQueue
		}
		delete(s.store.State().Patients, req.PatientID)
		if err := s.store.SaveLocked(); err != nil {
			return nil, err
		}
		return &protocol.DischargeResponse{
			Response: protocol.Response{Success: true, Message: "患者已从候床队列移除"},
		}, nil
	}

	if patient.BedID == nil {
		delete(s.store.State().Patients, req.PatientID)
		if err := s.store.SaveLocked(); err != nil {
			return nil, err
		}
		return &protocol.DischargeResponse{
			Response: protocol.Response{Success: true},
		}, nil
	}

	dept := s.store.State().Departments[*patient.DepartmentName]
	if dept == nil {
		return nil, fmt.Errorf("患者所在科室不存在")
	}

	ward := dept.Wards[*patient.WardName]
	if ward == nil {
		return nil, fmt.Errorf("患者所在病区不存在")
	}

	bed := ward.Beds[*patient.BedID]
	if bed == nil {
		return nil, fmt.Errorf("患者床位不存在")
	}

	bed.IsOccupied = false
	bed.PatientID = nil
	bed.PatientName = nil
	freedBedID := bed.ID

	if bed.IsExtra {
		delete(ward.Beds, bed.ID)
		ward.ExtraBedCount--
	}

	delete(s.store.State().Patients, req.PatientID)

	if err := s.processQueue(dept); err != nil {
		return nil, err
	}

	if err := s.store.SaveLocked(); err != nil {
		return nil, err
	}

	return &protocol.DischargeResponse{
		Response: protocol.Response{Success: true},
		BedID:    freedBedID,
	}, nil
}

func (s *BedService) processQueue(dept *model.Department) error {
	queueCopy := make([]*model.QueueEntry, len(dept.Queue))
	copy(queueCopy, dept.Queue)

	newQueue := []*model.QueueEntry{}
	allocations := 0

	for _, entry := range queueCopy {
		req := protocol.AdmissionRequest{
			PatientID:      entry.PatientID,
			PatientName:    entry.PatientName,
			DepartmentName: dept.Name,
			WardName:       entry.WardName,
			PreferredType:  entry.PreferredType,
		}

		bed, _, err := s.tryAllocateBed(dept, req)
		if err == nil && bed != nil {
			bed.IsOccupied = true
			bed.PatientID = &entry.PatientID
			bed.PatientName = &entry.PatientName

			patient := s.store.State().Patients[entry.PatientID]
			if patient != nil {
				patient.BedID = &bed.ID
				patient.WardName = &bed.WardName
				patient.BedType = &bed.Type
				patient.InQueue = false
				patient.QueuePosition = 0
			}
			allocations++
		} else {
			newQueue = append(newQueue, entry)
		}
	}

	for i, entry := range newQueue {
		patient := s.store.State().Patients[entry.PatientID]
		if patient != nil {
			patient.QueuePosition = i + 1
		}
	}

	dept.Queue = newQueue
	return nil
}

func (s *BedService) GetDepartmentStatus(deptName string) (*protocol.DepartmentStatusResponse, error) {
	if deptName == "" {
		return nil, fmt.Errorf("科室名称不能为空")
	}

	s.store.RLock()
	defer s.store.RUnlock()

	dept := s.store.State().Departments[deptName]
	if dept == nil {
		return nil, fmt.Errorf("科室[%s]不存在", deptName)
	}

	var statuses []protocol.BedStatus
	for _, ward := range dept.Wards {
		typeMap := make(map[protocol.BedType]struct{ total, available int })

		for _, bed := range ward.Beds {
			stat := typeMap[bed.Type]
			stat.total++
			if !bed.IsOccupied {
				stat.available++
			}
			typeMap[bed.Type] = stat
		}

		extraTotal := 0
		extraAvailable := 0
		if ward.ExtraBedLimit > 0 {
			extraTotal = ward.ExtraBedLimit
			for _, bed := range ward.Beds {
				if bed.Type == protocol.BedTypeExtra && bed.IsExtra {
					if !bed.IsOccupied {
						extraAvailable++
					}
				}
			}
			unusedExtra := ward.ExtraBedLimit - ward.ExtraBedCount
			extraAvailable += unusedExtra
		}

		for _, bedType := range []protocol.BedType{
			protocol.BedTypeSingle,
			protocol.BedTypeDouble,
			protocol.BedTypeTriple,
			protocol.BedTypeExtra,
		} {
			if bedType == protocol.BedTypeExtra {
				if ward.ExtraBedLimit > 0 {
					statuses = append(statuses, protocol.BedStatus{
						WardName:  ward.Name,
						BedType:   bedType,
						Total:     extraTotal,
						Available: extraAvailable,
					})
				}
			} else {
				stat := typeMap[bedType]
				var expectedTotal int
				switch bedType {
				case protocol.BedTypeSingle:
					expectedTotal = ward.SingleCount
				case protocol.BedTypeDouble:
					expectedTotal = ward.DoubleCount
				case protocol.BedTypeTriple:
					expectedTotal = ward.TripleCount
				}
				if expectedTotal > 0 {
					statuses = append(statuses, protocol.BedStatus{
						WardName:  ward.Name,
						BedType:   bedType,
						Total:     expectedTotal,
						Available: stat.available,
					})
				}
			}
		}
	}

	return &protocol.DepartmentStatusResponse{
		Response:       protocol.Response{Success: true},
		DepartmentName: deptName,
		BedStatuses:    statuses,
		QueueLength:    len(dept.Queue),
	}, nil
}

func (s *BedService) GetPatientBed(req protocol.PatientBedRequest) (*protocol.PatientBedResponse, error) {
	if req.PatientID == "" {
		return nil, fmt.Errorf("患者ID不能为空")
	}

	s.store.RLock()
	defer s.store.RUnlock()

	patient := s.store.State().Patients[req.PatientID]
	if patient == nil {
		return nil, fmt.Errorf("患者[%s]不存在", req.PatientID)
	}

	resp := &protocol.PatientBedResponse{
		Response:    protocol.Response{Success: true},
		PatientID:   patient.PatientID,
		PatientName: patient.PatientName,
		InQueue:     patient.InQueue,
	}

	if patient.DepartmentName != nil {
		resp.DepartmentName = *patient.DepartmentName
	}
	if patient.WardName != nil {
		resp.WardName = *patient.WardName
	}
	if patient.BedID != nil {
		resp.BedID = *patient.BedID
	}
	if patient.BedType != nil {
		resp.BedType = *patient.BedType
	}
	if patient.InQueue {
		resp.QueuePosition = patient.QueuePosition
	}

	return resp, nil
}

func (s *BedService) GetQueueStatus(deptName string) (*protocol.QueueStatusResponse, error) {
	if deptName == "" {
		return nil, fmt.Errorf("科室名称不能为空")
	}

	s.store.RLock()
	defer s.store.RUnlock()

	dept := s.store.State().Departments[deptName]
	if dept == nil {
		return nil, fmt.Errorf("科室[%s]不存在", deptName)
	}

	var items []protocol.QueueItem
	for i, entry := range dept.Queue {
		items = append(items, protocol.QueueItem{
			PatientID:     entry.PatientID,
			PatientName:   entry.PatientName,
			PreferredType: entry.PreferredType,
			Position:      i + 1,
		})
	}

	return &protocol.QueueStatusResponse{
		Response:       protocol.Response{Success: true},
		DepartmentName: deptName,
		QueueLength:    len(dept.Queue),
		QueueItems:     items,
	}, nil
}

func (s *BedService) TransferPatient(patientID, targetDept string, targetWard *string, preferredType protocol.BedType) error {
	s.store.Lock()
	defer s.store.Unlock()

	patient := s.store.State().Patients[patientID]
	if patient == nil {
		return fmt.Errorf("患者[%s]不存在", patientID)
	}

	if patient.InQueue {
		return fmt.Errorf("患者正在候床队列中，请先取消排队")
	}

	if patient.BedID == nil || patient.DepartmentName == nil || patient.WardName == nil {
		return fmt.Errorf("患者状态异常")
	}

	targetDeptObj := s.store.State().Departments[targetDept]
	if targetDeptObj == nil {
		return fmt.Errorf("目标科室[%s]不存在", targetDept)
	}

	if targetWard != nil {
		if _, exists := targetDeptObj.Wards[*targetWard]; !exists {
			return fmt.Errorf("目标病区[%s]不存在", *targetWard)
		}
	}

	sourceDept := s.store.State().Departments[*patient.DepartmentName]
	if sourceDept == nil {
		return fmt.Errorf("原科室不存在")
	}

	sourceWard := sourceDept.Wards[*patient.WardName]
	if sourceWard == nil {
		return fmt.Errorf("原病区不存在")
	}

	bed := sourceWard.Beds[*patient.BedID]
	if bed == nil {
		return fmt.Errorf("原床位不存在")
	}

	bed.IsOccupied = false
	bed.PatientID = nil
	bed.PatientName = nil

	if bed.IsExtra {
		delete(sourceWard.Beds, bed.ID)
		sourceWard.ExtraBedCount--
	}

	patient.BedID = nil
	patient.WardName = nil
	patient.BedType = nil

	admitReq := protocol.AdmissionRequest{
		PatientID:      patientID,
		PatientName:    patient.PatientName,
		DepartmentName: targetDept,
		WardName:       targetWard,
		PreferredType:  preferredType,
	}

	allocatedBed, _, err := s.tryAllocateBed(targetDeptObj, admitReq)
	if err == nil && allocatedBed != nil {
		allocatedBed.IsOccupied = true
		allocatedBed.PatientID = &patientID
		allocatedBed.PatientName = &patient.PatientName

		patient.DepartmentName = &targetDept
		patient.WardName = &allocatedBed.WardName
		patient.BedID = &allocatedBed.ID
		patient.BedType = &allocatedBed.Type
		patient.InQueue = false

		if err := s.processQueue(sourceDept); err != nil {
			return err
		}

		return s.store.SaveLocked()
	}

	entry := &model.QueueEntry{
		PatientID:     patientID,
		PatientName:   patient.PatientName,
		WardName:      targetWard,
		PreferredType: preferredType,
		QueueTime:     time.Now(),
	}
	targetDeptObj.Queue = append(targetDeptObj.Queue, entry)

	patient.DepartmentName = &targetDept
	patient.InQueue = true
	patient.QueuePosition = len(targetDeptObj.Queue)

	if err := s.processQueue(sourceDept); err != nil {
		return err
	}

	return s.store.SaveLocked()
}

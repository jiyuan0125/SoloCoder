package service

import (
	"errors"
	"sort"
	"time"

	"ambulance-scheduler/internal/models"
	"ambulance-scheduler/internal/store"
	"ambulance-scheduler/pkg/distance"
)

type DispatchService struct {
	store          *store.Store
	vehicleService *VehicleService
}

func NewDispatchService(store *store.Store, vehicleService *VehicleService) *DispatchService {
	return &DispatchService{
		store:          store,
		vehicleService: vehicleService,
	}
}

type DispatchRequest struct {
	CallID       string `json:"call_id"`
	VehicleID    string `json:"vehicle_id,omitempty"`
}

type RecommendedVehicle struct {
	Vehicle *models.Ambulance `json:"vehicle"`
	Distance float64           `json:"distance_km"`
}

func (s *DispatchService) RecommendVehicle(callID string) ([]RecommendedVehicle, error) {
	call, ok := s.store.GetCall(callID)
	if !ok {
		return nil, errors.New("求救记录不存在")
	}

	var candidates []*models.Ambulance
	vehicles := s.store.ListVehicles()

	isUrgent := call.SeverityLevel == models.SeverityLevel1 || call.SeverityLevel == models.SeverityLevel2

	for _, v := range vehicles {
		if v.CurrentStatus == models.VehicleStatusIdle {
			candidates = append(candidates, v)
		}
	}

	if len(candidates) == 0 && isUrgent {
		for _, v := range vehicles {
			if v.CurrentStatus == models.VehicleStatusMaintenance {
				candidates = append(candidates, v)
			}
		}
	}

	var matching []*models.Ambulance
	for _, v := range candidates {
		if s.isVehicleMatch(v, call) {
			matching = append(matching, v)
		}
	}

	if len(matching) == 0 {
		return nil, errors.New("无合适车辆")
	}

	recommended := make([]RecommendedVehicle, 0, len(matching))
	for _, v := range matching {
		dist := distance.CalculateDistance(v.CurrentLocation, call.Location)
		recommended = append(recommended, RecommendedVehicle{
			Vehicle:  v,
			Distance: dist,
		})
	}

	sort.Slice(recommended, func(i, j int) bool {
		return recommended[i].Distance < recommended[j].Distance
	})

	return recommended, nil
}

func (s *DispatchService) isVehicleMatch(vehicle *models.Ambulance, call *models.EmergencyCall) bool {
	if call.PatientAgeGroup == models.AgeGroupChild {
		if call.ChiefComplaint == "新生儿" || call.ChiefComplaint == "早产" {
			return vehicle.VehicleType == models.VehicleTypeNeonate
		}
	}

	return true
}

func (s *DispatchService) DispatchVehicle(req *DispatchRequest) (*models.DispatchRecord, error) {
	call, ok := s.store.GetCall(req.CallID)
	if !ok {
		return nil, errors.New("求救记录不存在")
	}

	if call.InQueue {
		return nil, errors.New("该求救正在队列中等待")
	}

	var targetVehicle *models.Ambulance

	if req.VehicleID != "" {
		vehicle, ok := s.store.GetVehicle(req.VehicleID)
		if !ok {
			return nil, errors.New("指定的车辆不存在")
		}
		targetVehicle = vehicle
	} else {
		recommended, err := s.RecommendVehicle(req.CallID)
		if err != nil {
			return nil, err
		}
		if len(recommended) == 0 {
			return nil, errors.New("无合适车辆")
		}
		targetVehicle = recommended[0].Vehicle
	}

	if targetVehicle.CurrentStatus == models.VehicleStatusMaintenance {
		isUrgent := call.SeverityLevel == models.SeverityLevel1 || call.SeverityLevel == models.SeverityLevel2
		if !isUrgent {
			return nil, errors.New("维护中的车辆不能派车")
		}
	}

	if targetVehicle.PatientCount+call.PatientCount > 3 {
		return nil, errors.New("车辆已达最大承载人数")
	}

	now := time.Now()
	dist, estTime := distance.CalculateRoute(targetVehicle.CurrentLocation, call.Location)

	record := s.store.AtomicallyDispatch(req.CallID, targetVehicle.ID, func(c *models.EmergencyCall, v *models.Ambulance) *models.DispatchRecord {
		if v.PatientCount+c.PatientCount > 3 {
			return nil
		}

		if !models.ValidVehicleStatusTransition(v.CurrentStatus, models.VehicleStatusInTransit) && 
		   v.CurrentStatus != models.VehicleStatusMaintenance {
			return nil
		}

		r := models.NewDispatchRecord()
		r.CallID = c.ID
		r.VehicleID = v.ID
		r.DispatchTime = now
		r.TargetLocation = c.Location
		r.EstimatedArrivalTime = estTime
		r.EstimatedDistance = dist

		c.AssignedVehicleID = v.ID
		c.DispatchTime = &now
		c.Status = models.CallStatusProcessing
		c.UpdateTime = now

		v.CurrentStatus = models.VehicleStatusInTransit
		v.CurrentLocation = c.Location
		v.PatientCount += c.PatientCount
		v.CurrentStatusHistory = append(v.CurrentStatusHistory, models.StatusHistory{
			Status: models.VehicleStatusInTransit,
			Time:   now,
		})
		v.UpdateTime = now

		return r
	})

	if record == nil {
		return nil, errors.New("派车失败，条件不满足")
	}

	return record, nil
}

func (s *DispatchService) AddToQueue(callID string) error {
	call, ok := s.store.GetCall(callID)
	if !ok {
		return errors.New("求救记录不存在")
	}

	now := time.Now()
	call.InQueue = true
	call.QueueTime = &now
	call.UpdateTime = now
	s.store.UpdateCall(call)

	return nil
}

func (s *DispatchService) CheckTimeoutVehicles() []string {
	now := time.Now()
	timeoutVehicles := make([]string, 0)

	vehicles := s.store.ListVehicles()
	for _, v := range vehicles {
		if v.CurrentStatus == models.VehicleStatusInTransit {
			lastStatusTime := v.CurrentStatusHistory[len(v.CurrentStatusHistory)-1].Time
			if now.Sub(lastStatusTime) > 60*time.Minute {
				timeoutVehicles = append(timeoutVehicles, v.ID)
			}
		}
	}

	return timeoutVehicles
}

func (s *DispatchService) UpgradeWaitingCalls() {
	now := time.Now()
	calls := s.store.ListCalls()

	for _, call := range calls {
		if call.InQueue && call.SeverityLevel == models.SeverityLevel4 {
			if call.QueueTime != nil && now.Sub(*call.QueueTime) > 10*time.Minute {
				if call.LastUpgradeTime == nil || now.Sub(*call.LastUpgradeTime) >= 10*time.Minute {
					_ = s.store.AtomicallyUpgradeSeverity(call.ID, func(c *models.EmergencyCall) bool {
						c.SeverityLevel = models.SeverityLevel3
						c.LastUpgradeTime = &now
						c.UpdateTime = now
						return true
					})
				}
			}
		}
	}
}

func (s *DispatchService) ListDispatchRecords() []*models.DispatchRecord {
	return s.store.ListDispatchRecords()
}

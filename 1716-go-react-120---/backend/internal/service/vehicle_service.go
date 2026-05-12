package service

import (
	"errors"
	"time"

	"ambulance-scheduler/internal/models"
	"ambulance-scheduler/internal/store"
)

type VehicleService struct {
	store *store.Store
}

func NewVehicleService(store *store.Store) *VehicleService {
	return &VehicleService{store: store}
}

type CreateVehicleRequest struct {
	VehicleNumber   string             `json:"vehicle_number"`
	VehicleType     models.VehicleType `json:"vehicle_type"`
	CurrentLocation string             `json:"current_location"`
	DoctorCount     int                `json:"doctor_count"`
	NurseCount      int                `json:"nurse_count"`
}

func (s *VehicleService) CreateVehicle(req *CreateVehicleRequest) (*models.Ambulance, error) {
	vehicle := models.NewAmbulance()
	vehicle.VehicleNumber = req.VehicleNumber
	vehicle.VehicleType = req.VehicleType
	vehicle.CurrentLocation = req.CurrentLocation
	vehicle.DoctorCount = req.DoctorCount
	vehicle.NurseCount = req.NurseCount

	s.store.CreateVehicle(vehicle)
	return vehicle, nil
}

func (s *VehicleService) GetVehicle(id string) (*models.Ambulance, bool) {
	return s.store.GetVehicle(id)
}

func (s *VehicleService) ListVehicles() []*models.Ambulance {
	return s.store.ListVehicles()
}

func (s *VehicleService) UpdateStatus(vehicleID string, newStatus models.VehicleStatus) error {
	_, ok := s.store.GetVehicle(vehicleID)
	if !ok {
		return errors.New("车辆不存在")
	}

	success := s.store.AtomicallyUpdateStatus(vehicleID, newStatus, func(v *models.Ambulance) bool {
		if !models.ValidVehicleStatusTransition(v.CurrentStatus, newStatus) {
			return false
		}

		v.CurrentStatus = newStatus
		v.CurrentStatusHistory = append(v.CurrentStatusHistory, models.StatusHistory{
			Status: newStatus,
			Time:   time.Now(),
		})
		v.UpdateTime = time.Now()

		return true
	})

	if !success {
		return errors.New("无法更新车辆状态，状态流转不合法")
	}

	return nil
}

func (s *VehicleService) SetMaintenance(vehicleID string) error {
	vehicle, ok := s.store.GetVehicle(vehicleID)
	if !ok {
		return errors.New("车辆不存在")
	}

	if vehicle.CurrentStatus != models.VehicleStatusIdle {
		return errors.New("只有空闲车辆才能进入维护状态")
	}

	success := s.store.AtomicallyUpdateStatus(vehicleID, models.VehicleStatusMaintenance, func(v *models.Ambulance) bool {
		v.CurrentStatus = models.VehicleStatusMaintenance
		v.CurrentStatusHistory = append(v.CurrentStatusHistory, models.StatusHistory{
			Status: models.VehicleStatusMaintenance,
			Time:   time.Now(),
		})
		v.UpdateTime = time.Now()
		return true
	})

	if !success {
		return errors.New("无法设置维护状态")
	}

	return nil
}

package core

import (
	"errors"
	"sync"
	"time"

	"vehicle-inspection/common"
)

type VehicleManager struct {
	vehicles map[string]*common.Vehicle
	mu       sync.RWMutex
}

func NewVehicleManager() *VehicleManager {
	return &VehicleManager{
		vehicles: make(map[string]*common.Vehicle),
	}
}

func (vm *VehicleManager) RegisterVehicle(req *common.RegisterVehicleRequest) (*common.Vehicle, error) {
	if err := ValidatePlateNumber(req.PlateNumber); err != nil {
		return nil, err
	}
	if err := ValidateVehicleType(req.VehicleType); err != nil {
		return nil, err
	}
	if req.RegisterDate.IsZero() {
		return nil, errors.New("注册日期不能为空")
	}
	
	vm.mu.Lock()
	defer vm.mu.Unlock()
	
	if _, exists := vm.vehicles[req.PlateNumber]; exists {
		return nil, errors.New("该车牌号已存在")
	}
	
	vehicle := &common.Vehicle{
		PlateNumber:        req.PlateNumber,
		VehicleType:        req.VehicleType,
		RegisterDate:       req.RegisterDate,
		LastInspectionDate: req.LastInspectionDate,
	}
	
	vm.vehicles[req.PlateNumber] = vehicle
	return vehicle, nil
}

func (vm *VehicleManager) GetVehicle(plateNumber string) (*common.Vehicle, error) {
	if err := ValidatePlateNumber(plateNumber); err != nil {
		return nil, err
	}
	
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	
	vehicle, exists := vm.vehicles[plateNumber]
	if !exists {
		return nil, errors.New("车辆不存在")
	}
	
	return vehicle, nil
}

func (vm *VehicleManager) GetVehicleInfo(plateNumber string) (*common.VehicleInfoResponse, error) {
	vehicle, err := vm.GetVehicle(plateNumber)
	if err != nil {
		return nil, err
	}
	
	now := time.Now()
	nextDate := CalculateNextInspectionDate(vehicle, now)
	isDue := IsDueForInspection(vehicle, now)
	daysUntilDue := CalculateDaysUntilDue(vehicle, now)
	
	return &common.VehicleInfoResponse{
		Vehicle:            *vehicle,
		NextInspectionDate: nextDate.Format("2006-01-02"),
		IsDueForInspection: isDue,
		DaysUntilDue:       daysUntilDue,
	}, nil
}

func (vm *VehicleManager) ListVehicles() []*common.Vehicle {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	
	vehicles := make([]*common.Vehicle, 0, len(vm.vehicles))
	for _, v := range vm.vehicles {
		vehicles = append(vehicles, v)
	}
	
	return vehicles
}

func (vm *VehicleManager) UpdateLastInspectionDate(plateNumber string, inspectionDate time.Time) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	
	vehicle, exists := vm.vehicles[plateNumber]
	if !exists {
		return errors.New("车辆不存在")
	}
	
	vehicle.LastInspectionDate = inspectionDate
	return nil
}

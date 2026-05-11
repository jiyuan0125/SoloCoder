package core

import (
	"errors"
	"fmt"
	"regexp"
	"sync"
	"taxisystem/common"
)

var (
	regularPlateRegex  = regexp.MustCompile(`^[京津沪渝冀豫云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼][A-Z][A-Z0-9]{5}$`)
	newEnergyPlateRegex = regexp.MustCompile(`^[京津沪渝冀豫云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼][A-Z][DF][A-Z0-9]{5}[0-9A-Z]$`)
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

func ValidatePlateNumber(plate string) error {
	if len(plate) == 7 {
		if regularPlateRegex.MatchString(plate) {
			return nil
		}
		return errors.New("普通车牌号格式错误")
	}
	if len(plate) == 8 {
		if newEnergyPlateRegex.MatchString(plate) {
			return nil
		}
		return errors.New("新能源车牌号格式错误")
	}
	return errors.New("车牌号长度必须为7位或8位")
}

func (vm *VehicleManager) AddVehicle(plate, driverName, driverPhone string, lat, lng float64) (*common.Vehicle, error) {
	if err := ValidatePlateNumber(plate); err != nil {
		return nil, err
	}
	if driverName == "" {
		return nil, errors.New("司机姓名不能为空")
	}
	if driverPhone == "" {
		return nil, errors.New("司机手机号不能为空")
	}

	vm.mu.Lock()
	defer vm.mu.Unlock()

	if _, exists := vm.vehicles[plate]; exists {
		return nil, errors.New("该车牌号已存在")
	}

	vehicle := &common.Vehicle{
		PlateNumber: plate,
		DriverName:  driverName,
		DriverPhone: driverPhone,
		Status:      common.VehicleStatusIdle,
		Latitude:    lat,
		Longitude:   lng,
	}

	vm.vehicles[plate] = vehicle
	return vehicle, nil
}

func (vm *VehicleManager) GetVehicle(plate string) (*common.Vehicle, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	vehicle, exists := vm.vehicles[plate]
	if !exists {
		return nil, fmt.Errorf("车辆不存在: %s", plate)
	}

	v := *vehicle
	return &v, nil
}

func (vm *VehicleManager) UpdateStatus(plate string, status common.VehicleStatus) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vehicle, exists := vm.vehicles[plate]
	if !exists {
		return fmt.Errorf("车辆不存在: %s", plate)
	}

	vehicle.Status = status
	return nil
}

func (vm *VehicleManager) UpdateLocation(plate string, lat, lng float64) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vehicle, exists := vm.vehicles[plate]
	if !exists {
		return fmt.Errorf("车辆不存在: %s", plate)
	}

	vehicle.Latitude = lat
	vehicle.Longitude = lng
	return nil
}

func (vm *VehicleManager) ListVehicles() []*common.Vehicle {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	list := make([]*common.Vehicle, 0, len(vm.vehicles))
	for _, v := range vm.vehicles {
		vc := *v
		list = append(list, &vc)
	}
	return list
}

func (vm *VehicleManager) GetIdleVehicles() []*common.Vehicle {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	list := make([]*common.Vehicle, 0)
	for _, v := range vm.vehicles {
		if v.Status == common.VehicleStatusIdle {
			vc := *v
			list = append(list, &vc)
		}
	}
	return list
}

func (vm *VehicleManager) SetDispatched(plate string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vehicle, exists := vm.vehicles[plate]
	if !exists {
		return fmt.Errorf("车辆不存在: %s", plate)
	}
	if vehicle.Status != common.VehicleStatusIdle {
		return errors.New("车辆不是空闲状态")
	}

	vehicle.Status = common.VehicleStatusReserved
	return nil
}

func (vm *VehicleManager) SetAccepted(plate string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vehicle, exists := vm.vehicles[plate]
	if !exists {
		return fmt.Errorf("车辆不存在: %s", plate)
	}

	vehicle.Status = common.VehicleStatusOccupied
	return nil
}

func (vm *VehicleManager) SetCompleted(plate string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vehicle, exists := vm.vehicles[plate]
	if !exists {
		return fmt.Errorf("车辆不存在: %s", plate)
	}

	vehicle.Status = common.VehicleStatusIdle
	return nil
}

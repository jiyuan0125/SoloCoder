package core

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
	"taxisystem/common"

	"github.com/google/uuid"
)

const (
	dispatchTimeoutSeconds = 30
	maxDispatchAttempts    = 3
	avgSpeedKmPerHour      = 40
)

type DispatchManager struct {
	vm          *VehicleManager
	fc          *FareCalculator
	orders      map[string]*common.Order
	pendingOrders map[string]string
	dispatchTimers map[string]*time.Timer
	mu          sync.RWMutex
}

func NewDispatchManager(vm *VehicleManager, fc *FareCalculator) *DispatchManager {
	return &DispatchManager{
		vm:              vm,
		fc:              fc,
		orders:          make(map[string]*common.Order),
		pendingOrders:   make(map[string]string),
		dispatchTimers:  make(map[string]*time.Timer),
	}
}

func CalculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371.0

	lat1Rad := lat1 * math.Pi / 180.0
	lng1Rad := lng1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	lng2Rad := lng2 * math.Pi / 180.0

	dLat := lat2Rad - lat1Rad
	dLng := lng2Rad - lng1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
		math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

type vehicleWithDistance struct {
	vehicle  *common.Vehicle
	distance float64
}

func (dm *DispatchManager) findNearestIdleVehicles(pickup common.Location) []*common.Vehicle {
	vehicles := dm.vm.GetIdleVehicles()

	vd := make([]vehicleWithDistance, 0, len(vehicles))
	for _, v := range vehicles {
		dist := CalculateDistance(
			pickup.Latitude, pickup.Longitude,
			v.Latitude, v.Longitude,
		)
		vd = append(vd, vehicleWithDistance{vehicle: v, distance: dist})
	}

	sort.Slice(vd, func(i, j int) bool {
		return vd[i].distance < vd[j].distance
	})

	result := make([]*common.Vehicle, 0, len(vd))
	for _, item := range vd {
		result = append(result, item.vehicle)
	}

	return result
}

func (dm *DispatchManager) EstimateFare(pickup, dest common.Location) (float64, float64, int64) {
	distance := CalculateDistance(
		pickup.Latitude, pickup.Longitude,
		dest.Latitude, dest.Longitude,
	)

	if distance < 0.1 {
		distance = 0.1
	}

	durationSeconds := int64((distance / avgSpeedKmPerHour) * 3600)
	if durationSeconds < 60 {
		durationSeconds = 60
	}

	fare := dm.fc.CalculateEstimatedFare(time.Now(), distance, durationSeconds)
	fare = math.Round(fare*100) / 100

	return fare, distance, durationSeconds
}

func (dm *DispatchManager) CreateOrder(passengerID, passengerPhone string, pickup, dest common.Location) (*common.Order, error) {
	dm.mu.Lock()
	if _, exists := dm.pendingOrders[passengerID]; exists {
		dm.mu.Unlock()
		return nil, errors.New("您有待处理的订单，无法重复下单")
	}
	dm.mu.Unlock()

	estimatedFare, estimatedDistance, estimatedDuration := dm.EstimateFare(pickup, dest)

	order := &common.Order{
		OrderID:           uuid.New().String(),
		PassengerID:       passengerID,
		PassengerPhone:    passengerPhone,
		PickupLocation:    pickup,
		DestLocation:      dest,
		Status:            common.OrderStatusPending,
		EstimatedDistance: estimatedDistance,
		EstimatedDuration: estimatedDuration,
		EstimatedFare:     estimatedFare,
		CreateTime:        time.Now().Unix(),
		DispatchAttempts:  0,
	}

	dm.mu.Lock()
	dm.orders[order.OrderID] = order
	dm.pendingOrders[passengerID] = order.OrderID
	dm.mu.Unlock()

	go dm.dispatchOrder(order.OrderID)

	return order, nil
}

func (dm *DispatchManager) dispatchOrder(orderID string) {
	for attempt := 1; attempt <= maxDispatchAttempts; attempt++ {
		dm.mu.Lock()
		order, exists := dm.orders[orderID]
		if !exists {
			dm.mu.Unlock()
			return
		}
		order.DispatchAttempts = attempt
		dm.mu.Unlock()

		vehicles := dm.findNearestIdleVehicles(order.PickupLocation)
		if len(vehicles) == 0 {
			dm.mu.Lock()
			order.Status = common.OrderStatusNoVehicle
			delete(dm.pendingOrders, order.PassengerID)
			dm.mu.Unlock()
			return
		}

		var targetVehicle *common.Vehicle
		for _, v := range vehicles {
			err := dm.vm.SetDispatched(v.PlateNumber)
			if err == nil {
				targetVehicle = v
				break
			}
		}

		if targetVehicle == nil {
			continue
		}

		dm.mu.Lock()
		order.VehiclePlate = targetVehicle.PlateNumber
		order.Status = common.OrderStatusDispatched
		dm.mu.Unlock()

		timer := time.NewTimer(time.Duration(dispatchTimeoutSeconds) * time.Second)
		dm.mu.Lock()
		dm.dispatchTimers[orderID] = timer
		dm.mu.Unlock()

		<-timer.C

		dm.mu.Lock()
		currentOrder, exists := dm.orders[orderID]
		if !exists {
			dm.mu.Unlock()
			return
		}

		if currentOrder.Status == common.OrderStatusAccepted ||
			currentOrder.Status == common.OrderStatusInProgress ||
			currentOrder.Status == common.OrderStatusCompleted {
			dm.mu.Unlock()
			return
		}

		dm.mu.Unlock()

		_ = dm.vm.SetCompleted(targetVehicle.PlateNumber)
	}

	dm.mu.Lock()
	order, exists := dm.orders[orderID]
	if exists {
		order.Status = common.OrderStatusNoVehicle
		delete(dm.pendingOrders, order.PassengerID)
	}
	dm.mu.Unlock()
}

func (dm *DispatchManager) AcceptOrder(orderID, plateNumber string) error {
	dm.mu.Lock()
	order, exists := dm.orders[orderID]
	if !exists {
		dm.mu.Unlock()
		return errors.New("订单不存在")
	}

	if order.Status != common.OrderStatusDispatched {
		dm.mu.Unlock()
		return errors.New("订单不在待接单状态")
	}

	if order.VehiclePlate != plateNumber {
		dm.mu.Unlock()
		return errors.New("该订单不是派给此车辆")
	}

	if timer, exists := dm.dispatchTimers[orderID]; exists {
		timer.Stop()
		delete(dm.dispatchTimers, orderID)
	}

	order.Status = common.OrderStatusAccepted
	order.AcceptTime = time.Now().Unix()
	dm.mu.Unlock()

	err := dm.vm.SetAccepted(plateNumber)
	if err != nil {
		return err
	}

	return nil
}

func (dm *DispatchManager) StartTrip(orderID string, actualDistance float64, actualDuration int64, lowSpeedMinutes int) error {
	dm.mu.Lock()
	order, exists := dm.orders[orderID]
	if !exists {
		dm.mu.Unlock()
		return errors.New("订单不存在")
	}

	if order.Status != common.OrderStatusAccepted {
		dm.mu.Unlock()
		return errors.New("订单不在待开始状态")
	}

	order.Status = common.OrderStatusInProgress
	order.ActualDistance = actualDistance
	order.ActualDuration = actualDuration
	order.LowSpeedMinutes = lowSpeedMinutes
	order.StartTime = time.Now().Unix()
	dm.mu.Unlock()

	return nil
}

func (dm *DispatchManager) CompleteTrip(orderID string) (*common.Order, error) {
	dm.mu.Lock()
	order, exists := dm.orders[orderID]
	if !exists {
		dm.mu.Unlock()
		return nil, errors.New("订单不存在")
	}

	if order.Status != common.OrderStatusInProgress {
		dm.mu.Unlock()
		return nil, errors.New("订单不在进行中状态")
	}

	startTime := time.Unix(order.StartTime, 0)
	endTime := time.Now()

	actualFare, isLargeOrder := dm.fc.CalculateActualFare(
		startTime,
		endTime,
		order.ActualDistance,
		order.LowSpeedMinutes,
	)

	order.Status = common.OrderStatusCompleted
	order.ActualFare = actualFare
	order.IsLargeOrder = isLargeOrder
	order.EndTime = endTime.Unix()

	plate := order.VehiclePlate
	delete(dm.pendingOrders, order.PassengerID)
	dm.mu.Unlock()

	_ = dm.vm.SetCompleted(plate)

	return order, nil
}

func (dm *DispatchManager) GetOrder(orderID string) (*common.Order, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	order, exists := dm.orders[orderID]
	if !exists {
		return nil, errors.New("订单不存在")
	}

	o := *order
	if order.Rating != nil {
		r := *order.Rating
		o.Rating = &r
	}
	return &o, nil
}

func (dm *DispatchManager) ListOrders() []*common.Order {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	list := make([]*common.Order, 0, len(dm.orders))
	for _, o := range dm.orders {
		order := *o
		if o.Rating != nil {
			r := *o.Rating
			order.Rating = &r
		}
		list = append(list, &order)
	}
	return list
}

func (dm *DispatchManager) GetPendingOrdersCount() int {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return len(dm.pendingOrders)
}

func (dm *DispatchManager) GetOrdersForMetrics() []*common.Order {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	list := make([]*common.Order, 0, len(dm.orders))
	for _, o := range dm.orders {
		order := *o
		if o.Rating != nil {
			r := *o.Rating
			order.Rating = &r
		}
		list = append(list, &order)
	}
	return list
}

func (dm *DispatchManager) UpdateOrderRating(orderID string, rating *common.Rating) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	order, exists := dm.orders[orderID]
	if !exists {
		return errors.New("订单不存在")
	}

	r := *rating
	order.Rating = &r
	return nil
}

func (dm *DispatchManager) ListComplaints() []*common.Complaint {
	return []*common.Complaint{}
}

func (dm *DispatchManager) GetPlateByOrder(orderID string) (string, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	order, exists := dm.orders[orderID]
	if !exists {
		return "", fmt.Errorf("订单不存在: %s", orderID)
	}

	return order.VehiclePlate, nil
}

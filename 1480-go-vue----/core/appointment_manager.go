package core

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"vehicle-inspection/common"
)

type slotKey struct {
	stationID string
	date      string
	timeSlot  string
}

type AppointmentManager struct {
	appointments   map[string]*common.Appointment
	slotBookings   map[slotKey]int
	vehicleManager *VehicleManager
	stationManager *StationManager
	mu             sync.RWMutex
	slotMu         sync.Map
	idGen          *IDGenerator
}

func NewAppointmentManager(vm *VehicleManager, sm *StationManager) *AppointmentManager {
	return &AppointmentManager{
		appointments:   make(map[string]*common.Appointment),
		slotBookings:   make(map[slotKey]int),
		vehicleManager: vm,
		stationManager: sm,
		idGen:          NewIDGenerator(),
	}
}

func (am *AppointmentManager) CreateAppointment(req *common.CreateAppointmentRequest) (*common.Appointment, error) {
	if req.PlateNumber == "" {
		return nil, errors.New("车牌号不能为空")
	}
	if req.StationID == "" {
		return nil, errors.New("检测站ID不能为空")
	}
	if req.TimeSlot == "" {
		return nil, errors.New("时间段不能为空")
	}
	if err := ValidateDateNotInPast(req.AppointmentDate); err != nil {
		return nil, err
	}

	vehicle, err := am.vehicleManager.GetVehicle(req.PlateNumber)
	if err != nil {
		return nil, err
	}

	canBook, reason := CanBookAppointment(vehicle, time.Now())
	if !canBook {
		return nil, errors.New(reason)
	}

	hasSchedule, schedule := am.stationManager.HasSchedule(req.StationID, req.AppointmentDate, req.TimeSlot)
	if !hasSchedule {
		return nil, errors.New("该检测站此时间段未开放预约")
	}

	key := slotKey{
		stationID: req.StationID,
		date:      req.AppointmentDate.Format("2006-01-02"),
		timeSlot:  req.TimeSlot,
	}

	slotMutex := am.getSlotMutex(key)
	slotMutex.Lock()
	defer slotMutex.Unlock()

	am.mu.RLock()
	currentBookings := am.slotBookings[key]
	am.mu.RUnlock()

	if currentBookings >= schedule.MaxVehicles {
		return nil, errors.New("该时段预约已满")
	}

	if am.hasExistingAppointment(req.PlateNumber, req.AppointmentDate) {
		return nil, errors.New("该车辆今日已有预约")
	}

	appointment := &common.Appointment{
		ID:              am.generateAppointmentID(),
		PlateNumber:     req.PlateNumber,
		StationID:       req.StationID,
		AppointmentDate: req.AppointmentDate,
		TimeSlot:        req.TimeSlot,
		CreatedAt:       time.Now(),
	}

	am.mu.Lock()
	am.appointments[appointment.ID] = appointment
	am.slotBookings[key] = currentBookings + 1
	am.mu.Unlock()

	return appointment, nil
}

func (am *AppointmentManager) getSlotMutex(key slotKey) *sync.Mutex {
	actual, _ := am.slotMu.LoadOrStore(key, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

func (am *AppointmentManager) hasExistingAppointment(plateNumber string, date time.Time) bool {
	for _, appt := range am.appointments {
		if appt.PlateNumber == plateNumber && isSameDay(appt.AppointmentDate, date) {
			return true
		}
	}
	return false
}

func (am *AppointmentManager) generateAppointmentID() string {
	ts := time.Now().UnixNano()
	return fmt.Sprintf("APT-%d", ts)
}

func (am *AppointmentManager) GetAppointment(id string) (*common.Appointment, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	appt, exists := am.appointments[id]
	if !exists {
		return nil, errors.New("预约不存在")
	}
	return appt, nil
}

func (am *AppointmentManager) ListAppointmentsByPlate(plateNumber string) []*common.Appointment {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var result []*common.Appointment
	for _, appt := range am.appointments {
		if appt.PlateNumber == plateNumber {
			result = append(result, appt)
		}
	}
	return result
}

func (am *AppointmentManager) ListAppointments() []*common.Appointment {
	am.mu.RLock()
	defer am.mu.RUnlock()

	appointments := make([]*common.Appointment, 0, len(am.appointments))
	for _, a := range am.appointments {
		appointments = append(appointments, a)
	}
	return appointments
}

func (am *AppointmentManager) GetSlotAvailability(stationID string, date time.Time, timeSlot string) (int, int, error) {
	hasSchedule, schedule := am.stationManager.HasSchedule(stationID, date, timeSlot)
	if !hasSchedule {
		return 0, 0, errors.New("该时间段未开放预约")
	}

	key := slotKey{
		stationID: stationID,
		date:      date.Format("2006-01-02"),
		timeSlot:  timeSlot,
	}

	am.mu.RLock()
	booked := am.slotBookings[key]
	am.mu.RUnlock()

	return booked, schedule.MaxVehicles, nil
}

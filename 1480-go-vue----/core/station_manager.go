package core

import (
	"errors"
	"sync"
	"time"

	"vehicle-inspection/common"
)

type StationManager struct {
	stations map[string]*common.Station
	mu       sync.RWMutex
	idGen    *IDGenerator
}

func NewStationManager() *StationManager {
	return &StationManager{
		stations: make(map[string]*common.Station),
		idGen:    NewIDGenerator(),
	}
}

func (sm *StationManager) CreateStation(req *common.CreateStationRequest) (*common.Station, error) {
	if req.Name == "" {
		return nil, errors.New("检测站名称不能为空")
	}
	
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	station := &common.Station{
		ID:        sm.idGen.Generate(),
		Name:      req.Name,
		Workshops: req.Workshops,
		Schedules: []common.Schedule{},
	}
	
	for i := range station.Workshops {
		station.Workshops[i].ID = sm.idGen.Generate()
	}
	
	sm.stations[station.ID] = station
	return station, nil
}

func (sm *StationManager) GetStation(stationID string) (*common.Station, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	station, exists := sm.stations[stationID]
	if !exists {
		return nil, errors.New("检测站不存在")
	}
	return station, nil
}

func (sm *StationManager) ListStations() []*common.Station {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	stations := make([]*common.Station, 0, len(sm.stations))
	for _, s := range sm.stations {
		stations = append(stations, s)
	}
	return stations
}

func (sm *StationManager) AddSchedule(req *common.AddScheduleRequest) (*common.Schedule, error) {
	if req.StationID == "" {
		return nil, errors.New("检测站ID不能为空")
	}
	if req.TimeSlot == "" {
		return nil, errors.New("时间段不能为空")
	}
	if req.MaxVehicles <= 0 {
		return nil, errors.New("最大车辆数必须大于0")
	}
	if err := ValidateDateNotInPast(req.Date); err != nil {
		return nil, err
	}
	
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	station, exists := sm.stations[req.StationID]
	if !exists {
		return nil, errors.New("检测站不存在")
	}
	
	for _, s := range station.Schedules {
		if isSameDay(s.Date, req.Date) && s.TimeSlot == req.TimeSlot {
			return nil, errors.New("该时间段已存在预约配置")
		}
	}
	
	schedule := common.Schedule{
		Date:        req.Date,
		TimeSlot:    req.TimeSlot,
		MaxVehicles: req.MaxVehicles,
	}
	
	station.Schedules = append(station.Schedules, schedule)
	return &schedule, nil
}

func (sm *StationManager) GetAvailableSlots(stationID string, date time.Time) []common.Schedule {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	station, exists := sm.stations[stationID]
	if !exists {
		return nil
	}
	
	slots := []common.Schedule{}
	for _, s := range station.Schedules {
		if isSameDay(s.Date, date) {
			slots = append(slots, s)
		}
	}
	
	return slots
}

func (sm *StationManager) HasSchedule(stationID string, date time.Time, timeSlot string) (bool, *common.Schedule) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	station, exists := sm.stations[stationID]
	if !exists {
		return false, nil
	}
	
	for i, s := range station.Schedules {
		if isSameDay(s.Date, date) && s.TimeSlot == timeSlot {
			return true, &station.Schedules[i]
		}
	}
	
	return false, nil
}

func isSameDay(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.Month() == t2.Month() && t1.Day() == t2.Day()
}

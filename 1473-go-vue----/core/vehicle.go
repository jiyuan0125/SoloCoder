package core

import (
	"fmt"
	"time"
)

func (ts *TransitSystem) AddVehicle(id, lineCode string, direction Direction) (*Vehicle, error) {
	ts.Mutex.Lock()
	defer ts.Mutex.Unlock()

	if id == "" {
		return nil, fmt.Errorf("vehicle ID cannot be empty")
	}

	if _, ok := ts.Vehicles[id]; ok {
		return nil, fmt.Errorf("vehicle %s already exists", id)
	}

	lineKey := getLineKey(lineCode, direction)
	line, ok := ts.Lines[lineKey]
	if !ok {
		return nil, fmt.Errorf("line %s (%s) not found", lineCode, direction.String())
	}

	vehicle := &Vehicle{
		ID:            id,
		Line:          line,
		CurrentStation: 0,
		Status:        VehicleStatusIdle,
		Latitude:      line.Stations[0].Latitude,
		Longitude:     line.Stations[0].Longitude,
	}

	ts.Vehicles[id] = vehicle
	return vehicle, nil
}

func (ts *TransitSystem) GetVehicle(id string) (*Vehicle, bool) {
	ts.Mutex.RLock()
	defer ts.Mutex.RUnlock()
	vehicle, ok := ts.Vehicles[id]
	return vehicle, ok
}

func (ts *TransitSystem) GetAllVehicles() []*Vehicle {
	ts.Mutex.RLock()
	defer ts.Mutex.RUnlock()
	vehicles := make([]*Vehicle, 0, len(ts.Vehicles))
	for _, vehicle := range ts.Vehicles {
		vehicles = append(vehicles, vehicle)
	}
	return vehicles
}

func (ts *TransitSystem) GetVehiclesByLine(lineCode string, direction Direction) []*Vehicle {
	ts.Mutex.RLock()
	defer ts.Mutex.RUnlock()
	lineKey := getLineKey(lineCode, direction)
	line, ok := ts.Lines[lineKey]
	if !ok {
		return nil
	}

	vehicles := make([]*Vehicle, 0)
	for _, vehicle := range ts.Vehicles {
		if vehicle.Line == line {
			vehicles = append(vehicles, vehicle)
		}
	}
	return vehicles
}

func (v *Vehicle) UpdatePosition() {
	v.Mutex.Lock()
	defer v.Mutex.Unlock()

	line := v.Line
	if line == nil || len(line.Stations) == 0 {
		return
	}

	currentTime := time.Now().Unix()
	currentHour, currentMinute := GetCurrentTime()
	isInServiceTime := IsTimeInRange(
		currentHour, currentMinute,
		line.FirstHour, line.FirstMinute,
		line.LastHour, line.LastMinute,
	)

	switch v.Status {
	case VehicleStatusIdle:
		if isInServiceTime {
			v.Status = VehicleStatusRunning
			v.CurrentStation = 0
			v.Latitude = line.Stations[0].Latitude
			v.Longitude = line.Stations[0].Longitude
		}
	case VehicleStatusRunning:
		nextStation := (v.CurrentStation + 1) % len(line.Stations)
		v.Latitude = line.Stations[nextStation].Latitude
		v.Longitude = line.Stations[nextStation].Longitude
		v.CurrentStation = nextStation
		v.Status = VehicleStatusAtStation
		v.StayAtStationAt = currentTime
	case VehicleStatusAtStation:
		if currentTime-v.StayAtStationAt >= 30 {
			if v.CurrentStation == len(line.Stations)-1 {
				if isInServiceTime {
					v.Status = VehicleStatusRunning
					v.CurrentStation = 0
					v.Latitude = line.Stations[0].Latitude
					v.Longitude = line.Stations[0].Longitude
				} else {
					v.Status = VehicleStatusStopped
				}
			} else {
				v.Status = VehicleStatusRunning
			}
		}
	case VehicleStatusStopped:
		if isInServiceTime {
			v.Status = VehicleStatusIdle
		}
	}
}

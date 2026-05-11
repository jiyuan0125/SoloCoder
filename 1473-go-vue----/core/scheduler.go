package core

import (
	"time"
)

type Scheduler struct {
	TransitSystem *TransitSystem
	Ticker        *time.Ticker
	StopChan      chan bool
	Running       bool
}

func NewScheduler(ts *TransitSystem) *Scheduler {
	return &Scheduler{
		TransitSystem: ts,
		StopChan:      make(chan bool, 1),
	}
}

func (s *Scheduler) Start() {
	if s.Running {
		return
	}
	s.Running = true
	s.Ticker = time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-s.Ticker.C:
				s.Update()
			case <-s.StopChan:
				s.Ticker.Stop()
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	if !s.Running {
		return
	}
	s.Running = false
	s.StopChan <- true
}

func (s *Scheduler) Update() {
	s.TransitSystem.Mutex.RLock()
	vehicles := make([]*Vehicle, 0, len(s.TransitSystem.Vehicles))
	for _, vehicle := range s.TransitSystem.Vehicles {
		vehicles = append(vehicles, vehicle)
	}
	s.TransitSystem.Mutex.RUnlock()

	for _, vehicle := range vehicles {
		vehicle.UpdatePosition()
	}

	s.RefreshLineStatus()
	s.TransitSystem.RefreshStationTypes()
}

func (s *Scheduler) RefreshLineStatus() {
	s.TransitSystem.Mutex.Lock()
	defer s.TransitSystem.Mutex.Unlock()

	currentHour, currentMinute := GetCurrentTime()

	for _, line := range s.TransitSystem.Lines {
		isInServiceTime := IsTimeInRange(
			currentHour, currentMinute,
			line.FirstHour, line.FirstMinute,
			line.LastHour, line.LastMinute,
		)

		hasVehicles := false
		for _, vehicle := range s.TransitSystem.Vehicles {
			if vehicle.Line == line {
				hasVehicles = true
				break
			}
		}

		line.IsOperational = isInServiceTime && hasVehicles

		if !hasVehicles {
			line.HasError = true
			line.ErrorReason = "线路无可用车辆"
		} else {
			line.HasError = false
			line.ErrorReason = ""
		}
	}
}

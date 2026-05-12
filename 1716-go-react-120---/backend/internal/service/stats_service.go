package service

import (
	"time"

	"ambulance-scheduler/internal/models"
	"ambulance-scheduler/internal/store"
)

type StatsService struct {
	store *store.Store
}

func NewStatsService(store *store.Store) *StatsService {
	return &StatsService{store: store}
}

func (s *StatsService) GetStatistics() *models.Statistics {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	calls := s.store.ListCalls()
	vehicles := s.store.ListVehicles()
	dispatchRecords := s.store.ListDispatchRecords()

	todayCalls := 0
	level12Today := 0
	level34Today := 0
	waitingInQueue := 0

	for _, call := range calls {
		if call.CreateTime.After(startOfDay) {
			todayCalls++
			if call.SeverityLevel == models.SeverityLevel1 || call.SeverityLevel == models.SeverityLevel2 {
				level12Today++
			} else {
				level34Today++
			}
		}
		if call.InQueue {
			waitingInQueue++
		}
	}

	idleVehicles := 0
	inTransitVehicles := 0
	maintenanceVehicles := 0
	totalVehicles := len(vehicles)
	totalStatusTime := 0.0
	activeTime := 0.0

	for _, v := range vehicles {
		switch v.CurrentStatus {
		case models.VehicleStatusIdle:
			idleVehicles++
		case models.VehicleStatusInTransit:
			inTransitVehicles++
		case models.VehicleStatusMaintenance:
			maintenanceVehicles++
		}

		if len(v.CurrentStatusHistory) > 0 {
			for i := 1; i < len(v.CurrentStatusHistory); i++ {
				duration := v.CurrentStatusHistory[i].Time.Sub(v.CurrentStatusHistory[i-1].Time).Minutes()
				totalStatusTime += duration
				if v.CurrentStatusHistory[i-1].Status == models.VehicleStatusInTransit ||
				   v.CurrentStatusHistory[i-1].Status == models.VehicleStatusReturning {
					activeTime += duration
				}
			}

			lastDuration := now.Sub(v.CurrentStatusHistory[len(v.CurrentStatusHistory)-1].Time).Minutes()
			totalStatusTime += lastDuration
			if v.CurrentStatus == models.VehicleStatusInTransit || v.CurrentStatus == models.VehicleStatusReturning {
				activeTime += lastDuration
			}
		}
	}

	var utilizationRate float64
	if totalStatusTime > 0 && totalVehicles > 0 {
		utilizationRate = (activeTime / totalStatusTime) * 100
	}

	var avgResponseTime float64
	responseCount := 0

	for _, dr := range dispatchRecords {
		if dr.ArrivalTime != nil {
			avgResponseTime += dr.ArrivalTime.Sub(dr.DispatchTime).Minutes()
			responseCount++
		}
	}

	if responseCount > 0 {
		avgResponseTime = avgResponseTime / float64(responseCount)
	}

	return &models.Statistics{
		TodayCallsCount:         todayCalls,
		TotalCallsCount:         len(calls),
		AverageResponseTime:     avgResponseTime,
		VehicleUtilizationRate:  utilizationRate,
		IdleVehiclesCount:       idleVehicles,
		InTransitVehiclesCount:  inTransitVehicles,
		MaintenanceVehiclesCount: maintenanceVehicles,
		WaitingInQueueCount:     waitingInQueue,
		Level12CallsToday:       level12Today,
		Level34CallsToday:       level34Today,
	}
}

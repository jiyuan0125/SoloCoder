package core

import (
	"errors"
	"fmt"
	"parking-system/common"
	"sync"
)

type SpotManager struct {
	spots map[string]*ParkingSpot
	mu    sync.RWMutex
}

func NewSpotManager() *SpotManager {
	return &SpotManager{
		spots: make(map[string]*ParkingSpot),
	}
}

func (m *SpotManager) AddSpot(id, area, number string, spotType common.ParkingSpotType) error {
	if id == "" || area == "" || number == "" {
		return errors.New("id, area, and number are required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.spots[id]; exists {
		return fmt.Errorf("spot with id %s already exists", id)
	}

	for _, spot := range m.spots {
		if spot.Area == area && spot.Number == number {
			return fmt.Errorf("spot number %s already exists in area %s", number, area)
		}
	}

	spot := &ParkingSpot{
		ID:     id,
		Area:   area,
		Number: number,
		Type:   spotType,
		Status: common.SpotStatusAvailable,
	}
	m.spots[id] = spot
	return nil
}

func (m *SpotManager) GetSpot(id string) (*ParkingSpot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	spot, ok := m.spots[id]
	return spot, ok
}

func (m *SpotManager) UpdateSpotStatus(id string, status common.ParkingSpotStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	spot, exists := m.spots[id]
	if !exists {
		return fmt.Errorf("spot with id %s not found", id)
	}
	spot.SetStatus(status)
	return nil
}

func (m *SpotManager) ListSpots(area, status string) []*ParkingSpot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*ParkingSpot
	for _, spot := range m.spots {
		matchesArea := area == "" || spot.Area == area
		matchesStatus := status == "" || string(spot.GetStatus()) == status
		if matchesArea && matchesStatus {
			result = append(result, spot)
		}
	}
	return result
}

func (m *SpotManager) GetGuidanceInfo(monthlyCard *MonthlyCardManager) common.ParkingGuidanceDTO {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalSpots := 0
	availableSpots := 0
	areaStats := make(map[string]*common.AreaInfo)

	for _, spot := range m.spots {
		totalSpots++
		if _, ok := areaStats[spot.Area]; !ok {
			areaStats[spot.Area] = &common.AreaInfo{Name: spot.Area}
		}
		areaStats[spot.Area].TotalSpots++

		status := spot.GetStatus()
		if status == common.SpotStatusAvailable {
			availableSpots++
			areaStats[spot.Area].AvailableSpots++
		}
	}

	var areas []common.AreaInfo
	for _, info := range areaStats {
		info.IsFull = info.AvailableSpots == 0
		areas = append(areas, *info)
	}

	availablePercent := 0.0
	if totalSpots > 0 {
		availablePercent = float64(availableSpots) / float64(totalSpots) * 100
	}

	return common.ParkingGuidanceDTO{
		TotalSpots:       totalSpots,
		AvailableSpots:   availableSpots,
		AvailablePercent: availablePercent,
		IsFull:           availableSpots == 0,
		IsTight:          availablePercent < 10.0,
		Areas:            areas,
	}
}

func (m *SpotManager) AllocateSpot(forMonthlyCard bool, preferredType common.ParkingSpotType, fixedSpotID string) (*ParkingSpot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if forMonthlyCard && fixedSpotID != "" {
		spot, exists := m.spots[fixedSpotID]
		if !exists {
			return nil, fmt.Errorf("fixed spot %s not found", fixedSpotID)
		}
		status := spot.GetStatus()
		if status == common.SpotStatusReserved || status == common.SpotStatusAvailable {
			spot.SetStatus(common.SpotStatusOccupied)
			return spot, nil
		}
		return nil, fmt.Errorf("fixed spot %s is not available (status: %s)", fixedSpotID, status)
	}

	for _, spot := range m.spots {
		status := spot.GetStatus()
		if forMonthlyCard && spot.Type == preferredType && status == common.SpotStatusAvailable {
			spot.SetStatus(common.SpotStatusOccupied)
			return spot, nil
		}
		if !forMonthlyCard && status == common.SpotStatusAvailable {
			spot.SetStatus(common.SpotStatusOccupied)
			return spot, nil
		}
	}

	return nil, errors.New("no available parking spots")
}

func (m *SpotManager) FreeSpot(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	spot, exists := m.spots[id]
	if !exists {
		return fmt.Errorf("spot with id %s not found", id)
	}
	spot.SetStatus(common.SpotStatusAvailable)
	return nil
}

func (m *SpotManager) MarkSpotAsReserved(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	spot, exists := m.spots[id]
	if !exists {
		return fmt.Errorf("spot with id %s not found", id)
	}
	if spot.GetStatus() == common.SpotStatusAvailable {
		spot.SetStatus(common.SpotStatusReserved)
	}
	return nil
}

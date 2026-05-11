package core

import (
	"fmt"
)

func (ts *TransitSystem) AddStation(code, name string, latitude, longitude float64) (*Station, error) {
	ts.Mutex.Lock()
	defer ts.Mutex.Unlock()

	if code == "" {
		return nil, fmt.Errorf("station code cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("station name cannot be empty")
	}

	if existing, ok := ts.Stations[code]; ok {
		return existing, nil
	}

	station := &Station{
		Code:      code,
		Name:      name,
		Latitude:  latitude,
		Longitude: longitude,
		Type:      StationTypeNormal,
	}

	ts.Stations[code] = station
	return station, nil
}

func (ts *TransitSystem) GetStation(code string) (*Station, bool) {
	ts.Mutex.RLock()
	defer ts.Mutex.RUnlock()
	station, ok := ts.Stations[code]
	return station, ok
}

func (ts *TransitSystem) GetAllStations() []*Station {
	ts.Mutex.RLock()
	defer ts.Mutex.RUnlock()
	stations := make([]*Station, 0, len(ts.Stations))
	for _, station := range ts.Stations {
		stations = append(stations, station)
	}
	return stations
}

func (ts *TransitSystem) RefreshStationTypes() {
	ts.Mutex.Lock()
	defer ts.Mutex.Unlock()

	stationLineCount := make(map[string]int)
	for _, line := range ts.Lines {
		seen := make(map[string]bool)
		for _, station := range line.Stations {
			normalizedName := NormalizeStationName(station.Name)
			if !seen[normalizedName] {
				seen[normalizedName] = true
				stationLineCount[normalizedName]++
			}
		}
	}

	for _, line := range ts.Lines {
		for i, station := range line.Stations {
			normalizedName := NormalizeStationName(station.Name)
			if i == 0 || i == len(line.Stations)-1 {
				station.Type = StationTypeTerminal
			} else if stationLineCount[normalizedName] >= 2 {
				station.Type = StationTypeTransfer
			} else {
				station.Type = StationTypeNormal
			}
		}
	}
}

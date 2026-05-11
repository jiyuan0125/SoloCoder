package core

import (
	"fmt"
)

func (ts *TransitSystem) AddLine(code, name string, direction Direction, firstTime, lastTime string, stationCodes []string) (*Line, error) {
	ts.Mutex.Lock()
	defer ts.Mutex.Unlock()

	if code == "" {
		return nil, fmt.Errorf("line code cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("line name cannot be empty")
	}

	lineKey := getLineKey(code, direction)
	if _, ok := ts.Lines[lineKey]; ok {
		return nil, fmt.Errorf("line %s (%s) already exists", code, direction.String())
	}

	firstHour, firstMinute, err := ParseTime(firstTime)
	if err != nil {
		return nil, fmt.Errorf("invalid first time: %v", err)
	}

	lastHour, lastMinute, err := ParseTime(lastTime)
	if err != nil {
		return nil, fmt.Errorf("invalid last time: %v", err)
	}

	firstMinutes := firstHour*60 + firstMinute
	lastMinutes := lastHour*60 + lastMinute
	if firstMinutes >= lastMinutes {
		return nil, fmt.Errorf("first time (%s) must be earlier than last time (%s)", firstTime, lastTime)
	}

	if len(stationCodes) < 2 {
		return nil, fmt.Errorf("line must have at least 2 stations")
	}

	stations := make([]*Station, 0, len(stationCodes))
	for _, stationCode := range stationCodes {
		station, ok := ts.Stations[stationCode]
		if !ok {
			return nil, fmt.Errorf("station %s not found", stationCode)
		}
		stations = append(stations, station)
	}

	line := &Line{
		Code:        code,
		Name:        name,
		Direction:   direction,
		FirstHour:   firstHour,
		FirstMinute: firstMinute,
		LastHour:    lastHour,
		LastMinute:  lastMinute,
		Stations:    stations,
	}

	ts.Lines[lineKey] = line
	return line, nil
}

func (ts *TransitSystem) GetLine(code string, direction Direction) (*Line, bool) {
	ts.Mutex.RLock()
	defer ts.Mutex.RUnlock()
	line, ok := ts.Lines[getLineKey(code, direction)]
	return line, ok
}

func (ts *TransitSystem) GetAllLines() []*Line {
	ts.Mutex.RLock()
	defer ts.Mutex.RUnlock()
	lines := make([]*Line, 0, len(ts.Lines))
	for _, line := range ts.Lines {
		lines = append(lines, line)
	}
	return lines
}

func (ts *TransitSystem) UpdateLine(code string, direction Direction, updates map[string]interface{}) error {
	ts.Mutex.Lock()
	defer ts.Mutex.Unlock()

	lineKey := getLineKey(code, direction)
	line, ok := ts.Lines[lineKey]
	if !ok {
		return fmt.Errorf("line %s (%s) not found", code, direction.String())
	}

	if name, ok := updates["name"]; ok {
		if nameStr, ok := name.(string); ok && nameStr != "" {
			line.Name = nameStr
		}
	}

	if firstTime, ok := updates["first_time"]; ok {
		if firstTimeStr, ok := firstTime.(string); ok {
			firstHour, firstMinute, err := ParseTime(firstTimeStr)
			if err != nil {
				return fmt.Errorf("invalid first time: %v", err)
			}
			line.FirstHour = firstHour
			line.FirstMinute = firstMinute
		}
	}

	if lastTime, ok := updates["last_time"]; ok {
		if lastTimeStr, ok := lastTime.(string); ok {
			lastHour, lastMinute, err := ParseTime(lastTimeStr)
			if err != nil {
				return fmt.Errorf("invalid last time: %v", err)
			}
			line.LastHour = lastHour
			line.LastMinute = lastMinute
		}
	}

	firstMinutes := line.FirstHour*60 + line.FirstMinute
	lastMinutes := line.LastHour*60 + line.LastMinute
	if firstMinutes >= lastMinutes {
		return fmt.Errorf("first time must be earlier than last time")
	}

	if stationCodes, ok := updates["station_codes"]; ok {
		if codes, ok := stationCodes.([]string); ok {
			if len(codes) < 2 {
				return fmt.Errorf("line must have at least 2 stations")
			}
			stations := make([]*Station, 0, len(codes))
			for _, stationCode := range codes {
				station, found := ts.Stations[stationCode]
				if !found {
					return fmt.Errorf("station %s not found", stationCode)
				}
				stations = append(stations, station)
			}
			line.Stations = stations
		}
	}

	return nil
}

func getLineKey(code string, direction Direction) string {
	return fmt.Sprintf("%s:%d", code, direction)
}

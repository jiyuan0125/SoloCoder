package core

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"obd-platform/pkg/api"
)

type Storage struct {
	mu               sync.RWMutex
	records          map[string]*OBDRecord
	seenHashes       map[string]bool
	dailyVehicle     map[string]*DailyVehicleStats
	dailyFleet       map[string]*DailyFleetStats
	alerts           map[string]*Alert
	vehicleAlerts    map[string]map[string]*Alert
	alertIndex       map[string]string
	devices          map[string]bool
}

func NewStorage() *Storage {
	return &Storage{
		records:        make(map[string]*OBDRecord),
		seenHashes:     make(map[string]bool),
		dailyVehicle:   make(map[string]*DailyVehicleStats),
		dailyFleet:     make(map[string]*DailyFleetStats),
		alerts:         make(map[string]*Alert),
		vehicleAlerts:  make(map[string]map[string]*Alert),
		alertIndex:     make(map[string]string),
		devices:        make(map[string]bool),
	}
}

func (s *Storage) getVehicleDayKey(deviceID string, date time.Time) string {
	return fmt.Sprintf("%s|%s", deviceID, truncateToDate(date).Format("2006-01-02"))
}

func (s *Storage) getDayKey(date time.Time) string {
	return truncateToDate(date).Format("2006-01-02")
}

func (s *Storage) getAlertKey(deviceID string, alertType AlertType, date time.Time) string {
	return fmt.Sprintf("%s|%s|%s", deviceID, alertType, truncateToDate(date).Format("2006-01-02"))
}

func (s *Storage) InsertPackets(packets []*api.OBDPacket) (int, int, error) {
	if len(packets) == 0 {
		return 0, 0, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var uploaded, deduplicated int

	for _, p := range packets {
		if p.DeviceID == "" {
			continue
		}

		hash := recordHash(p.DeviceID, p.Timestamp)
		if s.seenHashes[hash] {
			deduplicated++
			continue
		}

		s.seenHashes[hash] = true
		record := newRecordFromAPI(p)

		dayKey := s.getVehicleDayKey(record.DeviceID, record.Timestamp)
		vehicleDay, ok := s.dailyVehicle[dayKey]
		if !ok {
			vehicleDay = &DailyVehicleStats{
				Date:     truncateToDate(record.Timestamp),
				DeviceID: record.DeviceID,
				HasData:  true,
			}
			s.dailyVehicle[dayKey] = vehicleDay
		}

		if vehicleDay.LastRecord != nil {
			speedDiff := vehicleDay.LastRecord.Speed - record.Speed
			if speedDiff > 30 {
				record.IsSevereBrake = true
			}

			mileage := distance(
				vehicleDay.LastRecord.Latitude,
				vehicleDay.LastRecord.Longitude,
				record.Latitude,
				record.Longitude,
			)
			vehicleDay.TotalMileage += mileage

			fuel := mileage * record.FuelConsumption / 100.0
			vehicleDay.TotalFuel += fuel
		}

		record.PrevSpeed = vehicleDay.LastRecord.Speed
		vehicleDay.LastRecord = record
		vehicleDay.Packets++
		vehicleDay.HardAccelerations += record.HardAccelerations
		vehicleDay.HardBrakes += record.HardBrakes
		vehicleDay.HardTurns += record.HardTurns
		if record.IsSevereBrake {
			vehicleDay.SevereBrakes++
		}

		if vehicleDay.TotalMileage > 0 {
			vehicleDay.AvgFuelConsumption = vehicleDay.TotalFuel / vehicleDay.TotalMileage * 100.0
		}

		vehicleDay.SafetyScore = calculateSafetyScore(
			vehicleDay.HardAccelerations,
			vehicleDay.HardBrakes,
			vehicleDay.HardTurns,
		)

		fleetKey := s.getDayKey(record.Timestamp)
		_, ok = s.dailyFleet[fleetKey]
		if !ok {
			s.dailyFleet[fleetKey] = &DailyFleetStats{
				Date: truncateToDate(record.Timestamp),
			}
		}

		s.records[hash] = record
		s.devices[record.DeviceID] = true
		uploaded++
	}

	s.refreshFleetStats()

	return uploaded, deduplicated, nil
}

func (s *Storage) refreshFleetStats() {
	fleetDaily := make(map[string]*DailyFleetStats)
	vehiclePerDay := make(map[string]map[string]bool)

	for key, vStats := range s.dailyVehicle {
		dateKey := s.getDayKey(vStats.Date)

		if _, ok := fleetDaily[dateKey]; !ok {
			fleetDaily[dateKey] = &DailyFleetStats{Date: vStats.Date}
		}
		if _, ok := vehiclePerDay[dateKey]; !ok {
			vehiclePerDay[dateKey] = make(map[string]bool)
		}

		fleetDaily[dateKey].TotalMileage += vStats.TotalMileage
		fleetDaily[dateKey].TotalAbnormalEvents += vStats.HardAccelerations + vStats.HardBrakes + vStats.HardTurns
		vehiclePerDay[dateKey][vStats.DeviceID] = true

		_ = key
	}

	for dateKey, fStats := range fleetDaily {
		fStats.TotalVehicles = len(vehiclePerDay[dateKey])

		var totalFuel float64
		var totalMileage float64
		for _, vStats := range s.dailyVehicle {
			if s.getDayKey(vStats.Date) == dateKey && vStats.TotalMileage > 0 {
				totalFuel += vStats.TotalFuel
				totalMileage += vStats.TotalMileage
			}
		}

		if totalMileage > 0 {
			fStats.AvgFuelConsumption = totalFuel / totalMileage * 100.0
		}
	}

	s.dailyFleet = fleetDaily
}

func (s *Storage) GetVehicleDailyStats(deviceID string, startDate, endDate time.Time) []*api.VehicleDailyStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*api.VehicleDailyStats

	start := truncateToDate(startDate)
	end := truncateToDate(endDate)

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := s.getVehicleDayKey(deviceID, d)
		stats, ok := s.dailyVehicle[key]
		if !ok {
			result = append(result, &api.VehicleDailyStats{
				Date:        d,
				DeviceID:    deviceID,
				Mileage:     0,
				AvgFuel:     0,
				SafetyScore: 0,
				TotalFuel:   0,
			})
			continue
		}
		result = append(result, recordToAPIDailyStats(stats))
	}

	return result
}

func (s *Storage) GetFleetDailyStats(startDate, endDate time.Time) []*api.FleetDailyStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*api.FleetDailyStats

	start := truncateToDate(startDate)
	end := truncateToDate(endDate)

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := s.getDayKey(d)
		stats, ok := s.dailyFleet[key]
		if !ok {
			result = append(result, &api.FleetDailyStats{
				Date:                   d,
				TotalMileage:           0,
				AvgFuelConsumption:     0,
				AbnormalEvents:         0,
			})
			continue
		}
		result = append(result, recordsToFleetStats(stats))
	}

	return result
}

func (s *Storage) GetDevices() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var devices []string
	for d := range s.devices {
		devices = append(devices, d)
	}
	sort.Strings(devices)
	return devices
}

func (s *Storage) GetAllVehicleDailyStatsForMonth(year int, month time.Month) map[string][]*DailyVehicleStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]*DailyVehicleStats)

	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	days := daysInMonth(year, month)
	endDate := startDate.AddDate(0, 0, days-1)

	for deviceID := range s.devices {
		var stats []*DailyVehicleStats
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			key := s.getVehicleDayKey(deviceID, d)
			vs, ok := s.dailyVehicle[key]
			if !ok {
				stats = append(stats, &DailyVehicleStats{
					Date:     d,
					DeviceID: deviceID,
					HasData:  false,
				})
			} else {
				stats = append(stats, vs)
			}
		}
		result[deviceID] = stats
	}

	return result
}

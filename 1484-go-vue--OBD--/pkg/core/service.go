package core

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"obd-platform/pkg/api"
)

type Service struct {
	storage *Storage
}

func NewService(storage *Storage) *Service {
	if storage == nil {
		storage = NewStorage()
	}
	return &Service{storage: storage}
}

func (s *Service) UploadData(req *api.UploadDataRequest) (*api.UploadDataResponse, error) {
	if req == nil || len(req.Packets) == 0 {
		return &api.UploadDataResponse{
			Success: false,
			Message: "no packets provided",
		}, nil
	}

	for i := range req.Packets {
		if req.Packets[i].DeviceID == "" {
			req.Packets[i].DeviceID = req.DeviceID
		}
	}

	uploaded, deduplicated, err := s.storage.InsertPackets(req.Packets)
	if err != nil {
		return nil, err
	}

	return &api.UploadDataResponse{
		Success:      true,
		Message:      fmt.Sprintf("uploaded %d packets", uploaded),
		Uploaded:     uploaded,
		Deduplicated: deduplicated,
	}, nil
}

func (s *Service) GetVehicleDailyStats(deviceID string, startDate, endDate time.Time) (*api.VehicleDailyStatsResponse, error) {
	if deviceID == "" {
		return nil, errors.New("device ID is required")
	}

	stats := s.storage.GetVehicleDailyStats(deviceID, startDate, endDate)

	return &api.VehicleDailyStatsResponse{
		Success: true,
		Data:    stats,
	}, nil
}

func (s *Service) GetFleetDailyStats(startDate, endDate time.Time) (*api.FleetDailyStatsResponse, error) {
	stats := s.storage.GetFleetDailyStats(startDate, endDate)

	return &api.FleetDailyStatsResponse{
		Success: true,
		Data:    stats,
	}, nil
}

func (s *Service) ExportCSV(yearMonth string) ([]byte, error) {
	parts := strings.Split(yearMonth, "-")
	if len(parts) != 2 {
		return nil, errors.New("invalid year-month format, expected YYYY-MM")
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, errors.New("invalid year")
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil || month < 1 || month > 12 {
		return nil, errors.New("invalid month")
	}

	vehicleStatsMap := s.storage.GetAllVehicleDailyStatsForMonth(year, time.Month(month))

	buf := &bytes.Buffer{}

	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(buf)
	defer writer.Flush()

	header := []string{"设备ID", "日期", "行驶里程(km)", "总油耗(L)", "平均油耗(L/100km)", "安全评分"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	devices := make([]string, 0, len(vehicleStatsMap))
	for d := range vehicleStatsMap {
		devices = append(devices, d)
	}
	sort.Strings(devices)

	for _, deviceID := range devices {
		statsList := vehicleStatsMap[deviceID]
		for _, stats := range statsList {
			row := make([]string, 6)
			row[0] = deviceID
			row[1] = stats.Date.Format("2006-01-02")

			if stats.HasData {
				row[2] = fmt.Sprintf("%.2f", stats.TotalMileage)
				row[3] = fmt.Sprintf("%.2f", stats.TotalFuel)
				row[4] = fmt.Sprintf("%.2f", stats.AvgFuelConsumption)
				row[5] = fmt.Sprintf("%d", stats.SafetyScore)
			} else {
				row[2] = "0"
				row[3] = "0"
				row[4] = "0"
				row[5] = ""
			}

			if err := writer.Write(row); err != nil {
				return nil, err
			}
		}
	}

	return buf.Bytes(), nil
}

func (s *Service) GenerateAlerts() ([]*api.Alert, error) {
	devices := s.storage.GetDevices()
	var allAlerts []*api.Alert

	for _, deviceID := range devices {
		dailyScores := s.getLast30DaysScores(deviceID)
		if len(dailyScores) == 0 {
			continue
		}

		alerts := s.checkAndGenerateAlerts(deviceID, dailyScores)
		for _, a := range alerts {
			allAlerts = append(allAlerts, a.toAPI())
		}
	}

	return allAlerts, nil
}

func (s *Service) getLast30DaysScores(deviceID string) map[time.Time]int {
	scores := make(map[time.Time]int)

	now := time.Now()
	start := now.AddDate(0, 0, -30)
	end := now

	s.storage.mu.RLock()
	defer s.storage.mu.RUnlock()

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := s.storage.getVehicleDayKey(deviceID, d)
		stats, ok := s.storage.dailyVehicle[key]
		if ok && stats.HasData {
			scores[truncateToDate(d)] = stats.SafetyScore
		}
	}

	return scores
}

func (s *Service) checkAndGenerateAlerts(deviceID string, scores map[time.Time]int) []*Alert {
	var alerts []*Alert

	lowScores := make(map[time.Time]int)
	for d, score := range scores {
		if score < 60 {
			lowScores[d] = score
		}
	}

	sortedDates := make([]time.Time, 0, len(scores))
	for d := range scores {
		sortedDates = append(sortedDates, d)
	}
	sort.Slice(sortedDates, func(i, j int) bool {
		return sortedDates[i].Before(sortedDates[j])
	})

	for i := 2; i < len(sortedDates); i++ {
		d1 := sortedDates[i-2]
		d2 := sortedDates[i-1]
		d3 := sortedDates[i]

		if d2.Sub(d1) == 24*time.Hour && d3.Sub(d2) == 24*time.Hour {
			_, ok1 := lowScores[d1]
			_, ok2 := lowScores[d2]
			_, ok3 := lowScores[d3]

			if ok1 && ok2 && ok3 {
				alert := s.createAlertIfNotExists(deviceID, AlertTypeDrivingHabit, d3,
					fmt.Sprintf("车辆 %s 连续3天安全评分低于60分", deviceID))
				if alert != nil {
					alerts = append(alerts, alert)
				}
			}
		}
	}

	for d, score := range scores {
		statsKey := s.storage.getVehicleDayKey(deviceID, d)
		s.storage.mu.RLock()
		vs, ok := s.storage.dailyVehicle[statsKey]
		s.storage.mu.RUnlock()

		if !ok {
			continue
		}

		if vs.SevereBrakes > 0 {
			alert := s.createAlertIfNotExists(deviceID, AlertTypeSevereBrake, d,
				fmt.Sprintf("车辆 %s 在 %s 发生 %d 次严重急刹车", deviceID, d.Format("2006-01-02"), vs.SevereBrakes))
			if alert != nil {
				alerts = append(alerts, alert)
			}
		}

		if vs.HardBrakes > 10 {
			alert := s.createAlertIfNotExists(deviceID, AlertTypeHighBraking, d,
				fmt.Sprintf("车辆 %s 在 %s 急刹车次数 %d 次，超过阈值10次", deviceID, d.Format("2006-01-02"), vs.HardBrakes))
			if alert != nil {
				alerts = append(alerts, alert)
			}
		}

		_ = score
	}

	return alerts
}

func (s *Service) createAlertIfNotExists(deviceID string, alertType AlertType, date time.Time, message string) *Alert {
	s.storage.mu.Lock()
	defer s.storage.mu.Unlock()

	key := s.storage.getAlertKey(deviceID, alertType, date)
	if _, exists := s.storage.alertIndex[key]; exists {
		return nil
	}

	alertID := fmt.Sprintf("alert_%s_%s_%d", deviceID, alertType, date.Unix())

	alert := &Alert{
		ID:        alertID,
		DeviceID:  deviceID,
		Type:      alertType,
		Date:      truncateToDate(date),
		Message:   message,
		Status:    AlertStatusPending,
		CreatedAt: time.Now(),
	}

	s.storage.alerts[alertID] = alert
	s.storage.alertIndex[key] = alertID

	if _, ok := s.storage.vehicleAlerts[deviceID]; !ok {
		s.storage.vehicleAlerts[deviceID] = make(map[string]*Alert)
	}
	s.storage.vehicleAlerts[deviceID][alertID] = alert

	return alert
}

func (s *Service) GetAlerts(deviceID, status string) ([]*api.Alert, error) {
	s.storage.mu.RLock()
	defer s.storage.mu.RUnlock()

	var result []*api.Alert

	for _, a := range s.storage.alerts {
		if deviceID != "" && a.DeviceID != deviceID {
			continue
		}
		if status != "" && a.Status != status {
			continue
		}
		result = append(result, a.toAPI())
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result, nil
}

func (s *Service) ResolveAlerts(alertIDs []string) error {
	s.storage.mu.Lock()
	defer s.storage.mu.Unlock()

	for _, id := range alertIDs {
		if a, ok := s.storage.alerts[id]; ok {
			a.Status = AlertStatusResolved
		}
	}

	return nil
}

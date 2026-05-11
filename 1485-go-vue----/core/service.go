package core

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func generateTaskID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type Service struct {
	store    *Store
	alerts   *AlertManager
	stats    *StatsCalculator
	exporter *ReportExporter
}

func NewService() *Service {
	store := NewStore()
	return &Service{
		store:    store,
		alerts:   NewAlertManager(store),
		stats:    NewStatsCalculator(store),
		exporter: NewReportExporter(store),
	}
}

func (s *Service) CreateTask(deviceID string, cargoTypeStr string, startWarehouse, endWarehouse string) (*TransportTask, error) {
	cargoType, err := ValidateCargoType(cargoTypeStr)
	if err != nil {
		return nil, err
	}

	tempLimits, err := GetTemperatureLimits(cargoType)
	if err != nil {
		return nil, err
	}

	humidityLimits, err := GetHumidityLimits(cargoType)
	if err != nil {
		return nil, err
	}

	task := &TransportTask{
		ID:             generateTaskID(),
		DeviceID:       deviceID,
		CargoType:      cargoType,
		StartWarehouse: startWarehouse,
		EndWarehouse:   endWarehouse,
		StartTime:      time.Now(),
		Status:         TaskStatusActive,
		TempMin:        tempLimits.Min,
		TempMax:        tempLimits.Max,
		HumidityMin:    humidityLimits.Min,
		HumidityMax:    humidityLimits.Max,
	}

	if err := s.store.CreateTask(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *Service) GetTask(taskID string) (*TransportTask, error) {
	return s.store.GetTask(taskID)
}

func (s *Service) EndTask(taskID string) error {
	return s.store.EndTask(taskID)
}

func (s *Service) ListTasks() []*TransportTask {
	return s.store.ListTasks()
}

func (s *Service) SubmitSensorData(data *SensorData) (*Alert, error) {
	task, exists := s.store.GetActiveTaskByDevice(data.DeviceID)
	if !exists {
		return nil, nil
	}

	if task.Status != TaskStatusActive {
		return nil, nil
	}

	s.store.AddSensorData(task.ID, data)

	alert, err := s.alerts.CheckAndGenerateAlert(task.ID, data)
	if err != nil {
		return nil, err
	}

	return alert, nil
}

func (s *Service) GetSensorData(taskID string) ([]*SensorData, error) {
	return s.store.GetSensorData(taskID)
}

func (s *Service) GetAlerts(taskID string) ([]*Alert, error) {
	return s.store.GetAlerts(taskID)
}

func (s *Service) CalculateStats(taskID string) (*TaskStats, error) {
	return s.stats.CalculateTaskStats(taskID)
}

func (s *Service) GenerateReport(taskID string) (*TaskReport, error) {
	return s.exporter.GenerateTaskReport(taskID)
}

func (s *Service) Exporter() *ReportExporter {
	return s.exporter
}

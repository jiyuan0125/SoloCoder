package core

import (
	"time"
)

type Service struct {
	GasMonitor        *GasMonitor
	WastewaterMonitor *WastewaterMonitor
	SolidWasteManager *SolidWasteManager
	AlarmManager      *AlarmManager
}

func NewService() *Service {
	service := &Service{
		GasMonitor:        NewGasMonitor(),
		WastewaterMonitor: NewWastewaterMonitor(),
		SolidWasteManager: NewSolidWasteManager(),
		AlarmManager:      NewAlarmManager(),
	}
	
	service.initializeStandards()
	return service
}

func (s *Service) initializeStandards() {
	s.GasMonitor.SetStandard("SO2", 100, 1000, "mg/m³")
	s.GasMonitor.SetStandard("NOx", 200, 2000, "mg/m³")
	s.GasMonitor.SetStandard("PM", 30, 500, "mg/m³")
	s.GasMonitor.SetStandard("VOCs", 50, 1000, "mg/m³")
	
	s.WastewaterMonitor.SetStandard("COD", 100, 0, 10000, "mg/L")
	s.WastewaterMonitor.SetStandard("NH3-N", 15, 0, 1000, "mg/L")
	s.WastewaterMonitor.SetStandard("TP", 0.5, 0, 100, "mg/L")
	s.WastewaterMonitor.SetStandard("pH", 7, 0, 14, "")
}

func (s *Service) RegisterGasOutlet(id, location string) *GasOutlet {
	return s.GasMonitor.RegisterOutlet(id, location)
}

func (s *Service) RegisterWastewaterOutlet(id string) {
	s.WastewaterMonitor.RegisterOutlet(id)
}

func (s *Service) ProcessGasReport(outletID string, reportedAt time.Time, measurements map[string]float64) (*GasReport, error) {
	alarmCheck := func(entityType EntityType, entityID, factor string, alarmType AlarmType, now time.Time) bool {
		return s.AlarmManager.HasAlarmToday(entityType, entityID, factor, alarmType, now)
	}
	
	report, alarms, err := s.GasMonitor.ProcessReport(outletID, reportedAt, measurements, alarmCheck)
	if err != nil {
		return nil, err
	}
	
	s.AlarmManager.AddAlarms(alarms)
	return report, nil
}

func (s *Service) ProcessWastewaterReport(outletID string, reportedAt time.Time, measurements map[string]float64) (*WastewaterReport, error) {
	alarmCheck := func(entityType EntityType, entityID, dateKey string, alarmType AlarmType) bool {
		return s.AlarmManager.HasDailyAlarmToday(entityType, entityID, dateKey)
	}
	
	report, alarms, err := s.WastewaterMonitor.ProcessReport(outletID, reportedAt, measurements, alarmCheck)
	if err != nil {
		return nil, err
	}
	
	s.AlarmManager.AddAlarms(alarms)
	return report, nil
}

func (s *Service) AddSolidWasteRecord(name string, category WasteCategory, amount float64, storageLocation string, generatedAt time.Time) (*SolidWasteRecord, error) {
	record, err := s.SolidWasteManager.AddRecord(name, category, amount, storageLocation, generatedAt)
	if err != nil {
		return nil, err
	}
	
	alarms := s.SolidWasteManager.CheckExpiration()
	s.AlarmManager.AddAlarms(alarms)
	
	return record, nil
}

func (s *Service) DisposeSolidWaste(id, method string) (*SolidWasteRecord, error) {
	return s.SolidWasteManager.DisposeRecord(id, method)
}

func (s *Service) GetAlarms(entityType *EntityType, entityID string, level *AlarmLevel, resolved *bool) []*Alarm {
	return s.AlarmManager.GetAlarms(entityType, entityID, level, resolved)
}

func (s *Service) ResolveAlarm(id string) (*Alarm, bool) {
	return s.AlarmManager.ResolveAlarm(id)
}

func (s *Service) CheckSolidWasteExpiration() []*Alarm {
	alarms := s.SolidWasteManager.CheckExpiration()
	s.AlarmManager.AddAlarms(alarms)
	return alarms
}

func (s *Service) GetGasOutlet(id string) (*GasOutlet, bool) {
	return s.GasMonitor.GetOutlet(id)
}

func (s *Service) GetGasReports(outletID string, from, to time.Time) []*GasReport {
	return s.GasMonitor.GetReports(outletID, from, to)
}

func (s *Service) GetWastewaterReports(outletID string, from, to time.Time) []*WastewaterReport {
	return s.WastewaterMonitor.GetReports(outletID, from, to)
}

func (s *Service) GetWastewaterDailyReport(outletID string, date time.Time) (*DailyReport, bool) {
	return s.WastewaterMonitor.GetDailyReport(outletID, date)
}

func (s *Service) GetSolidWasteRecords(category *WasteCategory, status *WasteStatus) []*SolidWasteRecord {
	return s.SolidWasteManager.GetRecords(category, status)
}

func (s *Service) GetSolidWasteRecord(id string) (*SolidWasteRecord, bool) {
	return s.SolidWasteManager.GetRecord(id)
}

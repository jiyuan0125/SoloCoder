package core

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type WastewaterMonitor struct {
	outlets         map[string]struct{}
	standards       map[string]*WastewaterStandard
	reports         []*WastewaterReport
	dailyReports    map[string]map[string]*DailyReport
	lock            sync.RWMutex
}

func NewWastewaterMonitor() *WastewaterMonitor {
	return &WastewaterMonitor{
		outlets:      make(map[string]struct{}),
		standards:    make(map[string]*WastewaterStandard),
		reports:      make([]*WastewaterReport, 0),
		dailyReports: make(map[string]map[string]*DailyReport),
	}
}

func (m *WastewaterMonitor) RegisterOutlet(id string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.outlets[id] = struct{}{}
	if _, exists := m.dailyReports[id]; !exists {
		m.dailyReports[id] = make(map[string]*DailyReport)
	}
}

func (m *WastewaterMonitor) SetStandard(factor string, limit, minValid, maxValid float64, unit string) {
	m.standards[factor] = &WastewaterStandard{
		Factor:   factor,
		Limit:    limit,
		Unit:     unit,
		MinValid: minValid,
		MaxValid: maxValid,
	}
}

func (m *WastewaterMonitor) ProcessReport(outletID string, reportedAt time.Time, measurements map[string]float64, alarmExists func(entityType EntityType, entityID, dateKey string, alarmType AlarmType) bool) (*WastewaterReport, []*Alarm, error) {
	m.lock.RLock()
	_, exists := m.outlets[outletID]
	m.lock.RUnlock()
	
	if !exists {
		return nil, nil, errors.New("outlet not registered")
	}
	
	report := &WastewaterReport{
		ID:           uuid.New().String(),
		OutletID:     outletID,
		ReportedAt:   reportedAt,
		ReceivedAt:   time.Now(),
		Measurements: make(map[string]Measurement),
	}
	
	exceededFactors := make([]string, 0)
	abnormalFactors := make([]string, 0)
	alarms := make([]*Alarm, 0)
	
	for factor, value := range measurements {
		measurement := Measurement{
			Value:      value,
			Factor:     factor,
			IsValid:    true,
		}
		
		standard, exists := m.standards[factor]
		if !exists {
			measurement.IsAbnormal = true
			abnormalFactors = append(abnormalFactors, factor)
			report.Measurements[factor] = measurement
			continue
		}
		
		measurement.Unit = standard.Unit
		
		if value < standard.MinValid || value > standard.MaxValid {
			measurement.IsValid = false
			measurement.IsAbnormal = true
			abnormalFactors = append(abnormalFactors, factor)
			
			alarm := &Alarm{
				ID:         uuid.New().String(),
				Type:       AlarmTypeEquipmentFailure,
				Level:      AlarmLevelWarning,
				EntityType: EntityTypeWastewaterOutlet,
				EntityID:   outletID,
				RelatedData: map[string]interface{}{
					"factor": factor,
					"value":  value,
				},
				Message:  "设备异常数据",
				CreatedAt: time.Now(),
			}
			alarms = append(alarms, alarm)
		} else {
			if factor == "pH" {
				if value < 6 || value > 9 {
					measurement.IsExceeded = true
					exceededFactors = append(exceededFactors, factor)
				}
			} else {
				if value > standard.Limit {
					measurement.IsExceeded = true
					exceededFactors = append(exceededFactors, factor)
				}
			}
		}
		
		report.Measurements[factor] = measurement
	}
	
	if len(abnormalFactors) == len(measurements) {
		report.DataStatus = DataStatusAllAbnormal
	} else if len(abnormalFactors) > 0 {
		report.DataStatus = DataStatusPartialAbnormal
	} else {
		report.DataStatus = DataStatusNormal
	}
	
	report.ExceededFactors = exceededFactors
	report.AbnormalFactors = abnormalFactors
	
	m.lock.Lock()
	m.reports = append(m.reports, report)
	
	dailyAlarms := m.updateDailyReport(outletID, reportedAt, report.Measurements, alarmExists)
	alarms = append(alarms, dailyAlarms...)
	
	m.lock.Unlock()
	
	return report, alarms, nil
}

func (m *WastewaterMonitor) updateDailyReport(outletID string, reportTime time.Time, measurements map[string]Measurement, alarmExists func(entityType EntityType, entityID, dateKey string, alarmType AlarmType) bool) []*Alarm {
	dateKey := reportTime.Format("2006-01-02")
	
	dailyReports, exists := m.dailyReports[outletID]
	if !exists {
		dailyReports = make(map[string]*DailyReport)
		m.dailyReports[outletID] = dailyReports
	}
	
	dailyReport, exists := dailyReports[dateKey]
	if !exists {
		dailyReport = &DailyReport{
			Date:     time.Date(reportTime.Year(), reportTime.Month(), reportTime.Day(), 0, 0, 0, 0, reportTime.Location()),
			OutletID: outletID,
			Status:   ReportStatusNormal,
			Data:     make(map[string]DailyFactorData),
		}
		dailyReports[dateKey] = dailyReport
	}
	
	alarms := make([]*Alarm, 0)
	
	for factor, measurement := range measurements {
		if !measurement.IsValid {
			continue
		}
		
		data, exists := dailyReport.Data[factor]
		if !exists {
			data = DailyFactorData{
				Max:   measurement.Value,
				Min:   measurement.Value,
				Count: 1,
				Unit:  measurement.Unit,
			}
		} else {
			if measurement.Value > data.Max {
				data.Max = measurement.Value
			}
			if measurement.Value < data.Min {
				data.Min = measurement.Value
			}
			data.Count++
		}
		
		data.Average = ((data.Average * float64(data.Count-1)) + measurement.Value) / float64(data.Count)
		
		standard, exists := m.standards[factor]
		if exists {
			if factor == "pH" {
				data.Exceeded = data.Average < 6 || data.Average > 9
			} else {
				data.Exceeded = data.Average > standard.Limit
			}
		}
		
		dailyReport.Data[factor] = data
	}
	
	hasExceeded := false
	for _, data := range dailyReport.Data {
		if data.Exceeded {
			hasExceeded = true
			break
		}
	}
	
	if hasExceeded {
		dailyReport.Status = ReportStatusExceeded
		
		if alarmExists == nil || !alarmExists(EntityTypeDailyReport, outletID, dateKey, AlarmTypeExceeding) {
			alarm := &Alarm{
				ID:         uuid.New().String(),
				Type:       AlarmTypeExceeding,
				Level:      AlarmLevelWarning,
				EntityType: EntityTypeDailyReport,
				EntityID:   outletID,
				RelatedData: map[string]interface{}{
					"date": dateKey,
				},
				Message:  "日均值超标告警",
				CreatedAt: time.Now(),
			}
			alarms = append(alarms, alarm)
		}
	}
	
	return alarms
}

func (m *WastewaterMonitor) GetDailyReport(outletID string, date time.Time) (*DailyReport, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	
	dailyReports, exists := m.dailyReports[outletID]
	if !exists {
		return nil, false
	}
	
	dateKey := date.Format("2006-01-02")
	report, exists := dailyReports[dateKey]
	return report, exists
}

func (m *WastewaterMonitor) GetReports(outletID string, from, to time.Time) []*WastewaterReport {
	m.lock.RLock()
	defer m.lock.RUnlock()
	
	reports := make([]*WastewaterReport, 0)
	for _, report := range m.reports {
		if (outletID == "" || report.OutletID == outletID) &&
			report.ReportedAt.After(from) && report.ReportedAt.Before(to) {
			reports = append(reports, report)
		}
	}
	return reports
}

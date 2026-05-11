package core

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type GasMonitor struct {
	outlets         map[string]*GasOutlet
	standards       map[string]*GasStandard
	reports         []*GasReport
	outletLock      sync.RWMutex
	reportLock      sync.RWMutex
}

func NewGasMonitor() *GasMonitor {
	return &GasMonitor{
		outlets:   make(map[string]*GasOutlet),
		standards: make(map[string]*GasStandard),
		reports:   make([]*GasReport, 0),
	}
}

func (m *GasMonitor) RegisterOutlet(id, location string) *GasOutlet {
	m.outletLock.Lock()
	defer m.outletLock.Unlock()
	
	if existing, exists := m.outlets[id]; exists {
		return existing
	}
	
	outlet := &GasOutlet{
		ID:       id,
		Location: location,
		Status:   OutletStatusNormal,
	}
	
	m.outlets[id] = outlet
	return outlet
}

func (m *GasMonitor) GetOutlet(id string) (*GasOutlet, bool) {
	m.outletLock.RLock()
	defer m.outletLock.RUnlock()
	
	outlet, exists := m.outlets[id]
	return outlet, exists
}

func (m *GasMonitor) SetStandard(factor string, limit, maxValid float64, unit string) {
	m.standards[factor] = &GasStandard{
		Factor:   factor,
		Limit:    limit,
		Unit:     unit,
		MaxValid: maxValid,
	}
}

func (m *GasMonitor) ProcessReport(outletID string, reportedAt time.Time, measurements map[string]float64, alarmExists func(entityType EntityType, entityID, factor string, alarmType AlarmType, now time.Time) bool) (*GasReport, []*Alarm, error) {
	m.outletLock.RLock()
	outlet, exists := m.outlets[outletID]
	m.outletLock.RUnlock()
	
	if !exists {
		return nil, nil, errors.New("outlet not registered")
	}
	
	report := &GasReport{
		ID:           uuid.New().String(),
		OutletID:     outletID,
		ReportedAt:   reportedAt,
		ReceivedAt:   time.Now(),
		Measurements: make(map[string]Measurement),
	}
	
	exceededFactors := make([]string, 0)
	abnormalFactors := make([]string, 0)
	alarms := make([]*Alarm, 0)
	
	now := time.Now()
	
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
		
		if value < 0 || value > standard.MaxValid {
			measurement.IsValid = false
			measurement.IsAbnormal = true
			abnormalFactors = append(abnormalFactors, factor)
			
			alarm := &Alarm{
				ID:         uuid.New().String(),
				Type:       AlarmTypeEquipmentFailure,
				Level:      AlarmLevelWarning,
				EntityType: EntityTypeGasOutlet,
				EntityID:   outletID,
				RelatedData: map[string]interface{}{
					"factor": factor,
					"value":  value,
				},
				Message:  "设备异常数据",
				CreatedAt: now,
			}
			alarms = append(alarms, alarm)
		} else {
			if value > standard.Limit {
				measurement.IsExceeded = true
				exceededFactors = append(exceededFactors, factor)
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
	
	m.outletLock.Lock()
	now = time.Now()
	
	if len(exceededFactors) > 0 {
		if outlet.ContinuousExceedingStart == nil {
			start := now
			outlet.ContinuousExceedingStart = &start
		} else if now.Sub(*outlet.ContinuousExceedingStart) >= 24*time.Hour {
			for _, factor := range exceededFactors {
				if alarmExists == nil || !alarmExists(EntityTypeGasOutlet, outletID, factor, AlarmTypeEmergency, now) {
					alarm := &Alarm{
						ID:         uuid.New().String(),
						Type:       AlarmTypeEmergency,
						Level:      AlarmLevelEmergency,
						EntityType: EntityTypeGasOutlet,
						EntityID:   outletID,
						RelatedData: map[string]interface{}{
							"factor": factor,
						},
						Message:  "连续超标超过24小时，紧急告警",
						CreatedAt: now,
					}
					alarms = append(alarms, alarm)
				}
			}
			outlet.Status = OutletStatusEmergency
		} else {
			for _, factor := range exceededFactors {
				if alarmExists == nil || !alarmExists(EntityTypeGasOutlet, outletID, factor, AlarmTypeExceeding, now) {
					alarm := &Alarm{
						ID:         uuid.New().String(),
						Type:       AlarmTypeExceeding,
						Level:      AlarmLevelWarning,
						EntityType: EntityTypeGasOutlet,
						EntityID:   outletID,
						RelatedData: map[string]interface{}{
							"factor": factor,
						},
						Message:  "超标告警",
						CreatedAt: now,
					}
					alarms = append(alarms, alarm)
				}
			}
			outlet.Status = OutletStatusWarning
		}
	} else {
		if report.DataStatus != DataStatusAllAbnormal {
			outlet.ContinuousExceedingStart = nil
			outlet.Status = OutletStatusNormal
		}
	}
	
	outlet.LastReport = &now
	m.outletLock.Unlock()
	
	m.reportLock.Lock()
	m.reports = append(m.reports, report)
	m.reportLock.Unlock()
	
	return report, alarms, nil
}

func (m *GasMonitor) GetReports(outletID string, from, to time.Time) []*GasReport {
	m.reportLock.RLock()
	defer m.reportLock.RUnlock()
	
	reports := make([]*GasReport, 0)
	for _, report := range m.reports {
		if (outletID == "" || report.OutletID == outletID) &&
			report.ReportedAt.After(from) && report.ReportedAt.Before(to) {
			reports = append(reports, report)
		}
	}
	return reports
}

package core

import (
	"sync"
	"time"
)

type AlarmManager struct {
	alarms []*Alarm
	lock   sync.RWMutex
}

func NewAlarmManager() *AlarmManager {
	return &AlarmManager{
		alarms: make([]*Alarm, 0),
	}
}

func (m *AlarmManager) AddAlarm(alarm *Alarm) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.alarms = append(m.alarms, alarm)
}

func (m *AlarmManager) AddAlarms(alarms []*Alarm) {
	if len(alarms) == 0 {
		return
	}
	
	m.lock.Lock()
	defer m.lock.Unlock()
	m.alarms = append(m.alarms, alarms...)
}

func (m *AlarmManager) GetAlarms(entityType *EntityType, entityID string, level *AlarmLevel, resolved *bool) []*Alarm {
	m.lock.RLock()
	defer m.lock.RUnlock()
	
	alarms := make([]*Alarm, 0)
	for _, alarm := range m.alarms {
		if entityType != nil && alarm.EntityType != *entityType {
			continue
		}
		if entityID != "" && alarm.EntityID != entityID {
			continue
		}
		if level != nil && alarm.Level != *level {
			continue
		}
		if resolved != nil && alarm.IsResolved != *resolved {
			continue
		}
		alarms = append(alarms, alarm)
	}
	return alarms
}

func (m *AlarmManager) ResolveAlarm(id string) (*Alarm, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	
	for _, alarm := range m.alarms {
		if alarm.ID == id && !alarm.IsResolved {
			now := time.Now()
			alarm.ResolvedAt = &now
			alarm.IsResolved = true
			return alarm, true
		}
	}
	return nil, false
}

func (m *AlarmManager) GetActiveAlarms() []*Alarm {
	return m.GetAlarms(nil, "", nil, boolPtr(false))
}

func (m *AlarmManager) GetEmergencyAlarms() []*Alarm {
	level := AlarmLevelEmergency
	return m.GetAlarms(nil, "", &level, boolPtr(false))
}

func (m *AlarmManager) HasAlarmToday(entityType EntityType, entityID, factor string, alarmType AlarmType, now time.Time) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	for _, alarm := range m.alarms {
		if alarm.EntityType != entityType ||
			alarm.EntityID != entityID ||
			alarm.Type != alarmType {
			continue
		}
		
		if !alarm.CreatedAt.After(startOfDay) {
			continue
		}
		
		if alarmFactor, exists := alarm.RelatedData["factor"].(string); exists && alarmFactor == factor {
			return true
		}
	}
	return false
}

func (m *AlarmManager) HasDailyAlarmToday(entityType EntityType, entityID, dateKey string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	
	for _, alarm := range m.alarms {
		if alarm.EntityType != entityType ||
			alarm.EntityID != entityID ||
			alarm.Type != AlarmTypeExceeding {
			continue
		}
		
		if alarmDate, exists := alarm.RelatedData["date"].(string); exists && alarmDate == dateKey {
			return true
		}
	}
	return false
}

func boolPtr(b bool) *bool {
	return &b
}

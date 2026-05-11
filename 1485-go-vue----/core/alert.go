package core

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

const ConsecutiveAlertsThreshold = 3

func generateAlertID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type AlertManager struct {
	store *Store
}

func NewAlertManager(store *Store) *AlertManager {
	return &AlertManager{store: store}
}

func (am *AlertManager) CheckAndGenerateAlert(taskID string, data *SensorData) (*Alert, error) {
	task, err := am.store.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	shouldMonitor, err := ShouldMonitor(task.CargoType)
	if err != nil {
		return nil, err
	}

	if !shouldMonitor {
		return nil, nil
	}

	limits := TemperatureLimits{
		Min: task.TempMin,
		Max: task.TempMax,
	}

	if IsTemperatureWithinLimits(data.Temperature, limits) {
		return nil, nil
	}

	alertLevel, err := am.determineAlertLevel(taskID, data.Timestamp)
	if err != nil {
		return nil, err
	}

	alert := &Alert{
		ID:          generateAlertID(),
		TaskID:      taskID,
		Timestamp:   data.Timestamp,
		Temperature: data.Temperature,
		Level:       alertLevel,
	}

	added := am.store.AddAlert(taskID, alert)
	if !added {
		return nil, nil
	}

	return alert, nil
}

func (am *AlertManager) determineAlertLevel(taskID string, currentTime time.Time) (AlertLevel, error) {
	recentData, err := am.store.GetRecentSensorData(taskID, ConsecutiveAlertsThreshold)
	if err != nil {
		return "", err
	}

	if len(recentData) < ConsecutiveAlertsThreshold {
		return AlertLevelWarning, nil
	}

	task, err := am.store.GetTask(taskID)
	if err != nil {
		return "", err
	}

	limits := TemperatureLimits{
		Min: task.TempMin,
		Max: task.TempMax,
	}

	consecutiveOutOfRange := 0
	for i := len(recentData) - 1; i >= 0 && consecutiveOutOfRange < ConsecutiveAlertsThreshold; i-- {
		if !IsTemperatureWithinLimits(recentData[i].Temperature, limits) {
			consecutiveOutOfRange++
		} else {
			break
		}
	}

	if consecutiveOutOfRange >= ConsecutiveAlertsThreshold {
		return AlertLevelCritical, nil
	}

	return AlertLevelWarning, nil
}

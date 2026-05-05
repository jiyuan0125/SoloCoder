package server

import (
	"time"

	"go-mailing-list/pkg/protocol"
)

type AlertManager struct {
	store *Store
}

func NewAlertManager(store *Store) *AlertManager {
	return &AlertManager{
		store: store,
	}
}

func (am *AlertManager) CheckBounceRate(listID string) {
	list, exists := am.store.GetMailingList(listID)
	if !exists {
		return
	}

	if list.IsPaused {
		return
	}

	bounceRate := am.calculateBounceRate(listID)

	if bounceRate > protocol.BounceRateThreshold {
		am.triggerBounceAlert(listID, bounceRate)
		list.IsPaused = true
		am.store.UpdateMailingList(list)
	}
}

func (am *AlertManager) calculateBounceRate(listID string) float64 {
	tasks := am.store.ListTasks()

	var totalSent, totalBounced int

	for _, task := range tasks {
		if task.ListID != listID {
			continue
		}

		records := am.store.GetSendRecords(task.ID)
		for _, record := range records {
			if record.Status == protocol.SendStatusSuccess {
				totalSent++
			}
			if record.Status == protocol.SendStatusBounced {
				totalBounced++
				totalSent++
			}
		}
	}

	if totalSent == 0 {
		return 0.0
	}

	return float64(totalBounced) / float64(totalSent)
}

func (am *AlertManager) triggerBounceAlert(listID string, bounceRate float64) {
	activeAlerts := am.store.GetActiveAlerts()
	for _, alert := range activeAlerts {
		if alert.ListID == listID && alert.Message == "High bounce rate detected" {
			return
		}
	}

	alert := &protocol.Alert{
		ID:        generateID(),
		ListID:    listID,
		Level:     protocol.AlertLevelError,
		Message:   "High bounce rate detected",
		Resolved:  false,
		CreatedAt: time.Now(),
	}

	am.store.CreateAlert(alert)
}

func (am *AlertManager) CreateAlert(listID, taskID string, level protocol.AlertLevel, message string) {
	alert := &protocol.Alert{
		ID:        generateID(),
		ListID:    listID,
		TaskID:    taskID,
		Level:     level,
		Message:   message,
		Resolved:  false,
		CreatedAt: time.Now(),
	}

	am.store.CreateAlert(alert)
}

func (am *AlertManager) GetActiveAlerts() []*protocol.Alert {
	return am.store.GetActiveAlerts()
}

func (am *AlertManager) ResolveAlert(alertID string) bool {
	return am.store.ResolveAlert(alertID)
}

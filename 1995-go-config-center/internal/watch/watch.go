package watch

import (
	"bytes"
	"config-center/internal/model"
	"config-center/internal/store"
	"encoding/json"
	"net/http"
	"time"
)

var retryIntervals = []time.Duration{5 * time.Second, 15 * time.Second}

type Manager struct {
	s           *store.Store
	httpClient  *http.Client
}

func NewManager(s *store.Store) *Manager {
	return &Manager{
		s:          s,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *Manager) HandleChange(event store.ChangeEvent) {
	watches := m.s.GetActiveWatchesForEvent(event)
	for _, w := range watches {
		go m.notifyWatch(w, event)
	}
}

func (m *Manager) notifyWatch(w *model.WatchRegistration, event store.ChangeEvent) {
	payload := model.WatchPayload{
		Env:      event.Env,
		Project:  event.Project,
		Key:      event.Key,
		OldValue: event.OldValue,
		NewValue: event.NewValue,
		Version:  event.Version,
		Time:     event.Time,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		m.markFailure(w.ID)
		return
	}

	success := false
	for attempt := 0; attempt <= len(retryIntervals); attempt++ {
		if attempt > 0 {
			time.Sleep(retryIntervals[attempt-1])
		}

		req, err := http.NewRequest(http.MethodPost, w.Callback, bytes.NewReader(body))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := m.httpClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			success = true
			break
		}
	}

	if success {
		m.s.MarkWatchSuccess(w.ID)
	} else {
		m.markFailure(w.ID)
	}
}

func (m *Manager) markFailure(watchID string) {
	m.s.MarkWatchFailure(watchID)
}

package subscribers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"circuit-breaker/models"
)

type Manager struct {
	mu          sync.RWMutex
	subscribers map[string]map[string]models.Subscriber
	client      *http.Client
}

func NewManager() *Manager {
	return &Manager{
		subscribers: make(map[string]map[string]models.Subscriber),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (m *Manager) Subscribe(breakerName string, subscriber models.Subscriber) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, ok := m.subscribers[breakerName]; !ok {
		m.subscribers[breakerName] = make(map[string]models.Subscriber)
	}
	
	m.subscribers[breakerName][subscriber.ID] = subscriber
}

func (m *Manager) Unsubscribe(breakerName string, subscriberID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if subs, ok := m.subscribers[breakerName]; ok {
		delete(subs, subscriberID)
	}
}

func (m *Manager) List(breakerName string) []models.Subscriber {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	subs, ok := m.subscribers[breakerName]
	if !ok {
		return []models.Subscriber{}
	}
	
	list := make([]models.Subscriber, 0, len(subs))
	for _, s := range subs {
		list = append(list, s)
	}
	return list
}

func (m *Manager) Broadcast(breakerName string, event models.Event) {
	m.mu.RLock()
	subs, ok := m.subscribers[breakerName]
	if !ok {
		m.mu.RUnlock()
		return
	}
	
	list := make([]models.Subscriber, 0, len(subs))
	for _, s := range subs {
		list = append(list, s)
	}
	m.mu.RUnlock()
	
	for _, sub := range list {
		go m.notify(sub, event)
	}
}

func (m *Manager) notify(sub models.Subscriber, event models.Event) {
	body, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("failed to marshal event: %v\n", err)
		return
	}
	
	req, err := http.NewRequest(http.MethodPost, sub.URL, bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("failed to create request for subscriber %s: %v\n", sub.ID, err)
		return
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := m.client.Do(req)
	if err != nil {
		fmt.Printf("failed to notify subscriber %s: %v\n", sub.ID, err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		fmt.Printf("subscriber %s returned error status: %d\n", sub.ID, resp.StatusCode)
	}
}

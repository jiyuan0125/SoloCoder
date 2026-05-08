package watcher

import (
	"sync"
)

type Manager struct {
	watchers  map[string]*DirectoryWatcher
	watchersMu sync.RWMutex
	events    []Event
	eventsMu  sync.RWMutex
	maxEvents int
}

func NewManager() *Manager {
	return &Manager{
		watchers:  make(map[string]*DirectoryWatcher),
		events:    make([]Event, 0),
		maxEvents: 100,
	}
}

func (m *Manager) AddWatch(id string, config WatchConfig) error {
	watcher, err := NewDirectoryWatcher(config)
	if err != nil {
		return err
	}

	if err := watcher.Start(); err != nil {
		return err
	}

	m.watchersMu.Lock()
	m.watchers[id] = watcher
	m.watchersMu.Unlock()

	return nil
}

func (m *Manager) RemoveWatch(id string) error {
	m.watchersMu.Lock()
	defer m.watchersMu.Unlock()

	watcher, exists := m.watchers[id]
	if !exists {
		return nil
	}

	if err := watcher.Stop(); err != nil {
		return err
	}

	delete(m.watchers, id)
	return nil
}

func (m *Manager) GetWatch(id string) (*DirectoryWatcher, bool) {
	m.watchersMu.RLock()
	defer m.watchersMu.RUnlock()

	watcher, exists := m.watchers[id]
	return watcher, exists
}

func (m *Manager) ListWatches() map[string]*DirectoryWatcher {
	m.watchersMu.RLock()
	defer m.watchersMu.RUnlock()

	result := make(map[string]*DirectoryWatcher, len(m.watchers))
	for id, w := range m.watchers {
		result[id] = w
	}
	return result
}

func (m *Manager) StopAll() error {
	m.watchersMu.Lock()
	defer m.watchersMu.Unlock()

	var firstErr error
	for id, w := range m.watchers {
		if err := w.Stop(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(m.watchers, id)
	}

	return firstErr
}

func (m *Manager) AddEvent(event Event) {
	m.eventsMu.Lock()
	defer m.eventsMu.Unlock()

	m.events = append(m.events, event)
	if len(m.events) > m.maxEvents {
		m.events = m.events[len(m.events)-m.maxEvents:]
	}
}

func (m *Manager) GetRecentEvents() []Event {
	m.eventsMu.RLock()
	defer m.eventsMu.RUnlock()

	result := make([]Event, len(m.events))
	copy(result, m.events)
	return result
}

func (m *Manager) CreateCallback() func(event interface{}) {
	return func(event interface{}) {
		if e, ok := event.(*Event); ok {
			m.AddEvent(*e)
		}
	}
}

func (m *Manager) CreateCallbackWithCustom(custom func(event interface{})) func(event interface{}) {
	return func(event interface{}) {
		if e, ok := event.(*Event); ok {
			m.AddEvent(*e)
		}
		if custom != nil {
			custom(event)
		}
	}
}

type ManagerConfig struct {
	MaxEvents int
}

func NewManagerWithConfig(config ManagerConfig) *Manager {
	maxEvents := config.MaxEvents
	if maxEvents <= 0 {
		maxEvents = 100
	}

	return &Manager{
		watchers:  make(map[string]*DirectoryWatcher),
		events:    make([]Event, 0),
		maxEvents: maxEvents,
	}
}

func (m *Manager) WatchCount() int {
	m.watchersMu.RLock()
	defer m.watchersMu.RUnlock()
	return len(m.watchers)
}

func (m *Manager) EventCount() int {
	m.eventsMu.RLock()
	defer m.eventsMu.RUnlock()
	return len(m.events)
}

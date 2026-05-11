package config

import (
	"log"
	"strings"
	"sync"
	"time"
)

type listenerEntry struct {
	id       string
	prefix   string
	callback ListenerFunc
}

type Manager struct {
	options     Options
	config      map[string]interface{}
	configMu    sync.RWMutex
	listeners   map[string]listenerEntry
	listenersMu sync.RWMutex
	watcher     *fileWatcher
	history     []ChangeRecord
	historyMu   sync.Mutex
	historyCap  int
	refMap      map[string][]string
}

func NewManager(options Options) (*Manager, error) {
	if options.StableWaitTime <= 0 {
		options.StableWaitTime = 2 * time.Second
	}

	m := &Manager{
		options:    options,
		listeners:  make(map[string]listenerEntry),
		history:    make([]ChangeRecord, 0),
		historyCap: 100,
	}

	if err := m.reload(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) Start() error {
	w, err := newFileWatcher(m.options.FilePath, m.options.StableWaitTime, func() {
		m.reload()
	})
	if err != nil {
		return err
	}

	m.watcher = w
	return w.start()
}

func (m *Manager) Stop() {
	if m.watcher != nil {
		m.watcher.stop()
	}
}

func (m *Manager) Reload() error {
	return m.reload()
}

func (m *Manager) Get() map[string]interface{} {
	m.configMu.RLock()
	defer m.configMu.RUnlock()
	return deepCopy(m.config)
}

func (m *Manager) GetField(path string) (interface{}, bool) {
	m.configMu.RLock()
	defer m.configMu.RUnlock()
	val := getValueByPath(m.config, path)
	return val, val != nil
}

func (m *Manager) Subscribe(prefix string, callback ListenerFunc) string {
	m.listenersMu.Lock()
	defer m.listenersMu.Unlock()

	id := generateListenerID()
	m.listeners[id] = listenerEntry{
		id:       id,
		prefix:   prefix,
		callback: callback,
	}
	return id
}

func (m *Manager) Unsubscribe(id string) {
	m.listenersMu.Lock()
	defer m.listenersMu.Unlock()
	delete(m.listeners, id)
}

func (m *Manager) History() []ChangeRecord {
	m.historyMu.Lock()
	defer m.historyMu.Unlock()

	result := make([]ChangeRecord, len(m.history))
	copy(result, m.history)
	return result
}

func (m *Manager) reload() error {
	newConfig, newRefMap, err := parseFile(m.options.FilePath, m.options.Format, m.options.EnableRef)
	if err != nil {
		return err
	}

	m.configMu.RLock()
	oldConfig := m.config
	m.configMu.RUnlock()

	var changes []Change
	if oldConfig != nil {
		changes = computeDiff(oldConfig, newConfig, newRefMap)
	} else {
		changes = make([]Change, 0)
	}

	if oldConfig == nil || len(changes) > 0 {
		m.configMu.Lock()
		m.config = newConfig
		m.refMap = newRefMap
		m.configMu.Unlock()

		if len(changes) > 0 {
			m.addHistory(changes)
			m.notifyListeners(changes)
		}
	}

	return nil
}

func (m *Manager) addHistory(changes []Change) {
	m.historyMu.Lock()
	defer m.historyMu.Unlock()

	record := ChangeRecord{
		Timestamp: time.Now(),
		Changes:   changes,
	}

	m.history = append(m.history, record)

	if len(m.history) > m.historyCap {
		m.history = m.history[len(m.history)-m.historyCap:]
	}
}

func (m *Manager) notifyListeners(changes []Change) {
	m.listenersMu.RLock()
	listeners := make([]listenerEntry, 0, len(m.listeners))
	for _, l := range m.listeners {
		listeners = append(listeners, l)
	}
	m.listenersMu.RUnlock()

	for _, change := range changes {
		for _, l := range listeners {
			if matchesListener(change.Path, l.prefix) {
				safeInvoke(l.callback, change)
			}

			for _, affected := range change.Affected {
				if matchesListener(affected, l.prefix) {
					safeInvoke(l.callback, Change{
						Path: affected,
						Old:  nil,
						New:  nil,
						Type: ChangeTypeModify,
					})
				}
			}
		}
	}
}

func safeInvoke(fn ListenerFunc, change Change) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("config listener panic for path %s: %v", change.Path, r)
		}
	}()

	fn(change)
}

func matchesListener(path, prefix string) bool {
	if prefix == "" {
		return true
	}

	if path == prefix {
		return true
	}

	if strings.HasPrefix(path, prefix+".") {
		return true
	}

	return false
}

func generateListenerID() string {
	return time.Now().Format("20060102150405000000000") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	now := time.Now().UnixNano()
	for i := range b {
		b[i] = letters[int(now)%len(letters)]
		now = now >> 1
	}
	return string(b)
}

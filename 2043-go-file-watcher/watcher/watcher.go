package watcher

import (
	"bytes"
	"encoding/json"
	"file-watcher/models"
	"file-watcher/repository"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type WatcherManager struct {
	watcher        *fsnotify.Watcher
	watches        map[uint]*DirectoryWatcher
	mutex          sync.RWMutex
	eventChan      chan EventInfo
	stopChan       chan struct{}
	httpClient     *http.Client
}

type DirectoryWatcher struct {
	ID        uint
	Path      string
	Recursive bool
	Callback  string
	Status    string
	watching  bool
}

type EventInfo struct {
	WatchID   uint
	Path      string
	Event     string
	Timestamp time.Time
}

type pendingEvent struct {
	lastEvent string
	timer     *time.Timer
}

var (
	manager   *WatcherManager
	once      sync.Once
	pendingMu sync.Mutex
	pending   = make(map[string]*pendingEvent)
)

const EventMergeWindow = 5 * time.Second
const DirectoryCheckInterval = 10 * time.Second

func GetManager() *WatcherManager {
	once.Do(func() {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatalf("Failed to create watcher: %v", err)
		}

		manager = &WatcherManager{
			watcher:    w,
			watches:    make(map[uint]*DirectoryWatcher),
			eventChan:  make(chan EventInfo, 1000),
			stopChan:   make(chan struct{}),
			httpClient: &http.Client{Timeout: 10 * time.Second},
		}

		go manager.run()
		go manager.startDirectoryChecker()
	})
	return manager
}

func (m *WatcherManager) run() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			m.handleEvent(event)
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		case eventInfo := <-m.eventChan:
			m.processEvent(eventInfo)
		case <-m.stopChan:
			return
		}
	}
}

func (m *WatcherManager) handleEvent(event fsnotify.Event) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var matchedWatcher *DirectoryWatcher
	for _, dw := range m.watches {
		if dw.Status != "active" || !dw.watching {
			continue
		}
		if event.Name == dw.Path || isSubPath(event.Name, dw.Path) {
			matchedWatcher = dw
			break
		}
	}

	if matchedWatcher == nil {
		return
	}

	eventType := getEventType(event)
	if eventType == "" {
		return
	}

	if eventType == "create" && matchedWatcher.Recursive {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			m.watchNewDirectoryRecursively(event.Name)
		}
	}

	eventInfo := EventInfo{
		WatchID:   matchedWatcher.ID,
		Path:      event.Name,
		Event:     eventType,
		Timestamp: time.Now(),
	}

	m.eventChan <- eventInfo
}

func (m *WatcherManager) processEvent(eventInfo EventInfo) {
	key := getEventKey(eventInfo.WatchID, eventInfo.Path)

	pendingMu.Lock()
	defer pendingMu.Unlock()

	if pe, exists := pending[key]; exists {
		if !pe.timer.Stop() {
			select {
			case <-pe.timer.C:
			default:
			}
		}
		if eventInfo.Event == "delete" {
			pe.lastEvent = "delete"
		} else if eventInfo.Event == "create" {
			pe.lastEvent = "create"
		}
	} else {
		pending[key] = &pendingEvent{
			lastEvent: eventInfo.Event,
		}
	}

	pending[key].timer = time.AfterFunc(EventMergeWindow, func() {
		m.sendNotification(key, eventInfo.WatchID)
	})
}

func (m *WatcherManager) sendNotification(key string, watchID uint) {
	pendingMu.Lock()
	pe, exists := pending[key]
	if !exists {
		pendingMu.Unlock()
		return
	}
	delete(pending, key)
	pendingMu.Unlock()

	m.mutex.RLock()
	dw, exists := m.watches[watchID]
	m.mutex.RUnlock()

	if !exists || dw.Status != "active" {
		return
	}

	notification := models.NotificationEvent{
		WatchID:   watchID,
		Path:      key[len(string(watchID))+1:],
		Event:     pe.lastEvent,
		Timestamp: time.Now().Unix(),
	}

	go func() {
		body, err := json.Marshal(notification)
		if err != nil {
			log.Printf("Failed to marshal notification: %v", err)
			return
		}

		req, err := http.NewRequest("POST", dw.Callback, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("Failed to create request: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := m.httpClient.Do(req)
		if err != nil {
			log.Printf("Failed to send notification to %s: %v", dw.Callback, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			log.Printf("Notification to %s returned status %d", dw.Callback, resp.StatusCode)
		}
	}()
}

func (m *WatcherManager) startDirectoryChecker() {
	ticker := time.NewTicker(DirectoryCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkDirectories()
		case <-m.stopChan:
			return
		}
	}
}

func (m *WatcherManager) checkDirectories() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, dw := range m.watches {
		if _, err := os.Stat(dw.Path); os.IsNotExist(err) {
			if dw.Status == "active" {
				log.Printf("Directory %s not found, pausing watcher %d", dw.Path, dw.ID)
				dw.Status = "paused"
				if err := repository.UpdateWatchDirectoryStatus(dw.ID, "paused"); err != nil {
					log.Printf("Failed to update watcher status: %v", err)
				}
				m.removeWatchPaths(dw)
			}
		} else if err == nil {
			if dw.Status == "paused" {
				log.Printf("Directory %s recovered, resuming watcher %d", dw.Path, dw.ID)
				dw.Status = "active"
				if err := repository.UpdateWatchDirectoryStatus(dw.ID, "active"); err != nil {
					log.Printf("Failed to update watcher status: %v", err)
				}
				if err := m.addWatchPaths(dw); err != nil {
					log.Printf("Failed to resume watcher: %v", err)
				}
			}
		}
	}
}

func (m *WatcherManager) AddWatch(path string, recursive bool, callback string) (*DirectoryWatcher, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	dir, err := repository.AddWatchDirectory(path, recursive, callback)
	if err != nil {
		return nil, err
	}

	dw := &DirectoryWatcher{
		ID:        dir.ID,
		Path:      path,
		Recursive: recursive,
		Callback:  callback,
		Status:    "active",
		watching:  false,
	}

	m.watches[dir.ID] = dw

	if err := m.addWatchPaths(dw); err != nil {
		log.Printf("Warning: some paths may not be watched: %v", err)
	}

	return dw, nil
}

func (m *WatcherManager) addWatchPaths(dw *DirectoryWatcher) error {
	dw.watching = true

	if err := m.watcher.Add(dw.Path); err != nil {
		log.Printf("Failed to add watch for %s: %v", dw.Path, err)
		dw.watching = false
		return err
	}

	if dw.Recursive {
		m.walkAndWatch(dw)
	}

	return nil
}

func (m *WatcherManager) walkAndWatch(dw *DirectoryWatcher) {
	filepath.Walk(dw.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				log.Printf("Warning: permission denied for %s, skipping", path)
				return filepath.SkipDir
			}
			log.Printf("Warning: error walking %s: %v", path, err)
			return nil
		}

		if info.IsDir() && path != dw.Path {
			if err := m.watcher.Add(path); err != nil {
				log.Printf("Warning: failed to add watch for %s: %v", path, err)
			}
		}

		return nil
	})
}

func (m *WatcherManager) watchNewDirectoryRecursively(root string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				log.Printf("Warning: permission denied for %s, skipping", path)
				return filepath.SkipDir
			}
			log.Printf("Warning: error walking %s: %v", path, err)
			return nil
		}

		if info.IsDir() {
			if err := m.watcher.Add(path); err != nil {
				log.Printf("Warning: failed to add watch for %s: %v", path, err)
			} else {
				log.Printf("Added watch for newly created directory: %s", path)
			}
		}

		return nil
	})
}

func (m *WatcherManager) removeWatchPaths(dw *DirectoryWatcher) {
	dw.watching = false
	m.watcher.Remove(dw.Path)
}

func (m *WatcherManager) RemoveWatch(id uint) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	dw, exists := m.watches[id]
	if !exists {
		return nil
	}

	m.removeWatchPaths(dw)
	delete(m.watches, id)
	return repository.DeleteWatchDirectory(id)
}

func (m *WatcherManager) LoadExistingWatches() error {
	dirs, err := repository.GetAllActiveWatchDirectories()
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		dw := &DirectoryWatcher{
			ID:        dir.ID,
			Path:      dir.Path,
			Recursive: dir.Recursive,
			Callback:  dir.Callback,
			Status:    "active",
			watching:  false,
		}

		m.watches[dir.ID] = dw

		if _, err := os.Stat(dir.Path); err == nil {
			if err := m.addWatchPaths(dw); err != nil {
				log.Printf("Warning: some paths may not be watched: %v", err)
			}
		} else {
			log.Printf("Directory %s not found, watcher %d is paused", dir.Path, dir.ID)
			dw.Status = "paused"
		}
	}

	return nil
}

func (m *WatcherManager) Stop() {
	close(m.stopChan)
	m.watcher.Close()
}

func getEventType(event fsnotify.Event) string {
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		return "create"
	case event.Op&fsnotify.Write == fsnotify.Write:
		return "modify"
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		return "delete"
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		return "rename"
	default:
		return ""
	}
}

func isSubPath(path, base string) bool {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return false
	}
	return rel != "." && !filepath.IsAbs(rel)
}

func getEventKey(watchID uint, path string) string {
	return string(watchID) + ":" + path
}

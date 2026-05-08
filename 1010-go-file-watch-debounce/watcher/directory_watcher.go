package watcher

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type DirectoryWatcher struct {
	config     WatchConfig
	fsWatcher  *fsnotify.Watcher
	pending    map[string]*pendingEvent
	pendingMu  sync.Mutex
	active     bool
	activeMu   sync.RWMutex
	stopChan   chan struct{}
	renameFrom map[string]string
	renameMu   sync.Mutex
}

func NewDirectoryWatcher(config WatchConfig) (*DirectoryWatcher, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	info, err := os.Stat(config.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrPathNotExist
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, ErrPathNotDirectory
	}

	w := &DirectoryWatcher{
		config:     config,
		pending:    make(map[string]*pendingEvent),
		stopChan:   make(chan struct{}),
		renameFrom: make(map[string]string),
	}

	return w, nil
}

func (w *DirectoryWatcher) Start() error {
	w.activeMu.Lock()
	if w.active {
		w.activeMu.Unlock()
		return nil
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		w.activeMu.Unlock()
		return err
	}

	w.fsWatcher = fsWatcher
	w.active = true
	w.activeMu.Unlock()

	if err := fsWatcher.Add(w.config.Path); err != nil {
		w.Stop()
		return err
	}

	go w.run()
	return nil
}

func (w *DirectoryWatcher) Stop() error {
	w.activeMu.Lock()
	if !w.active {
		w.activeMu.Unlock()
		return nil
	}
	w.active = false
	w.activeMu.Unlock()

	close(w.stopChan)

	if w.fsWatcher != nil {
		err := w.fsWatcher.Close()
		w.fsWatcher = nil
		return err
	}

	w.pendingMu.Lock()
	for path, p := range w.pending {
		if p.debounceTimer != nil {
			p.debounceTimer.Stop()
		}
		if p.maxWaitTimer != nil {
			p.maxWaitTimer.Stop()
		}
		delete(w.pending, path)
	}
	w.pendingMu.Unlock()

	return nil
}

func (w *DirectoryWatcher) IsActive() bool {
	w.activeMu.RLock()
	defer w.activeMu.RUnlock()
	return w.active
}

func (w *DirectoryWatcher) Config() WatchConfig {
	return w.config
}

func (w *DirectoryWatcher) run() {
	for {
		select {
		case <-w.stopChan:
			return
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)
		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			_ = err
		}
	}
}

func (w *DirectoryWatcher) handleEvent(event fsnotify.Event) {
	eventName := event.Name
	baseName := filepath.Base(eventName)
	if baseName == "" || baseName[0] == '.' {
		return
	}

	info, err := os.Lstat(eventName)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		return
	}

	if event.Op&fsnotify.Rename != 0 {
		w.handleRenameEvent(eventName)
		return
	}

	w.processEvent(eventName, event.Op.String())
}

func (w *DirectoryWatcher) handleRenameEvent(path string) {
	w.renameMu.Lock()
	defer w.renameMu.Unlock()

	fromPath, exists := w.renameFrom[path]
	if exists {
		delete(w.renameFrom, path)
		w.processRenameComplete(fromPath, path)
	} else {
		w.renameFrom[path] = path
		go func() {
			time.Sleep(100 * time.Millisecond)
			w.renameMu.Lock()
			if currentFrom, ok := w.renameFrom[path]; ok && currentFrom == path {
				delete(w.renameFrom, path)
				w.processEvent(path, "RENAME")
			}
			w.renameMu.Unlock()
		}()
	}
}

func (w *DirectoryWatcher) processRenameComplete(fromPath, toPath string) {
	w.pendingMu.Lock()
	defer w.pendingMu.Unlock()

	event := &Event{
		Path:      toPath,
		OldPath:   fromPath,
		Type:      EventRename,
		Timestamp: time.Now(),
	}
	w.config.Callback(event)
}

func (w *DirectoryWatcher) processEvent(path string, eventType string) {
	w.pendingMu.Lock()
	defer w.pendingMu.Unlock()

	p, exists := w.pending[path]
	if !exists {
		p = &pendingEvent{
			path:      path,
			firstSeen: time.Now(),
		}
		w.pending[path] = p
	}

	p.addEventType(eventType)

	if p.debounceTimer != nil {
		p.debounceTimer.Stop()
	}

	p.debounceTimer = time.AfterFunc(w.config.Debounce, func() {
		w.triggerEvent(path)
	})

	if p.maxWaitTimer == nil {
		p.maxWaitTimer = time.AfterFunc(w.config.MaxWait, func() {
			w.triggerEvent(path)
		})
	}
}

func (w *DirectoryWatcher) triggerEvent(path string) {
	w.pendingMu.Lock()
	defer w.pendingMu.Unlock()

	p, exists := w.pending[path]
	if !exists {
		return
	}

	if p.debounceTimer != nil {
		p.debounceTimer.Stop()
		p.debounceTimer = nil
	}
	if p.maxWaitTimer != nil {
		p.maxWaitTimer.Stop()
		p.maxWaitTimer = nil
	}

	delete(w.pending, path)

	event := w.classifyEvent(p)
	if event != nil {
		w.config.Callback(event)
	}
}

func (w *DirectoryWatcher) classifyEvent(p *pendingEvent) *Event {
	hasCreate := p.hasEventType("CREATE")
	hasWrite := p.hasEventType("WRITE")
	hasRemove := p.hasEventType("REMOVE")
	hasRename := p.hasEventType("RENAME")
	hasChmod := p.hasEventType("CHMOD")

	event := &Event{
		Path:      p.path,
		OldPath:   p.oldPath,
		RawTypes:  p.getRawTypes(),
		Timestamp: time.Now(),
	}

	if hasRemove && hasCreate {
		event.Type = EventReplace
		return event
	}

	if hasRemove {
		event.Type = EventDelete
		return event
	}

	if hasCreate {
		event.Type = EventCreate
		return event
	}

	if hasWrite || hasChmod {
		event.Type = EventModify
		return event
	}

	if hasRename {
		event.Type = EventRename
		return event
	}

	event.Type = EventModify
	return event
}

package filewatcher

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	config        IgnoreConfig
	ignoreMatcher *IgnoreMatcher
	debouncer     *Debouncer
	fsWatcher     *fsnotify.Watcher
	done          chan struct{}
	wg            sync.WaitGroup
	watchedDirs   map[string]bool
	pendingRename map[string]bool
	mu            sync.Mutex
}

func NewWatcher() (*Watcher, error) {
	return NewWatcherWithConfig(IgnoreConfig{})
}

func NewWatcherWithConfig(config IgnoreConfig) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		config:        config,
		ignoreMatcher: NewIgnoreMatcher(config),
		debouncer:     NewDebouncer(),
		fsWatcher:     fsWatcher,
		done:          make(chan struct{}),
		watchedDirs:   make(map[string]bool),
		pendingRename: make(map[string]bool),
	}, nil
}

func (w *Watcher) Watch(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	if err := w.ignoreMatcher.LoadGitignore(absDir); err != nil {
		return err
	}

	if err := w.addDirRecursive(absDir); err != nil {
		return err
	}

	w.wg.Add(1)
	go w.run()

	return nil
}

func (w *Watcher) addDirRecursive(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if w.ignoreMatcher.ShouldIgnore(path, true) {
				return filepath.SkipDir
			}

			w.mu.Lock()
			if !w.watchedDirs[path] {
				if err := w.fsWatcher.Add(path); err != nil {
					w.mu.Unlock()
					return err
				}
				w.watchedDirs[path] = true
			}
			w.mu.Unlock()
		}

		return nil
	})
}

func (w *Watcher) run() {
	defer w.wg.Done()

	for {
		select {
		case <-w.done:
			return
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)
		case _, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	isDir := false
	info, err := os.Stat(event.Name)
	if err == nil {
		isDir = info.IsDir()
	} else if os.IsNotExist(err) {
		if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
			w.mu.Lock()
			if w.watchedDirs[event.Name] {
				delete(w.watchedDirs, event.Name)
				w.fsWatcher.Remove(event.Name)
			}
			w.mu.Unlock()
		}
	}

	if w.ignoreMatcher.ShouldIgnore(event.Name, isDir) {
		return
	}

	if event.Op&fsnotify.Rename != 0 {
		w.handleRenameEvent(event.Name, isDir, err == nil)
		return
	}

	if event.Op&fsnotify.Create != 0 {
		if isDir {
			w.addDirRecursive(event.Name)
		}

		w.mu.Lock()
		var oldPath string
		for pendingPath := range w.pendingRename {
			oldPath = pendingPath
			delete(w.pendingRename, pendingPath)
			break
		}
		w.mu.Unlock()

		if oldPath != "" {
			w.debouncer.Add(Rename, event.Name, oldPath)
		} else {
			w.debouncer.Add(Create, event.Name, "")
		}
		return
	}

	var eventType EventType
	switch {
	case event.Op&fsnotify.Write != 0:
		eventType = Write
	case event.Op&fsnotify.Remove != 0:
		eventType = Remove
	case event.Op&fsnotify.Chmod != 0:
		eventType = Chmod
	default:
		return
	}

	w.debouncer.Add(eventType, event.Name, "")
}

func (w *Watcher) handleRenameEvent(path string, isDir bool, exists bool) {
	if exists {
		w.mu.Lock()
		var oldPath string
		for pendingPath := range w.pendingRename {
			oldPath = pendingPath
			delete(w.pendingRename, pendingPath)
			break
		}
		w.mu.Unlock()

		if oldPath != "" {
			w.debouncer.Add(Rename, path, oldPath)
		} else {
			w.debouncer.Add(Rename, path, "")
		}
	} else {
		w.mu.Lock()
		w.pendingRename[path] = true
		w.mu.Unlock()

		w.debouncer.Add(Rename, "", path)
	}
}

func (w *Watcher) Events() <-chan Event {
	return w.debouncer.Events()
}

func (w *Watcher) Close() {
	close(w.done)
	w.debouncer.Close()
	w.fsWatcher.Close()
	w.wg.Wait()
}

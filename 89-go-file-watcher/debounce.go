package filewatcher

import (
	"sync"
	"time"
)

const DebounceDuration = 200 * time.Millisecond

type EventType int

const (
	Create EventType = iota
	Write
	Remove
	Rename
	Chmod
)

func (et EventType) String() string {
	switch et {
	case Create:
		return "Create"
	case Write:
		return "Write"
	case Remove:
		return "Remove"
	case Rename:
		return "Rename"
	case Chmod:
		return "Chmod"
	default:
		return "Unknown"
	}
}

type Event struct {
	Type    EventType
	Path    string
	OldPath string
}

type pendingEntry struct {
	events      []EventType
	path        string
	oldPath     string
	isRenameSrc bool
	timer       *time.Timer
}

type Debouncer struct {
	eventsOut    chan Event
	pending      map[string]*pendingEntry
	renameSrcMap map[string]bool
	mu           sync.Mutex
	done         chan struct{}
}

func NewDebouncer() *Debouncer {
	return &Debouncer{
		eventsOut:    make(chan Event, 100),
		pending:      make(map[string]*pendingEntry),
		renameSrcMap: make(map[string]bool),
		done:         make(chan struct{}),
	}
}

func (d *Debouncer) Events() <-chan Event {
	return d.eventsOut
}

func (d *Debouncer) Add(eventType EventType, path string, oldPath string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	select {
	case <-d.done:
		return
	default:
	}

	switch eventType {
	case Rename:
		d.handleRename(eventType, path, oldPath)
	default:
		d.handleNormal(eventType, path, oldPath)
	}
}

func (d *Debouncer) handleRename(eventType EventType, path string, oldPath string) {
	if oldPath != "" && path != "" {
		d.addCompleteRename(path, oldPath)
	} else if oldPath != "" && path == "" {
		d.addRenameSource(oldPath)
	} else if path != "" && oldPath == "" {
		d.addRenameDest(path)
	}
}

func (d *Debouncer) addCompleteRename(path string, oldPath string) {
	delete(d.renameSrcMap, oldPath)

	if oldEntry, ok := d.pending[oldPath]; ok {
		oldEntry.timer.Stop()
		delete(d.pending, oldPath)
	}

	entry, exists := d.pending[path]
	if exists {
		entry.timer.Stop()
		entry.events = append(entry.events, Rename)
		entry.oldPath = oldPath
	} else {
		entry = &pendingEntry{
			events:  []EventType{Rename},
			path:    path,
			oldPath: oldPath,
		}
		d.pending[path] = entry
	}

	entry.timer = time.AfterFunc(DebounceDuration, func() {
		d.process(path)
	})
}

func (d *Debouncer) addRenameSource(oldPath string) {
	d.renameSrcMap[oldPath] = true

	entry, exists := d.pending[oldPath]
	if exists {
		entry.timer.Stop()
		entry.events = append(entry.events, Rename)
		entry.isRenameSrc = true
	} else {
		entry = &pendingEntry{
			events:      []EventType{Rename},
			path:        "",
			oldPath:     oldPath,
			isRenameSrc: true,
		}
		d.pending[oldPath] = entry
	}

	entry.timer = time.AfterFunc(DebounceDuration, func() {
		d.process(oldPath)
	})
}

func (d *Debouncer) addRenameDest(path string) {
	var oldPath string
	for src := range d.renameSrcMap {
		oldPath = src
		delete(d.renameSrcMap, src)
		break
	}

	if oldPath != "" {
		if srcEntry, ok := d.pending[oldPath]; ok {
			srcEntry.timer.Stop()
			delete(d.pending, oldPath)
		}
	}

	entry, exists := d.pending[path]
	if exists {
		entry.timer.Stop()
		entry.events = append(entry.events, Rename)
		if oldPath != "" {
			entry.oldPath = oldPath
		}
	} else {
		entry = &pendingEntry{
			events:  []EventType{Rename},
			path:    path,
			oldPath: oldPath,
		}
		d.pending[path] = entry
	}

	entry.timer = time.AfterFunc(DebounceDuration, func() {
		d.process(path)
	})
}

func (d *Debouncer) handleNormal(eventType EventType, path string, oldPath string) {
	entry, exists := d.pending[path]
	if exists {
		entry.timer.Stop()
		entry.events = append(entry.events, eventType)
		if oldPath != "" {
			entry.oldPath = oldPath
		}
	} else {
		entry = &pendingEntry{
			events:  []EventType{eventType},
			path:    path,
			oldPath: oldPath,
		}
		d.pending[path] = entry
	}

	entry.timer = time.AfterFunc(DebounceDuration, func() {
		d.process(path)
	})
}

func (d *Debouncer) process(path string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	entry, exists := d.pending[path]
	if !exists {
		return
	}
	delete(d.pending, path)

	if entry.isRenameSrc {
		delete(d.renameSrcMap, path)
	}

	if len(entry.events) == 0 {
		return
	}

	finalType := d.mergeEvents(entry.events)

	if finalType == -1 {
		return
	}

	event := Event{
		Type: finalType,
		Path: entry.path,
	}

	if finalType == Rename {
		if entry.oldPath != "" {
			event.OldPath = entry.oldPath
		} else if entry.isRenameSrc {
			event.OldPath = path
		}
	}

	select {
	case <-d.done:
		return
	default:
		d.eventsOut <- event
	}
}

func (d *Debouncer) mergeEvents(events []EventType) EventType {
	if len(events) == 0 {
		return -1
	}

	hasCreate := false
	hasRemove := false
	hasWrite := false
	hasChmod := false
	hasRename := false

	for _, e := range events {
		switch e {
		case Create:
			hasCreate = true
		case Remove:
			hasRemove = true
		case Write:
			hasWrite = true
		case Chmod:
			hasChmod = true
		case Rename:
			hasRename = true
		}
	}

	if hasRename {
		return Rename
	}

	if hasCreate && hasRemove {
		return -1
	}

	if hasCreate && (hasWrite || hasChmod) {
		return Write
	}

	if hasRemove {
		return Remove
	}

	if hasWrite {
		return Write
	}

	if hasCreate {
		return Create
	}

	if hasChmod {
		return Chmod
	}

	return events[len(events)-1]
}

func (d *Debouncer) Close() {
	close(d.done)

	d.mu.Lock()
	for path, entry := range d.pending {
		if entry.timer != nil {
			entry.timer.Stop()
		}
		delete(d.pending, path)
	}
	for k := range d.renameSrcMap {
		delete(d.renameSrcMap, k)
	}
	d.mu.Unlock()

	close(d.eventsOut)
}

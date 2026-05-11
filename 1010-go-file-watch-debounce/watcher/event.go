package watcher

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

type EventType string

const (
	EventCreate  EventType = "create"
	EventModify  EventType = "modify"
	EventDelete  EventType = "delete"
	EventRename  EventType = "rename"
	EventReplace EventType = "replace"
)

type Event struct {
	Path      string
	OldPath   string
	Type      EventType
	RawOps    []fsnotify.Op
	Timestamp time.Time
}

type pendingEvent struct {
	path          string
	oldPath       string
	ops           fsnotify.Op
	firstSeen     time.Time
	debounceTimer *time.Timer
	maxWaitTimer  *time.Timer
}

func (p *pendingEvent) addOp(op fsnotify.Op) {
	p.ops |= op
}

func (p *pendingEvent) hasOp(op fsnotify.Op) bool {
	return p.ops&op != 0
}

func (p *pendingEvent) getRawOps() []fsnotify.Op {
	var ops []fsnotify.Op
	allOps := []fsnotify.Op{
		fsnotify.Create,
		fsnotify.Write,
		fsnotify.Remove,
		fsnotify.Rename,
		fsnotify.Chmod,
	}
	for _, op := range allOps {
		if p.hasOp(op) {
			ops = append(ops, op)
		}
	}
	return ops
}

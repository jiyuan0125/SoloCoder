package watcher

import "time"

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
	RawTypes  []string
	Timestamp time.Time
}

type pendingEvent struct {
	path         string
	oldPath      string
	eventTypes   map[string]struct{}
	firstSeen    time.Time
	debounceTimer *time.Timer
	maxWaitTimer *time.Timer
}

func (p *pendingEvent) addEventType(eventType string) {
	if p.eventTypes == nil {
		p.eventTypes = make(map[string]struct{})
	}
	p.eventTypes[eventType] = struct{}{}
}

func (p *pendingEvent) hasEventType(eventType string) bool {
	_, exists := p.eventTypes[eventType]
	return exists
}

func (p *pendingEvent) getRawTypes() []string {
	types := make([]string, 0, len(p.eventTypes))
	for t := range p.eventTypes {
		types = append(types, t)
	}
	return types
}

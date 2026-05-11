package eventbus

import (
	"encoding/json"
	"sync"
	"time"
)

const (
	DefaultPriority      = 0
	DefaultCleanupInterval = 30 * time.Second
	DefaultInactiveTimeout = 5 * time.Minute
)

type EventStatus int

const (
	EventPending EventStatus = iota
	EventProcessing
	EventDone
	EventPartial
	EventBlocked
)

func (s EventStatus) String() string {
	switch s {
	case EventPending:
		return "等待中"
	case EventProcessing:
		return "处理中"
	case EventDone:
		return "已完成"
	case EventPartial:
		return "部分完成"
	case EventBlocked:
		return "已拦截"
	default:
		return "未知"
	}
}

type Event struct {
	ID        string
	Topic     string
	Payload   json.RawMessage
	Timestamp time.Time
}

type EventContext struct {
	event       *Event
	intercepted bool
	mu          sync.Mutex
}

func newEventContext(evt *Event) *EventContext {
	return &EventContext{event: evt}
}

func (c *EventContext) Event() *Event {
	return c.event
}

func (c *EventContext) Intercept() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.intercepted = true
}

func (c *EventContext) IsIntercepted() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.intercepted
}

type HandlerResult struct {
	Success bool
	Error   error
}

type Handler func(ctx *EventContext) HandlerResult

type Subscriber struct {
	ID         string
	Priority   int
	Handler    Handler
	Endpoint   string
	Registered time.Time
	lastSeen   time.Time
	active     bool
	mu         sync.RWMutex
}

func (s *Subscriber) Touch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastSeen = time.Now()
	s.active = true
}

func (s *Subscriber) LastSeen() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastSeen
}

func (s *Subscriber) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

func (s *Subscriber) MarkInactive() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = false
}

type EventState struct {
	Event       *Event
	Status      EventStatus
	Total       int
	Success     int
	Failed      int
	Intercepted bool
	finished    chan struct{}
	mu          sync.RWMutex
}

func newEventState(evt *Event) *EventState {
	return &EventState{
		Event:    evt,
		Status:   EventPending,
		finished: make(chan struct{}),
	}
}

func (s *EventState) Wait() {
	<-s.finished
}

func (s *EventState) Done() {
	select {
	case <-s.finished:
	default:
		close(s.finished)
	}
}

type Options struct {
	CleanupInterval time.Duration
	InactiveTimeout time.Duration
}

func DefaultOptions() *Options {
	return &Options{
		CleanupInterval: DefaultCleanupInterval,
		InactiveTimeout: DefaultInactiveTimeout,
	}
}

func (o *Options) applyDefaults() {
	if o.CleanupInterval <= 0 {
		o.CleanupInterval = DefaultCleanupInterval
	}
	if o.InactiveTimeout <= 0 {
		o.InactiveTimeout = DefaultInactiveTimeout
	}
}

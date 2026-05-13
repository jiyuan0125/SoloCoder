package events

import (
	"sync"

	"serviceregistry/internal/model"
)

type Subscriber interface {
	OnEvent(event *model.ServiceEvent)
	Close()
}

type EventBus struct {
	subscribers map[string]map[Subscriber]struct{}
	lock        sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string]map[Subscriber]struct{}),
	}
}

func (eb *EventBus) Subscribe(serviceName string, sub Subscriber) {
	eb.lock.Lock()
	defer eb.lock.Unlock()

	if _, ok := eb.subscribers[serviceName]; !ok {
		eb.subscribers[serviceName] = make(map[Subscriber]struct{})
	}
	eb.subscribers[serviceName][sub] = struct{}{}
}

func (eb *EventBus) Unsubscribe(serviceName string, sub Subscriber) {
	eb.lock.Lock()
	defer eb.lock.Unlock()

	if subs, ok := eb.subscribers[serviceName]; ok {
		delete(subs, sub)
		if len(subs) == 0 {
			delete(eb.subscribers, serviceName)
		}
	}
}

func (eb *EventBus) Publish(event *model.ServiceEvent) {
	if event.Instance == nil {
		return
	}

	eb.lock.RLock()
	defer eb.lock.RUnlock()

	subs, ok := eb.subscribers[event.Instance.ServiceName]
	if !ok {
		return
	}

	for sub := range subs {
		sub.OnEvent(event)
	}
}

func (eb *EventBus) HasSubscribers(serviceName string) bool {
	eb.lock.RLock()
	defer eb.lock.RUnlock()

	subs, ok := eb.subscribers[serviceName]
	return ok && len(subs) > 0
}

type ChannelSubscriber struct {
	ch     chan *model.ServiceEvent
	closed chan struct{}
	once   sync.Once
}

func NewChannelSubscriber(bufferSize int) *ChannelSubscriber {
	return &ChannelSubscriber{
		ch:     make(chan *model.ServiceEvent, bufferSize),
		closed: make(chan struct{}),
	}
}

func (c *ChannelSubscriber) OnEvent(event *model.ServiceEvent) {
	select {
	case c.ch <- event:
	case <-c.closed:
	default:
	}
}

func (c *ChannelSubscriber) Close() {
	c.once.Do(func() {
		close(c.closed)
	})
}

func (c *ChannelSubscriber) Events() <-chan *model.ServiceEvent {
	return c.ch
}

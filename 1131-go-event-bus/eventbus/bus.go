package eventbus

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

var (
	ErrTopicExist    = errors.New("topic already exists")
	ErrTopicNotExist = errors.New("topic does not exist")
	ErrSubNotExist   = errors.New("subscriber does not exist")
)

type EventBus struct {
	topics        map[string][]*Subscriber
	subIndex      map[string]map[string]struct{}
	events        map[string]*EventState
	cleanupStop   chan struct{}
	cleanupTicker *time.Ticker
	opts          *Options
	mu            sync.RWMutex
}

func New(opts *Options) *EventBus {
	if opts == nil {
		opts = DefaultOptions()
	}
	opts.applyDefaults()

	bus := &EventBus{
		topics:      make(map[string][]*Subscriber),
		subIndex:    make(map[string]map[string]struct{}),
		events:      make(map[string]*EventState),
		cleanupStop: make(chan struct{}),
		opts:        opts,
	}
	bus.startCleanup()
	return bus
}

func (b *EventBus) Close() {
	select {
	case <-b.cleanupStop:
	default:
		close(b.cleanupStop)
	}
	if b.cleanupTicker != nil {
		b.cleanupTicker.Stop()
	}
}

func (b *EventBus) CreateTopic(topic string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.topics[topic]; ok {
		return ErrTopicExist
	}
	b.topics[topic] = []*Subscriber{}
	return nil
}

func (b *EventBus) DeleteTopic(topic string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs, ok := b.topics[topic]
	if !ok {
		return ErrTopicNotExist
	}

	for _, s := range subs {
		if idx, ok := b.subIndex[s.ID]; ok {
			delete(idx, topic)
			if len(idx) == 0 {
				delete(b.subIndex, s.ID)
			}
		}
	}
	delete(b.topics, topic)
	return nil
}

func (b *EventBus) HasTopic(topic string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.topics[topic]
	return ok
}

func (b *EventBus) Subscribe(topic string, sub *Subscriber) error {
	if sub == nil || sub.ID == "" {
		return errors.New("invalid subscriber")
	}
	if sub.Handler == nil {
		return errors.New("subscriber handler is required")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	subs, ok := b.topics[topic]
	if !ok {
		return ErrTopicNotExist
	}

	for _, s := range subs {
		if s.ID == sub.ID {
			return errors.New("subscriber already registered for this topic")
		}
	}

	sub.Touch()
	sub.Registered = time.Now()

	subs = append(subs, sub)
	b.sortSubs(subs)
	b.topics[topic] = subs

	if _, ok := b.subIndex[sub.ID]; !ok {
		b.subIndex[sub.ID] = make(map[string]struct{})
	}
	b.subIndex[sub.ID][topic] = struct{}{}

	return nil
}

func (b *EventBus) Unsubscribe(topic, subID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs, ok := b.topics[topic]
	if !ok {
		return ErrTopicNotExist
	}

	idx := -1
	for i, s := range subs {
		if s.ID == subID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrSubNotExist
	}

	subs[idx].MarkInactive()
	subs = append(subs[:idx], subs[idx+1:]...)
	b.topics[topic] = subs

	if idx, ok := b.subIndex[subID]; ok {
		delete(idx, topic)
		if len(idx) == 0 {
			delete(b.subIndex, subID)
		}
	}
	return nil
}

func (b *EventBus) ListSubscribers(topic string) ([]*Subscriber, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	subs, ok := b.topics[topic]
	if !ok {
		return nil, ErrTopicNotExist
	}

	result := make([]*Subscriber, 0, len(subs))
	for _, s := range subs {
		result = append(result, s)
	}
	return result, nil
}

func (b *EventBus) Publish(topic string, payload interface{}, async bool) (*EventState, error) {
	var rawPayload []byte
	switch v := payload.(type) {
	case []byte:
		rawPayload = v
	default:
		data, err := jsonMarshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		rawPayload = data
	}

	b.mu.RLock()
	subs, ok := b.topics[topic]
	b.mu.RUnlock()
	if !ok {
		return nil, ErrTopicNotExist
	}

	event := &Event{
		ID:        genEventID(),
		Topic:     topic,
		Payload:   rawPayload,
		Timestamp: time.Now(),
	}

	state := newEventState(event)
	state.Total = len(subs)

	b.mu.Lock()
	b.events[event.ID] = state
	b.mu.Unlock()

	if async {
		go b.dispatch(state, subs)
		return state, nil
	}

	b.dispatch(state, subs)
	return state, nil
}

func (b *EventBus) GetEventState(eventID string) (*EventState, error) {
	b.mu.RLock()
	state, ok := b.events[eventID]
	b.mu.RUnlock()
	if !ok {
		return nil, errors.New("event not found")
	}
	return state, nil
}

func (b *EventBus) dispatch(state *EventState, subs []*Subscriber) {
	state.mu.Lock()
	state.Status = EventProcessing
	state.mu.Unlock()

	ctx := newEventContext(state.Event)
	successCount := 0
	failedCount := 0
	intercepted := false

	for _, sub := range subs {
		if !sub.IsActive() {
			continue
		}

		if ctx.IsIntercepted() {
			intercepted = true
			break
		}

		sub.Touch()
		result := sub.Handler(ctx)
		if result.Success {
			successCount++
		} else {
			failedCount++
			if result.Error != nil {
				log.Printf("subscriber %s handle error: %v", sub.ID, result.Error)
			}
		}

		if ctx.IsIntercepted() {
			intercepted = true
			break
		}
	}

	state.mu.Lock()
	state.Success = successCount
	state.Failed = failedCount
	state.Intercepted = intercepted

	switch {
	case intercepted:
		state.Status = EventBlocked
	case failedCount == 0:
		state.Status = EventDone
	case successCount > 0:
		state.Status = EventPartial
	default:
		state.Status = EventPartial
	}
	state.mu.Unlock()

	state.Done()
}

func (b *EventBus) sortSubs(subs []*Subscriber) {
	sort.SliceStable(subs, func(i, j int) bool {
		return subs[i].Priority > subs[j].Priority
	})
}

func (b *EventBus) startCleanup() {
	b.cleanupTicker = time.NewTicker(b.opts.CleanupInterval)
	go func() {
		for {
			select {
			case <-b.cleanupStop:
				return
			case <-b.cleanupTicker.C:
				b.cleanupInactive()
			}
		}
	}()
}

func (b *EventBus) cleanupInactive() {
	now := time.Now()
	timeout := b.opts.InactiveTimeout

	b.mu.Lock()
	defer b.mu.Unlock()

	for topic, subs := range b.topics {
		filtered := subs[:0]
		for _, s := range subs {
			if !s.IsActive() || now.Sub(s.LastSeen()) > timeout {
				s.MarkInactive()
				if idx, ok := b.subIndex[s.ID]; ok {
					delete(idx, topic)
					if len(idx) == 0 {
						delete(b.subIndex, s.ID)
					}
				}
				continue
			}
			filtered = append(filtered, s)
		}
		b.topics[topic] = filtered
	}

	cutoff := now.Add(-1 * time.Hour)
	for id, state := range b.events {
		if state.Status != EventPending && state.Status != EventProcessing {
			if state.Event.Timestamp.Before(cutoff) {
				delete(b.events, id)
			}
		}
	}
}

func genEventID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

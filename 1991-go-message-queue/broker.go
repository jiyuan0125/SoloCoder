package main

import (
	"fmt"
	"sync"
	"time"
)

func NewBroker() *Broker {
	return &Broker{
		topics: make(map[string]*Topic),
		nextID: 1,
	}
}

func (b *Broker) GetOrCreateTopic(name string) *Topic {
	b.mu.RLock()
	t, ok := b.topics[name]
	b.mu.RUnlock()
	if ok {
		return t
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if t, ok = b.topics[name]; ok {
		return t
	}

	t = &Topic{
		Name:      name,
		Messages:  make([]Message, 0, DefaultTopicCapacity),
		StartID:   1,
		Capacity:  DefaultTopicCapacity,
		Groups:    make(map[string]*GroupState),
		DeadLetter: make([]Message, 0),
	}
	b.topics[name] = t
	return t
}

func (b *Broker) GetTopic(name string) (*Topic, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	t, ok := b.topics[name]
	return t, ok
}

func (b *Broker) nextMessageID() uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	id := b.nextID
	b.nextID++
	return id
}

func (b *Broker) Publish(topicName string, body []byte) uint64 {
	topic := b.GetOrCreateTopic(topicName)
	
	topic.mu.Lock()
	defer topic.mu.Unlock()

	id := b.nextMessageID()

	msg := Message{
		ID:        id,
		Topic:     topicName,
		Body:      append([]byte(nil), body...),
		Timestamp: time.Now().Unix(),
	}

	if len(topic.Messages) >= topic.Capacity {
		topic.Messages = topic.Messages[1:]
		topic.StartID++
	}

	topic.Messages = append(topic.Messages, msg)

	for _, g := range topic.Groups {
		if g.LastConsumed == 0 {
			g.LastConsumed = topic.StartID - 1
		}
	}

	return id
}

func (t *Topic) GetGroup(name string) (*GroupState, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	g, ok := t.Groups[name]
	return g, ok
}

func (t *Topic) GetOrCreateGroup(name string) *GroupState {
	t.mu.Lock()
	defer t.mu.Unlock()

	g, ok := t.Groups[name]
	if !ok {
		g = &GroupState{
			Name:         name,
			Topic:        t.Name,
			LastConsumed: t.StartID - 1,
			Consumers:    make(map[string]*Consumer),
			Pending:      make(map[uint64]*PendingMessage),
		}
		t.Groups[name] = g
	}
	return g
}

func (g *GroupState) RegisterConsumer(consumerID string) (*Consumer, error) {
	if consumerID == "" {
		return nil, fmt.Errorf("consumer ID cannot be empty")
	}

	g.muLock()
	defer g.muUnlock()

	c, exists := g.Consumers[consumerID]
	if !exists {
		c = &Consumer{
			ID:       consumerID,
			Group:    g.Name,
			Topic:    g.Topic,
			LastSeen: time.Now().Unix(),
		}
		g.Consumers[consumerID] = c
	} else {
		c.LastSeen = time.Now().Unix()
	}
	return c, nil
}

func (g *GroupState) UnregisterConsumer(consumerID string) {
	g.muLock()
	defer g.muUnlock()

	_, ok := g.Consumers[consumerID]
	if !ok {
		return
	}

	var toRequeue []uint64
	for id, pending := range g.Pending {
		if pending.ConsumerID == consumerID {
			toRequeue = append(toRequeue, id)
		}
	}

	for _, id := range toRequeue {
		delete(g.Pending, id)
		if g.LastConsumed >= id {
			g.LastConsumed = id - 1
		}
	}

	delete(g.Consumers, consumerID)
}

func (g *GroupState) Heartbeat(consumerID string) bool {
	g.muLock()
	defer g.muUnlock()

	c, ok := g.Consumers[consumerID]
	if !ok {
		return false
	}
	c.LastSeen = time.Now().Unix()
	return true
}

func (g *GroupState) CleanupExpiredConsumers() {
	g.muLock()
	defer g.muUnlock()

	now := time.Now().Unix()
	var expired []string

	for id, c := range g.Consumers {
		if now-c.LastSeen > DefaultConsumerTTL {
			expired = append(expired, id)
		}
	}

	for _, id := range expired {
		var toRequeue []uint64
		for msgID, pending := range g.Pending {
			if pending.ConsumerID == id {
				toRequeue = append(toRequeue, msgID)
			}
		}

		for _, msgID := range toRequeue {
			delete(g.Pending, msgID)
			if g.LastConsumed >= msgID {
				g.LastConsumed = msgID - 1
			}
		}

		delete(g.Consumers, id)
	}
}

func (g *GroupState) HasConsumer(consumerID string) bool {
	g.muRLock()
	defer g.muRUnlock()
	_, ok := g.Consumers[consumerID]
	return ok
}

func (g *GroupState) ConsumerCount() int {
	g.muRLock()
	defer g.muRUnlock()
	return len(g.Consumers)
}

var (
	groupMuMap   = make(map[string]*sync.RWMutex)
	groupMuMapMu sync.Mutex
)

func getGroupMu(key string) *sync.RWMutex {
	groupMuMapMu.Lock()
	defer groupMuMapMu.Unlock()
	mu, ok := groupMuMap[key]
	if !ok {
		mu = &sync.RWMutex{}
		groupMuMap[key] = mu
	}
	return mu
}

func (g *GroupState) muLock() {
	getGroupMu(g.Topic + ":" + g.Name).Lock()
}

func (g *GroupState) muUnlock() {
	getGroupMu(g.Topic + ":" + g.Name).Unlock()
}

func (g *GroupState) muRLock() {
	getGroupMu(g.Topic + ":" + g.Name).RLock()
}

func (g *GroupState) muRUnlock() {
	getGroupMu(g.Topic + ":" + g.Name).RUnlock()
}

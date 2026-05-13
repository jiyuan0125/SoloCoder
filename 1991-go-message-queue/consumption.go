package main

import (
	"time"
)

func (g *GroupState) Consume(topic *Topic, consumerID string, maxMessages int) []Message {
	g.muLock()
	defer g.muUnlock()

	if _, ok := g.Consumers[consumerID]; !ok {
		return nil
	}

	var result []Message
	if maxMessages <= 0 {
		maxMessages = 1
	}

	now := time.Now().Unix()

	topic.mu.RLock()
	messages := topic.Messages
	startID := topic.StartID
	topic.mu.RUnlock()

	for len(result) < maxMessages {
		nextID := g.LastConsumed + 1

		if nextID < startID {
			g.LastConsumed = startID - 1
			nextID = startID
		}

		if nextID-startID >= uint64(len(messages)) {
			break
		}

		msg := messages[nextID-startID]

		if pending, exists := g.Pending[msg.ID]; exists {
			if now >= pending.ExpireAt {
				pending.ConsumerID = consumerID
				pending.ExpireAt = now + int64(DefaultAckTimeout)
				result = append(result, pending.Message)
			}
			g.LastConsumed = msg.ID
			continue
		}

		pending := &PendingMessage{
			Message:    msg,
			ConsumerID: consumerID,
			ExpireAt:   now + int64(DefaultAckTimeout),
			RetryCount: 0,
		}
		g.Pending[msg.ID] = pending
		g.LastConsumed = msg.ID
		result = append(result, msg)
	}

	return result
}

func (g *GroupState) Ack(topic *Topic, consumerID string, msgID uint64) bool {
	g.muLock()
	defer g.muUnlock()

	pending, exists := g.Pending[msgID]
	if !exists {
		return false
	}

	if pending.ConsumerID != consumerID {
		return false
	}

	delete(g.Pending, msgID)
	return true
}

func (g *GroupState) Nack(topic *Topic, consumerID string, msgID uint64) bool {
	g.muLock()
	defer g.muUnlock()

	pending, exists := g.Pending[msgID]
	if !exists {
		return false
	}

	if pending.ConsumerID != consumerID {
		return false
	}

	pending.RetryCount++
	if pending.RetryCount >= maxRetryCount {
		topic.mu.Lock()
		topic.DeadLetter = append(topic.DeadLetter, pending.Message)
		topic.mu.Unlock()
		delete(g.Pending, msgID)
		return true
	}

	interval := retryIntervals[pending.RetryCount-1]
	pending.ExpireAt = time.Now().Unix() + int64(interval)
	return true
}

func (g *GroupState) RequeueExpired(topic *Topic) {
	g.muLock()
	defer g.muUnlock()

	now := time.Now().Unix()

	for id, pending := range g.Pending {
		if now >= pending.ExpireAt {
			pending.RetryCount++
			if pending.RetryCount >= maxRetryCount {
				topic.mu.Lock()
				topic.DeadLetter = append(topic.DeadLetter, pending.Message)
				topic.mu.Unlock()
				delete(g.Pending, id)
				continue
			}

			interval := retryIntervals[pending.RetryCount-1]
			pending.ExpireAt = now + int64(interval)

			if g.LastConsumed >= id {
				g.LastConsumed = id - 1
			}
		}
	}
}

func (t *Topic) GetDeadLetter() []Message {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]Message, len(t.DeadLetter))
	copy(result, t.DeadLetter)
	return result
}

func (t *Topic) RepublishDeadLetter(msgID uint64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	idx := -1
	var targetMsg Message
	for i, msg := range t.DeadLetter {
		if msg.ID == msgID {
			idx = i
			targetMsg = msg
			break
		}
	}

	if idx == -1 {
		return false
	}

	t.DeadLetter = append(t.DeadLetter[:idx], t.DeadLetter[idx+1:]...)

	if len(t.Messages) >= t.Capacity {
		t.Messages = t.Messages[1:]
		t.StartID++
	}

	newMsg := Message{
		ID:        msgID,
		Topic:     t.Name,
		Body:      append([]byte(nil), targetMsg.Body...),
		Timestamp: time.Now().Unix(),
	}

	insertPos := -1
	for i, msg := range t.Messages {
		if msg.ID > msgID {
			insertPos = i
			break
		}
	}

	if insertPos == -1 {
		t.Messages = append(t.Messages, newMsg)
	} else {
		t.Messages = append(t.Messages[:insertPos], append([]Message{newMsg}, t.Messages[insertPos:]...)...)
	}

	for _, g := range t.Groups {
		g.muLock()
		if g.LastConsumed >= msgID {
			g.LastConsumed = msgID - 1
		}
		g.muUnlock()
	}

	return true
}

func (t *Topic) GetMessagesByIDRange(start, end uint64) []Message {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.Messages) == 0 {
		return nil
	}

	if end == 0 {
		end = t.Messages[len(t.Messages)-1].ID
	}

	if start < t.StartID {
		start = t.StartID
	}

	startIdx := start - t.StartID
	endIdx := end - t.StartID

	if startIdx >= uint64(len(t.Messages)) {
		return nil
	}

	if endIdx >= uint64(len(t.Messages)) {
		endIdx = uint64(len(t.Messages)) - 1
	}

	result := make([]Message, 0, endIdx-startIdx+1)
	for i := startIdx; i <= endIdx && i < uint64(len(t.Messages)); i++ {
		result = append(result, t.Messages[i])
	}

	return result
}

func (t *Topic) GetStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["topic"] = t.Name
	stats["capacity"] = t.Capacity
	stats["total_messages"] = len(t.Messages)
	stats["dead_letter_size"] = len(t.DeadLetter)

	if len(t.Messages) > 0 {
		stats["start_id"] = t.StartID
		stats["latest_id"] = t.Messages[len(t.Messages)-1].ID
	} else {
		stats["start_id"] = uint64(0)
		stats["latest_id"] = uint64(0)
	}

	groupStats := make(map[string]interface{})
	for name, g := range t.Groups {
		gs := make(map[string]interface{})
		g.muRLock()
		gs["last_consumed_id"] = g.LastConsumed
		gs["consumer_count"] = len(g.Consumers)
		gs["pending_count"] = len(g.Pending)
		consumerIDs := make([]string, 0, len(g.Consumers))
		for id := range g.Consumers {
			consumerIDs = append(consumerIDs, id)
		}
		gs["consumers"] = consumerIDs
		g.muRUnlock()
		groupStats[name] = gs
	}
	stats["groups"] = groupStats

	return stats
}

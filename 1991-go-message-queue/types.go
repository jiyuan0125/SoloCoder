package main

import (
	"sync"
)

type Message struct {
	ID        uint64
	Topic     string
	Body      []byte
	Timestamp int64
}

type PendingMessage struct {
	Message
	ConsumerID string
	ExpireAt   int64
	RetryCount int
}

type Consumer struct {
	ID         string
	Group      string
	Topic      string
	LastSeen   int64
}

type GroupState struct {
	Name         string
	Topic        string
	LastConsumed uint64
	Consumers    map[string]*Consumer
	Pending      map[uint64]*PendingMessage
}

type Topic struct {
	Name      string
	Messages  []Message
	StartID   uint64
	Capacity  int
	Groups    map[string]*GroupState
	DeadLetter []Message
	mu        sync.RWMutex
}

type Broker struct {
	topics map[string]*Topic
	nextID uint64
	mu     sync.RWMutex
}

const (
	DefaultTopicCapacity = 1000000
	DefaultAckTimeout    = 30
	DefaultConsumerTTL   = 300
)

var (
	retryIntervals = []int{5, 15, 45}
	maxRetryCount  = 3
)

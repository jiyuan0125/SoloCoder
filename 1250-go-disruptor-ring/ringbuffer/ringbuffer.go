package ringbuffer

import (
	"errors"
	"sync"
	"sync/atomic"

	"disruptor/common"
)

const (
	initialSequence = int64(-1)
)

type RingBuffer struct {
	name          string
	capacity      int64
	mask          int64
	buffer        []string
	producerCursor int64
	consumerSeqs  sync.Map
	mu            sync.RWMutex
}

func nextPowerOfTwo(size int64) int64 {
	if size <= 0 {
		return 1
	}
	size--
	size |= size >> 1
	size |= size >> 2
	size |= size >> 4
	size |= size >> 8
	size |= size >> 16
	size |= size >> 32
	size++
	return size
}

func NewRingBuffer(name string, size int64) (*RingBuffer, error) {
	if size > common.MaxBufferSize {
		return nil, errors.New(common.ErrInvalidSize)
	}

	capacity := nextPowerOfTwo(size)
	if capacity > common.MaxBufferSize {
		return nil, errors.New(common.ErrInvalidSize)
	}

	return &RingBuffer{
		name:          name,
		capacity:      capacity,
		mask:          capacity - 1,
		buffer:        make([]string, capacity),
		producerCursor: initialSequence,
	}, nil
}

func (rb *RingBuffer) Name() string {
	return rb.name
}

func (rb *RingBuffer) Capacity() int64 {
	return rb.capacity
}

func (rb *RingBuffer) minConsumerSequence() int64 {
	minSeq := atomic.LoadInt64(&rb.producerCursor)
	rb.consumerSeqs.Range(func(key, value interface{}) bool {
		seq := value.(int64)
		if seq < minSeq {
			minSeq = seq
		}
		return true
	})
	return minSeq
}

func (rb *RingBuffer) Publish(message string) (int64, error) {
	currentProducer := atomic.LoadInt64(&rb.producerCursor)
	nextSeq := currentProducer + 1
	minConsumer := rb.minConsumerSequence()

	if nextSeq-minConsumer > rb.capacity {
		return 0, errors.New(common.ErrBufferFull)
	}

	if !atomic.CompareAndSwapInt64(&rb.producerCursor, currentProducer, nextSeq) {
		return rb.Publish(message)
	}

	index := nextSeq & rb.mask
	rb.buffer[index] = message
	return nextSeq, nil
}

func (rb *RingBuffer) BatchPublish(messages []string) (int64, int, error) {
	if len(messages) == 0 {
		return 0, 0, nil
	}

	currentProducer := atomic.LoadInt64(&rb.producerCursor)
	count := int64(len(messages))
	startSeq := currentProducer + 1
	endSeq := currentProducer + count
	minConsumer := rb.minConsumerSequence()

	if endSeq-minConsumer > rb.capacity {
		return 0, 0, errors.New(common.ErrBufferFull)
	}

	if !atomic.CompareAndSwapInt64(&rb.producerCursor, currentProducer, endSeq) {
		return rb.BatchPublish(messages)
	}

	for i, msg := range messages {
		seq := startSeq + int64(i)
		index := seq & rb.mask
		rb.buffer[index] = msg
	}

	return startSeq, len(messages), nil
}

func (rb *RingBuffer) getOrCreateConsumerSeq(consumerID string) int64 {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if seq, ok := rb.consumerSeqs.Load(consumerID); ok {
		return seq.(int64)
	}
	rb.consumerSeqs.Store(consumerID, initialSequence)
	return initialSequence
}

func (rb *RingBuffer) Consume(consumerID string) (int64, string, error) {
	consumerSeq := rb.getOrCreateConsumerSeq(consumerID)
	nextSeq := consumerSeq + 1
	producerCursor := atomic.LoadInt64(&rb.producerCursor)

	if nextSeq > producerCursor {
		return 0, "", errors.New(common.ErrNoData)
	}

	index := nextSeq & rb.mask
	message := rb.buffer[index]
	rb.consumerSeqs.Store(consumerID, nextSeq)
	return nextSeq, message, nil
}

func (rb *RingBuffer) BatchConsume(consumerID string, maxCount int) ([]common.ConsumeResponse, error) {
	if maxCount <= 0 {
		return nil, nil
	}

	consumerSeq := rb.getOrCreateConsumerSeq(consumerID)
	nextSeq := consumerSeq + 1
	producerCursor := atomic.LoadInt64(&rb.producerCursor)

	if nextSeq > producerCursor {
		return nil, errors.New(common.ErrNoData)
	}

	available := producerCursor - nextSeq + 1
	actualCount := int64(maxCount)
	if available < actualCount {
		actualCount = available
	}

	result := make([]common.ConsumeResponse, 0, actualCount)
	for i := int64(0); i < actualCount; i++ {
		seq := nextSeq + i
		index := seq & rb.mask
		result = append(result, common.ConsumeResponse{
			Sequence: seq,
			Message:  rb.buffer[index],
		})
	}

	lastSeq := nextSeq + actualCount - 1
	rb.consumerSeqs.Store(consumerID, lastSeq)
	return result, nil
}

func (rb *RingBuffer) Info() common.BufferInfo {
	consumers := make(map[string]int64)
	rb.consumerSeqs.Range(func(key, value interface{}) bool {
		consumers[key.(string)] = value.(int64)
		return true
	})

	producerCursor := atomic.LoadInt64(&rb.producerCursor)
	minConsumer := rb.minConsumerSequence()
	used := producerCursor - minConsumer
	if used < 0 {
		used = 0
	}

	return common.BufferInfo{
		Name:           rb.name,
		Capacity:       rb.capacity,
		Used:           used,
		ProducerCursor: producerCursor,
		Consumers:      consumers,
	}
}

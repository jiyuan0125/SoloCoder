package ringbuffer

import (
	"errors"
	"sync"

	"disruptor/common"
)

type Manager struct {
	buffers sync.Map
	mu      sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) CreateBuffer(name string, size int64) (int64, error) {
	if _, exists := m.buffers.Load(name); exists {
		return 0, errors.New(common.ErrBufferExists)
	}

	rb, err := NewRingBuffer(name, size)
	if err != nil {
		return 0, err
	}

	m.buffers.Store(name, rb)
	return rb.Capacity(), nil
}

func (m *Manager) DestroyBuffer(name string) error {
	if _, exists := m.buffers.Load(name); !exists {
		return errors.New(common.ErrBufferNotFound)
	}
	m.buffers.Delete(name)
	return nil
}

func (m *Manager) getBuffer(name string) (*RingBuffer, error) {
	if rb, exists := m.buffers.Load(name); exists {
		return rb.(*RingBuffer), nil
	}
	return nil, errors.New(common.ErrBufferNotFound)
}

func (m *Manager) Publish(bufferName string, message string) (int64, error) {
	rb, err := m.getBuffer(bufferName)
	if err != nil {
		return 0, err
	}
	return rb.Publish(message)
}

func (m *Manager) BatchPublish(bufferName string, messages []string) (int64, int, error) {
	rb, err := m.getBuffer(bufferName)
	if err != nil {
		return 0, 0, err
	}
	return rb.BatchPublish(messages)
}

func (m *Manager) Consume(bufferName string, consumerID string) (int64, string, error) {
	rb, err := m.getBuffer(bufferName)
	if err != nil {
		return 0, "", err
	}
	return rb.Consume(consumerID)
}

func (m *Manager) BatchConsume(bufferName string, consumerID string, maxCount int) ([]common.ConsumeResponse, error) {
	rb, err := m.getBuffer(bufferName)
	if err != nil {
		return nil, err
	}
	return rb.BatchConsume(consumerID, maxCount)
}

func (m *Manager) BufferInfo(bufferName string) (common.BufferInfo, error) {
	rb, err := m.getBuffer(bufferName)
	if err != nil {
		return common.BufferInfo{}, err
	}
	return rb.Info(), nil
}

func (m *Manager) ListBuffers() []string {
	names := make([]string, 0)
	m.buffers.Range(func(key, value interface{}) bool {
		names = append(names, key.(string))
		return true
	})
	return names
}

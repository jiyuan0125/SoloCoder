package leaf

import (
	"sync"
	"sync/atomic"
)

type MockCenterClient struct {
	mu         sync.Mutex
	nextStart  int64
	segmentSize int64
}

func NewMockCenterClient(segmentSize int64) *MockCenterClient {
	return &MockCenterClient{
		segmentSize: segmentSize,
		nextStart:   1,
	}
}

func (m *MockCenterClient) GetSegment() (*Segment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	start := atomic.LoadInt64(&m.nextStart)
	end := start + m.segmentSize - 1
	atomic.StoreInt64(&m.nextStart, end+1)
	return NewSegment(start, end), nil
}

func (m *MockCenterClient) ReportProgress(current, max int64) error {
	return nil
}

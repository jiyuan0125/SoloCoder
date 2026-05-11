package idgen

import (
	"sync"
	"time"

	"idgen/api"
)

type IDGenerator struct {
	mu         sync.Mutex
	nodeID     uint16
	lastTime   int64
	sequence   int64
}

func NewIDGenerator(nodeID uint16) *IDGenerator {
	return &IDGenerator{
		nodeID:   nodeID,
		lastTime: -1,
		sequence: 0,
	}
}

func (g *IDGenerator) NextID() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.nextID()
}

func (g *IDGenerator) BatchNextID(count int) ([]int64, error) {
	if count <= 0 || count > api.MaxBatchCount {
		return nil, ErrInvalidBatchCount
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	ids := make([]int64, 0, count)
	for i := 0; i < count; i++ {
		id, err := g.nextID()
		if err != nil {
			return nil, ErrBatchGenerationFail
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (g *IDGenerator) nextID() (int64, error) {
	now := currentTimestamp()

	if now < g.lastTime {
		backward := g.lastTime - now
		if backward <= api.ClockBackoffMax {
			time.Sleep(time.Duration(backward) * time.Millisecond)
			now = currentTimestamp()
			if now < g.lastTime {
				return 0, ErrClockBackward
			}
		} else {
			return 0, ErrClockBackward
		}
	}

	if now == g.lastTime {
		g.sequence = (g.sequence + 1) & MaxSequence
		if g.sequence == 0 {
			now = waitNextMillis(g.lastTime)
		}
	} else {
		g.sequence = 0
	}

	g.lastTime = now

	return ((now - Epoch) << (NodeIDBits + SequenceBits)) |
		(int64(g.nodeID) << SequenceBits) |
		g.sequence, nil
}

func currentTimestamp() int64 {
	return time.Now().UnixMilli()
}

func waitNextMillis(lastTime int64) int64 {
	now := currentTimestamp()
	for now <= lastTime {
		time.Sleep(1 * time.Millisecond)
		now = currentTimestamp()
	}
	return now
}

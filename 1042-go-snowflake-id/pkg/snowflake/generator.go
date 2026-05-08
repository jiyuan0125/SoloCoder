package snowflake

import (
	"sync"
	"time"
)

type Generator struct {
	mu          sync.Mutex
	config      *Config
	lastTime    int64
	sequence    int64
	machineID   int64
}

func NewGenerator(config *Config) (*Generator, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	machineID := config.MachineID
	if config.AutoMachineID {
		var err error
		machineID, err = GenerateMachineID()
		if err != nil {
			return nil, err
		}
	}

	return &Generator{
		config:    config,
		lastTime:  -1,
		sequence:  0,
		machineID: machineID,
	}, nil
}

func (g *Generator) Next() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.next()
}

func (g *Generator) NextBatch(count int) ([]int64, error) {
	if count <= 0 {
		return nil, nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	ids := make([]int64, 0, count)
	for i := 0; i < count; i++ {
		id, err := g.next()
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (g *Generator) next() (int64, error) {
	for {
		now := g.getCurrentTimestamp()

		if now < g.lastTime {
			backward := g.lastTime - now
			if backward <= g.config.MaxWaitMillis {
				time.Sleep(time.Duration(backward+1) * time.Millisecond)
				continue
			}
			return 0, ErrClockTooFarBack
		}

		if now == g.lastTime {
			g.sequence = (g.sequence + 1) & SequenceMask
			if g.sequence == 0 {
				for now <= g.lastTime {
					now = g.getCurrentTimestamp()
				}
			}
		} else {
			g.sequence = 0
		}

		g.lastTime = now
		break
	}

	id := g.assembleID(g.lastTime, g.machineID, g.sequence)
	return id, nil
}

func (g *Generator) assembleID(timestamp, machineID, sequence int64) int64 {
	return (timestamp << TimestampShift) |
		(machineID << MachineIDShift) |
		sequence
}

func (g *Generator) getCurrentTimestamp() int64 {
	return time.Now().UnixMilli() - g.config.Epoch
}

func (g *Generator) Status() Status {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := g.getCurrentTimestamp()
	return Status{
		LastTimestamp: g.lastTime,
		CurrentTimestamp: now,
		MachineID:     g.machineID,
		LastSequence:  g.sequence,
		Epoch:         g.config.Epoch,
		MaxSequence:   MaxSequence,
	}
}

func (g *Generator) MachineID() int64 {
	return g.machineID
}

func (g *Generator) Epoch() int64 {
	return g.config.Epoch
}

func (g *Generator) CurrentTime() int64 {
	return g.getCurrentTimestamp()
}

package serial

import (
	"fmt"
	"sync"
	"time"
)

type bizState struct {
	Date         string
	CurrentMax   int64
	Allocated    intervalList
	HistoryAlloc intervalList
}

type BizStatus struct {
	Date           string
	CurrentMax     int64
	TotalAllocated int64
}

type GapInfo struct {
	Start int64
	End   int64
}

type Generator struct {
	mu     sync.Mutex
	states map[string]*bizState
	nowFn  func() time.Time
}

func NewGenerator() *Generator {
	return &Generator{
		states: make(map[string]*bizState),
		nowFn:  time.Now,
	}
}

func (g *Generator) today() string {
	return g.nowFn().Format("20060102")
}

func (g *Generator) formatSerial(date string, seq int64) string {
	return fmt.Sprintf("%s-%06d", date, seq)
}

func (g *Generator) ensureBizState(bizType string) *bizState {
	if g.states[bizType] == nil {
		g.states[bizType] = &bizState{
			Date:         g.today(),
			CurrentMax:   0,
			Allocated:    newIntervalList(),
			HistoryAlloc: newIntervalList(),
		}
	}
	return g.states[bizType]
}

func (g *Generator) Next(bizType string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	state := g.ensureBizState(bizType)
	today := g.today()

	if state.Date != today {
		if len(state.Allocated) > 0 {
			start := state.Allocated[0].start
			end := state.Allocated[len(state.Allocated)-1].end
			state.HistoryAlloc.add(start, end)
		}
		state.Date = today
		state.CurrentMax = 0
		state.Allocated = newIntervalList()
	}

	state.CurrentMax++
	seq := state.CurrentMax
	state.Allocated.add(seq, seq)

	return g.formatSerial(state.Date, seq), nil
}

func (g *Generator) Batch(bizType string, count int) (startSeq int64, endSeq int64, date string, err error) {
	if count <= 0 {
		return 0, 0, "", fmt.Errorf("count must be positive")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	state := g.ensureBizState(bizType)
	today := g.today()

	if state.Date != today {
		if len(state.Allocated) > 0 {
			s := state.Allocated[0].start
			e := state.Allocated[len(state.Allocated)-1].end
			state.HistoryAlloc.add(s, e)
		}
		state.Date = today
		state.CurrentMax = 0
		state.Allocated = newIntervalList()
	}

	start := state.CurrentMax + 1
	end := state.CurrentMax + int64(count)
	state.CurrentMax = end
	state.Allocated.add(start, end)

	return start, end, state.Date, nil
}

func (g *Generator) Status(bizType string) *BizStatus {
	g.mu.Lock()
	defer g.mu.Unlock()

	state := g.ensureBizState(bizType)
	return &BizStatus{
		Date:           state.Date,
		CurrentMax:     state.CurrentMax,
		TotalAllocated: state.Allocated.count(),
	}
}

func (g *Generator) Check(bizType string) (hasGap bool, gaps []GapInfo, currentMax int64, total int64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	state := g.ensureBizState(bizType)
	currentMax = state.CurrentMax
	total = state.Allocated.count()

	rawGaps := state.Allocated.findGaps(state.CurrentMax)
	if len(rawGaps) == 0 {
		return false, nil, currentMax, total
	}

	gapInfos := make([]GapInfo, len(rawGaps))
	for i, g := range rawGaps {
		gapInfos[i] = GapInfo{Start: g.start, End: g.end}
	}
	return true, gapInfos, currentMax, total
}

func (g *Generator) Reset(bizType string, confirm bool) (string, error) {
	if !confirm {
		return "", fmt.Errorf("reset requires confirm=true")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	state := g.ensureBizState(bizType)
	if len(state.Allocated) > 0 {
		s := state.Allocated[0].start
		e := state.Allocated[len(state.Allocated)-1].end
		state.HistoryAlloc.add(s, e)
	}
	state.Date = g.today()
	state.CurrentMax = 0
	state.Allocated = newIntervalList()

	return state.Date, nil
}

func (g *Generator) TotalAllocated(bizType string) int64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	state := g.ensureBizState(bizType)
	return state.Allocated.count()
}

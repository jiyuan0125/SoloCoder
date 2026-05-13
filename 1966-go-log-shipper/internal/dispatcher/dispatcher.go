package dispatcher

import (
	"sync"
	"sync/atomic"
	"time"

	"log-shipper/internal/types"
)

type Dispatcher struct {
	targets   map[string]types.Target
	targetsMu sync.RWMutex

	totalReceived   int64
	receivedSamples []int64
	samplesMu       sync.Mutex
	lastSampleTime  time.Time

	startTime time.Time
}

func New() *Dispatcher {
	return &Dispatcher{
		targets:        make(map[string]types.Target),
		startTime:      time.Now(),
		lastSampleTime: time.Now(),
	}
}

func (d *Dispatcher) AddTarget(target types.Target) error {
	d.targetsMu.Lock()
	defer d.targetsMu.Unlock()

	if _, exists := d.targets[target.ID()]; exists {
		return &TargetExistsError{ID: target.ID()}
	}

	d.targets[target.ID()] = target
	if startable, ok := target.(interface{ Start() error }); ok {
		return startable.Start()
	}
	return nil
}

func (d *Dispatcher) RemoveTarget(id string) bool {
	d.targetsMu.Lock()
	defer d.targetsMu.Unlock()

	target, exists := d.targets[id]
	if !exists {
		return false
	}

	if stoppable, ok := target.(interface{ Stop() error }); ok {
		stoppable.Stop()
	}

	delete(d.targets, id)
	return true
}

func (d *Dispatcher) GetTarget(id string) (types.Target, bool) {
	d.targetsMu.RLock()
	defer d.targetsMu.RUnlock()
	t, ok := d.targets[id]
	return t, ok
}

func (d *Dispatcher) ListTargets() []types.TargetStatus {
	d.targetsMu.RLock()
	defer d.targetsMu.RUnlock()

	result := make([]types.TargetStatus, 0, len(d.targets))
	for _, t := range d.targets {
		result = append(result, t.Status())
	}
	return result
}

func (d *Dispatcher) Dispatch(entries []*types.LogEntry) {
	if len(entries) == 0 {
		return
	}

	atomic.AddInt64(&d.totalReceived, int64(len(entries)))
	d.recordSample(len(entries))

	d.targetsMu.RLock()
	defer d.targetsMu.RUnlock()

	var wg sync.WaitGroup
	for _, target := range d.targets {
		wg.Add(1)
		go func(t types.Target) {
			defer wg.Done()
			t.Write(entries)
		}(target)
	}
	wg.Wait()
}

func (d *Dispatcher) Stats() types.Stats {
	d.targetsMu.RLock()
	defer d.targetsMu.RUnlock()

	elapsed := time.Since(d.startTime).Seconds()
	var receiveRate float64
	if elapsed > 0 {
		receiveRate = float64(atomic.LoadInt64(&d.totalReceived)) / elapsed
	}

	recentRate := d.calculateRecentRate()
	if recentRate > 0 {
		receiveRate = recentRate
	}

	targetStats := make(map[string]types.TargetStats, len(d.targets))
	for id, target := range d.targets {
		status := target.Status()
		targetStats[id] = types.TargetStats{
			WriteRate:      status.WriteRate,
			SuccessPerSec:  status.WriteRate,
			TotalSuccess:   status.SuccessCount,
			TotalFailure:   status.FailureCount,
			PendingRetries: status.PendingRetryCount,
		}
	}

	return types.Stats{
		ReceivedPerSecond: receiveRate,
		TotalReceived:     atomic.LoadInt64(&d.totalReceived),
		TargetStats:       targetStats,
	}
}

func (d *Dispatcher) recordSample(count int) {
	d.samplesMu.Lock()
	defer d.samplesMu.Unlock()

	now := time.Now()
	if now.Sub(d.lastSampleTime) > 5*time.Second {
		d.receivedSamples = make([]int64, 0, 60)
		d.lastSampleTime = now
	}
	d.receivedSamples = append(d.receivedSamples, int64(count))
}

func (d *Dispatcher) calculateRecentRate() float64 {
	d.samplesMu.Lock()
	defer d.samplesMu.Unlock()

	if len(d.receivedSamples) < 2 {
		return 0
	}

	elapsed := time.Since(d.lastSampleTime).Seconds()
	if elapsed <= 0 {
		elapsed = 1.0
	}

	var total int64
	for _, s := range d.receivedSamples {
		total += s
	}
	return float64(total) / elapsed
}

func (d *Dispatcher) Stop() {
	d.targetsMu.Lock()
	defer d.targetsMu.Unlock()

	for _, target := range d.targets {
		if stoppable, ok := target.(interface{ Stop() error }); ok {
			stoppable.Stop()
		}
	}
}

type TargetExistsError struct {
	ID string
}

func (e *TargetExistsError) Error() string {
	return "target already exists: " + e.ID
}

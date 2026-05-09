package leaf

import (
	"sync"
	"sync/atomic"
	"time"

	"leaf-segment/pkg/common"
)

type Allocator struct {
	config           *Config
	centerClient     CenterClient
	currentSegment   atomic.Value
	preloadedSegment atomic.Value
	isPreloading     int32
	mu               sync.Mutex
	waitCh           chan struct{}
	lastTimestamp    int64
	seqOffset        int64
	seqMu            sync.Mutex
	reportCount      int64
}

func NewAllocator(config *Config, centerClient CenterClient) (*Allocator, error) {
	alloc := &Allocator{
		config:       config,
		centerClient: centerClient,
		waitCh:       make(chan struct{}),
	}
	seg, err := centerClient.GetSegment()
	if err != nil {
		return nil, err
	}
	alloc.currentSegment.Store(seg)
	alloc.reportProgress(seg)
	return alloc, nil
}

func (a *Allocator) Allocate() (string, error) {
	for {
		seg := a.getCurrentSegment()
		if seg == nil {
			if err := a.waitForSegment(); err != nil {
				return "", err
			}
			continue
		}

		seq, ok := seg.Next()
		if ok {
			a.checkPreload(seg)
			a.tryReportProgress(seg)
			return a.formatID(seq), nil
		}

		if a.trySwitchToPreloaded() {
			continue
		}

		if err := a.waitForSegment(); err != nil {
			return "", err
		}
	}
}

func (a *Allocator) AllocateBatch(n int) ([]string, error) {
	if n <= 0 {
		return nil, ErrInvalidCount
	}
	ids := make([]string, 0, n)
	for len(ids) < n {
		remaining := n - len(ids)
		seg := a.getCurrentSegment()
		if seg == nil {
			if err := a.waitForSegment(); err != nil {
				return nil, err
			}
			continue
		}
		batch, ok := seg.NextBatch(remaining)
		if !ok {
			if a.trySwitchToPreloaded() {
				continue
			}
			if err := a.waitForSegment(); err != nil {
				return nil, err
			}
			continue
		}
		for _, seq := range batch {
			ids = append(ids, a.formatID(seq))
		}
		a.checkPreload(seg)
		a.tryReportProgress(seg)
	}
	return ids, nil
}

func (a *Allocator) Status() *common.StatusResponse {
	current := a.getCurrentSegment()
	preloaded := a.getPreloadedSegment()
	isPreloading := atomic.LoadInt32(&a.isPreloading) == 1

	resp := &common.StatusResponse{
		HasPreloadedSegment: preloaded != nil,
		IsPreloading:        isPreloading,
	}
	if current != nil {
		resp.CurrentSegment = &common.SegmentInfo{
			Start:     current.Start,
			End:       current.End,
			Current:   atomic.LoadInt64(&current.current),
			Allocated: current.Allocated(),
			Remaining: current.Remaining(),
		}
	}
	if preloaded != nil {
		resp.PreloadedSegment = &common.SegmentInfo{
			Start:     preloaded.Start,
			End:       preloaded.End,
			Current:   atomic.LoadInt64(&preloaded.current),
			Allocated: preloaded.Allocated(),
			Remaining: preloaded.Remaining(),
		}
	}
	return resp
}

func (a *Allocator) getCurrentSegment() *Segment {
	v := a.currentSegment.Load()
	if v == nil {
		return nil
	}
	return v.(*Segment)
}

func (a *Allocator) setCurrentSegment(seg *Segment) {
	a.currentSegment.Store(seg)
}

func (a *Allocator) getPreloadedSegment() *Segment {
	v := a.preloadedSegment.Load()
	if v == nil {
		return nil
	}
	return v.(*Segment)
}

func (a *Allocator) setPreloadedSegment(seg *Segment) {
	if seg == nil {
		a.preloadedSegment.Store((*Segment)(nil))
	} else {
		a.preloadedSegment.Store(seg)
	}
}

func (a *Allocator) checkPreload(seg *Segment) {
	total := seg.End - seg.Start + 1
	remaining := seg.Remaining()
	threshold := int64(float64(total) * a.config.PreloadThresholdRatio)
	if remaining <= threshold {
		a.triggerPreload()
	}
}

func (a *Allocator) triggerPreload() {
	if atomic.CompareAndSwapInt32(&a.isPreloading, 0, 1) {
		go a.doPreload()
	}
}

func (a *Allocator) doPreload() {
	defer atomic.StoreInt32(&a.isPreloading, 0)
	seg, err := a.centerClient.GetSegment()
	if err != nil {
		return
	}
	a.setPreloadedSegment(seg)
	a.reportProgress(seg)
}

func (a *Allocator) trySwitchToPreloaded() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	current := a.getCurrentSegment()
	if current == nil || !current.IsExhausted() {
		return false
	}
	preloaded := a.getPreloadedSegment()
	if preloaded == nil {
		return false
	}
	a.setCurrentSegment(preloaded)
	a.setPreloadedSegment(nil)
	return true
}

func (a *Allocator) waitForSegment() error {
	current := a.getCurrentSegment()
	if current == nil || current.IsExhausted() {
		preloaded := a.getPreloadedSegment()
		if preloaded != nil {
			if a.trySwitchToPreloaded() {
				return nil
			}
		}
	}
	a.triggerPreload()
	timeout := time.After(a.config.WaitTimeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return ErrTimeout
		case <-ticker.C:
			if a.trySwitchToPreloaded() {
				return nil
			}
			a.triggerPreload()
		}
	}
}

func (a *Allocator) formatID(seq int64) string {
	a.seqMu.Lock()
	defer a.seqMu.Unlock()
	now := CurrentTimestampSec()
	if now == a.lastTimestamp {
		a.seqOffset++
	} else {
		a.lastTimestamp = now
		a.seqOffset = 0
	}
	effectiveSeq := seq*100000 + a.seqOffset
	return GenerateID(now, effectiveSeq)
}

func (a *Allocator) tryReportProgress(seg *Segment) {
	reportFreq := int64(100)
	cnt := atomic.AddInt64(&a.reportCount, 1)
	if cnt%reportFreq == 0 {
		go a.centerClient.ReportProgress(atomic.LoadInt64(&seg.current), seg.End)
	}
}

func (a *Allocator) reportProgress(seg *Segment) {
	go a.centerClient.ReportProgress(atomic.LoadInt64(&seg.current), seg.End)
}

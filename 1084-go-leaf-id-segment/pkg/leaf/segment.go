package leaf

import (
	"errors"
	"strconv"
	"sync/atomic"
	"time"
)

type Segment struct {
	Start   int64
	End     int64
	current int64
}

func NewSegment(start, end int64) *Segment {
	return &Segment{
		Start:   start,
		End:     end,
		current: start,
	}
}

func (s *Segment) Remaining() int64 {
	current := atomic.LoadInt64(&s.current)
	return s.End - current + 1
}

func (s *Segment) Allocated() int64 {
	current := atomic.LoadInt64(&s.current)
	if current < s.Start {
		return 0
	}
	return current - s.Start
}

func (s *Segment) IsExhausted() bool {
	current := atomic.LoadInt64(&s.current)
	return current > s.End
}

func (s *Segment) Next() (int64, bool) {
	for {
		current := atomic.LoadInt64(&s.current)
		if current > s.End {
			return 0, false
		}
		if atomic.CompareAndSwapInt64(&s.current, current, current+1) {
			return current, true
		}
	}
}

func (s *Segment) NextBatch(n int) ([]int64, bool) {
	if n <= 0 {
		return nil, false
	}
	for {
		current := atomic.LoadInt64(&s.current)
		available := s.End - current + 1
		if available <= 0 {
			return nil, false
		}
		take := n
		if int64(n) > available {
			take = int(available)
		}
		next := current + int64(take)
		if next > s.End+1 {
			next = s.End + 1
		}
		if atomic.CompareAndSwapInt64(&s.current, current, next) {
			ids := make([]int64, take)
			for i := 0; i < take; i++ {
				ids[i] = current + int64(i)
			}
			return ids, true
		}
	}
}

func GenerateID(timestampSec int64, sequence int64) string {
	return strconv.FormatInt(timestampSec, 10) + strconv.FormatInt(sequence, 10)
}

func CurrentTimestampSec() int64 {
	return time.Now().Unix()
}

var (
	ErrSegmentExhausted = errors.New("segment exhausted")
	ErrTimeout      = errors.New("timeout waiting for next segment")
	ErrInvalidCount = errors.New("invalid count: must be positive")
	ErrCenterDown   = errors.New("center node unavailable")
	ErrNoSegments   = errors.New("failed to get segment from center")
)

package commitlog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Log struct {
	dir        string
	config     Config
	segments   []*segment
	activeSeg  *segment
	nextOffset int64
	mu         sync.RWMutex
	closed     bool
}

func Open(dir string, config Config) (*Log, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	segments, err := openSegments(dir)
	if err != nil {
		return nil, err
	}

	l := &Log{
		dir:      dir,
		config:   config,
		segments: segments,
	}

	if len(segments) == 0 {
		seg, err := createSegment(dir, 0)
		if err != nil {
			return nil, err
		}
		l.segments = append(l.segments, seg)
		l.activeSeg = seg
		l.nextOffset = 0
	} else {
		l.activeSeg = segments[len(segments)-1]
		l.nextOffset = l.discoverNextOffset()
	}

	return l, nil
}

func (l *Log) discoverNextOffset() int64 {
	if l.activeSeg == nil {
		return 0
	}

	var maxOffset int64 = -1
	l.activeSeg.scanFrom(0, 10000, func(r *record) bool {
		if r.offset > maxOffset {
			maxOffset = r.offset
		}
		return true
	})

	if maxOffset < 0 {
		return 0
	}
	return maxOffset + 1
}

func (l *Log) Append(value []byte) (int64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return 0, fmt.Errorf("log closed")
	}

	rec := newRecord(l.nextOffset, value)

	if !l.activeSeg.canFit(rec, l.config.SegmentSize) {
		if err := l.roll(); err != nil {
			return 0, err
		}
	}

	offset, err := l.activeSeg.append(rec, l.config.IndexInterval)
	if err != nil {
		return 0, err
	}

	l.nextOffset++
	return offset, nil
}

func (l *Log) roll() error {
	if err := l.activeSeg.sync(); err != nil {
		return err
	}

	seg, err := createSegment(l.dir, l.nextOffset)
	if err != nil {
		return err
	}

	l.segments = append(l.segments, seg)
	l.activeSeg = seg
	return nil
}

func (l *Log) Fetch(startOffset int64, max int) ([]*record, int64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.closed {
		return nil, 0, fmt.Errorf("log closed")
	}

	if startOffset >= l.nextOffset {
		return nil, l.nextOffset, nil
	}

	if startOffset < 0 {
		startOffset = 0
	}

	segIdx := l.findSegment(startOffset)
	if segIdx < 0 {
		return nil, l.nextOffset, nil
	}

	var records []*record
	nextOffset := startOffset

	for segIdx < len(l.segments) && len(records) < max {
		seg := l.segments[segIdx]

		pos, err := seg.findPosition(startOffset)
		if err != nil {
			return nil, 0, err
		}

		err = seg.scanFrom(pos, max-len(records), func(r *record) bool {
			if r.offset >= startOffset {
				records = append(records, r)
				nextOffset = r.offset + 1
			}
			return len(records) < max
		})
		if err != nil {
			return nil, 0, err
		}

		segIdx++
	}

	return records, nextOffset, nil
}

func (l *Log) findSegment(offset int64) int {
	idx := sort.Search(len(l.segments), func(i int) bool {
		return l.segments[i].baseOffset > offset
	})
	if idx == 0 {
		if len(l.segments) > 0 && l.segments[0].baseOffset <= offset {
			return 0
		}
		return -1
	}
	return idx - 1
}

func (l *Log) NextOffset() int64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.nextOffset
}

func (l *Log) CleanupRetention(activeReaders map[int64]struct{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return fmt.Errorf("log closed")
	}

	if l.config.RetentionDuration <= 0 {
		return nil
	}

	cutoff := time.Now().Add(-l.config.RetentionDuration)
	var toDelete []*segment

	for i, seg := range l.segments {
		if i == len(l.segments)-1 {
			continue
		}

		if seg.lastWrite.After(cutoff) {
			continue
		}

		if _, ok := activeReaders[seg.baseOffset]; ok {
			continue
		}

		toDelete = append(toDelete, seg)
	}

	for _, seg := range toDelete {
		if err := seg.delete(l.dir); err != nil {
			return err
		}

		for i, s := range l.segments {
			if s == seg {
				l.segments = append(l.segments[:i], l.segments[i+1:]...)
				break
			}
		}
	}

	return nil
}

func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}
	l.closed = true

	var firstErr error
	for _, seg := range l.segments {
		if err := seg.close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (l *Log) Dir() string {
	return l.dir
}

func (l *Log) Config() Config {
	return l.config
}

func NewTopicLog(dataDir, topic string, config Config) (*Log, error) {
	topicDir := filepath.Join(dataDir, "topics", topic)
	return Open(topicDir, config)
}

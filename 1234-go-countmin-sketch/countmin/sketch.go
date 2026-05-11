package countmin

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
	"math"
)

const (
	DefaultWidth  = 1024
	DefaultDepth  = 5
)

type Sketch struct {
	width      int
	depth      int
	counters   [][]int64
	seeds      []uint64
	totalCount int64
}

func New(width, depth int) (*Sketch, error) {
	if width <= 0 {
		return nil, errors.New("width must be positive")
	}
	if depth <= 0 {
		return nil, errors.New("depth must be positive")
	}

	counters := make([][]int64, depth)
	for i := range counters {
		counters[i] = make([]int64, width)
	}

	seeds := make([]uint64, depth)
	for i := range seeds {
		seeds[i] = uint64(i + 1) * 0x9E3779B97F4A7C15
	}

	return &Sketch{
		width:      width,
		depth:      depth,
		counters:   counters,
		seeds:      seeds,
		totalCount: 0,
	}, nil
}

func NewDefault() *Sketch {
	s, _ := New(DefaultWidth, DefaultDepth)
	return s
}

func (s *Sketch) hash(item []byte, seed uint64) uint64 {
	h := fnv.New64a()
	h.Write(item)
	hashVal := h.Sum64()
	return hashVal ^ seed
}

func (s *Sketch) getIndex(item []byte, row int) int {
	h := s.hash(item, s.seeds[row])
	return int(h % uint64(s.width))
}

func (s *Sketch) Add(item []byte, count int64) {
	for i := 0; i < s.depth; i++ {
		idx := s.getIndex(item, i)
		s.counters[i][idx] += count
	}
	s.totalCount += count
}

func (s *Sketch) AddInt64(item int64, count int64) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(item))
	s.Add(buf, count)
}

func (s *Sketch) AddString(item string, count int64) {
	s.Add([]byte(item), count)
}

type Estimate struct {
	Frequency   int64
	ErrorBound  float64
}

func (s *Sketch) Query(item []byte) Estimate {
	if s.totalCount == 0 {
		return Estimate{Frequency: 0, ErrorBound: 0}
	}

	minCount := int64(math.MaxInt64)
	for i := 0; i < s.depth; i++ {
		idx := s.getIndex(item, i)
		if s.counters[i][idx] < minCount {
			minCount = s.counters[i][idx]
		}
	}

	errorBound := float64(s.totalCount) / float64(s.width)
	return Estimate{Frequency: minCount, ErrorBound: errorBound}
}

func (s *Sketch) QueryInt64(item int64) Estimate {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(item))
	return s.Query(buf)
}

func (s *Sketch) QueryString(item string) Estimate {
	return s.Query([]byte(item))
}

func (s *Sketch) TotalCount() int64 {
	return s.totalCount
}

func (s *Sketch) Width() int {
	return s.width
}

func (s *Sketch) Depth() int {
	return s.depth
}

func (s *Sketch) Merge(other *Sketch) error {
	if s.width != other.width {
		return errors.New("incompatible width for merge")
	}
	if s.depth != other.depth {
		return errors.New("incompatible depth for merge")
	}

	for i := 0; i < s.depth; i++ {
		for j := 0; j < s.width; j++ {
			s.counters[i][j] += other.counters[i][j]
		}
	}
	s.totalCount += other.totalCount
	return nil
}

func (s *Sketch) Reset() {
	for i := 0; i < s.depth; i++ {
		for j := 0; j < s.width; j++ {
			s.counters[i][j] = 0
		}
	}
	s.totalCount = 0
}

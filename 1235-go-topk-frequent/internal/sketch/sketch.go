package sketch

import (
	"encoding/binary"
	"hash/fnv"
	"math"
)

type CountMinSketch struct {
	width  int
	depth  int
	table  [][]uint64
	hashes []hashFunc
}

type hashFunc func(string) uint64

func New(width, depth int) *CountMinSketch {
	if width <= 0 {
		width = 1000
	}
	if depth <= 0 {
		depth = 5
	}

	cms := &CountMinSketch{
		width:  width,
		depth:  depth,
		table:  make([][]uint64, depth),
		hashes: make([]hashFunc, depth),
	}

	for i := 0; i < depth; i++ {
		cms.table[i] = make([]uint64, width)
		cms.hashes[i] = makeHash(uint64(i + 1))
	}

	return cms
}

func makeHash(seed uint64) hashFunc {
	return func(s string) uint64 {
		h := fnv.New64a()
		seedBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(seedBytes, seed)
		h.Write(seedBytes)
		h.Write([]byte(s))
		return h.Sum64()
	}
}

func (cms *CountMinSketch) Add(item string) {
	for i := 0; i < cms.depth; i++ {
		idx := cms.hashes[i](item) % uint64(cms.width)
		cms.table[i][idx]++
	}
}

func (cms *CountMinSketch) AddN(item string, n uint64) {
	if n == 0 {
		return
	}
	for i := 0; i < cms.depth; i++ {
		idx := cms.hashes[i](item) % uint64(cms.width)
		cms.table[i][idx] += n
	}
}

func (cms *CountMinSketch) Count(item string) uint64 {
	min := uint64(math.MaxUint64)
	for i := 0; i < cms.depth; i++ {
		idx := cms.hashes[i](item) % uint64(cms.width)
		if cms.table[i][idx] < min {
			min = cms.table[i][idx]
		}
	}
	return min
}

func (cms *CountMinSketch) Counts(item string) []uint64 {
	counts := make([]uint64, cms.depth)
	for i := 0; i < cms.depth; i++ {
		idx := cms.hashes[i](item) % uint64(cms.width)
		counts[i] = cms.table[i][idx]
	}
	return counts
}

func (cms *CountMinSketch) Reset() {
	for i := 0; i < cms.depth; i++ {
		for j := 0; j < cms.width; j++ {
			cms.table[i][j] = 0
		}
	}
}

func (cms *CountMinSketch) Width() int {
	return cms.width
}

func (cms *CountMinSketch) Depth() int {
	return cms.depth
}

package core

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
)

type Layer struct {
	index           int
	capacity        int
	count           int
	hashFunctions   int
	bits            []uint64
	counts          []uint32
	bitCount        int
	seed            uint32
	targetFPR       float64
	warningFPR      float64
}

func NewLayer(index int, config LayerConfig, warningFactor float64) *Layer {
	bits := CalculateOptimalBits(config.Capacity, config.TargetFPR)
	if bits < 64 {
		bits = 64
	}
	arraySize := (bits + 63) / 64
	return &Layer{
		index:          index,
		capacity:       config.Capacity,
		hashFunctions:  config.HashFunctions,
		bits:           make([]uint64, arraySize),
		counts:         make([]uint32, bits),
		bitCount:       bits,
		seed:           uint32(index*12345 + 6789),
		targetFPR:      config.TargetFPR,
		warningFPR:     config.TargetFPR * warningFactor,
	}
}

func (l *Layer) hash(data []byte, fn uint32) uint64 {
	h := fnv.New64a()
	seedBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(seedBytes[0:4], l.seed)
	binary.LittleEndian.PutUint32(seedBytes[4:8], fn)
	h.Write(seedBytes)
	h.Write(data)
	return h.Sum64()
}

func (l *Layer) getPositions(data []byte) []uint64 {
	positions := make([]uint64, l.hashFunctions)
	for i := 0; i < l.hashFunctions; i++ {
		hash := l.hash(data, uint32(i))
		positions[i] = hash % uint64(l.bitCount)
	}
	return positions
}

func (l *Layer) Add(data []byte) bool {
	if l.IsFull() {
		return false
	}
	positions := l.getPositions(data)
	for _, pos := range positions {
		idx := pos / 64
		bit := uint64(1) << (pos % 64)
		l.bits[idx] |= bit
		l.counts[pos]++
	}
	l.count++
	return true
}

func (l *Layer) Contains(data []byte) bool {
	if l.count == 0 {
		return false
	}
	positions := l.getPositions(data)
	for _, pos := range positions {
		idx := pos / 64
		bit := uint64(1) << (pos % 64)
		if (l.bits[idx] & bit) == 0 {
			return false
		}
	}
	return true
}

func (l *Layer) Remove(data []byte) bool {
	if !l.Contains(data) {
		return false
	}
	positions := l.getPositions(data)
	canRemove := true
	for _, pos := range positions {
		if l.counts[pos] == 0 {
			canRemove = false
			break
		}
	}
	if !canRemove {
		return false
	}
	for _, pos := range positions {
		l.counts[pos]--
		if l.counts[pos] == 0 {
			idx := pos / 64
			bit := ^(uint64(1) << (pos % 64))
			l.bits[idx] &= bit
		}
	}
	l.count--
	return true
}

func (l *Layer) IsFull() bool {
	return l.count >= l.capacity
}

func (l *Layer) Count() int {
	return l.count
}

func (l *Layer) Capacity() int {
	return l.capacity
}

func (l *Layer) CurrentFPR() float64 {
	return CalculateTheoreticalFPR(l.count, l.capacity, l.hashFunctions, l.bitCount)
}

func (l *Layer) TargetFPR() float64 {
	return l.targetFPR
}

func (l *Layer) HasWarning() bool {
	return l.CurrentFPR() > l.warningFPR
}

func (l *Layer) Reset() {
	for i := range l.bits {
		l.bits[i] = 0
	}
	for i := range l.counts {
		l.counts[i] = 0
	}
	l.count = 0
}

func (l *Layer) HashFunctions() int {
	return l.hashFunctions
}

func (l *Layer) BitCount() int {
	return l.bitCount
}

func (l *Layer) Index() int {
	return l.index
}

func (l *Layer) String() string {
	return fmt.Sprintf("Layer %d: count=%d/%d, fpr=%.6f, hash=%d",
		l.index, l.count, l.capacity, l.CurrentFPR(), l.hashFunctions)
}

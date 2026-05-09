package bloomfilter

import (
	"crypto/sha256"
	"encoding/binary"
	"log"
	"math"
	"sync"
)

type Config struct {
	Capacity      uint64
	FalsePositive float64
	CounterBits   uint8
}

type Filter struct {
	counters    []uint8
	numHashes   uint64
	numCounters uint64
	counterBits uint8
	counterMax  uint8
	inserted    uint64
	overflow    uint64
	mu          sync.RWMutex
}

func New(cfg Config) *Filter {
	if cfg.Capacity == 0 {
		cfg.Capacity = 1000
	}
	if cfg.FalsePositive <= 0 || cfg.FalsePositive >= 1 {
		cfg.FalsePositive = 0.01
	}
	if cfg.CounterBits < 1 || cfg.CounterBits > 8 {
		cfg.CounterBits = 4
	}

	m := float64(cfg.Capacity) * math.Abs(math.Log(cfg.FalsePositive)) / (math.Ln2 * math.Ln2)
	numCounters := uint64(math.Ceil(m))
	if numCounters < 1 {
		numCounters = 1
	}

	k := math.Ceil(m / float64(cfg.Capacity) * math.Ln2)
	numHashes := uint64(k)
	if numHashes < 1 {
		numHashes = 1
	}

	counterMax := uint8((1 << cfg.CounterBits) - 1)

	bytesNeeded := (numCounters*uint64(cfg.CounterBits) + 7) / 8
	if bytesNeeded < 1 {
		bytesNeeded = 1
	}

	return &Filter{
		counters:    make([]uint8, bytesNeeded),
		numHashes:   numHashes,
		numCounters: numCounters,
		counterBits: cfg.CounterBits,
		counterMax:  counterMax,
	}
}

func (f *Filter) Add(item string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	positions := f.hash(item)
	hadOverflow := false

	for _, pos := range positions {
		idx := pos * uint64(f.counterBits)
		byteIdx := idx / 8
		bitIdx := idx % 8

		current := f.readCounter(byteIdx, bitIdx)

		if current >= f.counterMax {
			hadOverflow = true
			continue
		}

		f.writeCounter(byteIdx, bitIdx, current+1)
	}

	if hadOverflow {
		f.overflow++
		log.Printf("WARN: Counter overflow occurred. Counter bits: %d, Max: %d", f.counterBits, f.counterMax)
	}

	f.inserted++
}

func (f *Filter) Delete(item string) {
	if !f.MayContain(item) {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	positions := f.hash(item)

	for _, pos := range positions {
		idx := pos * uint64(f.counterBits)
		byteIdx := idx / 8
		bitIdx := idx % 8

		current := f.readCounter(byteIdx, bitIdx)
		if current == 0 {
			continue
		}

		f.writeCounter(byteIdx, bitIdx, current-1)
	}
}

func (f *Filter) MayContain(item string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	positions := f.hash(item)

	for _, pos := range positions {
		idx := pos * uint64(f.counterBits)
		byteIdx := idx / 8
		bitIdx := idx % 8

		if f.readCounter(byteIdx, bitIdx) == 0 {
			return false
		}
	}

	return true
}

func (f *Filter) hash(item string) []uint64 {
	result := make([]uint64, f.numHashes)

	h := sha256.Sum256([]byte(item))
	h1 := binary.LittleEndian.Uint64(h[0:8])
	h2 := binary.LittleEndian.Uint64(h[8:16])

	for i := uint64(0); i < f.numHashes; i++ {
		combined := h1 + i*h2
		result[i] = combined % f.numCounters
	}

	return result
}

func (f *Filter) readCounter(byteIdx, bitIdx uint64) uint8 {
	cb := uint64(f.counterBits)
	totalBits := byteIdx*8 + bitIdx

	result := uint8(0)
	for i := uint64(0); i < cb; i++ {
		bIdx := (totalBits + i) / 8
		bOffset := (totalBits + i) % 8

		if bIdx >= uint64(len(f.counters)) {
			break
		}

		bit := (f.counters[bIdx] >> bOffset) & 0x01
		result |= bit << i
	}

	return result
}

func (f *Filter) writeCounter(byteIdx, bitIdx uint64, value uint8) {
	cb := uint64(f.counterBits)
	totalBits := byteIdx*8 + bitIdx

	for i := uint64(0); i < cb; i++ {
		bIdx := (totalBits + i) / 8
		bOffset := (totalBits + i) % 8

		if bIdx >= uint64(len(f.counters)) {
			break
		}

		bit := (value >> i) & 0x01
		if bit == 1 {
			f.counters[bIdx] |= 0x01 << bOffset
		} else {
			f.counters[bIdx] &= ^(0x01 << bOffset)
		}
	}
}

type Info struct {
	NumHashes       uint64
	CounterBits     uint8
	Capacity        uint64
	Inserted       uint64
	NonZeroCounters uint64
	OverflowCount  uint64
}

func (f *Filter) Info() Info {
	f.mu.RLock()
	defer f.mu.RUnlock()

	nonZero := uint64(0)
	for i := uint64(0); i < f.numCounters; i++ {
		idx := i * uint64(f.counterBits)
		byteIdx := idx / 8
		bitIdx := idx % 8
		if f.readCounter(byteIdx, bitIdx) > 0 {
			nonZero++
		}
	}

	optimalCapacity := uint64(float64(f.numCounters) * math.Ln2 / float64(f.numHashes))

	return Info{
		NumHashes:       f.numHashes,
		CounterBits:     f.counterBits,
		Capacity:        optimalCapacity,
		Inserted:       f.inserted,
		NonZeroCounters: nonZero,
		OverflowCount:  f.overflow,
	}
}

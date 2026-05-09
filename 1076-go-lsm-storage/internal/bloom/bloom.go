package bloom

import (
	"encoding/binary"
	"math"
)

type Filter struct {
	bits    []uint64
	bitSize uint
	k       uint
}

func OptimalBitSize(n uint, p float64) uint {
	if n == 0 || p <= 0 || p >= 1 {
		return 1024 * 8
	}
	m := float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)
	if m < 0 {
		m = -m
	}
	return uint(m) + 1
}

func OptimalHashCount(m, n uint) uint {
	if n == 0 || m == 0 {
		return 3
	}
	k := math.Round(float64(m) / float64(n) * math.Log(2))
	if k < 1 {
		k = 1
	}
	return uint(k)
}

func New(bitSize uint, k uint) *Filter {
	wordCount := (bitSize + 63) / 64
	if wordCount == 0 {
		wordCount = 1
	}
	return &Filter{
		bits:    make([]uint64, wordCount),
		bitSize: wordCount * 64,
		k:       k,
	}
}

func (f *Filter) Add(key []byte) {
	h1, h2 := hash(key)
	for i := uint(0); i < f.k; i++ {
		idx := uint(h1 + uint64(i)*h2) % f.bitSize
		wordIdx := idx / 64
		bitIdx := idx % 64
		f.bits[wordIdx] |= 1 << bitIdx
	}
}

func (f *Filter) MayContain(key []byte) bool {
	if f == nil || len(f.bits) == 0 {
		return true
	}
	h1, h2 := hash(key)
	for i := uint(0); i < f.k; i++ {
		idx := uint(h1 + uint64(i)*h2) % f.bitSize
		wordIdx := idx / 64
		bitIdx := idx % 64
		if f.bits[wordIdx]&(1<<bitIdx) == 0 {
			return false
		}
	}
	return true
}

func (f *Filter) Encode() []byte {
	buf := make([]byte, 8+8+len(f.bits)*8)
	binary.BigEndian.PutUint64(buf[0:8], uint64(f.bitSize))
	binary.BigEndian.PutUint64(buf[8:16], uint64(f.k))
	for i, w := range f.bits {
		binary.BigEndian.PutUint64(buf[16+i*8:16+i*8+8], w)
	}
	return buf
}

func Decode(data []byte) *Filter {
	if len(data) < 16 {
		return nil
	}
	bitSize := uint(binary.BigEndian.Uint64(data[0:8]))
	k := uint(binary.BigEndian.Uint64(data[8:16]))
	wordCount := (bitSize + 63) / 64
	if len(data) < 16+int(wordCount)*8 {
		return nil
	}
	bits := make([]uint64, wordCount)
	for i := 0; i < int(wordCount); i++ {
		bits[i] = binary.BigEndian.Uint64(data[16+i*8 : 16+i*8+8])
	}
	return &Filter{bits: bits, bitSize: bitSize, k: k}
}

func hash(b []byte) (uint64, uint64) {
	var h1, h2 uint64 = 0xcbf29ce484222325, 0x811c9dc5
	for _, v := range b {
		h1 ^= uint64(v)
		h1 *= 0x100000001b3
		h2 ^= uint64(v) + 1
		h2 *= 0x100000001b3
	}
	return h1, h2
}

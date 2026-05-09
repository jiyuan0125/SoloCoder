package lsm

import (
	"encoding/binary"
	"hash"
	"math"

	"github.com/spaolacci/murmur3"
)

type BloomFilter struct {
	bitArray []bool
	k        int
	m        int
}

func NewBloomFilter(n int, p float64) *BloomFilter {
	m := int(math.Ceil(-1 * float64(n) * math.Log(p) / (math.Log(2) * math.Log(2))))
	k := int(math.Ceil(math.Log(2) * float64(m) / float64(n)))
	if m < 1 {
		m = 1
	}
	if k < 1 {
		k = 1
	}
	return &BloomFilter{
		bitArray: make([]bool, m),
		k:        k,
		m:        m,
	}
}

func NewBloomFilterFromBytes(data []byte, k int) *BloomFilter {
	if len(data) < 8 {
		return &BloomFilter{bitArray: make([]bool, 1), k: 1, m: 1}
	}
	m := int(binary.LittleEndian.Uint64(data[:8]))
	bytes := data[8:]
	bitArray := make([]bool, m)
	for i := 0; i < m; i++ {
		byteIdx := i / 8
		bitIdx := uint(i % 8)
		if byteIdx < len(bytes) {
			bitArray[i] = (bytes[byteIdx] & (1 << bitIdx)) != 0
		}
	}
	return &BloomFilter{
		bitArray: bitArray,
		k:        k,
		m:        m,
	}
}

func (bf *BloomFilter) getHashFunctions(key string) []hash.Hash64 {
	hashes := make([]hash.Hash64, bf.k)
	for i := 0; i < bf.k; i++ {
		hashes[i] = murmur3.New64WithSeed(uint32(i))
	}
	return hashes
}

func (bf *BloomFilter) Add(key string) {
	for i := 0; i < bf.k; i++ {
		h := murmur3.Sum64WithSeed([]byte(key), uint32(i))
		idx := int(h % uint64(bf.m))
		bf.bitArray[idx] = true
	}
}

func (bf *BloomFilter) MayContain(key string) bool {
	for i := 0; i < bf.k; i++ {
		h := murmur3.Sum64WithSeed([]byte(key), uint32(i))
		idx := int(h % uint64(bf.m))
		if !bf.bitArray[idx] {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) Bytes() []byte {
	m := len(bf.bitArray)
	byteLen := (m + 7) / 8
	data := make([]byte, 8+byteLen)
	binary.LittleEndian.PutUint64(data[:8], uint64(m))
	for i := 0; i < m; i++ {
		if bf.bitArray[i] {
			byteIdx := i / 8
			bitIdx := uint(i % 8)
			data[8+byteIdx] |= 1 << bitIdx
		}
	}
	return data
}

func (bf *BloomFilter) K() int {
	return bf.k
}

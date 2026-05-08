package bloom

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"sync"
)

const (
	DefaultCapacity  = 100000
	DefaultFPR       = 0.01
	ExpandThreshold  = 0.7
	MagicNumber      = 0x424C4F4F4D
	Version          = 1
)

type Filter struct {
	mu           sync.RWMutex
	bitArray     []byte
	bitSize      uint64
	hashCount    uint64
	capacity     uint64
	count        uint64
	items        []string
	fpr          float64
	initialFPR   float64
}

func New(capacity uint64, fpr float64) (*Filter, error) {
	if capacity == 0 {
		capacity = DefaultCapacity
	}
	if fpr <= 0 || fpr >= 1 {
		fpr = DefaultFPR
	}

	bitSize := optimalBitSize(capacity, fpr)
	hashCount := optimalHashCount(capacity, bitSize)

	byteSize := (bitSize + 7) / 8
	if byteSize == 0 {
		byteSize = 1
	}

	return &Filter{
		bitArray:   make([]byte, byteSize),
		bitSize:    bitSize,
		hashCount:  hashCount,
		capacity:   capacity,
		count:      0,
		items:      make([]string, 0, capacity),
		fpr:        fpr,
		initialFPR: fpr,
	}, nil
}

func NewDefault() *Filter {
	f, _ := New(DefaultCapacity, DefaultFPR)
	return f
}

func optimalBitSize(capacity uint64, fpr float64) uint64 {
	if capacity == 0 {
		return 1
	}
	bitSize := float64(capacity) * math.Log(fpr) / math.Log(1.0/math.Pow(2.0, math.Log(2.0)))
	bitSize = math.Abs(bitSize)
	if bitSize < 1 {
		return 1
	}
	return uint64(math.Ceil(bitSize))
}

func optimalHashCount(capacity uint64, bitSize uint64) uint64 {
	if capacity == 0 || bitSize == 0 {
		return 1
	}
	hashCount := (float64(bitSize) / float64(capacity)) * math.Log(2.0)
	if hashCount < 1 {
		return 1
	}
	return uint64(math.Round(hashCount))
}

func calculateFPR(bitSize uint64, hashCount uint64, count uint64) float64 {
	if bitSize == 0 || count == 0 {
		return 0.0
	}
	ratio := float64(count) / float64(bitSize)
	return math.Pow(1.0-math.Exp(-float64(hashCount)*ratio), float64(hashCount))
}

func hash(item string) (uint64, uint64) {
	h1 := fnv.New64a()
	h1.Write([]byte(item))
	hash1 := h1.Sum64()

	h2 := fnv.New64()
	h2.Write([]byte(item))
	hash2 := h2.Sum64()

	return hash1, hash2
}

func (f *Filter) Add(item string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.checkWithLock(item) {
		return
	}

	h1, h2 := hash(item)
	for i := uint64(0); i < f.hashCount; i++ {
		index := (h1 + i*h2) % f.bitSize
		byteIndex := index / 8
		bitIndex := index % 8
		f.bitArray[byteIndex] |= 1 << bitIndex
	}

	f.items = append(f.items, item)
	f.count++

	if f.needExpand() {
		f.expand()
	}
}

func (f *Filter) AddBatch(items []string) {
	for _, item := range items {
		f.Add(item)
	}
}

func (f *Filter) Check(item string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.checkWithLock(item)
}

func (f *Filter) checkWithLock(item string) bool {
	h1, h2 := hash(item)
	for i := uint64(0); i < f.hashCount; i++ {
		index := (h1 + i*h2) % f.bitSize
		byteIndex := index / 8
		bitIndex := index % 8
		if (f.bitArray[byteIndex] & (1 << bitIndex)) == 0 {
			return false
		}
	}
	return true
}

func (f *Filter) needExpand() bool {
	if f.capacity == 0 {
		return false
	}
	usage := float64(f.count) / float64(f.capacity)
	return usage >= ExpandThreshold
}

func (f *Filter) expand() {
	newCapacity := f.capacity * 2
	newBitSize := optimalBitSize(newCapacity, f.initialFPR)
	newHashCount := optimalHashCount(newCapacity, newBitSize)

	newByteSize := (newBitSize + 7) / 8
	if newByteSize == 0 {
		newByteSize = 1
	}
	newBitArray := make([]byte, newByteSize)

	for _, item := range f.items {
		h1, h2 := hash(item)
		for i := uint64(0); i < newHashCount; i++ {
			index := (h1 + i*h2) % newBitSize
			byteIndex := index / 8
			bitIndex := index % 8
			newBitArray[byteIndex] |= 1 << bitIndex
		}
	}

	f.bitArray = newBitArray
	f.bitSize = newBitSize
	f.hashCount = newHashCount
	f.capacity = newCapacity
	f.fpr = calculateFPR(newBitSize, newHashCount, f.count)
}

func (f *Filter) Stats() Stats {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return Stats{
		Count:     f.count,
		Capacity:  f.capacity,
		FPR:       f.fpr,
		BitSize:   f.bitSize,
		HashCount: f.hashCount,
		BitUsage:  f.calculateBitUsage(),
	}
}

func (f *Filter) calculateBitUsage() float64 {
	if f.bitSize == 0 {
		return 0.0
	}

	setBits := uint64(0)
	for _, b := range f.bitArray {
		for b != 0 {
			setBits += uint64(b & 1)
			b >>= 1
		}
	}

	return float64(setBits) / float64(f.bitSize)
}

type Stats struct {
	Count     uint64  `json:"count"`
	Capacity  uint64  `json:"capacity"`
	FPR       float64 `json:"fpr"`
	BitSize   uint64  `json:"bit_size"`
	HashCount uint64  `json:"hash_count"`
	BitUsage  float64 `json:"bit_usage"`
}

func (f *Filter) Serialize() []byte {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, uint64(MagicNumber))
	binary.Write(&buf, binary.BigEndian, uint64(Version))

	binary.Write(&buf, binary.BigEndian, f.bitSize)
	binary.Write(&buf, binary.BigEndian, f.hashCount)
	binary.Write(&buf, binary.BigEndian, f.capacity)
	binary.Write(&buf, binary.BigEndian, f.count)
	binary.Write(&buf, binary.BigEndian, math.Float64bits(f.fpr))
	binary.Write(&buf, binary.BigEndian, math.Float64bits(f.initialFPR))

	bitArrayLen := len(f.bitArray)
	binary.Write(&buf, binary.BigEndian, uint64(bitArrayLen))
	buf.Write(f.bitArray)

	itemsLen := len(f.items)
	binary.Write(&buf, binary.BigEndian, uint64(itemsLen))
	for _, item := range f.items {
		itemBytes := []byte(item)
		binary.Write(&buf, binary.BigEndian, uint64(len(itemBytes)))
		buf.Write(itemBytes)
	}

	return buf.Bytes()
}

func Deserialize(data []byte) (*Filter, error) {
	if len(data) < 8*8 {
		return nil, errors.New("data too short")
	}

	reader := bytes.NewReader(data)

	var magic uint64
	if err := binary.Read(reader, binary.BigEndian, &magic); err != nil {
		return nil, err
	}
	if magic != MagicNumber {
		return nil, errors.New("invalid magic number")
	}

	var version uint64
	if err := binary.Read(reader, binary.BigEndian, &version); err != nil {
		return nil, err
	}
	if version != Version {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	var bitSize, hashCount, capacity, count uint64
	var fprBits, initialFPRBits uint64

	if err := binary.Read(reader, binary.BigEndian, &bitSize); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &hashCount); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &capacity); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &count); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &fprBits); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &initialFPRBits); err != nil {
		return nil, err
	}

	var bitArrayLen uint64
	if err := binary.Read(reader, binary.BigEndian, &bitArrayLen); err != nil {
		return nil, err
	}

	expectedByteSize := (bitSize + 7) / 8
	if expectedByteSize == 0 {
		expectedByteSize = 1
	}
	if bitArrayLen != expectedByteSize {
		return nil, fmt.Errorf("bit array size mismatch: expected %d, got %d", expectedByteSize, bitArrayLen)
	}

	bitArray := make([]byte, bitArrayLen)
	if _, err := reader.Read(bitArray); err != nil {
		return nil, err
	}

	var itemsLen uint64
	if err := binary.Read(reader, binary.BigEndian, &itemsLen); err != nil {
		return nil, err
	}

	items := make([]string, 0, itemsLen)
	for i := uint64(0); i < itemsLen; i++ {
		var itemLen uint64
		if err := binary.Read(reader, binary.BigEndian, &itemLen); err != nil {
			return nil, err
		}
		itemBytes := make([]byte, itemLen)
		if _, err := reader.Read(itemBytes); err != nil {
			return nil, err
		}
		items = append(items, string(itemBytes))
	}

	if uint64(len(items)) != count {
		return nil, fmt.Errorf("items count mismatch: expected %d, got %d", count, len(items))
	}

	return &Filter{
		bitArray:   bitArray,
		bitSize:    bitSize,
		hashCount:  hashCount,
		capacity:   capacity,
		count:      count,
		items:      items,
		fpr:        math.Float64frombits(fprBits),
		initialFPR: math.Float64frombits(initialFPRBits),
	}, nil
}

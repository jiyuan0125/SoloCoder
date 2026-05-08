package bitmap

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrInvalidSize  = errors.New("invalid size: must be positive")
	ErrIndexOutOfBounds = errors.New("index out of bounds")
	ErrSizeMismatch = errors.New("bitmaps have different sizes")
	ErrEmptyOperand = errors.New("operand list is empty")
)

const (
	bytesPerWord = 8
	bitsPerWord  = 64
)

type Bitmap struct {
	mu      sync.RWMutex
	size    uint64
	words   []uint64
}

func New(size uint64) (*Bitmap, error) {
	if size == 0 {
		return nil, ErrInvalidSize
	}
	numWords := (size + bitsPerWord - 1) / bitsPerWord
	return &Bitmap{
		size:  size,
		words: make([]uint64, numWords),
	}, nil
}

func (b *Bitmap) Size() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

func (b *Bitmap) Set(index uint64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if index >= b.size {
		return ErrIndexOutOfBounds
	}
	wordIdx := index / bitsPerWord
	bitIdx := index % bitsPerWord
	b.words[wordIdx] |= 1 << bitIdx
	return nil
}

func (b *Bitmap) Clear(index uint64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if index >= b.size {
		return ErrIndexOutOfBounds
	}
	wordIdx := index / bitsPerWord
	bitIdx := index % bitsPerWord
	b.words[wordIdx] &= ^(1 << bitIdx)
	return nil
}

func (b *Bitmap) Get(index uint64) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if index >= b.size {
		return false, ErrIndexOutOfBounds
	}
	wordIdx := index / bitsPerWord
	bitIdx := index % bitsPerWord
	return (b.words[wordIdx] & (1 << bitIdx)) != 0, nil
}

func (b *Bitmap) Count() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var count uint64
	for _, w := range b.words {
		count += uint64(popcount(w))
	}
	return count
}

func popcount(x uint64) int {
	cnt := 0
	for x != 0 {
		x &= x - 1
		cnt++
	}
	return cnt
}

func (b *Bitmap) Indices() []uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var indices []uint64
	for i, w := range b.words {
		if w == 0 {
			continue
		}
		base := uint64(i) * bitsPerWord
		for j := 0; j < bitsPerWord; j++ {
			if w&(1<<j) != 0 {
				idx := base + uint64(j)
				if idx < b.size {
					indices = append(indices, idx)
				}
			}
		}
	}
	return indices
}

func (b *Bitmap) Clone() *Bitmap {
	b.mu.RLock()
	defer b.mu.RUnlock()
	clone := &Bitmap{
		size:  b.size,
		words: make([]uint64, len(b.words)),
	}
	copy(clone.words, b.words)
	return clone
}

func (b *Bitmap) And(other *Bitmap) (*Bitmap, error) {
	return And(b, other)
}

func (b *Bitmap) Or(other *Bitmap) (*Bitmap, error) {
	return Or(b, other)
}

func (b *Bitmap) Xor(other *Bitmap) (*Bitmap, error) {
	return Xor(b, other)
}

func (b *Bitmap) Not() *Bitmap {
	return Not(b)
}

func (b *Bitmap) AndNot(mask *Bitmap) (*Bitmap, error) {
	return AndNot(b, mask)
}

func And(first *Bitmap, rest ...*Bitmap) (*Bitmap, error) {
	if first == nil {
		return nil, ErrEmptyOperand
	}
	first.mu.RLock()
	defer first.mu.RUnlock()

	if len(rest) == 0 {
		return first.Clone(), nil
	}

	result := first.Clone()
	for _, other := range rest {
		if other == nil {
			return nil, ErrEmptyOperand
		}
		other.mu.RLock()
		defer other.mu.RUnlock()
		if result.size != other.size {
			return nil, ErrSizeMismatch
		}
		for i := range result.words {
			result.words[i] &= other.words[i]
		}
	}
	return result, nil
}

func Or(first *Bitmap, rest ...*Bitmap) (*Bitmap, error) {
	if first == nil {
		return nil, ErrEmptyOperand
	}
	first.mu.RLock()
	defer first.mu.RUnlock()

	if len(rest) == 0 {
		return first.Clone(), nil
	}

	result := first.Clone()
	for _, other := range rest {
		if other == nil {
			return nil, ErrEmptyOperand
		}
		other.mu.RLock()
		defer other.mu.RUnlock()
		if result.size != other.size {
			return nil, ErrSizeMismatch
		}
		for i := range result.words {
			result.words[i] |= other.words[i]
		}
	}
	return result, nil
}

func Xor(first *Bitmap, rest ...*Bitmap) (*Bitmap, error) {
	if first == nil {
		return nil, ErrEmptyOperand
	}
	first.mu.RLock()
	defer first.mu.RUnlock()

	if len(rest) == 0 {
		return first.Clone(), nil
	}

	result := first.Clone()
	for _, other := range rest {
		if other == nil {
			return nil, ErrEmptyOperand
		}
		other.mu.RLock()
		defer other.mu.RUnlock()
		if result.size != other.size {
			return nil, ErrSizeMismatch
		}
		for i := range result.words {
			result.words[i] ^= other.words[i]
		}
	}
	return result, nil
}

func Not(b *Bitmap) *Bitmap {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := &Bitmap{
		size:  b.size,
		words: make([]uint64, len(b.words)),
	}
	for i, w := range b.words {
		result.words[i] = ^w
	}
	lastWordIdx := len(result.words) - 1
	remainingBits := b.size % bitsPerWord
	if remainingBits > 0 {
		mask := uint64((1 << remainingBits) - 1)
		result.words[lastWordIdx] &= mask
	}
	return result
}

func AndNot(b *Bitmap, mask *Bitmap) (*Bitmap, error) {
	if b == nil || mask == nil {
		return nil, ErrEmptyOperand
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	mask.mu.RLock()
	defer mask.mu.RUnlock()

	if b.size != mask.size {
		return nil, ErrSizeMismatch
	}

	result := &Bitmap{
		size:  b.size,
		words: make([]uint64, len(b.words)),
	}
	for i := range b.words {
		result.words[i] = b.words[i] & ^mask.words[i]
	}
	return result, nil
}

const (
	magicNumber   = 0x4249544D
	version       = 1
	headerSize    = 16
)

func (b *Bitmap) Serialize() ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	wordsBytes := len(b.words) * bytesPerWord
	totalSize := headerSize + wordsBytes
	buf := make([]byte, totalSize)

	binary.LittleEndian.PutUint32(buf[0:4], magicNumber)
	binary.LittleEndian.PutUint32(buf[4:8], version)
	binary.LittleEndian.PutUint64(buf[8:16], b.size)

	for i, w := range b.words {
		offset := headerSize + i*bytesPerWord
		binary.LittleEndian.PutUint64(buf[offset:offset+bytesPerWord], w)
	}

	return buf, nil
}

func Deserialize(data []byte) (*Bitmap, error) {
	if len(data) < headerSize {
		return nil, fmt.Errorf("data too short: expected at least %d bytes", headerSize)
	}

	if binary.LittleEndian.Uint32(data[0:4]) != magicNumber {
		return nil, fmt.Errorf("invalid magic number")
	}

	ver := binary.LittleEndian.Uint32(data[4:8])
	if ver != version {
		return nil, fmt.Errorf("unsupported version: %d", ver)
	}

	size := binary.LittleEndian.Uint64(data[8:16])
	numWords := (size + bitsPerWord - 1) / bitsPerWord
	expectedSize := headerSize + int(numWords)*bytesPerWord

	if len(data) < expectedSize {
		return nil, fmt.Errorf("data too short: expected %d bytes, got %d", expectedSize, len(data))
	}

	bm := &Bitmap{
		size:  size,
		words: make([]uint64, numWords),
	}

	for i := range bm.words {
		offset := headerSize + i*bytesPerWord
		bm.words[i] = binary.LittleEndian.Uint64(data[offset : offset+bytesPerWord])
	}

	return bm, nil
}

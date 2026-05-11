package bloomfilter

import (
	"encoding/base64"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"math"
)

var (
	ErrInvalidCapacity    = errors.New("capacity must be >= 0")
	ErrInvalidFalsePos    = errors.New("false positive rate must be > 0 and < 1")
	ErrInvalidBitArray    = errors.New("bit array size must be > 0")
	ErrEmptyBitArray      = errors.New("bit array is empty, cannot add elements")
	ErrInvalidDataSize    = errors.New("data size does not match expected bit array size")
	ErrUnsupportedVersion = errors.New("unsupported serialization version")
)

type BloomFilter struct {
	bitArray        []byte
	m               uint64
	k               uint64
	capacity        uint64
	falsePositive   float64
	elementsAdded   uint64
}

func New(capacity uint64, falsePositiveRate float64) (*BloomFilter, error) {
	if capacity == 0 {
		return &BloomFilter{
			bitArray:      []byte{},
			m:             0,
			k:             0,
			capacity:      0,
			falsePositive: falsePositiveRate,
			elementsAdded: 0,
		}, nil
	}

	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		return nil, ErrInvalidFalsePos
	}

	m := calculateM(capacity, falsePositiveRate)
	k := calculateK(m, capacity)

	if m <= 0 {
		return nil, ErrInvalidBitArray
	}

	byteSize := (m + 7) / 8

	return &BloomFilter{
		bitArray:      make([]byte, byteSize),
		m:             m,
		k:             k,
		capacity:      capacity,
		falsePositive: falsePositiveRate,
		elementsAdded: 0,
	}, nil
}

func calculateM(capacity uint64, falsePositiveRate float64) uint64 {
	ln2 := math.Ln2
	lnP := math.Log(falsePositiveRate)
	m := float64(capacity) * (-lnP) / (ln2 * ln2)
	return uint64(math.Ceil(m))
}

func calculateK(m, capacity uint64) uint64 {
	k := (float64(m) / float64(capacity)) * math.Ln2
	kInt := uint64(math.Round(k))
	if kInt < 1 {
		kInt = 1
	}
	return kInt
}

func hash1(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}

func hash2(data []byte) uint64 {
	h := fnv.New64()
	h.Write(data)
	return h.Sum64()
}

func (bf *BloomFilter) getPositions(data []byte) []uint64 {
	if bf.m == 0 || bf.k == 0 {
		return nil
	}

	h1 := hash1(data)
	h2 := hash2(data)

	positions := make([]uint64, bf.k)
	for i := uint64(0); i < bf.k; i++ {
		pos := (h1 + i*h2) % bf.m
		if pos < 0 {
			pos += bf.m
		}
		positions[i] = pos
	}
	return positions
}

func (bf *BloomFilter) Add(element string) error {
	if bf.m == 0 {
		return ErrEmptyBitArray
	}

	positions := bf.getPositions([]byte(element))
	for _, pos := range positions {
		bf.setBit(pos)
	}
	bf.elementsAdded++
	return nil
}

func (bf *BloomFilter) Contains(element string) (bool, error) {
	if bf.m == 0 {
		return false, nil
	}

	positions := bf.getPositions([]byte(element))
	for _, pos := range positions {
		if !bf.getBit(pos) {
			return false, nil
		}
	}
	return true, nil
}

func (bf *BloomFilter) setBit(pos uint64) {
	byteIndex := pos / 8
	bitIndex := pos % 8
	bf.bitArray[byteIndex] |= (1 << bitIndex)
}

func (bf *BloomFilter) getBit(pos uint64) bool {
	byteIndex := pos / 8
	bitIndex := pos % 8
	return (bf.bitArray[byteIndex] & (1 << bitIndex)) != 0
}

func (bf *BloomFilter) Clear() {
	byteSize := len(bf.bitArray)
	bf.bitArray = make([]byte, byteSize)
	bf.elementsAdded = 0
}

func (bf *BloomFilter) ElementsAdded() uint64 {
	return bf.elementsAdded
}

func (bf *BloomFilter) FillRate() float64 {
	if bf.m == 0 {
		return 0.0
	}

	setBits := uint64(0)
	for _, b := range bf.bitArray {
		for b != 0 {
			setBits += uint64(b & 1)
			b >>= 1
		}
	}
	return float64(setBits) / float64(bf.m)
}

func (bf *BloomFilter) Capacity() uint64 {
	return bf.capacity
}

func (bf *BloomFilter) FalsePositiveRate() float64 {
	return bf.falsePositive
}

func (bf *BloomFilter) CurrentFalsePositiveRate() float64 {
	if bf.m == 0 || bf.elementsAdded == 0 {
		return 0.0
	}
	return currentFalsePositiveRate(bf.m, bf.k, bf.elementsAdded)
}

func currentFalsePositiveRate(m, k, n uint64) float64 {
	exponent := -float64(k) * float64(n) / float64(m)
	base := 1 - math.Exp(exponent)
	return math.Pow(base, float64(k))
}

func (bf *BloomFilter) EstimateRemainingCapacity() (uint64, error) {
	if bf.m == 0 {
		return 0, nil
	}

	targetFPR := bf.falsePositive
	currentN := bf.elementsAdded

	for n := currentN; ; n++ {
		fpr := currentFalsePositiveRate(bf.m, bf.k, n)
		if fpr > targetFPR {
			if n == 0 {
				return 0, nil
			}
			return n - 1 - currentN, nil
		}
		if n > 10000000000 {
			return 10000000000 - currentN, nil
		}
	}
}

func (bf *BloomFilter) M() uint64 {
	return bf.m
}

func (bf *BloomFilter) K() uint64 {
	return bf.k
}

func (bf *BloomFilter) SerializeToBase64() (string, error) {
	data, err := bf.Serialize()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (bf *BloomFilter) Serialize() ([]byte, error) {
	buf := make([]byte, 0, 8+8+8+8+8+len(bf.bitArray))

	version := uint8(1)
	buf = append(buf, byte(version))

	buf = appendUint64(buf, bf.m)
	buf = appendUint64(buf, bf.k)
	buf = appendUint64(buf, bf.capacity)
	buf = appendUint64(buf, bf.elementsAdded)

	fpBytes := math.Float64bits(bf.falsePositive)
	buf = appendUint64(buf, fpBytes)

	buf = append(buf, bf.bitArray...)

	return buf, nil
}

func (bf *BloomFilter) SerializeToWriter(w io.Writer) error {
	data, err := bf.Serialize()
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func DeserializeFromBase64(encoded string) (*BloomFilter, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	return Deserialize(data)
}

func Deserialize(data []byte) (*BloomFilter, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("invalid data: too short")
	}

	version := data[0]
	if version != 1 {
		return nil, ErrUnsupportedVersion
	}

	offset := 1

	if len(data) < offset+8 {
		return nil, fmt.Errorf("invalid data: missing m")
	}
	m := readUint64(data[offset:])
	offset += 8

	if len(data) < offset+8 {
		return nil, fmt.Errorf("invalid data: missing k")
	}
	k := readUint64(data[offset:])
	offset += 8

	if len(data) < offset+8 {
		return nil, fmt.Errorf("invalid data: missing capacity")
	}
	capacity := readUint64(data[offset:])
	offset += 8

	if len(data) < offset+8 {
		return nil, fmt.Errorf("invalid data: missing elementsAdded")
	}
	elementsAdded := readUint64(data[offset:])
	offset += 8

	if len(data) < offset+8 {
		return nil, fmt.Errorf("invalid data: missing falsePositive")
	}
	fpBits := readUint64(data[offset:])
	falsePositive := math.Float64frombits(fpBits)
	offset += 8

	var bitArray []byte
	expectedByteSize := uint64(0)
	if m > 0 {
		expectedByteSize = (m + 7) / 8
	}
	actualByteSize := uint64(len(data) - offset)

	if m > 0 {
		if actualByteSize != expectedByteSize {
			return nil, fmt.Errorf("%w: expected %d bytes, got %d bytes",
				ErrInvalidDataSize, expectedByteSize, actualByteSize)
		}
		bitArray = make([]byte, actualByteSize)
		copy(bitArray, data[offset:])
	}

	return &BloomFilter{
		bitArray:      bitArray,
		m:             m,
		k:             k,
		capacity:      capacity,
		falsePositive: falsePositive,
		elementsAdded: elementsAdded,
	}, nil
}

func DeserializeFromReader(r io.Reader) (*BloomFilter, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Deserialize(data)
}

func appendUint64(buf []byte, val uint64) []byte {
	return append(buf,
		byte(val>>56),
		byte(val>>48),
		byte(val>>40),
		byte(val>>32),
		byte(val>>24),
		byte(val>>16),
		byte(val>>8),
		byte(val),
	)
}

func readUint64(data []byte) uint64 {
	return uint64(data[0])<<56 |
		uint64(data[1])<<48 |
		uint64(data[2])<<40 |
		uint64(data[3])<<32 |
		uint64(data[4])<<24 |
		uint64(data[5])<<16 |
		uint64(data[6])<<8 |
		uint64(data[7])
}

package hll

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
)

const (
	minPrecision = 4
	maxPrecision = 18
	version      = 1
	magic        = "HLL\001"
)

type HLL struct {
	precision       int
	numBuckets      int
	sparseData      map[uint64]uint8
	denseData       []uint8
	alpha           float64
	sparseThreshold int
	useSparse       bool
}

func NewHLL(precision int) (*HLL, error) {
	if precision < minPrecision || precision > maxPrecision {
		return nil, fmt.Errorf("precision must be between %d and %d", minPrecision, maxPrecision)
	}

	numBuckets := 1 << uint(precision)
	hll := &HLL{
		precision:  precision,
		numBuckets: numBuckets,
		alpha:      alpha(numBuckets),
		useSparse:  true,
	}

	hll.sparseThreshold = int(float64(numBuckets) * 0.3)
	if hll.sparseThreshold < 10 {
		hll.sparseThreshold = 10
	}

	hll.sparseData = make(map[uint64]uint8)

	return hll, nil
}

func (h *HLL) Add(value string) {
	hash := fnvHashString(value)
	h.addHash(hash)
}

func (h *HLL) addHash(hash uint64) {
	bucket := hash & (uint64(h.numBuckets) - 1)
	remainder := hash >> uint(h.precision)
	rho := countLeadingZeros(remainder, 64-h.precision)

	if h.useSparse {
		oldRho, exists := h.sparseData[bucket]
		if !exists || rho > oldRho {
			h.sparseData[bucket] = rho
		}

		if len(h.sparseData) >= h.sparseThreshold {
			h.convertToDense()
		}
	} else {
		if rho > h.denseData[bucket] {
			h.denseData[bucket] = rho
		}
	}
}

func (h *HLL) Count() float64 {
	if h.useSparse && len(h.sparseData) == 0 {
		return 0.0
	}

	var z float64
	var emptyBuckets int

	if h.useSparse {
		for i := 0; i < h.numBuckets; i++ {
			rho, exists := h.sparseData[uint64(i)]
			if !exists {
				emptyBuckets++
				z += 1.0
			} else {
				z += 1.0 / float64(int(1)<<uint(rho))
			}
		}
	} else {
		for i := 0; i < h.numBuckets; i++ {
			if h.denseData[i] == 0 {
				emptyBuckets++
				z += 1.0
			} else {
				z += 1.0 / float64(int(1)<<uint(h.denseData[i]))
			}
		}
	}

	m := float64(h.numBuckets)
	estimate := h.alpha * m * m / z

	if estimate <= 2.5*m && emptyBuckets > 0 {
		estimate = m * math.Log(m/float64(emptyBuckets))
	}

	return estimate
}

func (h *HLL) Merge(other *HLL) error {
	if h == nil || other == nil {
		return errors.New("HLL instance is nil")
	}

	if h.precision != other.precision {
		return errors.New("precision mismatch: cannot merge HLLs with different precisions")
	}

	if other.useSparse {
		for bucket, rho := range other.sparseData {
			if h.useSparse {
				if current, exists := h.sparseData[bucket]; !exists || rho > current {
					h.sparseData[bucket] = rho
				}
			} else {
				if rho > h.denseData[bucket] {
					h.denseData[bucket] = rho
				}
			}
		}

		if h.useSparse && len(h.sparseData) >= h.sparseThreshold {
			h.convertToDense()
		}
	} else {
		h.convertToDense()

		for i := 0; i < h.numBuckets; i++ {
			if other.denseData[i] > h.denseData[i] {
				h.denseData[i] = other.denseData[i]
			}
		}
	}

	return nil
}

func (h *HLL) Serialize() ([]byte, error) {
	data := make([]byte, 0)

	for _, b := range magic {
		data = append(data, byte(b))
	}

	precisionBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(precisionBuf, uint32(h.precision))
	data = append(data, precisionBuf...)

	var mode byte
	if h.useSparse {
		mode = 0
	} else {
		mode = 1
	}
	data = append(data, mode)

	if h.useSparse {
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(h.sparseData)))
		data = append(data, lenBuf...)

		keys := make([]uint64, 0, len(h.sparseData))
		for k := range h.sparseData {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

		for _, bucket := range keys {
			bucketBuf := make([]byte, 8)
			binary.BigEndian.PutUint64(bucketBuf, bucket)
			data = append(data, bucketBuf...)
			data = append(data, h.sparseData[bucket])
		}
	} else {
		data = append(data, h.denseData...)
	}

	return data, nil
}

func Deserialize(data []byte) (*HLL, error) {
	if len(data) < 9 {
		return nil, errors.New("invalid HLL data: too short")
	}

	if string(data[:4]) != magic {
		return nil, errors.New("invalid HLL data: wrong magic number")
	}

	precision := int(binary.BigEndian.Uint32(data[4:8]))
	mode := data[8]

	hll, err := NewHLL(precision)
	if err != nil {
		return nil, err
	}

	if mode == 0 {
		if len(data) < 13 {
			return nil, errors.New("invalid sparse HLL data")
		}
		sparseLen := int(binary.BigEndian.Uint32(data[9:13]))
		pos := 13

		if len(data) < pos+sparseLen*9 {
			return nil, errors.New("invalid sparse HLL data: insufficient length")
		}

		for i := 0; i < sparseLen; i++ {
			bucket := binary.BigEndian.Uint64(data[pos : pos+8])
			rho := data[pos+8]
			hll.sparseData[bucket] = rho
			pos += 9
		}
		hll.useSparse = true
	} else {
		hll.convertToDense()
		if len(data) < 9+hll.numBuckets {
			return nil, errors.New("invalid dense HLL data: insufficient length")
		}
		copy(hll.denseData, data[9:9+hll.numBuckets])
	}

	return hll, nil
}

func (h *HLL) Precision() int {
	return h.precision
}

func (h *HLL) convertToDense() {
	if !h.useSparse {
		return
	}

	h.denseData = make([]uint8, h.numBuckets)

	for bucket, rho := range h.sparseData {
		h.denseData[bucket] = rho
	}

	h.sparseData = nil
	h.useSparse = false
}

func fnvHashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func countLeadingZeros(x uint64, maxBits int) uint8 {
	if x == 0 {
		return uint8(maxBits)
	}

	var count uint8 = 1
	for i := maxBits - 1; i >= 0; i-- {
		if (x & (uint64(1) << uint(i))) == 0 {
			count++
		} else {
			break
		}
	}

	if count > uint8(maxBits) {
		count = uint8(maxBits)
	}

	return count
}

func alpha(m int) float64 {
	switch m {
	case 16:
		return 0.673
	case 32:
		return 0.697
	case 64:
		return 0.709
	default:
		return 0.7213 / (1.0 + 1.079/float64(m))
	}
}

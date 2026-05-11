package decomposition

import (
	"math"
)

type SqrtDecomposition struct {
	n          int
	blockSize  int
	blocks     int
	data       []int64
	blockSums  []int64
	blockAdds  []int64
}

func NewSqrtDecomposition(n int) *SqrtDecomposition {
	if n <= 0 {
		return &SqrtDecomposition{
			n:         0,
			blockSize: 0,
			blocks:    0,
			data:      []int64{},
			blockSums: []int64{},
			blockAdds: []int64{},
		}
	}

	blockSize := int(math.Sqrt(float64(n)))
	if blockSize == 0 {
		blockSize = 1
	}
	blocks := (n + blockSize - 1) / blockSize

	return &SqrtDecomposition{
		n:         n,
		blockSize: blockSize,
		blocks:    blocks,
		data:      make([]int64, n),
		blockSums: make([]int64, blocks),
		blockAdds: make([]int64, blocks),
	}
}

func (s *SqrtDecomposition) getBlockSize(blockIndex int) int {
	if blockIndex == s.blocks-1 {
		return s.n - blockIndex*s.blockSize
	}
	return s.blockSize
}

func (s *SqrtDecomposition) pushDown(blockIndex int) {
	if s.blockAdds[blockIndex] != 0 {
		add := s.blockAdds[blockIndex]
		size := s.getBlockSize(blockIndex)
		for i := 0; i < size; i++ {
			s.data[blockIndex*s.blockSize+i] += add
		}
		s.blockAdds[blockIndex] = 0
	}
}

func (s *SqrtDecomposition) RangeAdd(l, r int, val int64) {
	if l > r || s.n == 0 {
		return
	}
	if l < 0 {
		l = 0
	}
	if r >= s.n {
		r = s.n - 1
	}

	startBlock := l / s.blockSize
	endBlock := r / s.blockSize

	if startBlock == endBlock {
		s.pushDown(startBlock)
		for i := l; i <= r; i++ {
			s.data[i] += val
			s.blockSums[startBlock] += val
		}
		return
	}

	s.pushDown(startBlock)
	for i := l; i < (startBlock+1)*s.blockSize; i++ {
		s.data[i] += val
		s.blockSums[startBlock] += val
	}

	for b := startBlock + 1; b < endBlock; b++ {
		s.blockAdds[b] += val
		s.blockSums[b] += val * int64(s.getBlockSize(b))
	}

	s.pushDown(endBlock)
	for i := endBlock * s.blockSize; i <= r; i++ {
		s.data[i] += val
		s.blockSums[endBlock] += val
	}
}

func (s *SqrtDecomposition) RangeSum(l, r int) int64 {
	if l > r || s.n == 0 {
		return 0
	}
	if l < 0 {
		l = 0
	}
	if r >= s.n {
		r = s.n - 1
	}

	startBlock := l / s.blockSize
	endBlock := r / s.blockSize

	if startBlock == endBlock {
		s.pushDown(startBlock)
		sum := int64(0)
		for i := l; i <= r; i++ {
			sum += s.data[i]
		}
		return sum
	}

	s.pushDown(startBlock)
	sum := int64(0)
	for i := l; i < (startBlock+1)*s.blockSize; i++ {
		sum += s.data[i]
	}

	for b := startBlock + 1; b < endBlock; b++ {
		sum += s.blockSums[b]
	}

	s.pushDown(endBlock)
	for i := endBlock * s.blockSize; i <= r; i++ {
		sum += s.data[i]
	}

	return sum
}

func (s *SqrtDecomposition) Set(index int, val int64) {
	if index < 0 || index >= s.n {
		return
	}

	blockIndex := index / s.blockSize
	s.pushDown(blockIndex)
	oldVal := s.data[index]
	s.data[index] = val
	s.blockSums[blockIndex] += val - oldVal
}

func (s *SqrtDecomposition) Get(index int) int64 {
	if index < 0 || index >= s.n {
		return 0
	}

	blockIndex := index / s.blockSize
	s.pushDown(blockIndex)
	return s.data[index]
}

func (s *SqrtDecomposition) Flush() {
	for b := 0; b < s.blocks; b++ {
		s.pushDown(b)
	}
}

func (s *SqrtDecomposition) Dump() []int64 {
	s.Flush()
	result := make([]int64, s.n)
	copy(result, s.data)
	return result
}

func (s *SqrtDecomposition) Len() int {
	return s.n
}

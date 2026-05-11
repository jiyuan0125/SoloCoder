package radixsort

import (
	"math"
	"sync"
)

var (
	mu     sync.RWMutex
	global = &Sorter{
		radix: 256,
	}
)

type Stats struct {
	Passes          int
	TotalOperations int
	ArraySize       int
	Radix           int
}

type Sorter struct {
	radix int
	stats Stats
}

func NewSorter(radix int) *Sorter {
	if radix < 2 {
		radix = 256
	}
	return &Sorter{radix: radix}
}

func SetRadix(radix int) {
	mu.Lock()
	defer mu.Unlock()
	if radix >= 2 {
		global.radix = radix
	}
}

func GetRadix() int {
	mu.RLock()
	defer mu.RUnlock()
	return global.radix
}

func GetStats() Stats {
	mu.RLock()
	defer mu.RUnlock()
	return global.stats
}

func (s *Sorter) GetStats() Stats {
	return s.stats
}

func (s *Sorter) SetRadix(radix int) {
	if radix >= 2 {
		s.radix = radix
	}
}

func (s *Sorter) GetRadix() int {
	return s.radix
}

func SortInt32(arr []int32) []int32 {
	mu.Lock()
	defer mu.Unlock()
	return global.SortInt32(arr)
}

func SortInt32Slice(arr []int32) {
	mu.Lock()
	defer mu.Unlock()
	global.SortInt32Slice(arr)
}

func SortInt64(arr []int64) []int64 {
	mu.Lock()
	defer mu.Unlock()
	return global.SortInt64(arr)
}

func SortInt64Slice(arr []int64) {
	mu.Lock()
	defer mu.Unlock()
	global.SortInt64Slice(arr)
}

func (s *Sorter) SortInt32(arr []int32) []int32 {
	if len(arr) == 0 {
		s.stats = Stats{Radix: s.radix}
		return []int32{}
	}
	result := make([]int32, len(arr))
	copy(result, arr)
	s.SortInt32Slice(result)
	return result
}

func (s *Sorter) SortInt32Slice(arr []int32) {
	if len(arr) == 0 {
		s.stats = Stats{Radix: s.radix}
		return
	}
	n := len(arr)
	s.stats = Stats{
		ArraySize: n,
		Radix:     s.radix,
	}
	passes := int(math.Ceil(32.0 / math.Log2(float64(s.radix))))
	s.stats.Passes = passes
	aux := make([]int32, n)
	for pass := 0; pass < passes; pass++ {
		s.countingSortInt32(arr, aux, pass)
		if pass < passes-1 {
			arr, aux = aux, arr
		}
	}
}

func (s *Sorter) countingSortInt32(arr, aux []int32, pass int) {
	n := len(arr)
	radix := s.radix
	bits := int(math.Log2(float64(radix)))
	mask := uint32(radix - 1)
	shift := pass * bits
	isLast := pass == s.stats.Passes-1
	count := make([]int, radix)
	for _, v := range arr {
		b := s.extractBucketInt32(v, mask, shift, isLast)
		count[b]++
	}
	prefix := make([]int, radix)
	for i := 1; i < radix; i++ {
		prefix[i] = prefix[i-1] + count[i-1]
	}
	for i := 0; i < n; i++ {
		v := arr[i]
		b := s.extractBucketInt32(v, mask, shift, isLast)
		aux[prefix[b]] = v
		prefix[b]++
		s.stats.TotalOperations++
	}
}

func (s *Sorter) extractBucketInt32(v int32, mask uint32, shift int, isLast bool) int {
	u := uint32(v)
	if isLast {
		u ^= 0x80000000
	}
	b := (u >> shift) & mask
	return int(b)
}

func (s *Sorter) SortInt64(arr []int64) []int64 {
	if len(arr) == 0 {
		s.stats = Stats{Radix: s.radix}
		return []int64{}
	}
	result := make([]int64, len(arr))
	copy(result, arr)
	s.SortInt64Slice(result)
	return result
}

func (s *Sorter) SortInt64Slice(arr []int64) {
	if len(arr) == 0 {
		s.stats = Stats{Radix: s.radix}
		return
	}
	n := len(arr)
	s.stats = Stats{
		ArraySize: n,
		Radix:     s.radix,
	}
	passes := int(math.Ceil(64.0 / math.Log2(float64(s.radix))))
	s.stats.Passes = passes
	aux := make([]int64, n)
	for pass := 0; pass < passes; pass++ {
		s.countingSortInt64(arr, aux, pass)
		if pass < passes-1 {
			arr, aux = aux, arr
		}
	}
}

func (s *Sorter) countingSortInt64(arr, aux []int64, pass int) {
	n := len(arr)
	radix := s.radix
	bits := int(math.Log2(float64(radix)))
	mask := uint64(radix - 1)
	shift := pass * bits
	isLast := pass == s.stats.Passes-1
	count := make([]int, radix)
	for _, v := range arr {
		b := s.extractBucketInt64(v, mask, shift, isLast)
		count[b]++
	}
	prefix := make([]int, radix)
	for i := 1; i < radix; i++ {
		prefix[i] = prefix[i-1] + count[i-1]
	}
	for i := 0; i < n; i++ {
		v := arr[i]
		b := s.extractBucketInt64(v, mask, shift, isLast)
		aux[prefix[b]] = v
		prefix[b]++
		s.stats.TotalOperations++
	}
}

func (s *Sorter) extractBucketInt64(v int64, mask uint64, shift int, isLast bool) int {
	u := uint64(v)
	if isLast {
		u ^= 0x8000000000000000
	}
	b := (u >> shift) & mask
	return int(b)
}

type Int32Field struct {
	Name  string
	Value int32
}

func (s *Sorter) SortStructByField(arr []Int32Field, fieldName string) []Int32Field {
	if len(arr) == 0 {
		s.stats = Stats{Radix: s.radix}
		return []Int32Field{}
	}
	result := make([]Int32Field, len(arr))
	copy(result, arr)
	n := len(result)
	s.stats = Stats{
		ArraySize: n,
		Radix:     s.radix,
	}
	passes := int(math.Ceil(32.0 / math.Log2(float64(s.radix))))
	s.stats.Passes = passes
	aux := make([]Int32Field, n)
	for pass := 0; pass < passes; pass++ {
		s.countingSortStruct(result, aux, pass)
		if pass < passes-1 {
			result, aux = aux, result
		}
	}
	return result
}

func (s *Sorter) countingSortStruct(arr, aux []Int32Field, pass int) {
	n := len(arr)
	radix := s.radix
	bits := int(math.Log2(float64(radix)))
	mask := uint32(radix - 1)
	shift := pass * bits
	isLast := pass == s.stats.Passes-1
	count := make([]int, radix)
	for _, item := range arr {
		b := s.extractBucketInt32(item.Value, mask, shift, isLast)
		count[b]++
	}
	prefix := make([]int, radix)
	for i := 1; i < radix; i++ {
		prefix[i] = prefix[i-1] + count[i-1]
	}
	for i := 0; i < n; i++ {
		item := arr[i]
		b := s.extractBucketInt32(item.Value, mask, shift, isLast)
		aux[prefix[b]] = item
		prefix[b]++
		s.stats.TotalOperations++
	}
}

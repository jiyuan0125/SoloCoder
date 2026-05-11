package radixsort

import (
	"math"
	"reflect"
	"sort"
	"testing"
)

func TestSortInt32_Empty(t *testing.T) {
	arr := []int32{}
	result := SortInt32(arr)
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestSortInt32_Single(t *testing.T) {
	arr := []int32{42}
	result := SortInt32(arr)
	expected := []int32{42}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSortInt32_AllSame(t *testing.T) {
	arr := []int32{5, 5, 5, 5}
	result := SortInt32(arr)
	expected := []int32{5, 5, 5, 5}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSortInt32_Positive(t *testing.T) {
	arr := []int32{170, 45, 75, 90, 802, 24, 2, 66}
	result := SortInt32(arr)
	expected := []int32{2, 24, 45, 66, 75, 90, 170, 802}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSortInt32_Negative(t *testing.T) {
	arr := []int32{-5, -1, -10, -3}
	result := SortInt32(arr)
	expected := []int32{-10, -5, -3, -1}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSortInt32_Mixed(t *testing.T) {
	arr := []int32{-2, 5, -10, 0, 3, -1}
	result := SortInt32(arr)
	expected := []int32{-10, -2, -1, 0, 3, 5}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSortInt32_Boundary(t *testing.T) {
	arr := []int32{math.MaxInt32, math.MinInt32, 0, -1, 1}
	result := SortInt32(arr)
	expected := []int32{math.MinInt32, -1, 0, 1, math.MaxInt32}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSortInt32_Sorted(t *testing.T) {
	arr := []int32{1, 2, 3, 4, 5}
	result := SortInt32(arr)
	expected := []int32{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSortInt32_Reverse(t *testing.T) {
	arr := []int32{5, 4, 3, 2, 1}
	result := SortInt32(arr)
	expected := []int32{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSortInt32_Radix2(t *testing.T) {
	sorter := NewSorter(2)
	arr := []int32{-5, 10, 0, 3, -1}
	result := sorter.SortInt32(arr)
	expected := []int32{-5, -1, 0, 3, 10}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSortInt32_Radix16(t *testing.T) {
	sorter := NewSorter(16)
	arr := []int32{170, 45, 75, 90, 802, 24, 2, 66}
	result := sorter.SortInt32(arr)
	expected := []int32{2, 24, 45, 66, 75, 90, 170, 802}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSortInt64_Boundary(t *testing.T) {
	arr := []int64{math.MaxInt64, math.MinInt64, 0, -1, 1}
	result := SortInt64(arr)
	expected := []int64{math.MinInt64, -1, 0, 1, math.MaxInt64}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestSortInt64_LargeRandom(t *testing.T) {
	const size = 10000
	arr := make([]int64, size)
	expected := make([]int64, size)
	for i := 0; i < size; i++ {
		v := int64(i*314159 - size/2*314159)
		arr[i] = v
		expected[i] = v
	}
	sort.Slice(expected, func(i, j int) bool { return expected[i] < expected[j] })
	result := SortInt64(arr)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("large random array not sorted correctly")
	}
}

func TestSortInt64_Stats(t *testing.T) {
	sorter := NewSorter(256)
	arr := []int64{1, 2, 3, 4, 5}
	sorter.SortInt64(arr)
	stats := sorter.GetStats()
	if stats.Radix != 256 {
		t.Errorf("expected radix 256, got %d", stats.Radix)
	}
	if stats.Passes != 8 {
		t.Errorf("expected 8 passes for int64 with radix 256, got %d", stats.Passes)
	}
	if stats.TotalOperations != 5*8 {
		t.Errorf("expected 40 operations, got %d", stats.TotalOperations)
	}
}

func TestSortInt32Slice_InPlace(t *testing.T) {
	arr := []int32{3, 1, 4, 1, 5}
	SortInt32Slice(arr)
	expected := []int32{1, 1, 3, 4, 5}
	if !reflect.DeepEqual(arr, expected) {
		t.Errorf("expected %v, got %v", expected, arr)
	}
}

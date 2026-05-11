package externalsort

import (
	"reflect"
	"sort"
	"testing"
)

func TestSortInMemory(t *testing.T) {
	sorter := New()
	sorter.SetMemoryLimit(100)

	data := []int64{5, 3, 9, 1, 7, 2, 8, 4, 6}
	sorter.AddData(data)

	if err := sorter.Sort(); err != nil {
		t.Fatalf("sort failed: %v", err)
	}

	result := sorter.GetFinalResult()
	expected := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	stats := sorter.GetStats()
	if stats.ChunksCreated != 0 {
		t.Errorf("expected 0 chunks (in-memory sort), got %d", stats.ChunksCreated)
	}
	if stats.MergePasses != 0 {
		t.Errorf("expected 0 merge passes, got %d", stats.MergePasses)
	}
}

func TestSortExternally(t *testing.T) {
	sorter := New()
	sorter.SetMemoryLimit(3)

	data := []int64{5, 3, 9, 1, 7, 2, 8, 4, 6, 10}
	sorter.AddData(data)

	if err := sorter.Sort(); err != nil {
		t.Fatalf("sort failed: %v", err)
	}

	result := sorter.GetFinalResult()
	expected := make([]int64, len(data))
	copy(expected, data)
	sort.Slice(expected, func(i, j int) bool {
		return expected[i] < expected[j]
	})

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	stats := sorter.GetStats()
	if stats.ChunksCreated != 4 {
		t.Errorf("expected 4 chunks, got %d", stats.ChunksCreated)
	}
}

func TestEmptyData(t *testing.T) {
	sorter := New()
	sorter.SetMemoryLimit(10)

	if err := sorter.Sort(); err != nil {
		t.Fatalf("sort failed: %v", err)
	}

	result := sorter.GetFinalResult()
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestAllSameValues(t *testing.T) {
	sorter := New()
	sorter.SetMemoryLimit(3)

	data := make([]int64, 10)
	for i := range data {
		data[i] = 5
	}
	sorter.AddData(data)

	if err := sorter.Sort(); err != nil {
		t.Fatalf("sort failed: %v", err)
	}

	result := sorter.GetFinalResult()
	if !reflect.DeepEqual(result, data) {
		t.Errorf("expected %v, got %v", data, result)
	}
}

func TestExactMultiples(t *testing.T) {
	sorter := New()
	sorter.SetMemoryLimit(4)

	data := []int64{8, 5, 2, 1, 7, 6, 4, 3, 12, 9, 11, 10}
	sorter.AddData(data)

	if err := sorter.Sort(); err != nil {
		t.Fatalf("sort failed: %v", err)
	}

	result := sorter.GetFinalResult()
	expected := make([]int64, len(data))
	copy(expected, data)
	sort.Slice(expected, func(i, j int) bool {
		return expected[i] < expected[j]
	})

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	stats := sorter.GetStats()
	if stats.ChunksCreated != 3 {
		t.Errorf("expected 3 chunks, got %d", stats.ChunksCreated)
	}
}

func TestInvalidMemoryLimit(t *testing.T) {
	sorter := New()
	sorter.SetMemoryLimit(0)
	if sorter.GetMemoryLimit() != 1000 {
		t.Errorf("expected default memory limit 1000, got %d", sorter.GetMemoryLimit())
	}

	sorter.SetMemoryLimit(-5)
	if sorter.GetMemoryLimit() != 1000 {
		t.Errorf("expected default memory limit 1000, got %d", sorter.GetMemoryLimit())
	}
}

func TestInvalidMergeWays(t *testing.T) {
	sorter := New()
	sorter.SetMergeWays(0)
	if sorter.GetMergeWays() != 2 {
		t.Errorf("expected default merge ways 2, got %d", sorter.GetMergeWays())
	}

	sorter.SetMergeWays(1)
	if sorter.GetMergeWays() != 2 {
		t.Errorf("expected default merge ways 2 (ways=1 is invalid), got %d", sorter.GetMergeWays())
	}
}

func TestMultiwayMerge(t *testing.T) {
	sorter := New()
	sorter.SetMemoryLimit(2)
	sorter.SetMergeWays(4)

	data := []int64{9, 3, 7, 1, 8, 2, 6, 4, 5, 10}
	sorter.AddData(data)

	if err := sorter.Sort(); err != nil {
		t.Fatalf("sort failed: %v", err)
	}

	result := sorter.GetFinalResult()
	expected := make([]int64, len(data))
	copy(expected, data)
	sort.Slice(expected, func(i, j int) bool {
		return expected[i] < expected[j]
	})

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	stats := sorter.GetStats()
	if stats.ChunksCreated != 5 {
		t.Errorf("expected 5 chunks, got %d", stats.ChunksCreated)
	}
}

func TestReset(t *testing.T) {
	sorter := New()
	sorter.SetMemoryLimit(3)

	data := []int64{3, 1, 2}
	sorter.AddData(data)
	sorter.Sort()

	if !sorter.IsSorted() {
		t.Error("expected sorted")
	}

	sorter.Reset()

	if sorter.IsSorted() {
		t.Error("expected not sorted after reset")
	}

	if len(sorter.GetData()) != 0 {
		t.Error("expected empty data after reset")
	}
}

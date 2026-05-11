package skiplist

import (
	"fmt"
	"testing"
)

func TestUserScenario(t *testing.T) {
	sl := New()

	sl.Put("only", []byte("v0"))

	for i := 1; i <= 20; i++ {
		key := fmt.Sprintf("key%02d", i)
		sl.Put(key, []byte(fmt.Sprintf("v%d", i)))
	}

	for i := 1; i <= 20; i++ {
		key := fmt.Sprintf("key%02d", i)
		rank, found := sl.Rank(key)
		if !found {
			t.Errorf("key %s should be found", key)
			continue
		}
		if rank != i {
			t.Errorf("key %s: expected rank %d, got %d", key, i, rank)
		}
	}

	rank, found := sl.Rank("only")
	if !found {
		t.Error("key only should be found")
	}
	if rank != 21 {
		t.Errorf("key only: expected rank 21, got %d", rank)
	}
}

func TestBasicRank(t *testing.T) {
	sl := New()

	sl.Put("key01", []byte("v1"))
	sl.Put("key02", []byte("v2"))
	sl.Put("key03", []byte("v3"))
	sl.Put("key04", []byte("v4"))
	sl.Put("key05", []byte("v5"))

	expected := map[string]int{
		"key01": 1,
		"key02": 2,
		"key03": 3,
		"key04": 4,
		"key05": 5,
	}

	for key, expRank := range expected {
		rank, found := sl.Rank(key)
		if !found {
			t.Errorf("key %s should be found", key)
			continue
		}
		if rank != expRank {
			t.Errorf("key %s: expected rank %d, got %d", key, expRank, rank)
		}
	}
}

func TestRankWithRandomLevels(t *testing.T) {
	sl := New()

	keys := []string{"aaa", "apple", "banana", "cherry", "date", "zzz"}
	for i, k := range keys {
		sl.Put(k, []byte(fmt.Sprintf("v%d", i+1)))
	}

	expected := map[string]int{
		"aaa":    1,
		"apple":  2,
		"banana": 3,
		"cherry": 4,
		"date":   5,
		"zzz":    6,
	}

	for key, expRank := range expected {
		rank, found := sl.Rank(key)
		if !found {
			t.Errorf("key %s should be found", key)
			continue
		}
		if rank != expRank {
			t.Errorf("key %s: expected rank %d, got %d", key, expRank, rank)
		}
	}
}

func TestRankWith20Keys(t *testing.T) {
	sl := New()

	for i := 1; i <= 20; i++ {
		key := fmt.Sprintf("key%02d", i)
		sl.Put(key, []byte(fmt.Sprintf("v%d", i)))
	}

	for i := 1; i <= 20; i++ {
		key := fmt.Sprintf("key%02d", i)
		rank, found := sl.Rank(key)
		if !found {
			t.Errorf("key %s should be found", key)
			continue
		}
		if rank != i {
			t.Errorf("key %s: expected rank %d, got %d", key, i, rank)
		}
	}
}

func TestDeleteAndRank(t *testing.T) {
	sl := New()

	for i := 1; i <= 10; i++ {
		key := fmt.Sprintf("key%02d", i)
		sl.Put(key, []byte(fmt.Sprintf("v%d", i)))
	}

	sl.Delete("key05")

	for i := 1; i <= 4; i++ {
		key := fmt.Sprintf("key%02d", i)
		rank, _ := sl.Rank(key)
		if rank != i {
			t.Errorf("key %s: expected rank %d, got %d", key, i, rank)
		}
	}

	for i := 6; i <= 10; i++ {
		key := fmt.Sprintf("key%02d", i)
		rank, _ := sl.Rank(key)
		if rank != i-1 {
			t.Errorf("key %s: expected rank %d, got %d", key, i-1, rank)
		}
	}
}

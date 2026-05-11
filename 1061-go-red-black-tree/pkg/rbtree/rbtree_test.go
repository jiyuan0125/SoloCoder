package rbtree

import (
	"math/rand"
	"sort"
	"testing"
)

func TestInsertAndSize(t *testing.T) {
	tree := New()

	if tree.Size() != 0 {
		t.Errorf("Expected size 0, got %d", tree.Size())
	}

	tree.Insert(5)
	tree.Insert(3)
	tree.Insert(7)
	tree.Insert(3)

	if tree.Size() != 4 {
		t.Errorf("Expected size 4, got %d", tree.Size())
	}

	if !tree.validateRBProperties() {
		t.Error("RB properties violated after inserts")
	}
}

func TestRangeQuery(t *testing.T) {
	tree := New()
	tree.Insert(5)
	tree.Insert(3)
	tree.Insert(7)
	tree.Insert(3)
	tree.Insert(9)
	tree.Insert(1)

	result := tree.RangeQuery(3, 7)
	expected := []int{3, 3, 5, 7}

	if len(result) != len(expected) {
		t.Errorf("Expected %d elements, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
		}
	}

	emptyResult := tree.RangeQuery(10, 20)
	if len(emptyResult) != 0 {
		t.Errorf("Expected empty result, got %v", emptyResult)
	}

	invalidRange := tree.RangeQuery(10, 5)
	if len(invalidRange) != 0 {
		t.Errorf("Expected empty result for invalid range, got %v", invalidRange)
	}
}

func TestKthLargest(t *testing.T) {
	tree := New()
	tree.Insert(3)
	tree.Insert(3)
	tree.Insert(5)

	val, err := tree.KthLargest(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != 5 {
		t.Errorf("Expected 5, got %d", val)
	}

	val, err = tree.KthLargest(2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != 3 {
		t.Errorf("Expected 3, got %d", val)
	}

	val, err = tree.KthLargest(3)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if val != 3 {
		t.Errorf("Expected 3, got %d", val)
	}

	_, err = tree.KthLargest(4)
	if err == nil {
		t.Error("Expected error for k out of range")
	}

	emptyTree := New()
	_, err = emptyTree.KthLargest(1)
	if err == nil {
		t.Error("Expected error for empty tree")
	}
}

func TestDelete(t *testing.T) {
	tree := New()
	tree.Insert(5)
	tree.Insert(3)
	tree.Insert(7)
	tree.Insert(3)
	tree.Insert(9)

	if !tree.Delete(3) {
		t.Error("Expected Delete to return true")
	}

	if tree.Size() != 4 {
		t.Errorf("Expected size 4, got %d", tree.Size())
	}

	if !tree.validateRBProperties() {
		t.Error("RB properties violated after delete")
	}

	if !tree.Delete(5) {
		t.Error("Expected Delete to return true")
	}

	if tree.Size() != 3 {
		t.Errorf("Expected size 3, got %d", tree.Size())
	}

	if !tree.validateRBProperties() {
		t.Error("RB properties violated after delete")
	}

	if tree.Delete(100) {
		t.Error("Expected Delete to return false for non-existent key")
	}
}

func TestRandomOperations(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	tree := New()
	values := []int{}

	for i := 0; i < 100; i++ {
		val := rng.Intn(100)
		tree.Insert(val)
		values = append(values, val)
		if !tree.validateRBProperties() {
			t.Fatalf("RB properties violated after inserting %d", val)
		}
	}

	for i := 0; i < 50; i++ {
		idx := rng.Intn(len(values))
		val := values[idx]
		tree.Delete(val)
		values = append(values[:idx], values[idx+1:]...)
		if !tree.validateRBProperties() {
			t.Fatalf("RB properties violated after deleting %d (iteration %d)", val, i)
		}
	}

	sort.Ints(values)
	result := tree.RangeQuery(values[0], values[len(values)-1])

	if len(result) != len(values) {
		t.Errorf("Expected %d elements, got %d", len(values), len(result))
	}

	for i, v := range result {
		if v != values[i] {
			t.Errorf("Expected %d at index %d, got %d", values[i], i, v)
		}
	}
}

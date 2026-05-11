package rtree

import (
	"fmt"
	"testing"
)

func TestInsertAndDelete20Objects(t *testing.T) {
	tree := New()

	ids := make([]string, 0, 20)
	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("obj_%d", i)
		ids = append(ids, id)
		obj := NewSimpleObject(id, float64(i), float64(i), float64(i+5), float64(i+5), nil)
		err := tree.Insert(obj)
		if err != nil {
			t.Fatalf("Insert %s failed: %v", id, err)
		}
	}

	if tree.Count() != 20 {
		t.Errorf("Expected count 20, got %d", tree.Count())
	}

	t.Logf("Before delete - NodeCount: %d, Height: %d, Count: %d",
		tree.NodeCount(), tree.Height(), tree.Count())

	deleted := 0
	notFound := 0
	for _, id := range ids {
		err := tree.Delete(id)
		if err != nil {
			t.Logf("Delete %s failed: %v (count=%d)", id, err, tree.Count())
			notFound++
		} else {
			deleted++
			t.Logf("Deleted %s successfully (count=%d)", id, tree.Count())
		}
	}

	t.Logf("After delete - deleted=%d, notFound=%d, remaining=%d",
		deleted, notFound, tree.Count())

	if deleted != 20 {
		t.Errorf("Expected to delete 20 objects, but deleted=%d, notFound=%d, remaining=%d",
			deleted, notFound, tree.Count())
	}

	if tree.Count() != 0 {
		t.Errorf("Expected count 0 after deleting all, got %d", tree.Count())
	}
}

func TestDeleteOneByOne(t *testing.T) {
	for testNum := 0; testNum < 5; testNum++ {
		tree := New()

		ids := make([]string, 0, 20)
		for i := 0; i < 20; i++ {
			id := fmt.Sprintf("obj_%d", i)
			ids = append(ids, id)
			x := float64(i * 3)
			y := float64(i * 2)
			obj := NewSimpleObject(id, x, y, x+10, y+10, nil)
			err := tree.Insert(obj)
			if err != nil {
				t.Fatalf("Test %d: Insert %s failed: %v", testNum, id, err)
			}
		}

		if tree.Count() != 20 {
			t.Fatalf("Test %d: Expected count 20, got %d", testNum, tree.Count())
		}

		for i, id := range ids {
			err := tree.Delete(id)
			if err != nil {
				t.Errorf("Test %d: Delete %d (%s) failed: %v. Remaining count: %d",
					testNum, i, id, err, tree.Count())
			}
		}

		if tree.Count() != 0 {
			t.Errorf("Test %d: Expected count 0, got %d", testNum, tree.Count())
		}
	}
}

func TestCondenseTreeLeavesObjects(t *testing.T) {
	tree := New()

	for i := 0; i < 30; i++ {
		id := fmt.Sprintf("obj_%d", i)
		x := float64(i % 5)
		y := float64(i / 5)
		obj := NewSimpleObject(id, x, y, x+1, y+1, nil)
		err := tree.Insert(obj)
		if err != nil {
			t.Fatalf("Insert %s failed: %v", id, err)
		}
	}

	if tree.Count() != 30 {
		t.Errorf("Expected count 30, got %d", tree.Count())
	}

	for i := 0; i < 25; i++ {
		id := fmt.Sprintf("obj_%d", i)
		err := tree.Delete(id)
		if err != nil {
			t.Errorf("Delete %s failed: %v", id, err)
		}
	}

	remaining := tree.Search(-1000, -1000, 1000, 1000)
	t.Logf("Remaining objects: %d", len(remaining))

	for _, obj := range remaining {
		t.Logf("  - %s", obj.ID())
	}

	if tree.Count() != 5 {
		t.Errorf("Expected count 5, got %d", tree.Count())
	}

	expectedIds := make(map[string]bool)
	for i := 25; i < 30; i++ {
		expectedIds[fmt.Sprintf("obj_%d", i)] = true
	}

	for _, obj := range remaining {
		if !expectedIds[obj.ID()] {
			t.Errorf("Unexpected remaining object: %s", obj.ID())
		}
		delete(expectedIds, obj.ID())
	}

	if len(expectedIds) > 0 {
		t.Errorf("Missing expected objects: %v", expectedIds)
	}
}

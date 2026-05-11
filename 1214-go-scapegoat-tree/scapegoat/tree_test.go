package scapegoat

import "testing"

func TestBasicOperations(t *testing.T) {
	tree := NewTree()

	if tree.Size() != 0 {
		t.Errorf("expected size 0, got %d", tree.Size())
	}

	tree.Insert(5, 50)
	tree.Insert(3, 30)
	tree.Insert(7, 70)
	tree.Insert(2, 20)
	tree.Insert(4, 40)
	tree.Insert(6, 60)
	tree.Insert(8, 80)

	if tree.Size() != 7 {
		t.Errorf("expected size 7, got %d", tree.Size())
	}

	val, err := tree.Search(3)
	if err != nil || val != 30 {
		t.Errorf("expected 30, got %d, err: %v", val, err)
	}

	err = tree.Delete(5)
	if err != nil {
		t.Errorf("delete failed: %v", err)
	}

	if tree.Size() != 6 {
		t.Errorf("expected size 6 after delete, got %d", tree.Size())
	}

	_, err = tree.Search(5)
	if err == nil {
		t.Errorf("expected error for deleted key 5")
	}
}

func TestEmptyTreeDelete(t *testing.T) {
	tree := NewTree()
	err := tree.Delete(1)
	if err != ErrEmptyTree {
		t.Errorf("expected ErrEmptyTree, got %v", err)
	}
}

func TestSearchNotFound(t *testing.T) {
	tree := NewTree()
	tree.Insert(1, 10)
	_, err := tree.Search(2)
	if err != ErrKeyNotFound {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestReinsertDeletedKey(t *testing.T) {
	tree := NewTree()
	tree.Insert(1, 10)
	tree.Delete(1)
	if tree.Size() != 0 {
		t.Errorf("expected size 0 after delete")
	}

	tree.Insert(1, 100)
	if tree.Size() != 1 {
		t.Errorf("expected size 1 after reinsert")
	}

	val, err := tree.Search(1)
	if err != nil || val != 100 {
		t.Errorf("expected 100, got %d, err: %v", val, err)
	}
}

func TestRebuildOnInsert(t *testing.T) {
	tree := NewTree()

	for i := 0; i < 100; i++ {
		tree.Insert(int64(i), int64(i*10))
	}

	if tree.Size() != 100 {
		t.Errorf("expected size 100, got %d", tree.Size())
	}

	for i := 0; i < 100; i++ {
		val, err := tree.Search(int64(i))
		if err != nil || val != int64(i*10) {
			t.Errorf("search failed for %d: got %d, err: %v", i, val, err)
		}
	}
}

func TestFullRebuildOnDelete(t *testing.T) {
	tree := NewTree()

	for i := 0; i < 10; i++ {
		tree.Insert(int64(i), int64(i*10))
	}

	for i := 0; i < 6; i++ {
		tree.Delete(int64(i))
	}

	if tree.Size() != 4 {
		t.Errorf("expected size 4, got %d", tree.Size())
	}

	for i := 0; i < 6; i++ {
		_, err := tree.Search(int64(i))
		if err == nil {
			t.Errorf("expected error for deleted key %d", i)
		}
	}

	for i := 6; i < 10; i++ {
		val, err := tree.Search(int64(i))
		if err != nil || val != int64(i*10) {
			t.Errorf("search failed for %d: got %d, err: %v", i, val, err)
		}
	}
}

func TestUpdateExistingKey(t *testing.T) {
	tree := NewTree()
	tree.Insert(1, 10)

	val, _ := tree.Search(1)
	if val != 10 {
		t.Errorf("expected 10, got %d", val)
	}

	tree.Insert(1, 100)
	if tree.Size() != 1 {
		t.Errorf("expected size 1, got %d", tree.Size())
	}

	val, _ = tree.Search(1)
	if val != 100 {
		t.Errorf("expected 100, got %d", val)
	}
}

func TestDeleteAll(t *testing.T) {
	tree := NewTree()
	tree.Insert(1, 10)
	tree.Insert(2, 20)

	tree.Delete(1)
	tree.Delete(2)

	if tree.Size() != 0 {
		t.Errorf("expected size 0, got %d", tree.Size())
	}

	_, err := tree.Search(1)
	if err == nil {
		t.Errorf("expected error for deleted key 1")
	}

	_, err = tree.Search(2)
	if err == nil {
		t.Errorf("expected error for deleted key 2")
	}

	tree.Insert(3, 30)
	if tree.Size() != 1 {
		t.Errorf("expected size 1, got %d", tree.Size())
	}
}

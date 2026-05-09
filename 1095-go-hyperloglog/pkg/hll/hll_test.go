package hll

import (
	"testing"
)

func TestNewHLL(t *testing.T) {
	_, err := NewHLL(3)
	if err == nil {
		t.Error("expected error for precision 3")
	}

	_, err = NewHLL(19)
	if err == nil {
		t.Error("expected error for precision 19")
	}

	hll, err := NewHLL(14)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if hll.Precision() != 14 {
		t.Errorf("expected precision 14, got %d", hll.Precision())
	}
}

func TestEmptyHLLCount(t *testing.T) {
	hll, _ := NewHLL(14)
	count := hll.Count()
	if count != 0.0 {
		t.Errorf("expected 0 for empty HLL, got %f", count)
	}
}

func TestSingleElement(t *testing.T) {
	hll, _ := NewHLL(14)
	hll.Add("test")
	count := hll.Count()
	if count < 0.5 || count > 1.5 {
		t.Errorf("expected count close to 1, got %f", count)
	}
}

func TestSparseDenseTransition(t *testing.T) {
	hll, _ := NewHLL(4)

	for i := 0; i < 10; i++ {
		hll.Add(string(rune('a' + i)))
	}

	originalCount := hll.Count()

	serialized, _ := hll.Serialize()
	deserialized, _ := Deserialize(serialized)

	if deserialized.Count() != originalCount {
		t.Errorf("count mismatch after serialization/deserialization")
	}
}

func TestMerge(t *testing.T) {
	hll1, _ := NewHLL(14)
	hll2, _ := NewHLL(14)

	for i := 0; i < 100; i++ {
		hll1.Add(string(rune('a' + i%26)))
	}

	for i := 0; i < 100; i++ {
		hll2.Add(string(rune('A' + i%26)))
	}

	hll1.Merge(hll2)

	if hll1.Count() < 30 {
		t.Errorf("expected count > 30 after merge, got %f", hll1.Count())
	}
}

func TestPrecisionMismatch(t *testing.T) {
	hll1, _ := NewHLL(14)
	hll2, _ := NewHLL(15)

	err := hll1.Merge(hll2)
	if err == nil {
		t.Error("expected error for precision mismatch")
	}
}

func TestSerialization(t *testing.T) {
	hll1, _ := NewHLL(14)

	for i := 0; i < 1000; i++ {
		hll1.Add(string(rune(i)))
	}

	originalCount := hll1.Count()

	serialized, err := hll1.Serialize()
	if err != nil {
		t.Errorf("serialization failed: %v", err)
	}

	if len(serialized) == 0 {
		t.Error("expected non-empty serialized data")
	}

	hll2, err := Deserialize(serialized)
	if err != nil {
		t.Errorf("deserialization failed: %v", err)
	}

	if hll2.Count() != originalCount {
		t.Errorf("count mismatch: original %f, deserialized %f", originalCount, hll2.Count())
	}

	if hll2.Precision() != 14 {
		t.Errorf("precision mismatch: expected 14, got %d", hll2.Precision())
	}
}

func TestInvalidSerialization(t *testing.T) {
	_, err := Deserialize([]byte("invalid"))
	if err == nil {
		t.Error("expected error for invalid data")
	}
}

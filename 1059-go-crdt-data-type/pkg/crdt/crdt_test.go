package crdt

import (
	"testing"
)

func TestGCounter_Basic(t *testing.T) {
	gc := NewGCounter()

	if gc.Value() != 0 {
		t.Errorf("Expected initial value 0, got %d", gc.Value())
	}

	gc.Increment("node1")
	gc.Increment("node1")
	gc.Increment("node2")

	if gc.Value() != 3 {
		t.Errorf("Expected value 3, got %d", gc.Value())
	}

	if gc.Counters["node1"] != 2 {
		t.Errorf("Expected node1 counter 2, got %d", gc.Counters["node1"])
	}
	if gc.Counters["node2"] != 1 {
		t.Errorf("Expected node2 counter 1, got %d", gc.Counters["node2"])
	}
}

func TestGCounter_IncrementBy(t *testing.T) {
	gc := NewGCounter()
	gc.IncrementBy("node1", 5)

	if gc.Value() != 5 {
		t.Errorf("Expected value 5, got %d", gc.Value())
	}

	if err := gc.IncrementBy("node1", -1); err == nil {
		t.Errorf("Expected error for negative delta")
	}
}

func TestGCounter_Merge_Commutative(t *testing.T) {
	a := NewGCounter()
	b := NewGCounter()

	a.IncrementBy("node1", 3)
	b.IncrementBy("node2", 5)

	mergeAB := a.Merge(b)
	mergeBA := b.Merge(a)

	if !mergeAB.Equals(mergeBA) {
		t.Error("merge(A, B) should equal merge(B, A) (commutative)")
	}

	if mergeAB.Value() != 8 {
		t.Errorf("Expected merged value 8, got %d", mergeAB.Value())
	}
}

func TestGCounter_Merge_Associative(t *testing.T) {
	a := NewGCounter()
	b := NewGCounter()
	c := NewGCounter()

	a.IncrementBy("node1", 2)
	b.IncrementBy("node2", 3)
	c.IncrementBy("node3", 4)

	left := a.Merge(b).Merge(c)
	right := a.Merge(b.Merge(c))

	if !left.Equals(right) {
		t.Error("merge(merge(A, B), C) should equal merge(A, merge(B, C)) (associative)")
	}

	if left.Value() != 9 {
		t.Errorf("Expected merged value 9, got %d", left.Value())
	}
}

func TestGCounter_Merge_Idempotent(t *testing.T) {
	a := NewGCounter()
	b := NewGCounter()

	a.IncrementBy("node1", 5)
	b.IncrementBy("node2", 3)

	merged1 := a.Merge(b)
	merged2 := merged1.Merge(b)

	if !merged1.Equals(merged2) {
		t.Error("Merge should be idempotent (multiple merges don't change result)")
	}

	if merged1.Value() != 8 {
		t.Errorf("Expected merged value 8, got %d", merged1.Value())
	}
}

func TestGCounter_Merge_NodeIDs(t *testing.T) {
	a := NewGCounter()
	b := NewGCounter()

	a.Increment("alice")
	b.Increment("bob")

	merged := a.Merge(b)

	if _, exists := merged.Counters["alice"]; !exists {
		t.Error("Merged G-Counter should contain node 'alice'")
	}
	if _, exists := merged.Counters["bob"]; !exists {
		t.Error("Merged G-Counter should contain node 'bob'")
	}
}

func TestGCounter_Merge_NoDoubleCounting(t *testing.T) {
	a := NewGCounter()
	b := NewGCounter()

	a.IncrementBy("node1", 5)
	b.IncrementBy("node1", 3)

	merged := a.Merge(b)

	if merged.Value() != 5 {
		t.Errorf("Merge should use max(5, 3)=5, got %d. Simple addition would give 8 which is wrong!", merged.Value())
	}
}

func TestPNCounter_Basic(t *testing.T) {
	pn := NewPNCounter()

	if pn.Value() != 0 {
		t.Errorf("Expected initial value 0, got %d", pn.Value())
	}

	pn.Increment("node1")
	pn.Increment("node1")
	pn.Decrement("node1")

	if pn.Value() != 1 {
		t.Errorf("Expected value 1, got %d", pn.Value())
	}
}

func TestPNCounter_IncrementBy_DecrementBy(t *testing.T) {
	pn := NewPNCounter()
	pn.IncrementBy("node1", 10)
	pn.DecrementBy("node1", 3)

	if pn.Value() != 7 {
		t.Errorf("Expected value 7, got %d", pn.Value())
	}
}

func TestPNCounter_Merge_Commutative(t *testing.T) {
	a := NewPNCounter()
	b := NewPNCounter()

	a.IncrementBy("node1", 10)
	a.DecrementBy("node1", 2)

	b.IncrementBy("node2", 5)
	b.DecrementBy("node2", 1)

	mergeAB := a.Merge(b)
	mergeBA := b.Merge(a)

	if !mergeAB.Equals(mergeBA) {
		t.Error("PN-Counter merge should be commutative")
	}

	expectedValue := (10 + 5) - (2 + 1)
	if mergeAB.Value() != expectedValue {
		t.Errorf("Expected value %d, got %d", expectedValue, mergeAB.Value())
	}
}

func TestPNCounter_Merge_Associative(t *testing.T) {
	a := NewPNCounter()
	b := NewPNCounter()
	c := NewPNCounter()

	a.IncrementBy("node1", 10)
	a.DecrementBy("node1", 2)

	b.IncrementBy("node2", 5)
	b.DecrementBy("node2", 3)

	c.IncrementBy("node3", 8)
	c.DecrementBy("node3", 1)

	left := a.Merge(b).Merge(c)
	right := a.Merge(b.Merge(c))

	if !left.Equals(right) {
		t.Error("PN-Counter merge should be associative")
	}

	expectedValue := (10 + 5 + 8) - (2 + 3 + 1)
	if left.Value() != expectedValue {
		t.Errorf("Expected value %d, got %d", expectedValue, left.Value())
	}
}

func TestGSet_Basic(t *testing.T) {
	s := NewGSet()

	if s.Size() != 0 {
		t.Errorf("Expected size 0, got %d", s.Size())
	}

	s.Add("apple")
	s.Add("banana")
	s.Add("apple")

	if s.Size() != 2 {
		t.Errorf("Expected size 2, got %d", s.Size())
	}

	if !s.Contains("apple") {
		t.Error("Set should contain 'apple'")
	}
	if !s.Contains("banana") {
		t.Error("Set should contain 'banana'")
	}
	if s.Contains("cherry") {
		t.Error("Set should not contain 'cherry'")
	}
}

func TestGSet_ExactMatch(t *testing.T) {
	s := NewGSet()
	s.Add("123")

	if s.Contains("123") != true {
		t.Error("Set should contain string '123'")
	}
	if s.Contains(" 123") != false {
		t.Error("Set should use exact matching (leading space matters)")
	}
	if s.Contains("123 ") != false {
		t.Error("Set should use exact matching (trailing space matters)")
	}
}

func TestGSet_Merge_Commutative(t *testing.T) {
	a := NewGSet()
	b := NewGSet()

	a.Add("apple")
	a.Add("banana")

	b.Add("banana")
	b.Add("cherry")

	mergeAB := a.Merge(b)
	mergeBA := b.Merge(a)

	if !mergeAB.Equals(mergeBA) {
		t.Error("G-Set merge should be commutative")
	}

	if mergeAB.Size() != 3 {
		t.Errorf("Expected size 3 after merge, got %d", mergeAB.Size())
	}
}

func TestGSet_Merge_Associative(t *testing.T) {
	a := NewGSet()
	b := NewGSet()
	c := NewGSet()

	a.Add("apple")
	b.Add("banana")
	c.Add("cherry")

	left := a.Merge(b).Merge(c)
	right := a.Merge(b.Merge(c))

	if !left.Equals(right) {
		t.Error("G-Set merge should be associative")
	}

	if left.Size() != 3 {
		t.Errorf("Expected size 3, got %d", left.Size())
	}
}

func TestGSet_Merge_Idempotent(t *testing.T) {
	a := NewGSet()
	b := NewGSet()

	a.Add("apple")
	b.Add("banana")

	merged1 := a.Merge(b)
	merged2 := merged1.Merge(b)

	if !merged1.Equals(merged2) {
		t.Error("G-Set merge should be idempotent")
	}
}

func TestGCounter_StringNodeIDs(t *testing.T) {
	gc := NewGCounter()

	nodeIDs := []string{"node-1", "server-alpha", "user@example.com", "中文节点"}

	for i, id := range nodeIDs {
		gc.IncrementBy(id, i+1)
	}

	if gc.Value() != 10 {
		t.Errorf("Expected value 10, got %d", gc.Value())
	}

	for i, id := range nodeIDs {
		if gc.Counters[id] != i+1 {
			t.Errorf("Expected counter %d for node %s, got %d", i+1, id, gc.Counters[id])
		}
	}
}

func TestGCounter_Merge_DifferentNodes(t *testing.T) {
	a := NewGCounter()
	b := NewGCounter()

	a.IncrementBy("alice", 5)
	b.IncrementBy("bob", 3)

	merged := a.Merge(b)

	if len(merged.Counters) != 2 {
		t.Errorf("Expected 2 nodes in merged counter, got %d", len(merged.Counters))
	}

	if merged.Counters["alice"] != 5 {
		t.Errorf("Expected alice=5, got %d", merged.Counters["alice"])
	}
	if merged.Counters["bob"] != 3 {
		t.Errorf("Expected bob=3, got %d", merged.Counters["bob"])
	}
}

func TestGCounter_Merge_SameNodeMax(t *testing.T) {
	a := NewGCounter()
	b := NewGCounter()

	a.IncrementBy("node1", 10)
	b.IncrementBy("node1", 7)

	merged := a.Merge(b)

	if merged.Counters["node1"] != 10 {
		t.Errorf("Merge should take max(10, 7)=10, got %d", merged.Counters["node1"])
	}

	merged2 := b.Merge(a)
	if merged2.Counters["node1"] != 10 {
		t.Errorf("Merge should be commutative, expected 10, got %d", merged2.Counters["node1"])
	}
}

func TestGCounter_JsonSerialization(t *testing.T) {
	gc := NewGCounter()
	gc.IncrementBy("node1", 5)
	gc.IncrementBy("node2", 3)

	data, err := gc.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	decoded, err := UnmarshalGCounter(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !gc.Equals(decoded) {
		t.Error("Serialization round-trip failed")
	}
}

func TestPNCounter_JsonSerialization(t *testing.T) {
	pn := NewPNCounter()
	pn.IncrementBy("node1", 10)
	pn.DecrementBy("node1", 3)

	data, err := pn.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	decoded, err := UnmarshalPNCounter(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !pn.Equals(decoded) {
		t.Error("Serialization round-trip failed")
	}
}

func TestGSet_JsonSerialization(t *testing.T) {
	s := NewGSet()
	s.Add("apple")
	s.Add("banana")

	data, err := s.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	decoded, err := UnmarshalGSet(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !s.Equals(decoded) {
		t.Error("Serialization round-trip failed")
	}
}

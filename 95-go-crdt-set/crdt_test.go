package crdt

import (
	"encoding/json"
	"testing"
)

func TestORSet_BasicOperations(t *testing.T) {
	set := NewORSet()

	if set.Contains("x") {
		t.Error("Expected set to not contain 'x' initially")
	}

	set.Add("x")
	if !set.Contains("x") {
		t.Error("Expected set to contain 'x' after Add")
	}

	elements := set.Elements()
	if len(elements) != 1 || elements[0] != "x" {
		t.Errorf("Expected Elements() to return ['x'], got %v", elements)
	}

	set.Remove("x")
	if set.Contains("x") {
		t.Error("Expected set to not contain 'x' after Remove")
	}

	elements = set.Elements()
	if len(elements) != 0 {
		t.Errorf("Expected Elements() to return empty, got %v", elements)
	}
}

func TestORSet_AddRemoveAdd(t *testing.T) {
	set := NewORSet()

	set.Add("x")
	set.Remove("x")
	set.Add("x")

	if !set.Contains("x") {
		t.Error("Expected set to contain 'x' after Add-Remove-Add pattern")
	}
}

func TestORSet_Merge_Basic(t *testing.T) {
	setA := NewORSet()
	setB := NewORSet()

	setA.Add("x")
	setB.Add("y")

	setA.Merge(setB)

	if !setA.Contains("x") {
		t.Error("Expected setA to contain 'x' after merge")
	}
	if !setA.Contains("y") {
		t.Error("Expected setA to contain 'y' after merge")
	}
}

func TestORSet_Merge_ConcurrentAdd(t *testing.T) {
	setA := NewORSet()
	setB := NewORSet()

	setA.Add("x")
	setB.Add("x")

	mergedA := setA.Clone()
	mergedA.Merge(setB)

	elements := mergedA.Elements()
	if len(elements) != 1 {
		t.Errorf("Expected merge of concurrent adds to have 1 element, got %d", len(elements))
	}
	if elements[0] != "x" {
		t.Errorf("Expected element to be 'x', got %s", elements[0])
	}
}

func TestORSet_KeySemantic(t *testing.T) {
	nodeA := NewORSet()
	nodeB := NewORSet()

	nodeA.Add("x")
	nodeA.Remove("x")
	nodeA.Add("x")

	nodeAFinal := nodeA.Clone()

	nodeB.Add("x")
	nodeB.Remove("x")

	if !nodeA.Contains("x") {
		t.Error("Node A should see 'x' after add-remove-add")
	}

	if nodeB.Contains("x") {
		t.Error("Node B should NOT see 'x' after add-then-remove")
	}

	nodeB.Merge(nodeAFinal)
	if !nodeB.Contains("x") {
		t.Error("Node B should see 'x' after merging with Node A's state (A has tag=2 which is not removed)")
	}
}



func TestORSet_Clone(t *testing.T) {
	original := NewORSet()
	original.Add("x")
	original.Add("y")

	clone := original.Clone()
	clone.Remove("x")

	if !original.Contains("x") {
		t.Error("Original set should not be affected by clone modifications")
	}
	if clone.Contains("x") {
		t.Error("Clone should have 'x' removed")
	}
}

func TestORSet_JSONSerialization(t *testing.T) {
	set := NewORSet()
	set.Add("x")
	set.Add("y")
	set.Remove("x")
	set.Add("x")

	data, err := set.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	restored := NewORSet()
	if err := restored.UnmarshalJSON(data); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if !restored.Contains("x") {
		t.Error("Restored set should contain 'x'")
	}
	if !restored.Contains("y") {
		t.Error("Restored set should contain 'y'")
	}

	restored.Add("z")
	if !restored.Contains("z") {
		t.Error("Restored set should support adding new elements")
	}
}

func TestORSet_JSONFormat(t *testing.T) {
	set := NewORSet()
	set.Add("x")
	set.Remove("x")
	set.Add("x")

	data, err := set.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	elements, ok := obj["elements"].([]interface{})
	if !ok {
		t.Error("JSON should have 'elements' array")
	}

	if len(elements) != 1 {
		t.Errorf("Expected 1 element in JSON, got %d", len(elements))
	}

	elem := elements[0].(map[string]interface{})
	if elem["value"] != "x" {
		t.Errorf("Expected value 'x', got %v", elem["value"])
	}

	addTags, ok := elem["add_tags"].([]interface{})
	if !ok || len(addTags) != 2 {
		t.Errorf("Expected add_tags array with 2 elements, got %v", addTags)
	}

	removeTags, ok := elem["remove_tags"].([]interface{})
	if !ok || len(removeTags) != 1 {
		t.Errorf("Expected remove_tags array with 1 element, got %v", removeTags)
	}

	t.Logf("JSON: %s", string(data))
}

type MergeSimulatorResult struct {
	NodeA_Elements        []string
	NodeB_Elements        []string
	MergedA_Elements      []string
	MergedB_Elements      []string
	Converged             bool
}

func MergeSimulator(
	nodeAOperations func(*ORSet),
	nodeBOperations func(*ORSet),
) MergeSimulatorResult {
	nodeA := NewORSet()
	nodeB := NewORSet()

	nodeAOperations(nodeA)

	partialA := nodeA.Clone()

	nodeBOperations(nodeB)

	nodeB.Merge(partialA)

	mergedA := nodeA.Clone()
	mergedA.Merge(nodeB)

	mergedB := nodeB.Clone()
	mergedB.Merge(nodeA)

	elementsA := mergedA.Elements()
	elementsB := mergedB.Elements()

	converged := len(elementsA) == len(elementsB)
	if converged {
		for i, e := range elementsA {
			if elementsB[i] != e {
				converged = false
				break
			}
		}
	}

	return MergeSimulatorResult{
		NodeA_Elements:   nodeA.Elements(),
		NodeB_Elements:   nodeB.Elements(),
		MergedA_Elements: elementsA,
		MergedB_Elements: elementsB,
		Converged:        converged,
	}
}

func TestMergeSimulator_Basic(t *testing.T) {
	result := MergeSimulator(
		func(a *ORSet) {
			a.Add("x")
		},
		func(b *ORSet) {
			b.Add("y")
		},
	)

	if !result.Converged {
		t.Error("Merge should result in convergence")
	}

	if len(result.MergedA_Elements) != 2 {
		t.Errorf("Expected 2 elements after merge, got %d", len(result.MergedA_Elements))
	}
}

func TestMergeSimulator_ConcurrentAddSameElement(t *testing.T) {
	result := MergeSimulator(
		func(a *ORSet) {
			a.Add("x")
		},
		func(b *ORSet) {
			b.Add("x")
		},
	)

	if !result.Converged {
		t.Error("Merge should result in convergence")
	}

	if len(result.MergedA_Elements) != 1 {
		t.Errorf("Expected 1 element after concurrent adds merge, got %d", len(result.MergedA_Elements))
	}
	if result.MergedA_Elements[0] != "x" {
		t.Errorf("Expected element to be 'x', got %s", result.MergedA_Elements[0])
	}
}

func TestMergeSimulator_AddRemoveAdd(t *testing.T) {
	result := MergeSimulator(
		func(a *ORSet) {
			a.Add("x")
			a.Remove("x")
			a.Add("x")
		},
		func(b *ORSet) {
		},
	)

	if !result.Converged {
		t.Error("Merge should result in convergence")
	}

	if len(result.MergedA_Elements) != 1 {
		t.Errorf("Expected 1 element, got %d", len(result.MergedA_Elements))
	}
}

func TestMergeSimulator_KeySemantic(t *testing.T) {
	result := MergeSimulator(
		func(a *ORSet) {
			a.Add("x")
			a.Remove("x")
			a.Add("x")
		},
		func(b *ORSet) {
			b.Add("x")
			b.Remove("x")
		},
	)

	if !result.Converged {
		t.Error("Merge should result in convergence")
	}

	if len(result.MergedA_Elements) != 1 {
		t.Errorf("Expected 1 element after merge (Node A's tag=2 should not be removed), got %d: %v",
			len(result.MergedA_Elements), result.MergedA_Elements)
	}
}

func TestORSet_IdempotentAdd(t *testing.T) {
	set := NewORSet()

	set.Add("x")
	set.Add("x")

	elements := set.Elements()
	if len(elements) != 1 {
		t.Errorf("Expected 1 element after two adds, got %d", len(elements))
	}

	state := set.elements["x"]
	if len(state.AddTags) != 2 {
		t.Errorf("Expected 2 add_tags after two adds, got %d", len(state.AddTags))
	}
}

func TestORSet_RemoveIdempotent(t *testing.T) {
	set := NewORSet()

	set.Add("x")
	set.Remove("x")
	set.Remove("x")

	if set.Contains("x") {
		t.Error("Should not contain 'x' after remove")
	}

	state := set.elements["x"]
	if len(state.RemoveTags) != 1 {
		t.Errorf("Expected 1 remove_tag after two removes (idempotent), got %d", len(state.RemoveTags))
	}
}

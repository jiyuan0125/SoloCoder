package crdt

import (
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

	nodeBMid := nodeA.Clone()
	nodeB.Remove("x")
	nodeB.Add("x")

	if !nodeA.Contains("x") {
		t.Error("Node A should see 'x' after add-remove-add")
	}

	if nodeB.Contains("x") {
		t.Error("Node B should NOT see 'x' - it only saw remove and new add, but remove includes the new tag")
	}

	nodeB.Merge(nodeBMid)
	if !nodeB.Contains("x") {
		t.Error("Node B should see 'x' after merging with Node A's state")
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
			b.Remove("x")
			b.Add("x")
		},
	)

	if !result.Converged {
		t.Error("Merge should result in convergence")
	}

	if len(result.MergedA_Elements) != 1 {
		t.Errorf("Expected 1 element after merge (Node A's add tags should win), got %d: %v",
			len(result.MergedA_Elements), result.MergedA_Elements)
	}
}

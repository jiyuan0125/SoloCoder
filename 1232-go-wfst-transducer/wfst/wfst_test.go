package wfst

import "testing"

func TestSimpleWFST(t *testing.T) {
	f := New()
	
	for f.NumStates() <= 1 {
		f.AddState()
	}
	
	f.AddArc(0, Arc{Input: "a", Output: "x", Weight: 0.1, Next: 1})
	
	if f.NumStates() != 2 {
		t.Errorf("Expected 2 states, got %d", f.NumStates())
	}
	
	if !f.IsFinal(1) {
		t.Errorf("State 1 should be final")
	}
	
	result := Viterbi(f, []string{"a"})
	if result == nil {
		t.Errorf("Expected to find a path")
		return
	}
	
	if len(result.Path) != 2 || result.Path[0] != 0 || result.Path[1] != 1 {
		t.Errorf("Expected path [0, 1], got %v", result.Path)
	}
	
	if result.TotalWeight != 0.1 {
		t.Errorf("Expected weight 0.1, got %f", result.TotalWeight)
	}
}

func TestThreeStates(t *testing.T) {
	f := New()
	
	for f.NumStates() <= 2 {
		f.AddState()
	}
	
	f.AddArc(0, Arc{Input: "a", Output: "x", Weight: 0.1, Next: 1})
	f.AddArc(1, Arc{Input: "b", Output: "y", Weight: 0.2, Next: 2})
	
	if f.NumStates() != 3 {
		t.Errorf("Expected 3 states, got %d", f.NumStates())
	}
	
	if f.IsFinal(1) {
		t.Errorf("State 1 should NOT be final (has out arc)")
	}
	
	if !f.IsFinal(2) {
		t.Errorf("State 2 should be final")
	}
	
	result1 := Viterbi(f, []string{"a"})
	if result1 != nil {
		t.Errorf("Expected NO path for ['a'] (state 1 is not final), but got %v", result1.Path)
	}
	
	result2 := Viterbi(f, []string{"a", "b"})
	if result2 == nil {
		t.Errorf("Expected to find a path for ['a', 'b']")
		return
	}
	
	if len(result2.Path) != 3 || result2.Path[0] != 0 || result2.Path[1] != 1 || result2.Path[2] != 2 {
		t.Errorf("Expected path [0, 1, 2], got %v", result2.Path)
	}
}

func TestThreeStatesAllFinal(t *testing.T) {
	f := New()
	
	for f.NumStates() <= 2 {
		f.AddState()
	}
	
	f.AddArc(0, Arc{Input: "a", Output: "x", Weight: 0.1, Next: 1})
	
	if f.NumStates() != 3 {
		t.Errorf("Expected 3 states, got %d", f.NumStates())
	}
	
	if !f.IsFinal(1) {
		t.Errorf("State 1 should be final")
	}
	
	if !f.IsFinal(2) {
		t.Errorf("State 2 should be final")
	}
	
	result := Viterbi(f, []string{"a"})
	if result == nil {
		t.Errorf("Expected to find a path")
		return
	}
	
	if len(result.Path) != 2 || result.Path[0] != 0 || result.Path[1] != 1 {
		t.Errorf("Expected path [0, 1], got %v", result.Path)
	}
}

func TestServerCreateLogic(t *testing.T) {
	type testArc struct {
		From int
		Next int
	}
	
	tests := []struct {
		name          string
		arcs          []testArc
		expectedStates int
	}{
		{"2 states (0->1)", []testArc{{0, 1}}, 2},
		{"3 states (0->1, 1->2)", []testArc{{0, 1}, {1, 2}}, 3},
		{"3 states (0->2)", []testArc{{0, 2}}, 3},
		{"4 states (0->1, 2->3)", []testArc{{0, 1}, {2, 3}}, 4},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := New()
			
			maxStateID := 0
			for _, arc := range tt.arcs {
				if arc.From > maxStateID {
					maxStateID = arc.From
				}
				if arc.Next > maxStateID {
					maxStateID = arc.Next
				}
			}
			
			for f.NumStates() <= maxStateID {
				f.AddState()
			}
			
			if f.NumStates() != tt.expectedStates {
				t.Errorf("Expected %d states, got %d", tt.expectedStates, f.NumStates())
			}
		})
	}
}

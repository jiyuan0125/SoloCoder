package wfst

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

const Epsilon = ""

type Arc struct {
	Input  string
	Output string
	Weight float64
	Next   int
}

type WFST struct {
	states map[int][]Arc
	nextID int
}

func New() *WFST {
	fst := &WFST{
		states: make(map[int][]Arc),
		nextID: 0,
	}
	fst.AddState()
	return fst
}

func (f *WFST) AddState() int {
	id := f.nextID
	f.states[id] = []Arc{}
	f.nextID++
	return id
}

func (f *WFST) AddArc(from int, arc Arc) {
	if _, exists := f.states[from]; !exists {
		f.states[from] = []Arc{}
		if from >= f.nextID {
			f.nextID = from + 1
		}
	}
	f.states[from] = append(f.states[from], arc)
}

func (f *WFST) GetArcs(state int) []Arc {
	if arcs, exists := f.states[state]; exists {
		return arcs
	}
	return nil
}

func (f *WFST) NumStates() int {
	return len(f.states)
}

func (f *WFST) States() []int {
	states := make([]int, 0, len(f.states))
	for s := range f.states {
		states = append(states, s)
	}
	sort.Ints(states)
	return states
}

func (f *WFST) IsFinal(state int) bool {
	arcs, exists := f.states[state]
	if !exists {
		return false
	}
	return len(arcs) == 0
}

func (f *WFST) String() string {
	var builder strings.Builder
	states := f.States()
	for _, state := range states {
		builder.WriteString(fmt.Sprintf("State %d", state))
		if f.IsFinal(state) {
			builder.WriteString(" (final)")
		}
		builder.WriteString(":\n")
		for _, arc := range f.GetArcs(state) {
			inLabel := "ε"
			if arc.Input != Epsilon {
				inLabel = arc.Input
			}
			outLabel := "ε"
			if arc.Output != Epsilon {
				outLabel = arc.Output
			}
			builder.WriteString(fmt.Sprintf("  %s -> %s / %.4f -> %d\n",
				inLabel, outLabel, arc.Weight, arc.Next))
		}
	}
	return builder.String()
}

func (f *WFST) DeepCopy() *WFST {
	newFST := New()
	newFST.nextID = f.nextID
	for state, arcs := range f.states {
		newArcs := make([]Arc, len(arcs))
		copy(newArcs, arcs)
		newFST.states[state] = newArcs
	}
	return newFST
}

func InverseWeight(weight float64) float64 {
	if weight == 0 {
		return math.Inf(1)
	}
	return -math.Log(weight)
}

func AddWeights(a, b float64) float64 {
	if a == math.Inf(1) || b == math.Inf(1) {
		return math.Inf(1)
	}
	return a + b
}

func MinWeight(a, b float64) float64 {
	if a == math.Inf(1) {
		return b
	}
	if b == math.Inf(1) {
		return a
	}
	if a < b {
		return a
	}
	return b
}

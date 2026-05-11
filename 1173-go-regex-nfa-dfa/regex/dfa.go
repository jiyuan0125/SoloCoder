package regex

import (
	"fmt"
	"sort"
)

type DFAState struct {
	ID       int
	IsFinal  bool
	NFAState map[int]*NFAState
	Trans    map[rune]*DFAState
}

func NewDFAState(id int) *DFAState {
	return &DFAState{
		ID:       id,
		IsFinal:  false,
		NFAState: make(map[int]*NFAState),
		Trans:    make(map[rune]*DFAState),
	}
}

type DFABuilder struct {
	nextID int
}

func NewDFABuilder() *DFABuilder {
	return &DFABuilder{nextID: 0}
}

func (b *DFABuilder) newState() *DFAState {
	s := NewDFAState(b.nextID)
	b.nextID++
	return s
}

func stateSetKey(states map[int]*NFAState) string {
	ids := make([]int, 0, len(states))
	for id := range states {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return fmt.Sprintf("%v", ids)
}

func hasFinalState(states map[int]*NFAState) bool {
	for _, s := range states {
		if s.IsFinal {
			return true
		}
	}
	return false
}

func collectAllSymbols(nfaStart *NFAState) []rune {
	states := CollectAllStates(nfaStart)
	symbols := make(map[rune]bool)

	for _, s := range states {
		for sym := range s.Trans {
			if sym != Epsilon {
				symbols[sym] = true
			}
		}
	}

	result := make([]rune, 0, len(symbols))
	for sym := range symbols {
		result = append(result, sym)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func (b *DFABuilder) Build(nfaStart *NFAState) (*DFAState, int) {
	startClosure := EpsilonClosure(map[int]*NFAState{nfaStart.ID: nfaStart})

	startState := b.newState()
	startState.NFAState = startClosure
	startState.IsFinal = hasFinalState(startClosure)

	stateMap := make(map[string]*DFAState)
	stateMap[stateSetKey(startClosure)] = startState

	queue := []*DFAState{startState}
	symbols := collectAllSymbols(nfaStart)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, sym := range symbols {
			moveResult := Move(current.NFAState, sym)
			if len(moveResult) == 0 {
				continue
			}

			closure := EpsilonClosure(moveResult)
			key := stateSetKey(closure)

			target, exists := stateMap[key]
			if !exists {
				target = b.newState()
				target.NFAState = closure
				target.IsFinal = hasFinalState(closure)
				stateMap[key] = target
				queue = append(queue, target)
			}

			current.Trans[sym] = target
		}
	}

	return startState, b.nextID
}

type DFATransition struct {
	From   int
	Symbol string
	To     int
}

type DFARepresentation struct {
	StartState  int
	FinalStates []int
	States      int
	Transitions []DFATransition
}

func CollectAllDFAStates(start *DFAState) []*DFAState {
	visited := make(map[int]bool)
	states := []*DFAState{}
	queue := []*DFAState{start}

	for len(queue) > 0 {
		s := queue[0]
		queue = queue[1:]

		if visited[s.ID] {
			continue
		}
		visited[s.ID] = true
		states = append(states, s)

		for _, t := range s.Trans {
			if !visited[t.ID] {
				queue = append(queue, t)
			}
		}
	}

	return states
}

func BuildDFARepresentation(start *DFAState) *DFARepresentation {
	states := CollectAllDFAStates(start)
	repr := &DFARepresentation{
		StartState:  start.ID,
		FinalStates: []int{},
		States:      len(states),
		Transitions: []DFATransition{},
	}

	for _, s := range states {
		if s.IsFinal {
			repr.FinalStates = append(repr.FinalStates, s.ID)
		}

		for sym, t := range s.Trans {
			repr.Transitions = append(repr.Transitions, DFATransition{
				From:   s.ID,
				Symbol: string(sym),
				To:     t.ID,
			})
		}
	}

	return repr
}

func DFAIsEquivalent(a, b *DFAState, partition map[int]int) bool {
	if a.IsFinal != b.IsFinal {
		return false
	}

	for sym, tA := range a.Trans {
		tB, ok := b.Trans[sym]
		if !ok {
			return false
		}
		if partition[tA.ID] != partition[tB.ID] {
			return false
		}
	}

	for sym, tB := range b.Trans {
		_, ok := a.Trans[sym]
		if !ok {
			return false
		}
		_ = tB
	}

	return true
}

package regex

import (
	"sort"
)

type MinimizedDFAState struct {
	ID      int
	IsFinal bool
	Group   int
	Trans   map[rune]*MinimizedDFAState
}

func NewMinimizedDFAState(id int, group int, isFinal bool) *MinimizedDFAState {
	return &MinimizedDFAState{
		ID:      id,
		IsFinal: isFinal,
		Group:   group,
		Trans:   make(map[rune]*MinimizedDFAState),
	}
}

func HopcroftMinimize(start *DFAState) (*MinimizedDFAState, int) {
	states := CollectAllDFAStates(start)
	if len(states) == 0 {
		return nil, 0
	}

	symbols := collectDFASymbols(states)

	partition := make(map[int]int)
	for _, s := range states {
		if s.IsFinal {
			partition[s.ID] = 1
		} else {
			partition[s.ID] = 0
		}
	}

	for {
		changed := false
		newPartition := make(map[int]int)
		groupMap := make(map[string]int)
		nextGroup := 0

		for _, s := range states {
			sig := signature(s, partition, symbols)
			if g, ok := groupMap[sig]; ok {
				newPartition[s.ID] = g
			} else {
				groupMap[sig] = nextGroup
				newPartition[s.ID] = nextGroup
				nextGroup++
			}
		}

		for id, g := range partition {
			if newPartition[id] != g {
				changed = true
				break
			}
		}

		if !changed {
			break
		}
		partition = newPartition
	}

	groupToState := make(map[int]*MinimizedDFAState)
	nextID := 0

	for _, s := range states {
		g := partition[s.ID]
		if _, exists := groupToState[g]; !exists {
			groupToState[g] = NewMinimizedDFAState(nextID, g, s.IsFinal)
			nextID++
		}
	}

	for _, s := range states {
		g := partition[s.ID]
		minState := groupToState[g]

		for sym, target := range s.Trans {
			targetG := partition[target.ID]
			if minTarget, exists := groupToState[targetG]; exists {
				if _, exists := minState.Trans[sym]; !exists {
					minState.Trans[sym] = minTarget
				}
			}
		}
	}

	startG := partition[start.ID]
	return groupToState[startG], nextID
}

func signature(s *DFAState, partition map[int]int, symbols []rune) string {
	parts := []string{}

	if s.IsFinal {
		parts = append(parts, "F")
	} else {
		parts = append(parts, "N")
	}

	for _, sym := range symbols {
		if target, ok := s.Trans[sym]; ok {
			parts = append(parts, string(sym), fmtInt(partition[target.ID]))
		} else {
			parts = append(parts, string(sym), "-")
		}
	}

	result := ""
	for _, p := range parts {
		result += p + "|"
	}
	return result
}

func fmtInt(n int) string {
	if n == 0 {
		return "0"
	}
	isNeg := n < 0
	if isNeg {
		n = -n
	}
	digits := []rune{}
	for n > 0 {
		digits = append([]rune{rune('0' + n%10)}, digits...)
		n /= 10
	}
	if isNeg {
		digits = append([]rune{'-'}, digits...)
	}
	return string(digits)
}

func collectDFASymbols(states []*DFAState) []rune {
	symbols := make(map[rune]bool)
	for _, s := range states {
		for sym := range s.Trans {
			symbols[sym] = true
		}
	}

	result := make([]rune, 0, len(symbols))
	for sym := range symbols {
		result = append(result, sym)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

type MinimizedDFATransition struct {
	From   int
	Symbol string
	To     int
}

type MinimizedDFARepresentation struct {
	StartState  int
	FinalStates []int
	States      int
	Transitions []MinimizedDFATransition
}

func CollectAllMinimizedDFAStates(start *MinimizedDFAState) []*MinimizedDFAState {
	visited := make(map[int]bool)
	states := []*MinimizedDFAState{}
	queue := []*MinimizedDFAState{start}

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

func BuildMinimizedDFARepresentation(start *MinimizedDFAState) *MinimizedDFARepresentation {
	states := CollectAllMinimizedDFAStates(start)
	repr := &MinimizedDFARepresentation{
		StartState:  start.ID,
		FinalStates: []int{},
		States:      len(states),
		Transitions: []MinimizedDFATransition{},
	}

	for _, s := range states {
		if s.IsFinal {
			repr.FinalStates = append(repr.FinalStates, s.ID)
		}

		for sym, t := range s.Trans {
			repr.Transitions = append(repr.Transitions, MinimizedDFATransition{
				From:   s.ID,
				Symbol: string(sym),
				To:     t.ID,
			})
		}
	}

	return repr
}

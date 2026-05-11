package regex

import (
	"fmt"
)

const Epsilon rune = 0

type NFAState struct {
	ID      int
	IsFinal bool
	Trans   map[rune][]*NFAState
}

func NewNFAState(id int) *NFAState {
	return &NFAState{
		ID:      id,
		IsFinal: false,
		Trans:   make(map[rune][]*NFAState),
	}
}

func (s *NFAState) AddTransition(sym rune, target *NFAState) {
	s.Trans[sym] = append(s.Trans[sym], target)
}

type NFAFragment struct {
	Start *NFAState
	Out   []*NFAState
}

func NewNFAFragment(start *NFAState, out []*NFAState) *NFAFragment {
	return &NFAFragment{
		Start: start,
		Out:   out,
	}
}

func (f *NFAFragment) Patch(s *NFAState) {
	for _, outState := range f.Out {
		outState.AddTransition(Epsilon, s)
	}
}

type NFABuilder struct {
	nextID int
}

func NewNFABuilder() *NFABuilder {
	return &NFABuilder{nextID: 0}
}

func (b *NFABuilder) newState() *NFAState {
	s := NewNFAState(b.nextID)
	b.nextID++
	return s
}

func (b *NFABuilder) Build(ast ASTNode) (*NFAState, int) {
	if ast == nil {
		start := b.newState()
		start.IsFinal = true
		return start, 1
	}

	frag := b.buildFragment(ast)
	final := b.newState()
	final.IsFinal = true
	frag.Patch(final)
	return frag.Start, b.nextID
}

func (b *NFABuilder) buildFragment(ast ASTNode) *NFAFragment {
	switch node := ast.(type) {
	case *EmptyNode:
		s := b.newState()
		return NewNFAFragment(s, []*NFAState{s})

	case *CharNode:
		s := b.newState()
		out := b.newState()
		s.AddTransition(node.Value, out)
		return NewNFAFragment(s, []*NFAState{out})

	case *ConcatNode:
		left := b.buildFragment(node.Left)
		right := b.buildFragment(node.Right)
		left.Patch(right.Start)
		return NewNFAFragment(left.Start, right.Out)

	case *AltNode:
		s := b.newState()
		left := b.buildFragment(node.Left)
		right := b.buildFragment(node.Right)
		s.AddTransition(Epsilon, left.Start)
		s.AddTransition(Epsilon, right.Start)
		outStates := append([]*NFAState{}, left.Out...)
		outStates = append(outStates, right.Out...)
		return NewNFAFragment(s, outStates)

	case *StarNode:
		s := b.newState()
		child := b.buildFragment(node.Child)
		s.AddTransition(Epsilon, child.Start)
		child.Patch(s)
		return NewNFAFragment(s, []*NFAState{s})

	case *PlusNode:
		child := b.buildFragment(node.Child)
		s := b.newState()
		child.Patch(s)
		s.AddTransition(Epsilon, child.Start)
		return NewNFAFragment(child.Start, []*NFAState{s})

	case *QuestionNode:
		s := b.newState()
		child := b.buildFragment(node.Child)
		s.AddTransition(Epsilon, child.Start)
		outStates := append([]*NFAState{}, child.Out...)
		outStates = append(outStates, s)
		return NewNFAFragment(s, outStates)

	default:
		panic(fmt.Sprintf("unexpected AST node type: %T", ast))
	}
}

func EpsilonClosure(states map[int]*NFAState) map[int]*NFAState {
	closure := make(map[int]*NFAState)
	stack := make([]*NFAState, 0, len(states))

	for _, s := range states {
		closure[s.ID] = s
		stack = append(stack, s)
	}

	for len(stack) > 0 {
		s := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if targets, ok := s.Trans[Epsilon]; ok {
			for _, t := range targets {
				if _, exists := closure[t.ID]; !exists {
					closure[t.ID] = t
					stack = append(stack, t)
				}
			}
		}
	}

	return closure
}

func Move(states map[int]*NFAState, sym rune) map[int]*NFAState {
	result := make(map[int]*NFAState)

	for _, s := range states {
		if targets, ok := s.Trans[sym]; ok {
			for _, t := range targets {
				result[t.ID] = t
			}
		}
	}

	return result
}

func CollectAllStates(start *NFAState) []*NFAState {
	visited := make(map[int]bool)
	states := []*NFAState{}
	queue := []*NFAState{start}

	for len(queue) > 0 {
		s := queue[0]
		queue = queue[1:]

		if visited[s.ID] {
			continue
		}
		visited[s.ID] = true
		states = append(states, s)

		for _, targets := range s.Trans {
			for _, t := range targets {
				if !visited[t.ID] {
					queue = append(queue, t)
				}
			}
		}
	}

	return states
}

type NFATransition struct {
	From   int
	Symbol string
	To     []int
}

type NFARepresentation struct {
	StartState  int
	FinalStates []int
	States      int
	Transitions []NFATransition
}

func BuildNFARepresentation(start *NFAState) *NFARepresentation {
	states := CollectAllStates(start)
	repr := &NFARepresentation{
		StartState:  start.ID,
		FinalStates: []int{},
		States:      len(states),
		Transitions: []NFATransition{},
	}

	for _, s := range states {
		if s.IsFinal {
			repr.FinalStates = append(repr.FinalStates, s.ID)
		}

		for sym, targets := range s.Trans {
			trans := NFATransition{
				From:   s.ID,
				Symbol: symToString(sym),
				To:     []int{},
			}
			for _, t := range targets {
				trans.To = append(trans.To, t.ID)
			}
			repr.Transitions = append(repr.Transitions, trans)
		}
	}

	return repr
}

func symToString(sym rune) string {
	if sym == Epsilon {
		return "ε"
	}
	return string(sym)
}

package fst

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"
)

type KeyValue struct {
	Key   string
	Value uint64
}

type Transition struct {
	Char  rune
	Next  *State
	Value uint64
}

type State struct {
	ID          int
	IsFinal     bool
	FinalValue  uint64
	Transitions []Transition
}

func (s *State) addTransition(char rune, next *State, value uint64) {
	s.Transitions = append(s.Transitions, Transition{
		Char:  char,
		Next:  next,
		Value: value,
	})
}

func (s *State) getTransition(char rune) (*Transition, bool) {
	for i := range s.Transitions {
		if s.Transitions[i].Char == char {
			return &s.Transitions[i], true
		}
	}
	return nil, false
}

func (s *State) sortTransitions() {
	sort.Slice(s.Transitions, func(i, j int) bool {
		return s.Transitions[i].Char < s.Transitions[j].Char
	})
}

func (s *State) signature() string {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(s.IsFinal)
	enc.Encode(s.FinalValue)
	for _, t := range s.Transitions {
		enc.Encode(t.Char)
		enc.Encode(t.Next.ID)
		enc.Encode(t.Value)
	}
	return buf.String()
}

type FST struct {
	Root    *State
	States  []*State
	Count   int
}

type SearchResult struct {
	Key   string
	Value uint64
	Found bool
}

func NewFST() *FST {
	return &FST{
		States: make([]*State, 0),
	}
}

func (f *FST) newState(isFinal bool, finalValue uint64) *State {
	state := &State{
		ID:         len(f.States),
		IsFinal:    isFinal,
		FinalValue: finalValue,
	}
	f.States = append(f.States, state)
	return state
}

func (f *FST) Build(items []KeyValue) error {
	if len(items) == 0 {
		f.Root = f.newState(false, 0)
		f.Count = 0
		return nil
	}

	items = append([]KeyValue(nil), items...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].Key < items[j].Key
	})

	for i := 1; i < len(items); i++ {
		if items[i].Key == items[i-1].Key {
			return fmt.Errorf("duplicate key: %s", items[i].Key)
		}
	}

	f.States = f.States[:0]
	f.Root = f.newState(false, 0)
	f.Count = len(items)

	if len(items) == 1 {
		f.buildSingle(items[0])
		return nil
	}

	f.buildFromSorted(items)
	return nil
}

func (f *FST) buildSingle(item KeyValue) {
	current := f.Root
	runes := []rune(item.Key)
	for i, char := range runes {
		var next *State
		if i == len(runes)-1 {
			next = f.newState(true, item.Value)
		} else {
			next = f.newState(false, 0)
		}
		current.addTransition(char, next, 0)
		current = next
	}
}

func (f *FST) buildFromSorted(items []KeyValue) {
	prefixFunc := func(s1, s2 string) int {
		r1, r2 := []rune(s1), []rune(s2)
		minLen := len(r1)
		if len(r2) < minLen {
			minLen = len(r2)
		}
		for i := 0; i < minLen; i++ {
			if r1[i] != r2[i] {
				return i
			}
		}
		return minLen
	}

	states := make(map[string]*State)
	minimized := make(map[int]bool)

	currentKey := ""
	currentNode := f.Root
	nodePath := []*State{f.Root}

	for _, item := range items {
		common := prefixFunc(currentKey, item.Key)
		for len(nodePath) > common+1 {
			idx := len(nodePath) - 1
			node := nodePath[idx]
			node.sortTransitions()
			sig := node.signature()
			if existing, ok := states[sig]; ok {
				parent := nodePath[idx-1]
				for i := range parent.Transitions {
					if parent.Transitions[i].Next == node {
						parent.Transitions[i].Next = existing
						break
					}
				}
				minimized[idx] = true
			} else {
				states[sig] = node
			}
			nodePath = nodePath[:idx]
		}

		currentNode = nodePath[len(nodePath)-1]
		suffixRunes := []rune(item.Key)[common:]

		for i, char := range suffixRunes {
			var next *State
			if i == len(suffixRunes)-1 {
				next = f.newState(true, item.Value)
			} else {
				next = f.newState(false, 0)
			}
			currentNode.addTransition(char, next, 0)
			currentNode = next
			nodePath = append(nodePath, currentNode)
		}

		currentKey = item.Key
	}

	for len(nodePath) > 1 {
		idx := len(nodePath) - 1
		node := nodePath[idx]
		node.sortTransitions()
		sig := node.signature()
		if existing, ok := states[sig]; ok {
			parent := nodePath[idx-1]
			for i := range parent.Transitions {
				if parent.Transitions[i].Next == node {
					parent.Transitions[i].Next = existing
					break
				}
			}
			minimized[idx] = true
		} else {
			states[sig] = node
		}
		nodePath = nodePath[:idx]
	}

	f.Root.sortTransitions()
}

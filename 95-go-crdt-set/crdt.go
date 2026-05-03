package crdt

import (
	"encoding/json"
	"sort"
	"sync"
)

type ElementState struct {
	Value      string   `json:"value"`
	AddTags    []uint64 `json:"add_tags"`
	RemoveTags []uint64 `json:"remove_tags"`
}

type ORSet struct {
	mu        sync.RWMutex
	elements  map[string]*ElementState
	nextTag   uint64
}

func NewORSet() *ORSet {
	return &ORSet{
		elements: make(map[string]*ElementState),
		nextTag:  1,
	}
}

func (s *ORSet) Add(element string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, exists := s.elements[element]
	if !exists {
		state = &ElementState{
			Value:      element,
			AddTags:    []uint64{},
			RemoveTags: []uint64{},
		}
		s.elements[element] = state
	}

	state.AddTags = append(state.AddTags, s.nextTag)
	s.nextTag++
}

func (s *ORSet) Remove(element string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, exists := s.elements[element]
	if !exists {
		return
	}

	for _, tag := range state.AddTags {
		if !containsUint64(state.RemoveTags, tag) {
			state.RemoveTags = append(state.RemoveTags, tag)
		}
	}
}

func (s *ORSet) Contains(element string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, exists := s.elements[element]
	if !exists {
		return false
	}

	for _, addTag := range state.AddTags {
		if !containsUint64(state.RemoveTags, addTag) {
			return true
		}
	}
	return false
}

func (s *ORSet) Elements() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []string
	for value, state := range s.elements {
		for _, addTag := range state.AddTags {
			if !containsUint64(state.RemoveTags, addTag) {
				result = append(result, value)
				break
			}
		}
	}

	sort.Strings(result)
	return result
}

func (s *ORSet) Merge(other *ORSet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	for value, otherState := range other.elements {
		currentState, exists := s.elements[value]
		if !exists {
			currentState = &ElementState{
				Value:      value,
				AddTags:    []uint64{},
				RemoveTags: []uint64{},
			}
			s.elements[value] = currentState
		}

		for _, tag := range otherState.AddTags {
			if !containsUint64(currentState.AddTags, tag) {
				currentState.AddTags = append(currentState.AddTags, tag)
			}
		}

		for _, tag := range otherState.RemoveTags {
			if !containsUint64(currentState.RemoveTags, tag) {
				currentState.RemoveTags = append(currentState.RemoveTags, tag)
			}
		}
	}

	if other.nextTag > s.nextTag {
		s.nextTag = other.nextTag
	}
}

func (s *ORSet) Clone() *ORSet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clone := NewORSet()
	clone.nextTag = s.nextTag

	for value, state := range s.elements {
		cloneState := &ElementState{
			Value:      value,
			AddTags:    make([]uint64, len(state.AddTags)),
			RemoveTags: make([]uint64, len(state.RemoveTags)),
		}
		copy(cloneState.AddTags, state.AddTags)
		copy(cloneState.RemoveTags, state.RemoveTags)
		clone.elements[value] = cloneState
	}

	return clone
}

func (s *ORSet) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var elements []ElementState
	for _, state := range s.elements {
		elements = append(elements, *state)
	}

	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Value < elements[j].Value
	})

	return json.Marshal(map[string][]ElementState{
		"elements": elements,
	})
}

func (s *ORSet) UnmarshalJSON(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var wrapper struct {
		Elements []ElementState `json:"elements"`
	}

	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}

	s.elements = make(map[string]*ElementState)
	s.nextTag = 1

	for _, elem := range wrapper.Elements {
		state := &ElementState{
			Value:      elem.Value,
			AddTags:    make([]uint64, len(elem.AddTags)),
			RemoveTags: make([]uint64, len(elem.RemoveTags)),
		}
		copy(state.AddTags, elem.AddTags)
		copy(state.RemoveTags, elem.RemoveTags)
		s.elements[elem.Value] = state

		for _, tag := range elem.AddTags {
			if tag >= s.nextTag {
				s.nextTag = tag + 1
			}
		}
	}

	return nil
}

func containsUint64(slice []uint64, val uint64) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

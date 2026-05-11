package wfst

import (
	"container/list"
	"math"
)

type statePair struct {
	a int
	b int
}

func newStatePair(a, b int) statePair {
	return statePair{a, b}
}

func pairKey(p statePair) int64 {
	return (int64(p.a) << 32) | (int64(p.b) & 0xFFFFFFFF)
}

func ComputeEpsilonClosure(fst *WFST, startState int) map[int]float64 {
	closure := make(map[int]float64)
	queue := list.New()
	
	closure[startState] = 0.0
	queue.PushBack(startState)
	
	for queue.Len() > 0 {
		state := queue.Remove(queue.Front()).(int)
		currentWeight := closure[state]
		
		for _, arc := range fst.GetArcs(state) {
			if arc.Input == Epsilon && arc.Output == Epsilon {
				continue
			}
			if arc.Output == Epsilon {
				newWeight := AddWeights(currentWeight, arc.Weight)
				if existing, exists := closure[arc.Next]; !exists || newWeight < existing {
					closure[arc.Next] = newWeight
					queue.PushBack(arc.Next)
				}
			}
		}
	}
	
	return closure
}

func ComputeEpsilonClosureB(fst *WFST, startState int) map[int]float64 {
	closure := make(map[int]float64)
	queue := list.New()
	
	closure[startState] = 0.0
	queue.PushBack(startState)
	
	for queue.Len() > 0 {
		state := queue.Remove(queue.Front()).(int)
		currentWeight := closure[state]
		
		for _, arc := range fst.GetArcs(state) {
			if arc.Input == Epsilon && arc.Output == Epsilon {
				continue
			}
			if arc.Input == Epsilon {
				newWeight := AddWeights(currentWeight, arc.Weight)
				if existing, exists := closure[arc.Next]; !exists || newWeight < existing {
					closure[arc.Next] = newWeight
					queue.PushBack(arc.Next)
				}
			}
		}
	}
	
	return closure
}

func Compose(a, b *WFST) *WFST {
	result := New()
	
	stateMap := make(map[int64]int)
	queue := list.New()
	
	startPair := newStatePair(0, 0)
	startKey := pairKey(startPair)
	stateMap[startKey] = 0
	queue.PushBack(startPair)
	
	for queue.Len() > 0 {
		current := queue.Remove(queue.Front()).(statePair)
		currentID := stateMap[pairKey(current)]
		
		closureA := ComputeEpsilonClosure(a, current.a)
		for stateA, weightA := range closureA {
			for _, arcA := range a.GetArcs(stateA) {
				if arcA.Output != Epsilon {
					continue
				}
				
				nextPair := newStatePair(arcA.Next, current.b)
				nextKey := pairKey(nextPair)
				
				var nextID int
				if id, exists := stateMap[nextKey]; exists {
					nextID = id
				} else {
					nextID = result.AddState()
					stateMap[nextKey] = nextID
					queue.PushBack(nextPair)
				}
				
				totalWeight := AddWeights(weightA, arcA.Weight)
				result.AddArc(currentID, Arc{
					Input:  arcA.Input,
					Output: Epsilon,
					Weight: totalWeight,
					Next:   nextID,
				})
			}
		}
		
		closureB := ComputeEpsilonClosureB(b, current.b)
		for stateB, weightB := range closureB {
			for _, arcB := range b.GetArcs(stateB) {
				if arcB.Input != Epsilon {
					continue
				}
				
				nextPair := newStatePair(current.a, arcB.Next)
				nextKey := pairKey(nextPair)
				
				var nextID int
				if id, exists := stateMap[nextKey]; exists {
					nextID = id
				} else {
					nextID = result.AddState()
					stateMap[nextKey] = nextID
					queue.PushBack(nextPair)
				}
				
				totalWeight := AddWeights(weightB, arcB.Weight)
				result.AddArc(currentID, Arc{
					Input:  Epsilon,
					Output: arcB.Output,
					Weight: totalWeight,
					Next:   nextID,
				})
			}
		}
		
		for _, arcA := range a.GetArcs(current.a) {
			for _, arcB := range b.GetArcs(current.b) {
				if arcA.Output != arcB.Input {
					continue
				}
				if arcA.Output == Epsilon && arcB.Input == Epsilon {
					continue
				}
				
				nextPair := newStatePair(arcA.Next, arcB.Next)
				nextKey := pairKey(nextPair)
				
				var nextID int
				if id, exists := stateMap[nextKey]; exists {
					nextID = id
				} else {
					nextID = result.AddState()
					stateMap[nextKey] = nextID
					queue.PushBack(nextPair)
				}
				
				totalWeight := AddWeights(arcA.Weight, arcB.Weight)
				result.AddArc(currentID, Arc{
					Input:  arcA.Input,
					Output: arcB.Output,
					Weight: totalWeight,
					Next:   nextID,
				})
			}
		}
	}
	
	return result
}

func RemoveDeadStates(fst *WFST) *WFST {
	result := New()
	result.states = make(map[int][]Arc)
	
	reachable := make(map[int]bool)
	queue := list.New()
	
	reachable[0] = true
	queue.PushBack(0)
	
	for queue.Len() > 0 {
		state := queue.Remove(queue.Front()).(int)
		for _, arc := range fst.GetArcs(state) {
			if !reachable[arc.Next] {
				reachable[arc.Next] = true
				queue.PushBack(arc.Next)
			}
		}
	}
	
	for state := range reachable {
		var arcs []Arc
		for _, arc := range fst.GetArcs(state) {
			if reachable[arc.Next] {
				arcs = append(arcs, arc)
			}
		}
		result.states[state] = arcs
		if state >= result.nextID {
			result.nextID = state + 1
		}
	}
	
	if len(result.states) == 0 {
		result.AddState()
	}
	
	return result
}

func Prune(fst *WFST, beam float64) *WFST {
	if beam <= 0 || beam == math.Inf(1) {
		return fst.DeepCopy()
	}
	
	result := New()
	result.states = make(map[int][]Arc)
	
	bestWeight := make(map[int]float64)
	reachable := make(map[int]bool)
	queue := list.New()
	
	for i := range fst.states {
		bestWeight[i] = math.Inf(1)
	}
	bestWeight[0] = 0.0
	reachable[0] = true
	queue.PushBack(0)
	
	for queue.Len() > 0 {
		state := queue.Remove(queue.Front()).(int)
		currentWeight := bestWeight[state]
		
		for _, arc := range fst.GetArcs(state) {
			newWeight := AddWeights(currentWeight, arc.Weight)
			if newWeight-bestWeight[0] > beam {
				continue
			}
			
			if newWeight < bestWeight[arc.Next] {
				bestWeight[arc.Next] = newWeight
				if !reachable[arc.Next] {
					reachable[arc.Next] = true
					queue.PushBack(arc.Next)
				}
			}
		}
	}
	
	for state := range reachable {
		var arcs []Arc
		for _, arc := range fst.GetArcs(state) {
			if reachable[arc.Next] {
				arcWeight := AddWeights(bestWeight[state], arc.Weight)
				if arcWeight-bestWeight[0] <= beam {
					arcs = append(arcs, arc)
				}
			}
		}
		result.states[state] = arcs
		if state >= result.nextID {
			result.nextID = state + 1
		}
	}
	
	if len(result.states) == 0 {
		result.AddState()
	}
	
	return result
}

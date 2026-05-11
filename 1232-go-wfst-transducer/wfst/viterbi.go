package wfst

import (
	"container/list"
	"math"
)

type SearchResult struct {
	Path        []int
	InputSeq    []string
	OutputSeq   []string
	Weights     []float64
	TotalWeight float64
}

func Viterbi(fst *WFST, input []string) *SearchResult {
	if len(input) == 0 {
		return nil
	}
	
	n := len(input)
	
	type viterbiState struct {
		state  int
		weight float64
		prev   *viterbiState
		step   int
		input  string
		output string
	}
	
	tables := make([]map[int]*viterbiState, n+1)
	for i := range tables {
		tables[i] = make(map[int]*viterbiState)
	}
	
	tables[0][0] = &viterbiState{
		state:  0,
		weight: 0.0,
		step:   0,
	}
	
	for i := 0; i < n; i++ {
		epsilonQueue := list.New()
		epsilonVisited := make(map[int]float64)
		
		for state, vs := range tables[i] {
			if existing, exists := epsilonVisited[state]; !exists || vs.weight < existing {
				epsilonVisited[state] = vs.weight
				epsilonQueue.PushBack(vs)
			}
		}
		
		for epsilonQueue.Len() > 0 {
			current := epsilonQueue.Remove(epsilonQueue.Front()).(*viterbiState)
			currentWeight := current.weight
			
			if existing, exists := epsilonVisited[current.state]; exists && currentWeight > existing {
				continue
			}
			
			for _, arc := range fst.GetArcs(current.state) {
				if arc.Input != Epsilon {
					continue
				}
				
				newWeight := AddWeights(currentWeight, arc.Weight)
				
				if existing, exists := tables[i][arc.Next]; !exists || newWeight < existing.weight {
					newVs := &viterbiState{
						state:  arc.Next,
						weight: newWeight,
						prev:   current,
						step:   i,
						input:  Epsilon,
						output: arc.Output,
					}
					tables[i][arc.Next] = newVs
					
					if existing, exists := epsilonVisited[arc.Next]; !exists || newWeight < existing {
						epsilonVisited[arc.Next] = newWeight
						epsilonQueue.PushBack(newVs)
					}
				}
			}
		}
		
		currentLabel := input[i]
		
		for state, vs := range tables[i] {
			currentWeight := vs.weight
			
			labelFound := false
			
			for _, arc := range fst.GetArcs(state) {
				if arc.Input != currentLabel {
					continue
				}
				labelFound = true
				
				newWeight := AddWeights(currentWeight, arc.Weight)
				
				if existing, exists := tables[i+1][arc.Next]; !exists || newWeight < existing.weight {
					tables[i+1][arc.Next] = &viterbiState{
						state:  arc.Next,
						weight: newWeight,
						prev:   vs,
						step:   i + 1,
						input:  currentLabel,
						output: arc.Output,
					}
				}
			}
			
			if !labelFound {
				if existing, exists := tables[i+1][state]; !exists || currentWeight < existing.weight {
					tables[i+1][state] = &viterbiState{
						state:  state,
						weight: currentWeight,
						prev:   vs,
						step:   i + 1,
						input:  currentLabel,
						output: "",
					}
				}
			}
		}
	}
	
	finalTable := tables[n]
	if len(finalTable) == 0 {
		return nil
	}
	
	epsilonQueue := list.New()
	epsilonVisited := make(map[int]float64)
	
	for state, vs := range finalTable {
		if existing, exists := epsilonVisited[state]; !exists || vs.weight < existing {
			epsilonVisited[state] = vs.weight
			epsilonQueue.PushBack(vs)
		}
	}
	
	for epsilonQueue.Len() > 0 {
		current := epsilonQueue.Remove(epsilonQueue.Front()).(*viterbiState)
		currentWeight := current.weight
		
		if existing, exists := epsilonVisited[current.state]; exists && currentWeight > existing {
			continue
		}
		
		for _, arc := range fst.GetArcs(current.state) {
			if arc.Input != Epsilon {
				continue
			}
			
			newWeight := AddWeights(currentWeight, arc.Weight)
			
			if existing, exists := finalTable[arc.Next]; !exists || newWeight < existing.weight {
				newVs := &viterbiState{
					state:  arc.Next,
					weight: newWeight,
					prev:   current,
					step:   n,
					input:  Epsilon,
					output: arc.Output,
				}
				finalTable[arc.Next] = newVs
				
				if existing, exists := epsilonVisited[arc.Next]; !exists || newWeight < existing {
					epsilonVisited[arc.Next] = newWeight
					epsilonQueue.PushBack(newVs)
				}
			}
		}
	}
	
	bestFinal := math.Inf(1)
	var bestVs *viterbiState
	
	for state, vs := range finalTable {
		if vs.weight < bestFinal && fst.IsFinal(state) {
			bestFinal = vs.weight
			bestVs = vs
		}
	}
	
	if bestVs == nil {
		return nil
	}
	
	result := &SearchResult{
		TotalWeight: bestFinal,
	}
	
	current := bestVs
	for current != nil {
		result.Path = append([]int{current.state}, result.Path...)
		if current.step > 0 {
			if current.input != "" {
				result.InputSeq = append([]string{current.input}, result.InputSeq...)
			}
			if current.output != "" {
				result.OutputSeq = append([]string{current.output}, result.OutputSeq...)
			}
			result.Weights = append([]float64{current.weight}, result.Weights...)
		}
		current = current.prev
	}
	
	return result
}

func FindBestPath(fst *WFST) *SearchResult {
	bestWeight := make(map[int]float64)
	prevState := make(map[int]int)
	prevArc := make(map[int]Arc)
	
	queue := list.New()
	inQueue := make(map[int]bool)
	
	for state := range fst.states {
		bestWeight[state] = math.Inf(1)
	}
	bestWeight[0] = 0.0
	queue.PushBack(0)
	inQueue[0] = true
	
	for queue.Len() > 0 {
		state := queue.Remove(queue.Front()).(int)
		inQueue[state] = false
		
		for _, arc := range fst.GetArcs(state) {
			newWeight := AddWeights(bestWeight[state], arc.Weight)
			
			if newWeight < bestWeight[arc.Next] {
				bestWeight[arc.Next] = newWeight
				prevState[arc.Next] = state
				prevArc[arc.Next] = arc
				
				if !inQueue[arc.Next] {
					queue.PushBack(arc.Next)
					inQueue[arc.Next] = true
				}
			}
		}
	}
	
	bestFinal := math.Inf(1)
	bestFinalState := -1
	
	for state := range fst.states {
		if fst.IsFinal(state) && bestWeight[state] < bestFinal {
			bestFinal = bestWeight[state]
			bestFinalState = state
		}
	}
	
	if bestFinalState == -1 {
		return nil
	}
	
	result := &SearchResult{
		TotalWeight: bestFinal,
	}
	
	path := []int{bestFinalState}
	var inputSeq []string
	var outputSeq []string
	var weights []float64
	
	current := bestFinalState
	for current != 0 {
		prev := prevState[current]
		arc := prevArc[current]
		
		path = append([]int{prev}, path...)
		if arc.Input != "" {
			inputSeq = append([]string{arc.Input}, inputSeq...)
		}
		if arc.Output != "" {
			outputSeq = append([]string{arc.Output}, outputSeq...)
		}
		weights = append([]float64{bestWeight[current]}, weights...)
		
		current = prev
	}
	
	result.Path = path
	result.InputSeq = inputSeq
	result.OutputSeq = outputSeq
	result.Weights = weights
	
	return result
}

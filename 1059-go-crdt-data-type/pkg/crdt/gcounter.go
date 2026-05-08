package crdt

import (
	"encoding/json"
	"fmt"
)

type GCounter struct {
	Counters map[string]int `json:"counters"`
}

func NewGCounter() *GCounter {
	return &GCounter{
		Counters: make(map[string]int),
	}
}

func (gc *GCounter) Increment(nodeID string) error {
	if nodeID == "" {
		return fmt.Errorf("nodeID cannot be empty")
	}
	gc.Counters[nodeID]++
	return nil
}

func (gc *GCounter) IncrementBy(nodeID string, delta int) error {
	if nodeID == "" {
		return fmt.Errorf("nodeID cannot be empty")
	}
	if delta <= 0 {
		return fmt.Errorf("delta must be positive")
	}
	gc.Counters[nodeID] += delta
	return nil
}

func (gc *GCounter) Value() int {
	total := 0
	for _, v := range gc.Counters {
		total += v
	}
	return total
}

func (gc *GCounter) Merge(other *GCounter) *GCounter {
	result := NewGCounter()

	for nodeID, v := range gc.Counters {
		result.Counters[nodeID] = v
	}

	for nodeID, v := range other.Counters {
		if current, exists := result.Counters[nodeID]; exists {
			if v > current {
				result.Counters[nodeID] = v
			}
		} else {
			result.Counters[nodeID] = v
		}
	}

	return result
}

func (gc *GCounter) Equals(other *GCounter) bool {
	if len(gc.Counters) != len(other.Counters) {
		return false
	}

	for nodeID, v := range gc.Counters {
		if other.Counters[nodeID] != v {
			return false
		}
	}

	return true
}

func (gc *GCounter) Marshal() ([]byte, error) {
	return json.Marshal(gc)
}

func UnmarshalGCounter(data []byte) (*GCounter, error) {
	var gc GCounter
	if err := json.Unmarshal(data, &gc); err != nil {
		return nil, err
	}
	if gc.Counters == nil {
		gc.Counters = make(map[string]int)
	}
	return &gc, nil
}

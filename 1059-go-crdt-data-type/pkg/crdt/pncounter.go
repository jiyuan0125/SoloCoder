package crdt

import "encoding/json"

type PNCounter struct {
	Positive *GCounter `json:"positive"`
	Negative *GCounter `json:"negative"`
}

func NewPNCounter() *PNCounter {
	return &PNCounter{
		Positive: NewGCounter(),
		Negative: NewGCounter(),
	}
}

func (pn *PNCounter) Increment(nodeID string) error {
	return pn.Positive.Increment(nodeID)
}

func (pn *PNCounter) IncrementBy(nodeID string, delta int) error {
	return pn.Positive.IncrementBy(nodeID, delta)
}

func (pn *PNCounter) Decrement(nodeID string) error {
	return pn.Negative.Increment(nodeID)
}

func (pn *PNCounter) DecrementBy(nodeID string, delta int) error {
	return pn.Negative.IncrementBy(nodeID, delta)
}

func (pn *PNCounter) Value() int {
	return pn.Positive.Value() - pn.Negative.Value()
}

func (pn *PNCounter) Merge(other *PNCounter) *PNCounter {
	return &PNCounter{
		Positive: pn.Positive.Merge(other.Positive),
		Negative: pn.Negative.Merge(other.Negative),
	}
}

func (pn *PNCounter) Equals(other *PNCounter) bool {
	return pn.Positive.Equals(other.Positive) && pn.Negative.Equals(other.Negative)
}

func (pn *PNCounter) Marshal() ([]byte, error) {
	return json.Marshal(pn)
}

func UnmarshalPNCounter(data []byte) (*PNCounter, error) {
	var pn PNCounter
	if err := json.Unmarshal(data, &pn); err != nil {
		return nil, err
	}
	if pn.Positive == nil {
		pn.Positive = NewGCounter()
	}
	if pn.Negative == nil {
		pn.Negative = NewGCounter()
	}
	if pn.Positive.Counters == nil {
		pn.Positive.Counters = make(map[string]int)
	}
	if pn.Negative.Counters == nil {
		pn.Negative.Counters = make(map[string]int)
	}
	return &pn, nil
}

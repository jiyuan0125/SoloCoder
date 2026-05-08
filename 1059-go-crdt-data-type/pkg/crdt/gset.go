package crdt

import "encoding/json"

type GSet struct {
	Elements map[string]struct{} `json:"elements"`
}

func NewGSet() *GSet {
	return &GSet{
		Elements: make(map[string]struct{}),
	}
}

func (s *GSet) Add(element string) {
	s.Elements[element] = struct{}{}
}

func (s *GSet) Contains(element string) bool {
	_, exists := s.Elements[element]
	return exists
}

func (s *GSet) Size() int {
	return len(s.Elements)
}

func (s *GSet) ElementsList() []string {
	result := make([]string, 0, len(s.Elements))
	for e := range s.Elements {
		result = append(result, e)
	}
	return result
}

func (s *GSet) Merge(other *GSet) *GSet {
	result := NewGSet()

	for e := range s.Elements {
		result.Elements[e] = struct{}{}
	}

	for e := range other.Elements {
		result.Elements[e] = struct{}{}
	}

	return result
}

func (s *GSet) Equals(other *GSet) bool {
	if len(s.Elements) != len(other.Elements) {
		return false
	}

	for e := range s.Elements {
		if !other.Contains(e) {
			return false
		}
	}

	return true
}

func (s *GSet) Marshal() ([]byte, error) {
	return json.Marshal(s)
}

func UnmarshalGSet(data []byte) (*GSet, error) {
	var gset GSet
	if err := json.Unmarshal(data, &gset); err != nil {
		return nil, err
	}
	if gset.Elements == nil {
		gset.Elements = make(map[string]struct{})
	}
	return &gset, nil
}

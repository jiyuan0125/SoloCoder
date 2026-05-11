package phonetic

import (
	"sort"
	"sync"
)

type InMemoryStore struct {
	mu            sync.RWMutex
	names         map[string]bool
	soundexIndex  map[string][]string
	metaphoneIndex map[string][]string
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		names:         make(map[string]bool),
		soundexIndex:  make(map[string][]string),
		metaphoneIndex: make(map[string][]string),
	}
}

func (s *InMemoryStore) Add(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanName := filterLetters(name)
	if cleanName == "" {
		return false
	}

	if s.names[cleanName] {
		return false
	}

	s.names[cleanName] = true
	code := Encode(cleanName)

	if code.Soundex != "" {
		s.soundexIndex[code.Soundex] = append(s.soundexIndex[code.Soundex], cleanName)
	}
	if code.Metaphone != "" {
		s.metaphoneIndex[code.Metaphone] = append(s.metaphoneIndex[code.Metaphone], cleanName)
	}

	return true
}

func (s *InMemoryStore) Remove(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanName := filterLetters(name)
	if cleanName == "" || !s.names[cleanName] {
		return false
	}

	delete(s.names, cleanName)
	code := Encode(cleanName)

	if code.Soundex != "" {
		if names, ok := s.soundexIndex[code.Soundex]; ok {
			newNames := make([]string, 0, len(names)-1)
			for _, n := range names {
				if n != cleanName {
					newNames = append(newNames, n)
				}
			}
			if len(newNames) == 0 {
				delete(s.soundexIndex, code.Soundex)
			} else {
				s.soundexIndex[code.Soundex] = newNames
			}
		}
	}

	if code.Metaphone != "" {
		if names, ok := s.metaphoneIndex[code.Metaphone]; ok {
			newNames := make([]string, 0, len(names)-1)
			for _, n := range names {
				if n != cleanName {
					newNames = append(newNames, n)
				}
			}
			if len(newNames) == 0 {
				delete(s.metaphoneIndex, code.Metaphone)
			} else {
				s.metaphoneIndex[code.Metaphone] = newNames
			}
		}
	}

	return true
}

func (s *InMemoryStore) SearchBySoundex(code string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0)
	if found, ok := s.soundexIndex[code]; ok {
		names = append(names, found...)
	}
	sort.Strings(names)
	return names
}

func (s *InMemoryStore) SearchByMetaphone(code string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0)
	if found, ok := s.metaphoneIndex[code]; ok {
		names = append(names, found...)
	}
	sort.Strings(names)
	return names
}

func (s *InMemoryStore) Total() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.names)
}

func (s *InMemoryStore) UniqueSoundexCodes() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.soundexIndex)
}

func (s *InMemoryStore) UniqueMetaphoneCodes() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.metaphoneIndex)
}

func (s *InMemoryStore) TopSoundexCodes(n int) []CodeCount {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]CodeCount, 0, len(s.soundexIndex))
	for code, names := range s.soundexIndex {
		sortedNames := make([]string, len(names))
		copy(sortedNames, names)
		sort.Strings(sortedNames)
		result = append(result, CodeCount{
			Code:  code,
			Count: len(names),
			Names: sortedNames,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	if len(result) > n {
		result = result[:n]
	}
	return result
}

func (s *InMemoryStore) TopMetaphoneCodes(n int) []CodeCount {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]CodeCount, 0, len(s.metaphoneIndex))
	for code, names := range s.metaphoneIndex {
		sortedNames := make([]string, len(names))
		copy(sortedNames, names)
		sort.Strings(sortedNames)
		result = append(result, CodeCount{
			Code:  code,
			Count: len(names),
			Names: sortedNames,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	if len(result) > n {
		result = result[:n]
	}
	return result
}

func (s *InMemoryStore) Exists(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cleanName := filterLetters(name)
	return s.names[cleanName]
}

package fuzzy

import (
	"sort"
	"sync"
)

type Dictionary struct {
	mu    sync.RWMutex
	words map[string]bool
}

type SearchResult struct {
	Word     string
	Distance int
}

func NewDictionary() *Dictionary {
	return &Dictionary{
		words: make(map[string]bool),
	}
}

func (d *Dictionary) Add(word string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.words[word] {
		return false
	}
	d.words[word] = true
	return true
}

func (d *Dictionary) Remove(word string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.words[word] {
		return false
	}
	delete(d.words, word)
	return true
}

func (d *Dictionary) Contains(word string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.words[word]
}

func (d *Dictionary) Size() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.words)
}

func (d *Dictionary) GetAll() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	words := make([]string, 0, len(d.words))
	for word := range d.words {
		words = append(words, word)
	}
	sort.Strings(words)
	return words
}

func (d *Dictionary) Import(words []string) int {
	count := 0
	for _, word := range words {
		if d.Add(word) {
			count++
		}
	}
	return count
}

func (d *Dictionary) Search(query string, threshold int) []SearchResult {
	results := []SearchResult{}
	
	queryRunes := LenRunes(query)
	
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	for word := range d.words {
		wordRunes := LenRunes(word)
		if abs(queryRunes-wordRunes) > threshold {
			continue
		}
		
		if threshold == 0 {
			if query == word {
				results = append(results, SearchResult{Word: word, Distance: 0})
			}
			continue
		}
		
		distance := LevenshteinDistance(query, word)
		if distance <= threshold {
			results = append(results, SearchResult{Word: word, Distance: distance})
		}
	}
	
	sortResults(results)
	return results
}

func (d *Dictionary) WildcardSearch(pattern string, threshold int) []SearchResult {
	results := []SearchResult{}
	
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	for word := range d.words {
		matched, distance := WildcardMatch(pattern, word, threshold)
		if matched {
			results = append(results, SearchResult{Word: word, Distance: distance})
		}
	}
	
	sortResults(results)
	return results
}

func sortResults(results []SearchResult) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Distance == results[j].Distance {
			return results[i].Word < results[j].Word
		}
		return results[i].Distance < results[j].Distance
	})
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

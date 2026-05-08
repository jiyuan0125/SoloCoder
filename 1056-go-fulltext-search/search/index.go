package search

import (
	"math"
	"sort"
	"sync"
)

type Posting struct {
	DocID    string
	TermFreq int
	Positions []int
}

type InvertedIndex struct {
	mu        sync.RWMutex
	postings  map[string][]*Posting
	docTokens map[string]int
	docs      map[string]string
}

func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		postings:  make(map[string][]*Posting),
		docTokens: make(map[string]int),
		docs:      make(map[string]string),
	}
}

func (idx *InvertedIndex) AddDoc(docID, content string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.DeleteDocLocked(docID)

	tokens := Tokenize(content)
	idx.docTokens[docID] = len(tokens)
	idx.docs[docID] = content

	tokenPositions := make(map[string][]int)
	for i, token := range tokens {
		tokenPositions[token] = append(tokenPositions[token], i)
	}

	for token, positions := range tokenPositions {
		posting := &Posting{
			DocID:    docID,
			TermFreq: len(positions),
			Positions: positions,
		}
		idx.postings[token] = append(idx.postings[token], posting)
	}
}

func (idx *InvertedIndex) DeleteDoc(docID string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.DeleteDocLocked(docID)
}

func (idx *InvertedIndex) DeleteDocLocked(docID string) {
	if _, exists := idx.docs[docID]; !exists {
		return
	}

	delete(idx.docs, docID)
	delete(idx.docTokens, docID)

	for token, postings := range idx.postings {
		newPostings := postings[:0]
		for _, p := range postings {
			if p.DocID != docID {
				newPostings = append(newPostings, p)
			}
		}
		if len(newPostings) == 0 {
			delete(idx.postings, token)
		} else {
			idx.postings[token] = newPostings
		}
	}
}

func (idx *InvertedIndex) GetPostings(token string) []*Posting {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.postings[token]
}

func (idx *InvertedIndex) DocCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.docs)
}

func (idx *InvertedIndex) DocTokenCount(docID string) int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.docTokens[docID]
}

func (idx *InvertedIndex) GetAllDocIDs() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	ids := make([]string, 0, len(idx.docs))
	for id := range idx.docs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (idx *InvertedIndex) ComputeTFIDF(docID, token string) float64 {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	postings := idx.postings[token]
	if postings == nil {
		return 0
	}

	var termFreq int
	for _, p := range postings {
		if p.DocID == docID {
			termFreq = p.TermFreq
			break
		}
	}
	if termFreq == 0 {
		return 0
	}

	docTokenCount := idx.docTokens[docID]
	if docTokenCount == 0 {
		return 0
	}

	tf := float64(termFreq) / float64(docTokenCount)

	docCount := len(idx.docs)
	if docCount == 0 {
		return 0
	}

	df := len(postings)
	idf := math.Log(float64(docCount) / float64(df))

	return tf * idf
}

package textstat

import (
	"math"
	"strings"
)

type TFIDFVector map[string]float64

type Corpus struct {
	docs        []WordFreqMap
	docCount    int
	wordDocFreq map[string]int
}

func NewCorpus() *Corpus {
	return &Corpus{
		docs:        []WordFreqMap{},
		docCount:    0,
		wordDocFreq: make(map[string]int),
	}
}

func (c *Corpus) AddDocument(freqMap WordFreqMap) {
	if len(freqMap) == 0 {
		return
	}
	c.docs = append(c.docs, freqMap)
	c.docCount++
	
	seen := make(map[string]bool)
	for word := range freqMap {
		if !seen[word] {
			seen[word] = true
			c.wordDocFreq[word]++
		}
	}
}

func (c *Corpus) IDF(word string) float64 {
	if c.docCount == 0 {
		return 1
	}
	docFreq := c.wordDocFreq[word]
	if docFreq == 0 {
		docFreq = 1
	}
	return math.Log((float64(c.docCount)+1)/(float64(docFreq)+1)) + 1
}

func (c *Corpus) ComputeTFIDF(freqMap WordFreqMap) TFIDFVector {
	if len(freqMap) == 0 {
		return make(TFIDFVector)
	}
	
	totalWords := freqMap.TotalWords()
	vector := make(TFIDFVector)
	
	for word, tf := range freqMap {
		tfNorm := float64(tf) / float64(totalWords)
		idf := c.IDF(word)
		vector[word] = tfNorm * idf
	}
	
	return vector
}

func ComputeTFIDFFromTwoTexts(text1, text2 string, stopWords *StopWords) (TFIDFVector, TFIDFVector) {
	if text1 == "" && text2 == "" {
		return make(TFIDFVector), make(TFIDFVector)
	}
	
	freqMap1 := CountWordFrequencies(text1, stopWords)
	freqMap2 := CountWordFrequencies(text2, stopWords)
	
	corpus := NewCorpus()
	corpus.AddDocument(freqMap1)
	corpus.AddDocument(freqMap2)
	
	vector1 := corpus.ComputeTFIDF(freqMap1)
	vector2 := corpus.ComputeTFIDF(freqMap2)
	
	return vector1, vector2
}

func CosineSimilarity(v1, v2 TFIDFVector) float64 {
	if len(v1) == 0 || len(v2) == 0 {
		return 0.0
	}
	
	dotProduct := 0.0
	magnitude1 := 0.0
	magnitude2 := 0.0
	
	for word, val1 := range v1 {
		magnitude1 += val1 * val1
		if val2, exists := v2[word]; exists {
			dotProduct += val1 * val2
		}
	}
	
	for _, val2 := range v2 {
		magnitude2 += val2 * val2
	}
	
	if magnitude1 == 0 || magnitude2 == 0 {
		return 0.0
	}
	
	similarity := dotProduct / (math.Sqrt(magnitude1) * math.Sqrt(magnitude2))
	
	if similarity < 0 {
		similarity = 0
	}
	if similarity > 1 {
		similarity = 1
	}
	
	return similarity
}

func TextSimilarity(text1, text2 string, stopWords *StopWords) float64 {
	if strings.TrimSpace(text1) == "" || strings.TrimSpace(text2) == "" {
		return 0.0
	}
	
	v1, v2 := ComputeTFIDFFromTwoTexts(text1, text2, stopWords)
	return CosineSimilarity(v1, v2)
}

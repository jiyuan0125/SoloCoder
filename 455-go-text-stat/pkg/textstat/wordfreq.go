package textstat

import (
	"sort"
	"strings"
)

type WordFrequency struct {
	Word  string
	Count int
}

type WordFreqMap map[string]int

func CountWordFrequencies(text string, stopWords *StopWords) WordFreqMap {
	if text == "" {
		return make(WordFreqMap)
	}
	
	words := SplitWords(text)
	freqMap := make(WordFreqMap)
	
	for _, word := range words {
		lowerWord := strings.ToLower(word)
		if stopWords != nil && stopWords.IsStopWord(lowerWord) {
			continue
		}
		freqMap[lowerWord]++
	}
	
	return freqMap
}

func GetTopNWords(freqMap WordFreqMap, n int) []WordFrequency {
	if len(freqMap) == 0 || n <= 0 {
		return []WordFrequency{}
	}
	
	wordList := make([]WordFrequency, 0, len(freqMap))
	for word, count := range freqMap {
		wordList = append(wordList, WordFrequency{Word: word, Count: count})
	}
	
	sort.Slice(wordList, func(i, j int) bool {
		if wordList[i].Count == wordList[j].Count {
			return wordList[i].Word < wordList[j].Word
		}
		return wordList[i].Count > wordList[j].Count
	})
	
	if n > len(wordList) {
		n = len(wordList)
	}
	
	return wordList[:n]
}

func (m WordFreqMap) TotalWords() int {
	total := 0
	for _, count := range m {
		total += count
	}
	return total
}

func (m WordFreqMap) UniqueWords() int {
	return len(m)
}

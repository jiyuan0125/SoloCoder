package utils

import (
	"strings"
	"sync"
)

var sensitiveWords = []string{
	"垃圾", "废物", "蠢", "笨", "傻", "笨蛋", "白痴", "弱智",
	"去死", "滚", "去死吧", "该死", "混蛋", "王八蛋", "人渣",
}

var sensitiveTrie *Trie
var once sync.Once

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
}

type Trie struct {
	root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{root: &TrieNode{children: make(map[rune]*TrieNode)}}
}

func (t *Trie) Insert(word string) {
	node := t.root
	for _, ch := range word {
		if node.children[ch] == nil {
			node.children[ch] = &TrieNode{children: make(map[rune]*TrieNode)}
		}
		node = node.children[ch]
	}
	node.isEnd = true
}

func FilterSensitiveWords(text string) string {
	once.Do(func() {
		sensitiveTrie = NewTrie()
		for _, word := range sensitiveWords {
			sensitiveTrie.Insert(word)
		}
	})

	runes := []rune(text)
	result := make([]rune, 0, len(runes))
	i := 0

	for i < len(runes) {
		node := sensitiveTrie.root
		j := i
		matchLen := 0

		for j < len(runes) && node.children[runes[j]] != nil {
			node = node.children[runes[j]]
			j++
			if node.isEnd {
				matchLen = j - i
			}
		}

		if matchLen > 0 {
			for k := 0; k < matchLen; k++ {
				result = append(result, '*')
			}
			i += matchLen
		} else {
			result = append(result, runes[i])
			i++
		}
	}

	return string(result)
}

func RoundToTwoDecimals(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func TruncateComment(comment string, maxLen int) string {
	if len([]rune(comment)) <= maxLen {
		return comment
	}
	runes := []rune(comment)
	return string(runes[:maxLen])
}

func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

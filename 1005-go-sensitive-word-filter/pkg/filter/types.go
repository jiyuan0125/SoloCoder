package filter

import (
	"sensitive-word-filter/pkg/common"
	"sync"
)

type Filter struct {
	mu           sync.RWMutex
	trie         *TrieNode
	wildcards    []WildcardPattern
	whitelist    map[string]bool
	whitelistTrie *TrieNode
	
	dailyStats   map[string]*DailyCount
	hitStats     map[string]int64
}

type DailyCount struct {
	Total    int64
	HitCount int64
}

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	word     string
}

type WildcardPattern struct {
	original string
	parts    []string
}

type matchContext struct {
	text     []rune
	original string
}

type matchResult struct {
	word      string
	matchType common.MatchType
	positions []common.Position
	startIdx  int
	endIdx    int
}

package bktree

import (
	"sort"
	"sync"
)

type node struct {
	word     string
	children map[int]*node
}

type BKTree struct {
	root *node
	mu   sync.RWMutex
}

type Match struct {
	Word     string
	Distance int
}

func New() *BKTree {
	return &BKTree{}
}

func EditDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	m, n := len(s1), len(s2)
	prev := make([]int, n+1)
	curr := make([]int, n+1)

	for j := 0; j <= n; j++ {
		prev[j] = j
	}

	for i := 1; i <= m; i++ {
		curr[0] = i
		for j := 1; j <= n; j++ {
			if s1[i-1] == s2[j-1] {
				curr[j] = prev[j-1]
			} else {
				curr[j] = min(min(prev[j], curr[j-1]), prev[j-1]) + 1
			}
		}
		prev, curr = curr, prev
	}

	return prev[n]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (t *BKTree) Insert(word string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.root == nil {
		t.root = &node{word: word, children: make(map[int]*node)}
		return
	}

	current := t.root
	for {
		d := EditDistance(word, current.word)
		if d == 0 {
			current.word = word
			return
		}
		if child, exists := current.children[d]; exists {
			current = child
		} else {
			current.children[d] = &node{word: word, children: make(map[int]*node)}
			return
		}
	}
}

func (t *BKTree) Search(query string, maxDistance int) []Match {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if maxDistance < 0 || t.root == nil {
		return []Match{}
	}

	var results []Match
	t.searchRecursive(t.root, query, maxDistance, &results)

	sort.Slice(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})

	return results
}

func (t *BKTree) searchRecursive(current *node, query string, maxDistance int, results *[]Match) {
	d := EditDistance(query, current.word)
	if d <= maxDistance {
		*results = append(*results, Match{Word: current.word, Distance: d})
	}

	for dist, child := range current.children {
		if dist >= d-maxDistance && dist <= d+maxDistance {
			t.searchRecursive(child, query, maxDistance, results)
		}
	}
}

func (t *BKTree) Remove(word string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.root == nil {
		return false
	}

	if EditDistance(word, t.root.word) == 0 {
		oldRoot := t.root
		t.root = nil
		t.reinsertAll(oldRoot)
		return true
	}

	return t.removeRecursive(t.root, word)
}

func (t *BKTree) removeRecursive(current *node, word string) bool {
	for dist, child := range current.children {
		if EditDistance(word, child.word) == 0 {
			delete(current.children, dist)
			t.reinsertAll(child)
			return true
		}
		if t.removeRecursive(child, word) {
			return true
		}
	}
	return false
}

func (t *BKTree) reinsertAll(n *node) {
	var words []string
	var collect func(*node)
	collect = func(node *node) {
		for _, child := range node.children {
			words = append(words, child.word)
			collect(child)
		}
	}
	collect(n)

	for _, w := range words {
		t.insertInternal(w)
	}
}

func (t *BKTree) insertInternal(word string) {
	if t.root == nil {
		t.root = &node{word: word, children: make(map[int]*node)}
		return
	}

	current := t.root
	for {
		d := EditDistance(word, current.word)
		if d == 0 {
			return
		}
		if child, exists := current.children[d]; exists {
			current = child
		} else {
			current.children[d] = &node{word: word, children: make(map[int]*node)}
			return
		}
	}
}

func (t *BKTree) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.root == nil {
		return 0
	}

	count := 0
	var traverse func(*node)
	traverse = func(n *node) {
		count++
		for _, child := range n.children {
			traverse(child)
		}
	}
	traverse(t.root)
	return count
}

func (t *BKTree) Depth() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.root == nil {
		return 0
	}

	var maxDepth int
	var traverse func(*node, int)
	traverse = func(n *node, depth int) {
		if depth > maxDepth {
			maxDepth = depth
		}
		for _, child := range n.children {
			traverse(child, depth+1)
		}
	}
	traverse(t.root, 1)
	return maxDepth
}

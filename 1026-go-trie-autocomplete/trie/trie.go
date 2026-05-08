package trie

import (
	"strings"
	"sync/atomic"
)

type node struct {
	children map[rune]*node
	isEnd    bool
	word     string
}

type Trie struct {
	root atomic.Value
}

type rootWrapper struct {
	root *node
}

func newNode() *node {
	return &node{
		children: make(map[rune]*node),
	}
}

func New() *Trie {
	t := &Trie{}
	t.root.Store(&rootWrapper{root: newNode()})
	return t
}

func (t *Trie) getRoot() *node {
	w := t.root.Load().(*rootWrapper)
	return w.root
}

func (t *Trie) Insert(word string) {
	newRoot := t.pathCloneInsert(word)
	t.root.Store(&rootWrapper{root: newRoot})
}

func (t *Trie) Delete(word string) {
	newRoot := t.pathCloneDelete(word)
	if newRoot != nil {
		t.root.Store(&rootWrapper{root: newRoot})
	}
}

func (t *Trie) BuildFromList(words []string) {
	newRoot := buildTrieFromList(words)
	t.root.Store(&rootWrapper{root: newRoot})
}

func buildTrieFromList(words []string) *node {
	root := newNode()
	for _, word := range words {
		lowerWord := strings.ToLower(word)
		current := root
		for _, c := range lowerWord {
			if child, exists := current.children[c]; exists {
				current = child
			} else {
				newChild := newNode()
				current.children[c] = newChild
				current = newChild
			}
		}
		current.isEnd = true
		current.word = word
	}
	return root
}

func (t *Trie) pathCloneInsert(word string) *node {
	oldRoot := t.getRoot()
	lowerWord := strings.ToLower(word)

	path := []*node{}
	current := oldRoot
	exists := true

	for _, c := range lowerWord {
		path = append(path, current)
		if child, ok := current.children[c]; ok {
			current = child
		} else {
			exists = false
			break
		}
	}

	if exists {
		path = append(path, current)
		if current.isEnd && current.word == word {
			return oldRoot
		}
	}

	newRoot := cloneNode(oldRoot)
	newCurrent := newRoot

	for i := 0; i < len(path)-1; i++ {
		c := []rune(lowerWord)[i]
		oldChild := path[i].children[c]
		newChild := cloneNode(oldChild)
		newCurrent.children[c] = newChild
		newCurrent = newChild
	}

	remainingRunes := []rune(lowerWord)[len(path)-1:]
	if exists {
		newCurrent.isEnd = true
		newCurrent.word = word
	} else {
		for _, c := range remainingRunes {
			newChild := newNode()
			newCurrent.children[c] = newChild
			newCurrent = newChild
		}
		newCurrent.isEnd = true
		newCurrent.word = word
	}

	return newRoot
}

func (t *Trie) pathCloneDelete(word string) *node {
	oldRoot := t.getRoot()
	lowerWord := strings.ToLower(word)

	path := []*node{}
	current := oldRoot

	for _, c := range lowerWord {
		path = append(path, current)
		if child, ok := current.children[c]; ok {
			current = child
		} else {
			return nil
		}
	}

	if !current.isEnd {
		return nil
	}

	path = append(path, current)

	if len(current.children) > 0 {
		newRoot := cloneNode(oldRoot)
		newCurrent := newRoot
		for i := 0; i < len(path)-1; i++ {
			c := []rune(lowerWord)[i]
			oldChild := path[i].children[c]
			newChild := cloneNode(oldChild)
			newCurrent.children[c] = newChild
			newCurrent = newChild
		}
		newCurrent.isEnd = false
		newCurrent.word = ""
		return newRoot
	}

	trimIdx := len(path) - 2
	for trimIdx >= 0 {
		node := path[trimIdx+1]
		if len(node.children) > 1 || (node.isEnd && trimIdx+1 < len(path)-1) {
			break
		}
		trimIdx--
	}

	newRoot := cloneNode(oldRoot)
	newCurrent := newRoot

	for i := 0; i <= trimIdx; i++ {
		c := []rune(lowerWord)[i]
		oldChild := path[i].children[c]
		newChild := cloneNode(oldChild)
		newCurrent.children[c] = newChild
		newCurrent = newChild
	}

	if trimIdx >= 0 {
		c := []rune(lowerWord)[trimIdx+1]
		delete(newCurrent.children, c)
	}

	if trimIdx < 0 {
		return newNode()
	}

	return newRoot
}

func cloneNode(n *node) *node {
	newN := newNode()
	newN.isEnd = n.isEnd
	newN.word = n.word
	for c, child := range n.children {
		newN.children[c] = child
	}
	return newN
}

func (t *Trie) Search(prefix string, limit int) []string {
	if limit <= 0 {
		return []string{}
	}

	root := t.getRoot()
	lowerPrefix := strings.ToLower(prefix)

	if lowerPrefix == "" {
		return t.getAll(root, limit)
	}

	current := root
	for _, c := range lowerPrefix {
		if child, exists := current.children[c]; exists {
			current = child
		} else {
			return []string{}
		}
	}

	return t.collect(current, limit)
}

func (t *Trie) getAll(root *node, limit int) []string {
	return t.collect(root, limit)
}

func (t *Trie) collect(start *node, limit int) []string {
	results := []string{}
	count := 0

	var dfs func(*node)
	dfs = func(n *node) {
		if count >= limit {
			return
		}

		if n.isEnd {
			results = append(results, n.word)
			count++
			if count >= limit {
				return
			}
		}

		for _, child := range n.children {
			dfs(child)
			if count >= limit {
				return
			}
		}
	}

	dfs(start)
	return results
}

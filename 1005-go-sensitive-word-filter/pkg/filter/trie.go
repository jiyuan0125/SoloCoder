package filter

func newTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[rune]*TrieNode),
	}
}

func (t *TrieNode) insert(word string) {
	current := t
	for _, ch := range word {
		if _, exists := current.children[ch]; !exists {
			current.children[ch] = newTrieNode()
		}
		current = current.children[ch]
	}
	current.isEnd = true
	current.word = word
}

func (t *TrieNode) remove(word string) {
	var removeHelper func(node *TrieNode, word string, index int) bool
	removeHelper = func(node *TrieNode, word string, index int) bool {
		if index == len([]rune(word)) {
			if !node.isEnd {
				return false
			}
			node.isEnd = false
			node.word = ""
			return len(node.children) == 0
		}
		ch := []rune(word)[index]
		child, exists := node.children[ch]
		if !exists {
			return false
		}
		shouldRemove := removeHelper(child, word, index+1)
		if shouldRemove {
			delete(node.children, ch)
			return len(node.children) == 0 && !node.isEnd
		}
		return false
	}
	removeHelper(t, word, 0)
}

func (t *TrieNode) contains(word string) bool {
	current := t
	for _, ch := range word {
		child, exists := current.children[ch]
		if !exists {
			return false
		}
		current = child
	}
	return current.isEnd
}

func (t *TrieNode) listWords() []string {
	var words []string
	var collect func(node *TrieNode)
	collect = func(node *TrieNode) {
		if node.isEnd {
			words = append(words, node.word)
		}
		for _, child := range node.children {
			collect(child)
		}
	}
	collect(t)
	return words
}

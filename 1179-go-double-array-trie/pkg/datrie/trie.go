package datrie

import (
	"errors"
	"fmt"
	"sort"
)

const (
	RootIndex          = 1
	BaseRoot           = 1
	CheckRoot          = -1
	CharStart     byte = 32
	CharEnd       byte = 126
	CharRange          = int(CharEnd - CharStart + 1)
	NotFound           = -1
	InitialArraySize   = 1024
	ExpandFactor       = 2
)

var (
	ErrInvalidChar  = errors.New("invalid character")
	ErrEmptyTrie    = errors.New("trie is empty")
	ErrEmptyWord    = errors.New("empty word")
	ErrWordExists   = errors.New("word already exists")
	ErrWordNotFound = errors.New("word not found")
)

type Stat struct {
	BaseSize    int
	CheckSize   int
	UsedSlots   int
	WordCount   int
	Utilization float64
}

type Trie struct {
	base     []int
	check    []int
	terminal []bool
	count    int
	size     int
}

func NewTrie() *Trie {
	t := &Trie{
		base:     make([]int, InitialArraySize),
		check:    make([]int, InitialArraySize),
		terminal: make([]bool, InitialArraySize),
		count:    0,
		size:     InitialArraySize,
	}
	t.base[0] = BaseRoot
	t.check[0] = CheckRoot
	return t
}

func (t *Trie) validateWord(word string) error {
	if len(word) == 0 {
		return nil
	}
	for i := 0; i < len(word); i++ {
		c := word[i]
		if c < CharStart || c > CharEnd {
			return fmt.Errorf("%w: '%c' at position %d", ErrInvalidChar, c, i)
		}
	}
	return nil
}

func (t *Trie) charCode(c byte) int {
	return int(c - CharStart) + 1
}

func (t *Trie) expand(requiredSize int) {
	if t.size >= requiredSize {
		return
	}
	newSize := t.size
	for newSize < requiredSize {
		newSize *= ExpandFactor
	}
	newBase := make([]int, newSize)
	newCheck := make([]int, newSize)
	newTerminal := make([]bool, newSize)
	copy(newBase, t.base)
	copy(newCheck, t.check)
	copy(newTerminal, t.terminal)
	t.base = newBase
	t.check = newCheck
	t.terminal = newTerminal
	t.size = newSize
}

func (t *Trie) Insert(word string) error {
	if err := t.validateWord(word); err != nil {
		return err
	}

	if len(word) == 0 {
		if t.terminal[RootIndex] {
			return ErrWordExists
		}
		t.terminal[RootIndex] = true
		t.count++
		return nil
	}

	s := RootIndex
	for i := 0; i < len(word); i++ {
		c := t.charCode(word[i])
		t.expand(t.base[s] + c + 1)
		next := t.base[s] + c
		if t.check[next] == 0 {
			t.check[next] = s
			t.base[next] = next + CharRange
		} else if t.check[next] != s {
			if err := t.resolveConflict(s); err != nil {
				return err
			}
			next = t.base[s] + c
			t.check[next] = s
			t.base[next] = next + CharRange
		}
		s = next
	}

	if t.terminal[s] {
		return ErrWordExists
	}
	t.terminal[s] = true
	t.count++
	return nil
}

func (t *Trie) resolveConflict(s int) error {
	children := t.collectChildren(s)

	var newBase int
	for newBase = RootIndex; ; newBase++ {
		valid := true
		for _, c := range children {
			next := newBase + c
			t.expand(next + 1)
			if t.check[next] != 0 {
				valid = false
				break
			}
		}
		if valid {
			break
		}
	}

	for _, c := range children {
		oldNext := t.base[s] + c
		newNext := newBase + c
		t.check[newNext] = s
		t.base[newNext] = t.base[oldNext]
		t.terminal[newNext] = t.terminal[oldNext]

		oldChildren := t.collectChildren(oldNext)
		for _, childC := range oldChildren {
			childNext := t.base[oldNext] + childC
			t.check[childNext] = newNext
		}

		t.check[oldNext] = 0
		t.base[oldNext] = 0
		t.terminal[oldNext] = false
	}

	t.base[s] = newBase
	return nil
}

func (t *Trie) collectChildren(s int) []int {
	var children []int
	for c := 1; c <= CharRange; c++ {
		next := t.base[s] + c
		if next >= t.size {
			continue
		}
		if t.check[next] == s {
			children = append(children, c)
		}
	}
	return children
}

func (t *Trie) Search(word string) bool {
	if err := t.validateWord(word); err != nil {
		return false
	}

	if len(word) == 0 {
		return t.terminal[RootIndex]
	}

	s := RootIndex
	for i := 0; i < len(word); i++ {
		c := t.charCode(word[i])
		next := t.base[s] + c
		if next >= t.size || t.check[next] != s {
			return false
		}
		s = next
	}
	return t.terminal[s]
}

func (t *Trie) Prefix(prefix string) []string {
	if err := t.validateWord(prefix); err != nil {
		return nil
	}

	s := RootIndex
	for i := 0; i < len(prefix); i++ {
		c := t.charCode(prefix[i])
		next := t.base[s] + c
		if next >= t.size || t.check[next] != s {
			return nil
		}
		s = next
	}

	var results []string
	t.collectWords(s, prefix, &results)
	sort.Strings(results)
	return results
}

func (t *Trie) collectWords(s int, prefix string, results *[]string) {
	if t.terminal[s] {
		*results = append(*results, prefix)
	}

	for c := 1; c <= CharRange; c++ {
		next := t.base[s] + c
		if next >= t.size {
			continue
		}
		if t.check[next] == s {
			char := CharStart + byte(c-1)
			t.collectWords(next, prefix+string(char), results)
		}
	}
}

func (t *Trie) Delete(word string) error {
	if err := t.validateWord(word); err != nil {
		return err
	}

	if len(word) == 0 {
		if !t.terminal[RootIndex] {
			return ErrWordNotFound
		}
		t.terminal[RootIndex] = false
		t.count--
		return nil
	}

	s := RootIndex
	for i := 0; i < len(word); i++ {
		c := t.charCode(word[i])
		next := t.base[s] + c
		if next >= t.size || t.check[next] != s {
			return ErrWordNotFound
		}
		s = next
	}

	if !t.terminal[s] {
		return ErrWordNotFound
	}

	t.terminal[s] = false
	t.count--
	return nil
}

func (t *Trie) Stat() *Stat {
	used := 0
	for i := 0; i < t.size; i++ {
		if t.check[i] != 0 {
			used++
		}
	}

	utilization := 0.0
	if t.size > 0 {
		utilization = float64(used) / float64(t.size) * 100
	}

	return &Stat{
		BaseSize:    t.size,
		CheckSize:   t.size,
		UsedSlots:   used,
		WordCount:   t.count,
		Utilization: utilization,
	}
}

func (t *Trie) Build(words []string) error {
	sortedWords := make([]string, len(words))
	copy(sortedWords, words)
	sort.Strings(sortedWords)

	for _, word := range sortedWords {
		if err := t.Insert(word); err != nil {
			if err == ErrWordExists {
				continue
			}
			return err
		}
	}
	return nil
}

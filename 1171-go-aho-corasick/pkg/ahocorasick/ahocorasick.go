package ahocorasick

import (
	"errors"
	"sort"
)

type Match struct {
	Pattern string
	Start   int
	End     int
}

type node struct {
	children map[rune]*node
	fail     *node
	output   []string
}

type Matcher struct {
	root        *node
	patterns    map[string]struct{}
	patternList []string
	built       bool
}

func NewMatcher() *Matcher {
	return &Matcher{
		root:     &node{children: make(map[rune]*node)},
		patterns: make(map[string]struct{}),
		built:    false,
	}
}

var ErrEmptyPattern = errors.New("empty pattern is not allowed")

func (m *Matcher) Add(pattern string) error {
	if pattern == "" {
		return ErrEmptyPattern
	}

	if _, exists := m.patterns[pattern]; exists {
		return nil
	}

	m.patterns[pattern] = struct{}{}
	m.patternList = append(m.patternList, pattern)
	m.built = false

	current := m.root
	for _, ch := range pattern {
		if current.children == nil {
			current.children = make(map[rune]*node)
		}
		if next, ok := current.children[ch]; ok {
			current = next
		} else {
			newNode := &node{children: make(map[rune]*node)}
			current.children[ch] = newNode
			current = newNode
		}
	}
	current.output = append(current.output, pattern)

	return nil
}

func (m *Matcher) AddPatterns(patterns []string) (added int, skipped int, err error) {
	for _, pattern := range patterns {
		if pattern == "" {
			skipped++
			continue
		}
		if _, exists := m.patterns[pattern]; exists {
			skipped++
			continue
		}
		addErr := m.Add(pattern)
		if addErr != nil {
			if addErr == ErrEmptyPattern {
				skipped++
			} else {
				err = addErr
			}
			continue
		}
		added++
	}
	return
}

func (m *Matcher) Build() {
	if m.built {
		return
	}

	queue := make([]*node, 0)

	m.root.fail = nil

	for _, child := range m.root.children {
		child.fail = m.root
		queue = append(queue, child)
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for ch, child := range current.children {
			failNode := current.fail

			for failNode != nil && failNode.children[ch] == nil {
				failNode = failNode.fail
			}

			if failNode != nil {
				child.fail = failNode.children[ch]
			} else {
				child.fail = m.root
			}

			if child.fail != nil && len(child.fail.output) > 0 {
				child.output = append(child.output, child.fail.output...)
			}

			queue = append(queue, child)
		}
	}

	m.built = true
}

func (m *Matcher) Search(text string) []Match {
	if text == "" {
		return []Match{}
	}

	if len(m.patterns) == 0 {
		return []Match{}
	}

	if !m.built {
		m.Build()
	}

	matches := make([]Match, 0)
	current := m.root

	runes := []rune(text)
	index := 0

	for i, ch := range runes {
		for current != m.root && current.children[ch] == nil {
			current = current.fail
		}

		if next, ok := current.children[ch]; ok {
			current = next
		}

		if len(current.output) > 0 {
			for _, pattern := range current.output {
				patternRunes := []rune(pattern)
				patternLen := len(patternRunes)
				start := i - patternLen + 1
				if start < 0 {
					continue
				}

				byteStart := 0
				for j := 0; j < start; j++ {
					byteStart += len(string(runes[j]))
				}
				byteEnd := byteStart + len(pattern)

				matches = append(matches, Match{
					Pattern: pattern,
					Start:   byteStart,
					End:     byteEnd,
				})
			}
		}
		index++
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Start != matches[j].Start {
			return matches[i].Start < matches[j].Start
		}
		return len(matches[i].Pattern) > len(matches[j].Pattern)
	})

	return matches
}

func (m *Matcher) Clear() {
	m.root = &node{children: make(map[rune]*node)}
	m.patterns = make(map[string]struct{})
	m.patternList = []string{}
	m.built = false
}

func (m *Matcher) Count() int {
	return len(m.patternList)
}

func (m *Matcher) Patterns() []string {
	patterns := make([]string, len(m.patternList))
	copy(patterns, m.patternList)
	return patterns
}

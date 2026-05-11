package rabinkarp

import (
	"reflect"
	"sort"
	"testing"
)

func TestAddAndPatterns(t *testing.T) {
	m := NewMatcher()

	if !m.Add("hello") {
		t.Error("Expected Add to return true for new pattern")
	}

	if m.Add("hello") {
		t.Error("Expected Add to return false for duplicate pattern")
	}

	if m.Add("") {
		t.Error("Expected Add to return false for empty pattern")
	}

	patterns := m.Patterns()
	if len(patterns) != 1 {
		t.Errorf("Expected 1 pattern, got %d", len(patterns))
	}
}

func TestSearchBasic(t *testing.T) {
	m := NewMatcher()
	m.Add("hello")
	m.Add("world")

	matches := m.Search("hello world hello")
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}

	patterns := make([]string, len(matches))
	for i, m := range matches {
		patterns[i] = m.Pattern
	}
	sort.Strings(patterns)
	if !reflect.DeepEqual(patterns, []string{"hello", "hello", "world"}) {
		t.Errorf("Expected [hello, hello, world], got %v", patterns)
	}
}

func TestSearchMultipleLengths(t *testing.T) {
	m := NewMatcher()
	m.Add("ab")
	m.Add("abab")

	matches := m.Search("ababab")
	if len(matches) != 5 {
		t.Errorf("Expected 5 matches, got %d: %v", len(matches), matches)
	}

	countAB := 0
	countABAB := 0
	for _, m := range matches {
		if m.Pattern == "ab" {
			countAB++
		} else if m.Pattern == "abab" {
			countABAB++
		}
	}

	if countAB != 3 {
		t.Errorf("Expected 3 'ab' matches, got %d", countAB)
	}
	if countABAB != 2 {
		t.Errorf("Expected 2 'abab' matches, got %d", countABAB)
	}
}

func TestSearchSameCharacters(t *testing.T) {
	m := NewMatcher()
	m.Add("aaa")

	matches := m.Search("aaaaa")
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches for 'aaa' in 'aaaaa', got %d", len(matches))
	}

	expectedIndices := []int{0, 1, 2}
	for i, match := range matches {
		if match.Index != expectedIndices[i] {
			t.Errorf("Match %d: expected index %d, got %d", i, expectedIndices[i], match.Index)
		}
	}
}

func TestSearchPatternLongerThanText(t *testing.T) {
	m := NewMatcher()
	m.Add("thisisalongpattern")

	matches := m.Search("short")
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}
}

func TestSearchEmptyText(t *testing.T) {
	m := NewMatcher()
	m.Add("hello")

	matches := m.Search("")
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for empty text, got %d", len(matches))
	}
}

func TestHashFunction(t *testing.T) {
	h1 := Hash("hello")
	h2 := Hash("hello")
	if h1 != h2 {
		t.Error("Same string should have same hash")
	}

	h3 := Hash("world")
	if h1 == h3 {
		t.Error("Different strings should have different hash (high probability)")
	}
}

func TestRemove(t *testing.T) {
	m := NewMatcher()
	m.Add("hello")
	m.Add("world")

	if !m.Remove("hello") {
		t.Error("Expected Remove to return true for existing pattern")
	}

	if m.Remove("hello") {
		t.Error("Expected Remove to return false for non-existing pattern")
	}

	if m.Remove("") {
		t.Error("Expected Remove to return false for empty pattern")
	}

	patterns := m.Patterns()
	if len(patterns) != 1 || patterns[0] != "world" {
		t.Errorf("Expected only 'world' to remain, got %v", patterns)
	}
}

func TestAddBatch(t *testing.T) {
	m := NewMatcher()
	patterns := []string{"hello", "world", "hello", "", "test"}

	added, skipped := m.AddBatch(patterns)
	if added != 3 {
		t.Errorf("Expected 3 added, got %d", added)
	}
	if skipped != 2 {
		t.Errorf("Expected 2 skipped, got %d", skipped)
	}
}

func TestPatternsByLength(t *testing.T) {
	m := NewMatcher()
	m.Add("a")
	m.Add("ab")
	m.Add("abc")
	m.Add("abcd")

	byLen := m.PatternsByLength()
	if len(byLen) != 4 {
		t.Errorf("Expected 4 length groups, got %d", len(byLen))
	}
}

func TestSearchMultiplePatternsAtSamePosition(t *testing.T) {
	m := NewMatcher()
	m.Add("ab")
	m.Add("abc")
	m.Add("abcd")

	matches := m.Search("abcd")
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d: %v", len(matches), matches)
	}
}

func TestCollisionTracking(t *testing.T) {
	m := NewMatcher()
	m.Add("hello")
	m.Add("world")

	_ = m.Search("hello world test hello")

	stats := m.CollisionStats()
	if stats.TotalChecks == 0 {
		t.Error("Expected total checks to be tracked")
	}
}

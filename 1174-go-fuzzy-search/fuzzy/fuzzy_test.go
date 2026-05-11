package fuzzy

import "testing"

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		source, target string
		expected       int
	}{
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"hello", "hello", 0},
		{"hello", "hallo", 1},
		{"中国", "中", 1},
		{"你好", "你好世界", 2},
	}

	for _, tt := range tests {
		result := LevenshteinDistance(tt.source, tt.target)
		if result != tt.expected {
			t.Errorf("LevenshteinDistance(%q, %q) = %d, want %d", tt.source, tt.target, result, tt.expected)
		}
	}
}

func TestNormalizeWildcards(t *testing.T) {
	tests := []struct {
		pattern  string
		expected string
	}{
		{"a*b", "a*b"},
		{"a**b", "a*b"},
		{"a***b", "a*b"},
		{"**a", "*a"},
		{"a**", "a*"},
		{"**", "*"},
		{"a?b", "a?b"},
	}

	for _, tt := range tests {
		result := NormalizeWildcards(tt.pattern)
		if result != tt.expected {
			t.Errorf("NormalizeWildcards(%q) = %q, want %q", tt.pattern, result, tt.expected)
		}
	}
}

func TestWildcardMatch(t *testing.T) {
	tests := []struct {
		pattern, text string
		threshold     int
		expectedMatch bool
		expectedDist  int
	}{
		{"h?llo", "hello", 0, true, 0},
		{"h?llo", "hallo", 0, true, 0},
		{"h*llo", "hello", 0, true, 0},
		{"h*llo", "hllo", 0, true, 0},
		{"h*llo", "haxyzllo", 0, true, 0},
		{"test", "test", 0, true, 0},
		{"test", "tast", 1, true, 1},
		{"test", "tast", 0, false, 0},
		{"*", "anything", 0, true, 0},
		{"?", "a", 0, true, 0},
		{"a", "", 0, false, 0},
	}

	for _, tt := range tests {
		match, dist := WildcardMatch(tt.pattern, tt.text, tt.threshold)
		if match != tt.expectedMatch || (match && dist != tt.expectedDist) {
			t.Errorf("WildcardMatch(%q, %q, %d) = (%v, %d), want (%v, %d)",
				tt.pattern, tt.text, tt.threshold, match, dist, tt.expectedMatch, tt.expectedDist)
		}
	}
}

func TestDictionary(t *testing.T) {
	d := NewDictionary()

	if d.Size() != 0 {
		t.Errorf("New dictionary size should be 0, got %d", d.Size())
	}

	if !d.Add("hello") {
		t.Error("Should be able to add new word 'hello'")
	}

	if d.Add("hello") {
		t.Error("Should not be able to add duplicate word 'hello'")
	}

	if d.Size() != 1 {
		t.Errorf("Dictionary size should be 1, got %d", d.Size())
	}

	if !d.Contains("hello") {
		t.Error("Dictionary should contain 'hello'")
	}

	if d.Contains("world") {
		t.Error("Dictionary should not contain 'world'")
	}

	if !d.Remove("hello") {
		t.Error("Should be able to remove existing word 'hello'")
	}

	if d.Remove("hello") {
		t.Error("Should not be able to remove non-existing word")
	}

	if d.Size() != 0 {
		t.Errorf("Dictionary size should be 0 after removal, got %d", d.Size())
	}
}

func TestSearch(t *testing.T) {
	d := NewDictionary()
	d.Add("hello")
	d.Add("hallo")
	d.Add("help")
	d.Add("world")
	d.Add("")

	results := d.Search("helo", 2)
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if len(results) > 0 {
		if results[0].Word != "hello" || results[0].Distance != 1 {
			t.Errorf("First result should be 'hello' with distance 1, got %q with distance %d",
				results[0].Word, results[0].Distance)
		}
	}

	results = d.Search("hello", 0)
	if len(results) != 1 || results[0].Word != "hello" {
		t.Errorf("Exact search for 'hello' should return exactly 'hello'")
	}

	results = d.Search("", 0)
	if len(results) != 1 || results[0].Word != "" {
		t.Errorf("Exact search for empty string should return empty string")
	}
}

func TestWildcardSearch(t *testing.T) {
	d := NewDictionary()
	d.Add("hello")
	d.Add("hallo")
	d.Add("help")
	d.Add("world")

	results := d.WildcardSearch("h?llo", 0)
	if len(results) != 2 {
		t.Errorf("Expected 2 results for h?llo, got %d", len(results))
	}

	results = d.WildcardSearch("h*", 0)
	if len(results) != 3 {
		t.Errorf("Expected 3 results for h*, got %d", len(results))
	}
}

func TestImport(t *testing.T) {
	d := NewDictionary()
	words := []string{"hello", "world", "hello", "test"}
	count := d.Import(words)
	
	if count != 3 {
		t.Errorf("Should import 3 unique words, got %d", count)
	}
	
	if d.Size() != 3 {
		t.Errorf("Dictionary size should be 3, got %d", d.Size())
	}
}

func TestGetAll(t *testing.T) {
	d := NewDictionary()
	d.Add("banana")
	d.Add("apple")
	d.Add("cherry")

	words := d.GetAll()
	if len(words) != 3 {
		t.Errorf("Expected 3 words, got %d", len(words))
	}

	expected := []string{"apple", "banana", "cherry"}
	for i, word := range expected {
		if words[i] != word {
			t.Errorf("Word %d should be %q, got %q", i, word, words[i])
		}
	}
}

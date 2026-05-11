package bmsearch

import (
	"reflect"
	"testing"
)

func TestSearchBasic(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pattern  string
		expected []int
	}{
		{"simple match", "abcabcabc", "abc", []int{0, 3, 6}},
		{"no match", "abcabcabc", "xyz", []int{}},
		{"pattern longer than text", "abc", "abcdef", []int{}},
		{"single char match", "aaaaa", "a", []int{0, 1, 2, 3, 4}},
		{"single char no match", "aaaaa", "b", []int{}},
		{"all same pattern", "aaaaa", "aaa", []int{0, 1, 2}},
		{"text empty", "", "abc", []int{}},
		{"pattern empty", "abc", "", []int{}},
		{"both empty", "", "", []int{}},
		{"overlapping matches", "abababa", "aba", []int{0, 2, 4}},
		{"pattern at end", "hello world", "world", []int{6}},
		{"pattern at start", "hello world", "hello", []int{0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Search(tt.text, tt.pattern)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Search(%q, %q) = %v, want %v", tt.text, tt.pattern, got, tt.expected)
			}
		})
	}
}

func TestBuildBadCharTable(t *testing.T) {
	pattern := "abac"
	bmBc := BuildBadCharTable(pattern)
	if bmBc['a'] != 1 {
		t.Errorf("bmBc['a'] = %d, want 1", bmBc['a'])
	}
	if bmBc['b'] != 2 {
		t.Errorf("bmBc['b'] = %d, want 2", bmBc['b'])
	}
	if bmBc['c'] != 0 {
		t.Errorf("bmBc['c'] = %d, want 0", bmBc['c'])
	}
}

func TestPreprocess(t *testing.T) {
	_, err := Preprocess("")
	if err == nil {
		t.Error("Preprocess('') should return error")
	}

	result, err := Preprocess("test")
	if err != nil {
		t.Errorf("Preprocess('test') returned error: %v", err)
	}
	if result.Pattern != "test" {
		t.Errorf("Pattern = %q, want 'test'", result.Pattern)
	}
	if len(result.BmBc) != ASCIIMax {
		t.Errorf("BmBc length = %d, want %d", len(result.BmBc), ASCIIMax)
	}
	if len(result.BmGs) != 4 {
		t.Errorf("BmGs length = %d, want 4", len(result.BmGs))
	}
}

func TestHelloWorld(t *testing.T) {
	text := "hello world"
	pattern := "world"
	result := Search(text, pattern)
	expected := []int{6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Search(%q, %q) = %v, want %v", text, pattern, result, expected)
	}
}

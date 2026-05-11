package datrie

import (
	"testing"
)

func TestNewTrie(t *testing.T) {
	trie := NewTrie()
	if trie == nil {
		t.Fatal("NewTrie returned nil")
	}
	if trie.base[0] != BaseRoot {
		t.Errorf("Expected base[0] = %d, got %d", BaseRoot, trie.base[0])
	}
	if trie.check[0] != CheckRoot {
		t.Errorf("Expected check[0] = %d, got %d", CheckRoot, trie.check[0])
	}
}

func TestInsertAndSearch(t *testing.T) {
	trie := NewTrie()

	words := []string{"apple", "banana", "app", "application"}
	for _, word := range words {
		if err := trie.Insert(word); err != nil {
			t.Fatalf("Failed to insert '%s': %v", word, err)
		}
	}

	for _, word := range words {
		if !trie.Search(word) {
			t.Errorf("Expected '%s' to exist", word)
		}
	}

	nonExistent := []string{"appl", "bananas", "orange", ""}
	for _, word := range nonExistent {
		if trie.Search(word) {
			t.Errorf("Expected '%s' to not exist", word)
		}
	}
}

func TestInsertDuplicate(t *testing.T) {
	trie := NewTrie()
	if err := trie.Insert("hello"); err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}
	if err := trie.Insert("hello"); err != ErrWordExists {
		t.Errorf("Expected ErrWordExists, got %v", err)
	}
}

func TestEmptyWord(t *testing.T) {
	trie := NewTrie()

	if trie.Search("") {
		t.Error("Empty string should not exist initially")
	}

	if err := trie.Insert(""); err != nil {
		t.Fatalf("Failed to insert empty string: %v", err)
	}

	if !trie.Search("") {
		t.Error("Empty string should exist after insertion")
	}
}

func TestDelete(t *testing.T) {
	trie := NewTrie()
	trie.Insert("test")
	trie.Insert("testing")

	if !trie.Search("test") {
		t.Error("test should exist")
	}

	if err := trie.Delete("test"); err != nil {
		t.Fatalf("Failed to delete 'test': %v", err)
	}

	if trie.Search("test") {
		t.Error("test should not exist after deletion")
	}

	if !trie.Search("testing") {
		t.Error("testing should still exist")
	}

	if err := trie.Delete("nonexistent"); err != ErrWordNotFound {
		t.Errorf("Expected ErrWordNotFound, got %v", err)
	}
}

func TestPrefix(t *testing.T) {
	trie := NewTrie()
	words := []string{"app", "apple", "application", "apples", "banana"}
	for _, word := range words {
		trie.Insert(word)
	}

	results := trie.Prefix("app")
	expected := []string{"app", "apple", "apples", "application"}
	if len(results) != len(expected) {
		t.Errorf("Expected %d results, got %d", len(expected), len(results))
	}

	results = trie.Prefix("appl")
	expected = []string{"apple", "apples", "application"}
	if len(results) != len(expected) {
		t.Errorf("Expected %d results for 'appl', got %d", len(expected), len(results))
	}

	results = trie.Prefix("xyz")
	if results != nil {
		t.Errorf("Expected nil for non-existent prefix, got %v", results)
	}
}

func TestInvalidCharacter(t *testing.T) {
	trie := NewTrie()

	invalidWords := []string{"hello\x00", "test\t", "abc\n"}
	for _, word := range invalidWords {
		if err := trie.Insert(word); err == nil {
			t.Errorf("Expected error for invalid word '%s'", word)
		}
		if trie.Search(word) {
			t.Errorf("Expected Search to return false for invalid word '%s'", word)
		}
	}
}

func TestBuild(t *testing.T) {
	trie := NewTrie()
	words := []string{"banana", "apple", "cherry", "apple", "date"}
	if err := trie.Build(words); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	expected := []string{"apple", "banana", "cherry", "date"}
	for _, word := range expected {
		if !trie.Search(word) {
			t.Errorf("Expected '%s' to exist after Build", word)
		}
	}

	stat := trie.Stat()
	if stat.WordCount != 4 {
		t.Errorf("Expected 4 words (duplicate ignored), got %d", stat.WordCount)
	}
}

func TestStat(t *testing.T) {
	trie := NewTrie()

	stat := trie.Stat()
	if stat.WordCount != 0 {
		t.Errorf("Expected 0 words, got %d", stat.WordCount)
	}

	trie.Insert("hello")
	stat = trie.Stat()
	if stat.WordCount != 1 {
		t.Errorf("Expected 1 word, got %d", stat.WordCount)
	}
	if stat.BaseSize != stat.CheckSize {
		t.Error("BaseSize and CheckSize should be equal")
	}
}

func TestConflictResolution(t *testing.T) {
	trie := NewTrie()

	words := []string{"a", "b", "c", "ab", "ac", "ad", "abc", "abd", "abe"}
	for _, word := range words {
		if err := trie.Insert(word); err != nil {
			t.Fatalf("Failed to insert '%s': %v", word, err)
		}
	}

	for _, word := range words {
		if !trie.Search(word) {
			t.Errorf("Expected '%s' to exist after conflict resolution", word)
		}
	}
}

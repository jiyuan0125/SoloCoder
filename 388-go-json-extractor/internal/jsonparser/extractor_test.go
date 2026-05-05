package jsonparser

import (
	"strings"
	"testing"
)

func TestFieldNotFound(t *testing.T) {
	extractor := NewExtractor()
	
	jsonData := `{"name": "test", "value": 123}`
	reader := strings.NewReader(jsonData)
	
	results, err := extractor.ExtractFromReader(reader, []string{"nonexistent.field"}, false, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	expected := "nonexistent.field="
	if results[0] != expected {
		t.Errorf("Expected '%s', got '%s'", expected, results[0])
	}
}

func TestArrayIndexOutOfBounds(t *testing.T) {
	extractor := NewExtractor()
	
	jsonData := `{"users": [{"name": "Alice"}, {"name": "Bob"}]}`
	reader := strings.NewReader(jsonData)
	
	results, err := extractor.ExtractFromReader(reader, []string{"users.5.name"}, false, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	expected := "users.5.name="
	if results[0] != expected {
		t.Errorf("Expected '%s', got '%s'", expected, results[0])
	}
}

func TestPathInterruptedByNull(t *testing.T) {
	extractor := NewExtractor()
	
	jsonData := `{"a": null}`
	reader := strings.NewReader(jsonData)
	
	results, err := extractor.ExtractFromReader(reader, []string{"a.b.c"}, false, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	expected := "a.b.c="
	if results[0] != expected {
		t.Errorf("Expected '%s', got '%s'", expected, results[0])
	}
}

func TestActualNullValue(t *testing.T) {
	extractor := NewExtractor()
	
	jsonData := `{"name": null}`
	reader := strings.NewReader(jsonData)
	
	results, err := extractor.ExtractFromReader(reader, []string{"name"}, false, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	expected := "name=null"
	if results[0] != expected {
		t.Errorf("Expected '%s', got '%s'", expected, results[0])
	}
}

func TestCompactModeNotFound(t *testing.T) {
	extractor := NewExtractor()
	
	jsonData := `{"name": "test"}`
	reader := strings.NewReader(jsonData)
	
	results, err := extractor.ExtractFromReader(reader, []string{"nonexistent"}, false, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	expected := ""
	if results[0] != expected {
		t.Errorf("Expected '%s' (empty string), got '%s'", expected, results[0])
	}
}

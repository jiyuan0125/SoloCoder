package phonetic

import "testing"

func TestSoundex(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Robert", "R163"},
		{"Rupert", "R163"},
		{"Adams", "A352"},
		{"aeiou", "A000"},
		{"Smith", "S530"},
		{"Schmidt", "S530"},
		{"", ""},
		{"12345", ""},
		{"John Smith", "J525"},
		{"bcdfg", "B231"},
		{"Adams", "A352"},
	}

	for _, test := range tests {
		result := Soundex(test.input)
		if result != test.expected {
			t.Errorf("Soundex(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestMetaphone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"phone", "FN"},
		{"PHONE", "FN"},
		{"Smith", "SM0"},
		{"Schmidt", "XMTT"},
		{"", ""},
		{"123", ""},
	}

	for _, test := range tests {
		result := Metaphone(test.input)
		if result != test.expected {
			t.Errorf("Metaphone(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestStore(t *testing.T) {
	store := NewInMemoryStore()

	if !store.Add("Smith") {
		t.Error("Add Smith failed")
	}

	if store.Add("Smith") {
		t.Error("Duplicate add should return false")
	}

	if !store.Add("Schmidt") {
		t.Error("Add Schmidt failed")
	}

	if store.Total() != 2 {
		t.Errorf("Total should be 2, got %d", store.Total())
	}

	matches := store.SearchBySoundex("S530")
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}

	if store.Remove("Smith") != true {
		t.Error("Remove Smith failed")
	}

	if store.Total() != 1 {
		t.Errorf("Total should be 1, got %d", store.Total())
	}
}

package uninorm

import (
	"testing"
)

func TestNormalizeNFD(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"pure ASCII", "hello", "hello"},
		{"é precomposed", "\u00e9", "e\u0301"},
		{"É precomposed", "\u00c9", "E\u0301"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input, NFD)
			if result != tt.expected {
				t.Errorf("Normalize(%q, NFD) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeNFC(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"pure ASCII", "hello", "hello"},
		{"e + combining acute", "e\u0301", "\u00e9"},
		{"E + combining acute", "E\u0301", "\u00c9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input, NFC)
			if result != tt.expected {
				t.Errorf("Normalize(%q, NFC) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeNFKC(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"fullwidth A", "\uff21", "A"},
		{"fullwidth a", "\uff41", "a"},
		{"roman numeral I", "\u2160", "I"},
		{"fi ligature", "\ufb01", "fi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input, NFKC)
			if result != tt.expected {
				t.Errorf("Normalize(%q, NFKC) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeNFKD(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"fullwidth A", "\uff21", "A"},
		{"fi ligature", "\ufb01", "fi"},
		{"fraction 1/2", "\u00bd", "1\u20442"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input, NFKD)
			if result != tt.expected {
				t.Errorf("Normalize(%q, NFKD) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseForm(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Form
		wantErr  bool
	}{
		{"NFC uppercase", "NFC", NFC, false},
		{"NFC lowercase", "nfc", NFC, false},
		{"NFD", "NFD", NFD, false},
		{"NFKC", "NFKC", NFKC, false},
		{"NFKD", "NFKD", NFKD, false},
		{"invalid", "INVALID", NFC, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseForm(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseForm(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("ParseForm(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAnalyzeChanges(t *testing.T) {
	changes, normalized := AnalyzeChanges("\u00e9", NFC)
	
	if normalized != "\u00e9" {
		t.Errorf("AnalyzeChanges normalized = %q, want %q", normalized, "\u00e9")
	}
	
	if len(changes) != 0 {
		t.Errorf("AnalyzeChanges should return 0 changes for already normalized text, got %d", len(changes))
	}
	
	changes, normalized = AnalyzeChanges("\uff21", NFKC)
	
	if normalized != "A" {
		t.Errorf("AnalyzeChanges normalized = %q, want %q", normalized, "A")
	}
	
	if len(changes) != 1 {
		t.Errorf("AnalyzeChanges should return 1 change, got %d", len(changes))
	}
}

func TestIsNormalized(t *testing.T) {
	if !IsNormalized("hello", NFC) {
		t.Error("IsNormalized(\"hello\", NFC) should be true")
	}
	
	if !IsNormalized("\u00e9", NFC) {
		t.Error("IsNormalized(é, NFC) should be true")
	}
	
	if IsNormalized("e\u0301", NFC) {
		t.Error("IsNormalized(e+combining, NFC) should be false")
	}
}

func TestFormString(t *testing.T) {
	tests := []struct {
		form     Form
		expected string
	}{
		{NFC, "NFC"},
		{NFD, "NFD"},
		{NFKC, "NFKC"},
		{NFKD, "NFKD"},
		{Form(99), "Unknown"},
	}

	for _, tt := range tests {
		result := tt.form.String()
		if result != tt.expected {
			t.Errorf("Form(%d).String() = %q, want %q", tt.form, result, tt.expected)
		}
	}
}

func TestReorderCombiningMarks(t *testing.T) {
	input := []rune{'e', 0x0327, 0x0301}
	result := reorderCombiningMarks(input)
	
	expected := []rune{'e', 0x0327, 0x0301}
	
	if len(result) != len(expected) {
		t.Errorf("reorderCombiningMarks length = %d, want %d", len(result), len(expected))
		return
	}
	
	for i, r := range result {
		if r != expected[i] {
			t.Errorf("reorderCombiningMarks[%d] = U+%04X, want U+%04X", i, r, expected[i])
		}
	}
}

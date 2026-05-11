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
		{"ñ precomposed", "\u00f1", "n\u0303"},
		{"å precomposed", "\u00e5", "a\u030a"},
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
		{"n + combining tilde", "n\u0303", "\u00f1"},
		{"a + combining ring above", "a\u030a", "\u00e5"},
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
	t.Run("already normalized NFC", func(t *testing.T) {
		changes, normalized := AnalyzeChanges("\u00e9", NFC)
		if normalized != "\u00e9" {
			t.Errorf("AnalyzeChanges normalized = %q, want %q", normalized, "\u00e9")
		}
		if len(changes) != 0 {
			t.Errorf("AnalyzeChanges should return 0 changes for already normalized text, got %d", len(changes))
		}
	})

	t.Run("decomposed to composed NFC", func(t *testing.T) {
		input := "e\u0301"
		changes, normalized := AnalyzeChanges(input, NFC)
		if normalized != "\u00e9" {
			t.Errorf("AnalyzeChanges normalized = %q, want %q", normalized, "\u00e9")
		}
		if len(changes) != 1 {
			t.Errorf("AnalyzeChanges should return 1 change, got %d", len(changes))
		}
		if len(changes) > 0 {
			if changes[0].OriginalStr != input {
				t.Errorf("OriginalStr = %q, want %q", changes[0].OriginalStr, input)
			}
			if changes[0].NormalizedStr != "\u00e9" {
				t.Errorf("NormalizedStr = %q, want %q", changes[0].NormalizedStr, "\u00e9")
			}
		}
	})

	t.Run("fullwidth to halfwidth NFKC", func(t *testing.T) {
		changes, normalized := AnalyzeChanges("\uff21", NFKC)
		if normalized != "A" {
			t.Errorf("AnalyzeChanges normalized = %q, want %q", normalized, "A")
		}
		if len(changes) != 1 {
			t.Errorf("AnalyzeChanges should return 1 change, got %d", len(changes))
		}
	})

	t.Run("multiple changes", func(t *testing.T) {
		input := "e\u0301\uff21"
		changes, _ := AnalyzeChanges(input, NFKC)
		if len(changes) != 2 {
			t.Errorf("AnalyzeChanges should return 2 changes, got %d", len(changes))
		}
	})
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

	if !IsNormalized("e\u0301", NFD) {
		t.Error("IsNormalized(e+combining, NFD) should be true")
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

func TestMoreCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		nfcForm  string
		nfdForm  string
	}{
		{"ñ", "\u00f1", "\u00f1", "n\u0303"},
		{"õ", "\u00f5", "\u00f5", "o\u0303"},
		{"å", "\u00e5", "\u00e5", "a\u030a"},
		{"ø", "\u00f8", "\u00f8", "\u00f8"},
		{"ü", "\u00fc", "\u00fc", "u\u0308"},
		{"ö", "\u00f6", "\u00f6", "o\u0308"},
		{"ä", "\u00e4", "\u00e4", "a\u0308"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_NFC", func(t *testing.T) {
			result := Normalize(tt.input, NFC)
			if result != tt.nfcForm {
				t.Errorf("NFC: got %q, want %q", result, tt.nfcForm)
			}
		})

		t.Run(tt.name+"_NFD", func(t *testing.T) {
			result := Normalize(tt.input, NFD)
			if result != tt.nfdForm {
				t.Errorf("NFD: got %q, want %q", result, tt.nfdForm)
			}
		})
	}
}

func TestExtractCluster(t *testing.T) {
	runes := []rune{'e', 0x0301, 'h', 'e', 'l', 'l', 'o'}
	cluster := extractCluster(runes, 0)
	if len(cluster) != 2 {
		t.Errorf("extractCluster length = %d, want 2", len(cluster))
	}
	if cluster[0] != 'e' || cluster[1] != 0x0301 {
		t.Errorf("extractCluster returned wrong runes")
	}

	cluster2 := extractCluster(runes, 2)
	if len(cluster2) != 1 {
		t.Errorf("extractCluster length = %d, want 1", len(cluster2))
	}
}

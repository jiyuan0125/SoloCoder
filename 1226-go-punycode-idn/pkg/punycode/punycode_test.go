package punycode

import (
	"strings"
	"testing"
)

func stringsEqualIgnoreCase(a, b string) bool {
	return strings.EqualFold(a, b)
}

func TestEncodeBasic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example", "example"},
		{"a", "a"},
		{"test-123", "test-123"},
		{"EXAMPLE", "example"},
		{"Test-123", "test-123"},
	}

	for _, tt := range tests {
		result, err := Encode(tt.input)
		if err != nil {
			t.Errorf("Encode(%q) error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("Encode(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestDecodeBasic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example", "example"},
		{"a", "a"},
		{"test-123", "test-123"},
	}

	for _, tt := range tests {
		result, err := Decode(tt.input)
		if err != nil {
			t.Errorf("Decode(%q) error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("Decode(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestEncodeDecodeRoundtrip(t *testing.T) {
	tests := []string{
		"例子",
		"测试",
		"中文",
		"日本語",
		"한국어",
		"العربية",
		"münchen",
		"über",
	}

	for _, input := range tests {
		encoded, err := Encode(input)
		if err != nil {
			t.Errorf("Encode(%q) error: %v", input, err)
			continue
		}
		t.Logf("Encoded: %q -> %q", input, encoded)

		decoded, err := Decode(encoded)
		if err != nil {
			t.Errorf("Decode(%q) error: %v", encoded, err)
			continue
		}
		if decoded != input {
			t.Errorf("Decode(Encode(%q)) = %q, want %q", input, decoded, input)
		}
	}
}

func TestEncodeDecodeDomain(t *testing.T) {
	domains := []string{
		"例子.测试.com",
		"www.example.com",
		"中文.test",
		"日本語.jp",
		"한국어.kr",
		"例えば.co.jp",
	}

	for _, domain := range domains {
		encoded, err := EncodeDomain(domain)
		if err != nil {
			t.Errorf("EncodeDomain(%q) error: %v", domain, err)
			continue
		}
		t.Logf("Encoded: %q -> %q", domain, encoded)

		decoded, err := DecodeDomain(encoded)
		if err != nil {
			t.Errorf("DecodeDomain(%q) error: %v", encoded, err)
			continue
		}
		if decoded != domain {
			t.Errorf("DecodeDomain(EncodeDomain(%q)) = %q, want %q", domain, decoded, domain)
		}
	}
}

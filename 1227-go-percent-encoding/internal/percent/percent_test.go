package percent

import (
	"testing"
)

func TestEncodeSpace(t *testing.T) {
	tests := []struct {
		input    string
		mode     Mode
		expected string
	}{
		{"hello world", ModeURL, "hello%20world"},
		{"hello world", ModeURI, "hello%20world"},
		{"hello world", ModeForm, "hello+world"},
	}

	for _, tt := range tests {
		result := Encode(tt.input, tt.mode)
		if result != tt.expected {
			t.Errorf("Encode(%q, %s) = %q, expected %q", tt.input, tt.mode, result, tt.expected)
		}
	}
}

func TestEncodeTilde(t *testing.T) {
	tests := []struct {
		input    string
		mode     Mode
		expected string
	}{
		{"home~user", ModeURL, "home~user"},
		{"home~user", ModeForm, "home~user"},
		{"home~user", ModeURI, "home%7Euser"},
	}

	for _, tt := range tests {
		result := Encode(tt.input, tt.mode)
		if result != tt.expected {
			t.Errorf("Encode(%q, %s) = %q, expected %q", tt.input, tt.mode, result, tt.expected)
		}
	}
}

func TestEncodePlus(t *testing.T) {
	tests := []struct {
		input    string
		mode     Mode
		expected string
	}{
		{"a+b", ModeURL, "a%2Bb"},
		{"a+b", ModeURI, "a%2Bb"},
		{"a+b", ModeForm, "a%2Bb"},
	}

	for _, tt := range tests {
		result := Encode(tt.input, tt.mode)
		if result != tt.expected {
			t.Errorf("Encode(%q, %s) = %q, expected %q", tt.input, tt.mode, result, tt.expected)
		}
	}
}

func TestEncodeQuestionAndHash(t *testing.T) {
	tests := []struct {
		input     string
		mode      Mode
		component Component
		expected  string
	}{
		{"a?b#c", ModeURL, ComponentPath, "a%3Fb%23c"},
		{"a?b#c", ModeURL, ComponentQuery, "a?b#c"},
		{"a?b#c", ModeURL, ComponentAll, "a%3Fb%23c"},
	}

	for _, tt := range tests {
		result := Encode(tt.input, tt.mode, EncodeOptions{Component: tt.component})
		if result != tt.expected {
			t.Errorf("Encode(%q, %s, %s) = %q, expected %q", tt.input, tt.mode, tt.component, result, tt.expected)
		}
	}
}

func TestDecodeSpaceForm(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello+world", "hello world"},
		{"hello%20world", "hello world"},
		{"hello+%20world", "hello  world"},
	}

	for _, tt := range tests {
		result, err := Decode(tt.input, ModeForm)
		if err != nil {
			t.Errorf("Decode(%q, Form) error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("Decode(%q, Form) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestDecodeSpaceURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello+world", "hello+world"},
		{"hello%20world", "hello world"},
	}

	for _, tt := range tests {
		result, err := Decode(tt.input, ModeURL)
		if err != nil {
			t.Errorf("Decode(%q, URL) error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("Decode(%q, URL) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestDecodeHexCase(t *testing.T) {
	tests := []struct {
		input    string
		mode     Mode
		expected string
	}{
		{"%2f", ModeURL, "/"},
		{"%2F", ModeURL, "/"},
		{"%2f", ModeForm, "/"},
	}

	for _, tt := range tests {
		result, err := Decode(tt.input, tt.mode)
		if err != nil {
			t.Errorf("Decode(%q, %s) error: %v", tt.input, tt.mode, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("Decode(%q, %s) = %q, expected %q", tt.input, tt.mode, result, tt.expected)
		}
	}
}

func TestEncodeHexUppercase(t *testing.T) {
	result := Encode("/", ModeURL)
	expected := "%2F"
	if result != expected {
		t.Errorf("Encode slash should be uppercase: got %q, expected %q", result, expected)
	}
}

func TestFormDecodeOrder(t *testing.T) {
	input := "%2B"
	expected := "+"
	result, err := Decode(input, ModeForm)
	if err != nil {
		t.Errorf("Decode error: %v", err)
		return
	}
	if result != expected {
		t.Errorf("Decode %q in Form mode = %q, expected %q", input, result, expected)
	}
}

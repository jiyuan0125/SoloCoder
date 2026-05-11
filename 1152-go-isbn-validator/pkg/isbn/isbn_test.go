package isbn

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"978-7-115-54291-3", "9787115542913"},
		{"  978 7 115 54291 3  ", "9787115542913"},
		{"0-306-40615-2", "0306406152"},
		{"030640615X", "030640615X"},
		{"abc123x", "123X"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := CleanInput(tt.input)
			if result != tt.expected {
				t.Errorf("CleanInput(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCalculateCheckDigit10(t *testing.T) {
	tests := []struct {
		first9   string
		expected string
	}{
		{"030640615", "2"},
		{"030640615", "2"},
		{"012345678", "9"},
	}

	for _, tt := range tests {
		t.Run(tt.first9, func(t *testing.T) {
			result, err := CalculateCheckDigit10(tt.first9)
			if err != nil {
				t.Errorf("CalculateCheckDigit10(%q) error = %v", tt.first9, err)
				return
			}
			if result != tt.expected {
				t.Errorf("CalculateCheckDigit10(%q) = %q, want %q", tt.first9, result, tt.expected)
			}
		})
	}
}

func TestCalculateCheckDigit13(t *testing.T) {
	tests := []struct {
		first12  string
		expected string
	}{
		{"978711554291", "5"},
		{"978030640615", "7"},
	}

	for _, tt := range tests {
		t.Run(tt.first12, func(t *testing.T) {
			result, err := CalculateCheckDigit13(tt.first12)
			if err != nil {
				t.Errorf("CalculateCheckDigit13(%q) error = %v", tt.first12, err)
				return
			}
			if result != tt.expected {
				t.Errorf("CalculateCheckDigit13(%q) = %q, want %q", tt.first12, result, tt.expected)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		input    string
		isValid  bool
		isbnType ISBNType
	}{
		{"978-7-115-54291-5", true, ISBN13},
		{"0-306-40615-2", true, ISBN10},
		{"9787115542915", true, ISBN13},
		{"0306406152", true, ISBN10},
		{"9787115542914", false, ISBN13},
		{"0306406153", false, ISBN10},
		{"123456789", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Validate(tt.input)
			if result.IsValid != tt.isValid {
				t.Errorf("Validate(%q).IsValid = %v, want %v", tt.input, result.IsValid, tt.isValid)
			}
			if tt.isValid && result.ISBNType != tt.isbnType {
				t.Errorf("Validate(%q).ISBNType = %v, want %v", tt.input, result.ISBNType, tt.isbnType)
			}
		})
	}
}

func TestConvert10To13(t *testing.T) {
	tests := []struct {
		isbn10   string
		expected string
	}{
		{"0-306-40615-2", "9780306406157"},
		{"0306406152", "9780306406157"},
	}

	for _, tt := range tests {
		t.Run(tt.isbn10, func(t *testing.T) {
			result, err := Convert10To13(tt.isbn10)
			if err != nil {
				t.Errorf("Convert10To13(%q) error = %v", tt.isbn10, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Convert10To13(%q) = %q, want %q", tt.isbn10, result, tt.expected)
			}
		})
	}
}

func TestConvert13To10(t *testing.T) {
	tests := []struct {
		isbn13   string
		expected string
	}{
		{"978-0-306-40615-7", "0306406152"},
		{"9780306406157", "0306406152"},
	}

	for _, tt := range tests {
		t.Run(tt.isbn13, func(t *testing.T) {
			result, err := Convert13To10(tt.isbn13)
			if err != nil {
				t.Errorf("Convert13To10(%q) error = %v", tt.isbn13, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Convert13To10(%q) = %q, want %q", tt.isbn13, result, tt.expected)
			}
		})
	}
}

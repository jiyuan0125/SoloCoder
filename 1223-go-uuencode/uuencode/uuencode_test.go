package uuencode

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		filename string
		mode     int
	}{
		{
			name:     "simple text",
			data:     []byte("Hello, World!"),
			filename: "test.txt",
			mode:     644,
		},
		{
			name:     "binary data",
			data:     []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD, 0x00, 0x00},
			filename: "binary.bin",
			mode:     755,
		},
		{
			name:     "exactly 45 bytes",
			data:     bytes.Repeat([]byte{0x41}, 45),
			filename: "45bytes.txt",
			mode:     644,
		},
		{
			name:     "46 bytes",
			data:     bytes.Repeat([]byte{0x41}, 46),
			filename: "46bytes.txt",
			mode:     644,
		},
		{
			name:     "empty data",
			data:     []byte{},
			filename: "empty.txt",
			mode:     644,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := Encode(tc.data, tc.filename, tc.mode)

			result, err := Decode(encoded)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if !bytes.Equal(result.Data, tc.data) {
				t.Errorf("Data mismatch. Expected %v, got %v", tc.data, result.Data)
			}

			if result.Filename != tc.filename {
				t.Errorf("Filename mismatch. Expected %q, got %q", tc.filename, result.Filename)
			}

			if result.Mode != tc.mode {
				t.Errorf("Mode mismatch. Expected %d, got %d", tc.mode, result.Mode)
			}
		})
	}
}

func TestLengthIndicator(t *testing.T) {
	data := bytes.Repeat([]byte{0x41}, 45)
	encoded := Encode(data, "test.txt", 644)

	lines := []string{}
	for _, line := range splitLines(encoded) {
		if len(line) > 0 && line[0] != 'b' && line[0] != 'e' && line[0] != '`' {
			lines = append(lines, line)
		}
	}

	if len(lines) < 1 {
		t.Fatal("Expected at least one data line")
	}

	firstChar := lines[0][0]
	expected := byte(45 + 32)
	if firstChar != expected {
		t.Errorf("Length indicator for 45 bytes should be 'M' (77), got %c (%d)", firstChar, firstChar)
	}
}

func splitLines(s string) []string {
	var lines []string
	var current []byte
	for _, b := range []byte(s) {
		if b == '\n' {
			lines = append(lines, string(current))
			current = current[:0]
		} else {
			current = append(current, b)
		}
	}
	if len(current) > 0 {
		lines = append(lines, string(current))
	}
	return lines
}

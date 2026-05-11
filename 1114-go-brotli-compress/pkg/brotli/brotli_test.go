package brotli

import (
	"bytes"
	"strings"
	"testing"
)

func TestEmptyInput(t *testing.T) {
	compressed, err := Encode([]byte{}, 6)
	if err != nil {
		t.Fatalf("Encode empty failed: %v", err)
	}

	decompressed, err := Decode(compressed)
	if err != nil {
		t.Fatalf("Decode empty failed: %v", err)
	}

	if len(decompressed) != 0 {
		t.Fatalf("Expected empty output, got %d bytes", len(decompressed))
	}
}

func TestSmallInput(t *testing.T) {
	testCases := []string{
		"a",
		"hello",
		"1234567890123456",
	}

	for _, tc := range testCases {
		compressed, err := Encode([]byte(tc), 6)
		if err != nil {
			t.Fatalf("Encode '%s' failed: %v", tc, err)
		}

		decompressed, err := Decode(compressed)
		if err != nil {
			t.Fatalf("Decode '%s' failed: %v", tc, err)
		}

		if string(decompressed) != tc {
			t.Fatalf("Round-trip failed for '%s'. Expected '%s', got '%s'",
				tc, tc, string(decompressed))
		}
	}
}

func TestLargeInputRoundTrip(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
	}{
		{
			name:  "simple text",
			input: "The quick brown fox jumps over the lazy dog",
		},
		{
			name:  "repeated text",
			input: strings.Repeat("hello world! ", 100),
		},
		{
			name: "json config",
			input: `{
  "name": "test-app",
  "version": "1.0.0",
  "config": {
    "maxConnections": 100,
    "timeout": 30000,
    "debug": true
  }
}`,
		},
		{
			name:  "html fragment",
			input: `<div class="container"><h1>Hello World</h1><p>This is a test paragraph.</p></div>`,
		},
		{
			name:  "long repeated pattern",
			input: strings.Repeat("abcdefghijklmnopqrstuvwxyz", 50),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := []byte(tc.input)
			compressed, err := Encode(input, 6)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			decompressed, err := Decode(compressed)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if !bytes.Equal(input, decompressed) {
				t.Fatalf("Round-trip failed. Input size: %d, output size: %d",
					len(input), len(decompressed))
			}

			t.Logf("Compressed: %d -> %d bytes (%.2f%%)",
				len(input), len(compressed),
				float64(len(compressed))/float64(len(input))*100)
		})
	}
}

func TestQualityLevels(t *testing.T) {
	input := []byte(strings.Repeat("The quick brown fox jumps over the lazy dog. ", 50))

	compressedQ1, err := Encode(input, 1)
	if err != nil {
		t.Fatalf("Encode Q1 failed: %v", err)
	}

	compressedQ11, err := Encode(input, 11)
	if err != nil {
		t.Fatalf("Encode Q11 failed: %v", err)
	}

	t.Logf("Quality 1: %d bytes", len(compressedQ1))
	t.Logf("Quality 11: %d bytes", len(compressedQ11))

	if bytes.Equal(compressedQ1, compressedQ11) {
		t.Fatal("Quality 1 and 11 should produce different results")
	}

	if len(compressedQ11) > len(compressedQ1) {
		t.Logf("Note: Q11 (%d) > Q1 (%d) - Q11 may have more overhead on this input",
			len(compressedQ11), len(compressedQ1))
	}

	decompressed, err := Decode(compressedQ1)
	if err != nil {
		t.Fatalf("Decode Q1 failed: %v", err)
	}
	if !bytes.Equal(input, decompressed) {
		t.Fatal("Q1 round-trip failed")
	}

	decompressed, err = Decode(compressedQ11)
	if err != nil {
		t.Fatalf("Decode Q11 failed: %v", err)
	}
	if !bytes.Equal(input, decompressed) {
		t.Fatal("Q11 round-trip failed")
	}
}

func TestAllQualityLevels(t *testing.T) {
	input := []byte(strings.Repeat("hello world hello hello hello world ", 30))

	for q := 1; q <= 11; q++ {
		compressed, err := Encode(input, q)
		if err != nil {
			t.Fatalf("Encode Q%d failed: %v", q, err)
		}

		decompressed, err := Decode(compressed)
		if err != nil {
			t.Fatalf("Decode Q%d failed: %v", q, err)
		}

		if !bytes.Equal(input, decompressed) {
			t.Fatalf("Quality %d round-trip failed", q)
		}

		t.Logf("Quality %d: %d -> %d bytes (%.2f%%)",
			q, len(input), len(compressed),
			float64(len(compressed))/float64(len(input))*100)
	}
}

func TestBinaryData(t *testing.T) {
	binaryData := make([]byte, 1000)
	for i := range binaryData {
		binaryData[i] = byte(i % 256)
	}

	compressed, err := Encode(binaryData, 6)
	if err != nil {
		t.Fatalf("Encode binary failed: %v", err)
	}

	decompressed, err := Decode(compressed)
	if err != nil {
		t.Fatalf("Decode binary failed: %v", err)
	}

	if !bytes.Equal(binaryData, decompressed) {
		t.Fatal("Binary round-trip failed")
	}
}

func TestQualityDifferences(t *testing.T) {
	t.Run("search depth", func(t *testing.T) {
		if getSearchDepth(1) >= getSearchDepth(11) {
			t.Fatal("Q1 should have smaller search depth than Q11")
		}
		t.Logf("Search depth Q1=%d, Q11=%d", getSearchDepth(1), getSearchDepth(11))
	})

	t.Run("min match", func(t *testing.T) {
		if getMinMatch(1) <= getMinMatch(11) {
			t.Fatal("Q1 should have larger min match than Q11")
		}
		t.Logf("Min match Q1=%d, Q11=%d", getMinMatch(1), getMinMatch(11))
	})

	t.Run("window size", func(t *testing.T) {
		wb1 := getWindowBits(1)
		wb11 := getWindowBits(11)
		t.Logf("Window bits Q1=%d, Q11=%d", wb1, wb11)
		if wb1 > wb11 {
			t.Fatal("Q1 should not have larger window than Q11")
		}
	})
}

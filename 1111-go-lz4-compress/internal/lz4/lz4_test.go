package lz4

import (
	"bytes"
	"strings"
	"testing"
)

func TestEmptyFrameRoundTrip(t *testing.T) {
	empty := []byte{}

	compressed, err := CompressFrame(empty)
	if err != nil {
		t.Fatalf("CompressFrame failed: %v", err)
	}

	if len(compressed) != 11 {
		t.Errorf("Expected 11 bytes for empty frame, got %d", len(compressed))
	}

	decompressed, err := DecompressFrame(compressed)
	if err != nil {
		t.Fatalf("DecompressFrame failed: %v", err)
	}

	if !bytes.Equal(decompressed, empty) {
		t.Errorf("Round-trip failed: expected empty, got %v", decompressed)
	}
}

func TestRepeatedBytesCompression(t *testing.T) {
	original := strings.Repeat("A", 10000)
	originalBytes := []byte(original)

	compressed, err := CompressFrame(originalBytes)
	if err != nil {
		t.Fatalf("CompressFrame failed: %v", err)
	}

	if len(compressed) >= len(originalBytes) {
		t.Errorf("Compression ineffective: %d -> %d bytes (should be smaller)", len(originalBytes), len(compressed))
	}

	decompressed, err := DecompressFrame(compressed)
	if err != nil {
		t.Fatalf("DecompressFrame failed: %v", err)
	}

	if !bytes.Equal(decompressed, originalBytes) {
		t.Errorf("Round-trip failed for repeated bytes")
	}

	t.Logf("Compression ratio: %d -> %d (%.2f%%)", len(originalBytes), len(compressed), 
		float64(len(compressed))/float64(len(originalBytes))*100)
}

func TestOverlapMatch(t *testing.T) {
	original := "AAAAA"
	originalBytes := []byte(original)

	compressed, err := CompressFrame(originalBytes)
	if err != nil {
		t.Fatalf("CompressFrame failed: %v", err)
	}

	decompressed, err := DecompressFrame(compressed)
	if err != nil {
		t.Fatalf("DecompressFrame failed: %v", err)
	}

	if !bytes.Equal(decompressed, originalBytes) {
		t.Errorf("Round-trip failed: expected %q, got %q", original, string(decompressed))
	}
}

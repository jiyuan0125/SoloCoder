package snappy

import (
	"bytes"
	"testing"
)

func TestSingleByte(t *testing.T) {
	data := []byte{0x41}
	compressed := Encode(data)
	decompressed, err := Decode(compressed)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(data, decompressed) {
		t.Errorf("Expected %v, got %v", data, decompressed)
	}
}

func TestAllRepeatedBytes(t *testing.T) {
	data := make([]byte, 100)
	for i := range data {
		data[i] = 'A'
	}
	compressed := Encode(data)
	t.Logf("Original: %d bytes", len(data))
	t.Logf("Compressed: %d bytes, hex: %x", len(compressed), compressed)
	decompressed, err := Decode(compressed)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(data, decompressed) {
		t.Errorf("Mismatch after decompression")
	}
}

func TestEmptyData(t *testing.T) {
	data := []byte{}
	compressed := Encode(data)
	decompressed, err := Decode(compressed)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if len(decompressed) != 0 {
		t.Errorf("Expected empty, got %v", decompressed)
	}
}

func TestShortLiteral(t *testing.T) {
	data := []byte("hello world")
	compressed := Encode(data)
	decompressed, err := Decode(compressed)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(data, decompressed) {
		t.Errorf("Expected %q, got %q", data, decompressed)
	}
}

func TestRepeatedPattern(t *testing.T) {
	data := []byte("abcabcabcabc")
	compressed := Encode(data)
	t.Logf("Original: %d bytes, Compressed: %d bytes", len(data), len(compressed))
	t.Logf("Compressed hex: %x", compressed)
	decompressed, err := Decode(compressed)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if !bytes.Equal(data, decompressed) {
		t.Errorf("Expected %q, got %q", data, decompressed)
	}
}

func TestCopyEncoding(t *testing.T) {
	testCases1Byte := []struct {
		offset int
		length int
	}{
		{1, 4},
		{100, 10},
		{2047, 11},
	}

	for _, tc := range testCases1Byte {
		encoded := appendCopy1Byte(nil, tc.offset, tc.length)
		t.Logf("1-byte: offset=%d, length=%d => encoded: %x", tc.offset, tc.length, encoded)

		tag := encoded[0]
		elementType := tag & 0x03
		if elementType != 0x01 {
			t.Errorf("Wrong element type: got %d, want 1", elementType)
		}
		decodedLen := int((tag>>2)&0x07) + 4
		decodedOff := (int(tag&0xe0) << 3) | int(encoded[1])
		t.Logf("  decoded: offset=%d, length=%d", decodedOff, decodedLen)
		if decodedOff != tc.offset || decodedLen != tc.length {
			t.Errorf("Mismatch: got (%d,%d), want (%d,%d)", decodedOff, decodedLen, tc.offset, tc.length)
		}
	}

	testCases2Byte := []struct {
		offset int
		length int
	}{
		{4, 64},
		{2048, 67},
		{65535, 4},
	}

	for _, tc := range testCases2Byte {
		encoded := appendCopy2Byte(nil, tc.offset, tc.length)
		t.Logf("2-byte: offset=%d, length=%d => encoded: %x", tc.offset, tc.length, encoded)

		tag := encoded[0]
		elementType := tag & 0x03
		if elementType != 0x02 {
			t.Errorf("Wrong element type: got %d, want 2", elementType)
		}
		decodedLen := int((tag>>2)&0x3f) + 4
		decodedOff := int(encoded[1]) | int(encoded[2])<<8
		t.Logf("  decoded: offset=%d, length=%d", decodedOff, decodedLen)
		if decodedOff != tc.offset || decodedLen != tc.length {
			t.Errorf("Mismatch: got (%d,%d), want (%d,%d)", decodedOff, decodedLen, tc.offset, tc.length)
		}
	}
}

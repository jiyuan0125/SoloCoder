package codec

import "testing"

func TestZigZagEncodeDecode(t *testing.T) {
	tests := []struct {
		input  int32
		output uint32
	}{
		{0, 0},
		{-1, 1},
		{1, 2},
		{-2, 3},
		{2, 4},
		{-63, 125},
		{63, 126},
	}

	for _, tt := range tests {
		encoded := ZigZagEncodeInt32(tt.input)
		if encoded != tt.output {
			t.Errorf("ZigZagEncodeInt32(%d) = %d, want %d", tt.input, encoded, tt.output)
		}

		decoded := ZigZagDecodeUint32(tt.output)
		if decoded != tt.input {
			t.Errorf("ZigZagDecodeUint32(%d) = %d, want %d", tt.output, decoded, tt.input)
		}
	}
}

func TestVarintEncodeDecodeUint32(t *testing.T) {
	tests := []uint32{0, 1, 127, 128, 255, 256, 16384, 2097151, 2097152, 268435455, 268435456}

	for _, tt := range tests {
		encoded := EncodeVarintUint32(tt)
		decoded, n, err := DecodeVarintUint32(encoded)
		if err != nil {
			t.Errorf("DecodeVarintUint32 failed for %d: %v", tt, err)
			continue
		}
		if decoded != tt {
			t.Errorf("Varint round-trip failed: %d -> %d", tt, decoded)
		}
		if n != len(encoded) {
			t.Errorf("Consumed bytes mismatch: %d vs %d", n, len(encoded))
		}
	}
}

func TestLEB128EncodeDecodeUint32(t *testing.T) {
	tests := []uint32{0, 1, 127, 128, 255, 256, 16384, 2097151, 2097152, 268435455, 268435456}

	for _, tt := range tests {
		encoded := EncodeLEB128Uint32(tt)
		decoded, n, err := DecodeLEB128Uint32(encoded)
		if err != nil {
			t.Errorf("DecodeLEB128Uint32 failed for %d: %v", tt, err)
			continue
		}
		if decoded != tt {
			t.Errorf("LEB128 round-trip failed: %d -> %d", tt, decoded)
		}
		if n != len(encoded) {
			t.Errorf("Consumed bytes mismatch: %d vs %d", n, len(encoded))
		}
	}
}

func TestVarintInt32WithZigZag(t *testing.T) {
	tests := []int32{0, -1, 1, -2, 2, -100, 100, -1000, 1000}

	for _, tt := range tests {
		encoded := EncodeVarintInt32(tt)
		decoded, _, err := DecodeVarintInt32(encoded)
		if err != nil {
			t.Errorf("DecodeVarintInt32 failed for %d: %v", tt, err)
			continue
		}
		if decoded != tt {
			t.Errorf("Varint+ZigZag round-trip failed: %d -> %d", tt, decoded)
		}
	}
}

func TestOverflowDetection(t *testing.T) {
	overflowData := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x01}
	_, _, err := DecodeVarintUint64(overflowData)
	if err != ErrOverflow {
		t.Errorf("Expected overflow error")
	}

	truncatedData := []byte{0xFF, 0xFF}
	_, _, err = DecodeVarintUint64(truncatedData)
	if err != ErrTruncated {
		t.Errorf("Expected truncated error")
	}
}

func TestBatchEncodeDecode(t *testing.T) {
	values := []int64{0, -1, 1, -100, 100, -10000, 10000, 123456789}

	encoded := EncodeBatchInt64(values, ModeVarint)
	decoded, err := DecodeBatchInt64(encoded, ModeVarint)
	if err != nil {
		t.Fatalf("Batch decode failed: %v", err)
	}
	if len(decoded) != len(values) {
		t.Fatalf("Length mismatch: %d vs %d", len(decoded), len(values))
	}
	for i, v := range values {
		if decoded[i] != v {
			t.Errorf("Batch round-trip failed at %d: %d -> %d", i, v, decoded[i])
		}
	}
}

func TestValidation(t *testing.T) {
	validData := EncodeBatchInt64([]int64{1, 2, 3}, ModeVarint)
	valid, _ := ValidateVarint(validData, TypeInt64)
	if !valid {
		t.Error("Expected valid encoding")
	}

	invalidData := []byte{0xFF, 0xFF}
	valid, _ = ValidateVarint(invalidData, TypeInt64)
	if valid {
		t.Error("Expected invalid encoding for truncated data")
	}
}

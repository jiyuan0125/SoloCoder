package ascii85

import (
	"bytes"
	"testing"
)

func TestZCompression(t *testing.T) {
	input := []byte{0, 0, 0, 0}
	result, err := Encode(input, ModeAdobe)
	if err != nil {
		t.Errorf("Encode failed: %v", err)
	}
	if result != "<~z~>" {
		t.Errorf("Encode zeros = %q, want <~z~>", result)
	}

	decoded, _, err := Decode("<~z~>")
	if err != nil {
		t.Errorf("Decode z failed: %v", err)
	}
	if !bytes.Equal(decoded, input) {
		t.Errorf("Decode z = %v, want %v", decoded, input)
	}
}

func TestRoundTrip(t *testing.T) {
	testCases := [][]byte{
		[]byte(""),
		[]byte("a"),
		[]byte("ab"),
		[]byte("abc"),
		[]byte("abcd"),
		[]byte("abcde"),
		[]byte("Hello World"),
		[]byte{0, 0, 0, 0, 1, 2, 3, 4},
		[]byte{0x00, 0x01, 0x02, 0x03},
	}

	for _, data := range testCases {
		encoded, err := Encode(data, ModeAdobe)
		if err != nil {
			t.Errorf("Encode failed: %v", err)
			continue
		}

		decoded, mode, err := Decode(encoded)
		if err != nil {
			t.Errorf("Decode failed: %v", err)
			continue
		}
		if mode != ModeAdobe {
			t.Errorf("Expected Adobe mode, got %s", mode)
		}
		if !bytes.Equal(decoded, data) {
			t.Errorf("Round trip failed: %v -> %q -> %v", data, encoded, decoded)
		}
	}
}

func TestBtoaMode(t *testing.T) {
	data := []byte("Hello")

	encoded, err := Encode(data, ModeBtoa)
	if err != nil {
		t.Errorf("Btoa encode failed: %v", err)
	}

	decoded, mode, err := Decode(encoded)
	if err != nil {
		t.Errorf("Btoa decode failed: %v", err)
	}

	if mode != ModeBtoa {
		t.Errorf("Expected Btoa mode, got %s", mode)
	}

	if !bytes.Equal(decoded, data) {
		t.Errorf("Btoa round trip failed: %v -> %q -> %v", data, encoded, decoded)
	}
}

func TestWhitespaceInAdobe(t *testing.T) {
	data := []byte("Hello World")
	encoded, _ := Encode(data, ModeAdobe)

	withWhitespace := encoded[:2] + "\n\t  " + encoded[2:len(encoded)-2] + "  \n" + encoded[len(encoded)-2:]
	decoded, mode, err := Decode(withWhitespace)
	if err != nil {
		t.Errorf("Decode with whitespace failed: %v", err)
	}

	if mode != ModeAdobe {
		t.Errorf("Expected Adobe mode")
	}

	if !bytes.Equal(decoded, data) {
		t.Errorf("Whitespace handling failed")
	}
}

func TestAutoDetectMode(t *testing.T) {
	adobeData := "<~87cURD]i~>"
	_, mode, err := Decode(adobeData)
	if err != nil {
		t.Errorf("Adobe decode failed: %v", err)
	}
	if mode != ModeAdobe {
		t.Errorf("Expected Adobe mode, got %s", mode)
	}

	btoaData := `xbtoa Begin
87cURD]i
xbtoa End`
	_, mode, err = Decode(btoaData)
	if err != nil {
		t.Errorf("Btoa decode failed: %v", err)
	}
	if mode != ModeBtoa {
		t.Errorf("Expected Btoa mode, got %s", mode)
	}
}

func TestEmptyData(t *testing.T) {
	encoded, err := Encode([]byte{}, ModeAdobe)
	if err != nil {
		t.Errorf("Empty encode failed: %v", err)
	}
	if encoded != "<~~>" {
		t.Errorf("Empty encode = %q, want <~~>", encoded)
	}

	encoded, err = Encode([]byte{}, ModeBtoa)
	if err != nil {
		t.Errorf("Empty btoa encode failed: %v", err)
	}
}

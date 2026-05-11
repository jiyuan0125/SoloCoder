package frameparser

import (
	"testing"
)

func buildLiteralHeader(name, value string, incremental bool) []byte {
	result := []byte{}

	if incremental {
		result = append(result, 0x40)
	} else {
		result = append(result, 0x00)
	}

	result = append(result, byte(len(name)))
	result = append(result, []byte(name)...)

	result = append(result, byte(len(value)))
	result = append(result, []byte(value)...)

	return result
}

func TestDecodeLiteralWithIncremental(t *testing.T) {
	decoder := NewHPACKDecoder()

	testData := buildLiteralHeader("custom-header", "value1", true)

	headers, err := decoder.Decode(testData)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(headers) != 1 {
		t.Fatalf("Expected 1 header, got %d", len(headers))
	}

	if headers[0].Name != "custom-header" {
		t.Errorf("Expected name 'custom-header', got '%s'", headers[0].Name)
	}

	if headers[0].Value != "value1" {
		t.Errorf("Expected value 'value1', got '%s'", headers[0].Value)
	}
}

func TestDecodeMultipleLiteralHeaders(t *testing.T) {
	decoder := NewHPACKDecoder()

	testData := []byte{}
	testData = append(testData, buildLiteralHeader("host", "example.com", true)...)
	testData = append(testData, buildLiteralHeader("path", "/some/path.html", true)...)

	headers, err := decoder.Decode(testData)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(headers) != 2 {
		t.Fatalf("Expected 2 headers, got %d", len(headers))
	}

	if headers[0].Name != "host" || headers[0].Value != "example.com" {
		t.Errorf("First header mismatch: %s: %s", headers[0].Name, headers[0].Value)
	}

	if headers[1].Name != "path" || headers[1].Value != "/some/path.html" {
		t.Errorf("Second header mismatch: %s: %s", headers[1].Name, headers[1].Value)
	}
}

func TestDecodeIndexedAndLiteral(t *testing.T) {
	decoder := NewHPACKDecoder()

	testData := []byte{0x82}
	testData = append(testData, buildLiteralHeader("x-custom-hdr", "test1", true)...)

	headers, err := decoder.Decode(testData)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(headers) != 2 {
		t.Fatalf("Expected 2 headers, got %d", len(headers))
	}

	if headers[0].Name != ":method" || headers[0].Value != "GET" {
		t.Errorf("First header (indexed) mismatch: %s: %s", headers[0].Name, headers[0].Value)
	}

	if headers[1].Name != "x-custom-hdr" || headers[1].Value != "test1" {
		t.Errorf("Second header (literal) mismatch: %s: %s", headers[1].Name, headers[1].Value)
	}
}

func TestDecodeNeverIndexed(t *testing.T) {
	decoder := NewHPACKDecoder()

	testData := buildLiteralHeader("authorization", "secret-token", false)

	headers, err := decoder.Decode(testData)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(headers) != 1 {
		t.Fatalf("Expected 1 header, got %d", len(headers))
	}

	if headers[0].Name != "authorization" {
		t.Errorf("Expected name 'authorization', got '%s'", headers[0].Name)
	}

	if headers[0].Value != "secret-token" {
		t.Errorf("Expected value 'secret-token', got '%s'", headers[0].Value)
	}
}

package asn1

import (
	"encoding/hex"
	"testing"
)

func TestEncodeINTEGER42(t *testing.T) {
	jt := JSONTLV{
		Type:  "INTEGER",
		Value: float64(42),
	}

	result, err := EncodeDER(jt)
	if err != nil {
		t.Fatalf("EncodeDER failed: %v", err)
	}

	expected := "02012a"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestEncodeUTF8StringHello(t *testing.T) {
	jt := JSONTLV{
		Type:  "UTF8String",
		Value: "Hello",
	}

	result, err := EncodeDER(jt)
	if err != nil {
		t.Fatalf("EncodeDER failed: %v", err)
	}

	expected := "0c0548656c6c6f"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestEncodeSEQUENCE(t *testing.T) {
	jt := JSONTLV{
		Type: "SEQUENCE",
		Children: []JSONTLV{
			{
				Type:  "INTEGER",
				Value: float64(123),
			},
			{
				Type:  "UTF8String",
				Value: "test",
			},
		},
	}

	result, err := EncodeDER(jt)
	if err != nil {
		t.Fatalf("EncodeDER failed: %v", err)
	}

	// SEQUENCE tag=0x30, length should contain INTEGER + UTF8String
	// INTEGER 123 = 02 01 7B
	// UTF8String "test" = 0C 04 74 65 73 74
	// Total children length = 3 + 6 = 9
	// SEQUENCE = 30 09 + children
	expected := "300902017b0c0474657374"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestFromJSONBasic(t *testing.T) {
	jt := JSONTLV{
		Type:  "INTEGER",
		Value: float64(42),
	}

	tlv, err := FromJSON(jt)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if tlv.Class != TagClassUniversal {
		t.Errorf("Expected Class=Universal, got %d", tlv.Class)
	}
	if tlv.Number != TagInteger {
		t.Errorf("Expected Number=%d, got %d", TagInteger, tlv.Number)
	}
	if tlv.Constructed {
		t.Error("Expected Constructed=false")
	}

	// 42 = 0x2A
	if len(tlv.Value) != 1 || tlv.Value[0] != 0x2A {
		t.Errorf("Expected Value=[0x2A], got %v", tlv.Value)
	}
}

func TestEncodeBasicTLV(t *testing.T) {
	// Direct test of Encode function with a manually created TLV
	tlv := NewTLV(TagClassUniversal, TagInteger, false)
	tlv.Value = []byte{0x2A} // 42

	bytes, err := Encode(tlv, DefaultEncodeOptions(ModeDER))
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	result := hex.EncodeToString(bytes)
	expected := "02012a"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestEncodeBERIndefinite(t *testing.T) {
	jt := JSONTLV{
		Type: "SEQUENCE",
		Children: []JSONTLV{
			{
				Type:  "INTEGER",
				Value: float64(42),
			},
		},
	}

	result, err := EncodeBER(jt)
	if err != nil {
		t.Fatalf("EncodeBER failed: %v", err)
	}

	// SEQUENCE with indefinite length:
	// Tag = 0x30, Length = 0x80 (indefinite), Value = INTEGER, EOC = 00 00
	// 30 80 02 01 2a 00 00
	expected := "308002012a0000"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestEncodeNull(t *testing.T) {
	jt := JSONTLV{
		Type: "NULL",
	}

	result, err := EncodeDER(jt)
	if err != nil {
		t.Fatalf("EncodeDER failed: %v", err)
	}

	// NULL = 05 00
	expected := "0500"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestEncodeBoolean(t *testing.T) {
	jt := JSONTLV{
		Type:  "BOOLEAN",
		Value: true,
	}

	result, err := EncodeDER(jt)
	if err != nil {
		t.Fatalf("EncodeDER failed: %v", err)
	}

	// BOOLEAN TRUE = 01 01 FF
	expected := "0101ff"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

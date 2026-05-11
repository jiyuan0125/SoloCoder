package asn1

import (
	"encoding/hex"
	"testing"
)

func TestParseTagType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantNum     TagNumber
		wantClass   TagClass
		wantConstructed bool
		wantErr     bool
	}{
		{"INTEGER", "INTEGER", TagInteger, TagClassUniversal, false, false},
		{"integer lowercase", "integer", TagInteger, TagClassUniversal, false, false},
		{"Integer mixed", "Integer", TagInteger, TagClassUniversal, false, false},
		{"UTF8String", "UTF8String", TagUTF8String, TagClassUniversal, false, false},
		{"utf8string lowercase", "utf8string", TagUTF8String, TagClassUniversal, false, false},
		{"SEQUENCE", "SEQUENCE", TagSequence, TagClassUniversal, true, false},
		{"sequence lowercase", "sequence", TagSequence, TagClassUniversal, true, false},
		{"NULL", "NULL", TagNull, TagClassUniversal, false, false},
		{"BOOLEAN", "BOOLEAN", TagBoolean, TagClassUniversal, false, false},
		{"context specific", "[0]", TagNumber(0), TagClassContextSpecific, true, false},
		{"context specific 1", "[1]", TagNumber(1), TagClassContextSpecific, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num, class, constructed, err := parseTagType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTagType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if num != tt.wantNum {
				t.Errorf("parseTagType() num = %v, want %v", num, tt.wantNum)
			}
			if class != tt.wantClass {
				t.Errorf("parseTagType() class = %v, want %v", class, tt.wantClass)
			}
			if constructed != tt.wantConstructed {
				t.Errorf("parseTagType() constructed = %v, want %v", constructed, tt.wantConstructed)
			}
		})
	}
}

func TestTagBytes(t *testing.T) {
	tests := []struct {
		name     string
		tlv      *TLV
		expected string
	}{
		{"INTEGER", NewTLV(TagClassUniversal, TagInteger, false), "02"},
		{"UTF8String", NewTLV(TagClassUniversal, TagUTF8String, false), "0c"},
		{"SEQUENCE", NewTLV(TagClassUniversal, TagSequence, true), "30"},
		{"SET", NewTLV(TagClassUniversal, TagSet, true), "31"},
		{"NULL", NewTLV(TagClassUniversal, TagNull, false), "05"},
		{"Context [0] primitive", NewTLV(TagClassContextSpecific, 0, false), "80"},
		{"Context [0] constructed", NewTLV(TagClassContextSpecific, 0, true), "a0"},
		{"Context [1] primitive", NewTLV(TagClassContextSpecific, 1, false), "81"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes := tt.tlv.TagBytes()
			result := hex.EncodeToString(bytes)
			if result != tt.expected {
				t.Errorf("TagBytes() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestEncodeLength(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		definite bool
		expected string
	}{
		{"short 0", 0, true, "00"},
		{"short 1", 1, true, "01"},
		{"short 127", 127, true, "7f"},
		{"long 128", 128, true, "8180"},
		{"long 255", 255, true, "81ff"},
		{"long 256", 256, true, "820100"},
		{"indefinite", 0, false, "80"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes := EncodeLength(tt.length, tt.definite)
			result := hex.EncodeToString(bytes)
			if result != tt.expected {
				t.Errorf("EncodeLength() = %s, want %s", result, tt.expected)
			}
		})
	}
}

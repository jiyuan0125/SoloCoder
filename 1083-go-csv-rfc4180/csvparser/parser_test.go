package csvparser

import (
	"reflect"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	testCases := []struct {
		name string
		data [][]string
	}{
		{"single empty record", [][]string{{""}}},
		{"single record single field", [][]string{{"hello"}}},
		{"two records", [][]string{{"a"}, {"b"}}},
		{"record with comma", [][]string{{"a,b"}}},
		{"record with quote", [][]string{{`He said "hello"`}}},
		{"record with newline", [][]string{{"line1\nline2"}}},
		{"mixed", [][]string{{"name", "desc"}, {"Alice", `He said "hi"`}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serialized := Serialize(tc.data)
			result, err := Parse(serialized)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			if !reflect.DeepEqual(result.Records, tc.data) {
				t.Errorf("Round-trip failed.\nInput:    %v\nSerialized: %q\nOutput:   %v", tc.data, serialized, result.Records)
			}
		})
	}
}

func TestSerializeDistinguishEmpty(t *testing.T) {
	empty := Serialize([][]string{})
	singleEmptyRecord := Serialize([][]string{{""}})

	if empty == singleEmptyRecord {
		t.Errorf("Serialize([]) and Serialize([[]]) both return %q, cannot distinguish", empty)
	}

	t.Logf("Serialize([]) = %q", empty)
	t.Logf("Serialize([[]]) = %q", singleEmptyRecord)
}

func TestParseEmptyString(t *testing.T) {
	result, err := Parse("")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	expected := [][]string{{""}}
	if !reflect.DeepEqual(result.Records, expected) {
		t.Errorf("Parse(\"\") should return %v, got %v", expected, result.Records)
	}
}

func TestParseTrailingCRLF(t *testing.T) {
	result, err := Parse("hello\r\n")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	expected := [][]string{{"hello"}}
	if !reflect.DeepEqual(result.Records, expected) {
		t.Errorf("Expected %v, got %v", expected, result.Records)
	}
}

func TestParseOnlyCRLF(t *testing.T) {
	result, err := Parse("\r\n")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	expected := [][]string{{""}}
	if !reflect.DeepEqual(result.Records, expected) {
		t.Errorf("Expected %v, got %v", expected, result.Records)
	}
}

func TestSerializeMultipleEmpty(t *testing.T) {
	data := [][]string{{""}, {""}, {""}}
	serialized := Serialize(data)
	expected := "\r\n\r\n\r\n"
	if serialized != expected {
		t.Errorf("Expected %q, got %q", expected, serialized)
	}

	result, err := Parse(serialized)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if !reflect.DeepEqual(result.Records, data) {
		t.Errorf("Expected %v, got %v", data, result.Records)
	}
}

package csvparser

import (
	"reflect"
	"testing"
)

func TestR2Issue1_EmptyStringIsOneRecord(t *testing.T) {
	result, err := Parse("")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	expected := [][]string{{""}}
	if !reflect.DeepEqual(result.Records, expected) {
		t.Errorf("Parse(\"\") should return %v, got %v", expected, result.Records)
	}
}

func TestR2Issue2_ValidateTrailingCRLF(t *testing.T) {
	csv := "name,age\r\nAlice,30\r\n"
	errs := Validate(csv, ParserOption{WithHeader: true})
	if len(errs) != 0 {
		for _, e := range errs {
			t.Errorf("Unexpected validation error: %s", e.String())
		}
	}
}

func TestR2Issue3_UnclosedQuoteShouldError(t *testing.T) {
	_, err := Parse("\"")
	if err == nil {
		t.Error("Parse(\"\\\"\") should return error for unclosed quote")
	}
}

func TestR2Issue3_UnclosedQuoteInField(t *testing.T) {
	_, err := Parse("\"hello")
	if err == nil {
		t.Error("Parse(\"\\\"hello\") should return error for unclosed quote")
	}
}

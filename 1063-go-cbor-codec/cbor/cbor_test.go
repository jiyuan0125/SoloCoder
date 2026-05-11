package cbor

import (
	"encoding/hex"
	"math"
	"math/big"
	"reflect"
	"testing"
)

func TestIntegerEncoding(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  string
	}{
		{"uint 0", uint64(0), "00"},
		{"uint 1", uint64(1), "01"},
		{"uint 10", uint64(10), "0a"},
		{"uint 23", uint64(23), "17"},
		{"uint 24", uint64(24), "1818"},
		{"uint 255", uint64(255), "18ff"},
		{"uint 256", uint64(256), "190100"},
		{"uint 65535", uint64(65535), "19ffff"},
		{"uint 65536", uint64(65536), "1a00010000"},
		{"int -1", int64(-1), "20"},
		{"int -10", int64(-10), "29"},
		{"int 0", int64(0), "00"},
		{"int 42", int64(42), "182a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if hex.EncodeToString(got) != tt.want {
				t.Errorf("Marshal() = %v, want %v", hex.EncodeToString(got), tt.want)
			}
		})
	}
}

func TestIntegerDecoding(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Value
	}{
		{"uint 0", "00", uint64(0)},
		{"uint 1", "01", uint64(1)},
		{"uint 10", "0a", uint64(10)},
		{"uint 23", "17", uint64(23)},
		{"uint 24", "1818", uint64(24)},
		{"uint 255", "18ff", uint64(255)},
		{"uint 256", "190100", uint64(256)},
		{"int -1", "20", int64(-1)},
		{"int -10", "29", int64(-10)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := hex.DecodeString(tt.input)
			got, err := Unmarshal(data)
			if err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestStrings(t *testing.T) {
	tests := []struct {
		name  string
		input Value
		want  string
	}{
		{"empty string", "", "60"},
		{"a", "a", "6161"},
		{"IETF", "IETF", "6449455446"},
		{"hello", "hello", "6568656c6c6f"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if hex.EncodeToString(got) != tt.want {
				t.Errorf("Marshal() = %v, want %v", hex.EncodeToString(got), tt.want)
			}
		})
	}
}

func TestByteString(t *testing.T) {
	input := []byte{0x01, 0x02, 0x03, 0x04}
	encoded, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	want := "4401020304"
	if hex.EncodeToString(encoded) != want {
		t.Errorf("Marshal() = %v, want %v", hex.EncodeToString(encoded), want)
	}
}

func TestBooleans(t *testing.T) {
	tests := []struct {
		name  string
		input bool
		want  string
	}{
		{"false", false, "f4"},
		{"true", true, "f5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if hex.EncodeToString(got) != tt.want {
				t.Errorf("Marshal() = %v, want %v", hex.EncodeToString(got), tt.want)
			}
		})
	}
}

func TestNull(t *testing.T) {
	got, err := Marshal(nil)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	want := "f6"
	if hex.EncodeToString(got) != want {
		t.Errorf("Marshal() = %v, want %v", hex.EncodeToString(got), want)
	}
}

func TestSimpleValue(t *testing.T) {
	tests := []struct {
		name  string
		input SimpleValue
		want  string
	}{
		{"simple 0", SimpleValue(0), "e0"},
		{"simple 16", SimpleValue(16), "f0"},
		{"simple 19", SimpleValue(19), "f3"},
		{"simple 32", SimpleValue(32), "f820"},
		{"simple 255", SimpleValue(255), "f8ff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if hex.EncodeToString(got) != tt.want {
				t.Errorf("Marshal() = %v, want %v", hex.EncodeToString(got), tt.want)
			}
		})
	}
}

func TestFloat64(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{"1.5", 1.5, "fb3ff8000000000000"},
		{"0.0", 0.0, "fb0000000000000000"},
		{"-4.1", -4.1, "fbc010666666666666"},
		{"Infinity", math.Inf(1), "fb7ff0000000000000"},
		{"-Infinity", math.Inf(-1), "fbfff0000000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if hex.EncodeToString(got) != tt.want {
				t.Errorf("Marshal() = %v, want %v", hex.EncodeToString(got), tt.want)
			}
		})
	}
}

func TestFloatDecoding(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{"half 1.5", "f93e00", 1.5},
		{"half 0.0", "f90000", 0.0},
		{"half -0.0", "f98000", math.Copysign(0.0, -1)},
		{"half 65504", "f97bff", 65504.0},
		{"half Infinity", "f97c00", math.Inf(1)},
		{"single 1.5", "fa3fc00000", 1.5},
		{"double 1.5", "fb3ff8000000000000", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := hex.DecodeString(tt.input)
			got, err := Unmarshal(data)
			if err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			f, ok := got.(float64)
			if !ok {
				t.Fatalf("expected float64, got %T", got)
			}
			if math.IsNaN(tt.want) {
				if !math.IsNaN(f) {
					t.Errorf("expected NaN, got %v", f)
				}
			} else if math.Float64bits(f) != math.Float64bits(tt.want) {
				t.Errorf("Unmarshal() = %v, want %v", f, tt.want)
			}
		})
	}
}

func TestArray(t *testing.T) {
	input := []Value{uint64(1), uint64(2), uint64(3)}
	encoded, err := Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	want := "83010203"
	if hex.EncodeToString(encoded) != want {
		t.Errorf("Marshal() = %v, want %v", hex.EncodeToString(encoded), want)
	}
}

func TestMap(t *testing.T) {
	m := NewMap()
	m.Add("a", uint64(1))
	m.Add("b", uint64(2))

	encoded, err := Marshal(m)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	want := "a2616101616202"
	if hex.EncodeToString(encoded) != want {
		t.Errorf("Marshal() = %v, want %v", hex.EncodeToString(encoded), want)
	}
}

func TestBigInt(t *testing.T) {
	bigInt := big.NewInt(0)
	bigInt.SetString("18446744073709551617", 10)
	bi := NewBigInt(false, bigInt)

	encoded, err := Marshal(bi)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	want := "c249010000000000000001"
	if hex.EncodeToString(encoded) != want {
		t.Errorf("Marshal() = %v, want %v", hex.EncodeToString(encoded), want)
	}

	decoded, err := Unmarshal(encoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	decodedBI, ok := decoded.(*BigInt)
	if !ok {
		t.Fatalf("expected *BigInt, got %T", decoded)
	}
	if decodedBI.Negative {
		t.Errorf("expected non-negative")
	}
	if decodedBI.Int.Cmp(bigInt) != 0 {
		t.Errorf("got %s, want %s", decodedBI.Int.String(), bigInt.String())
	}
}

func TestIndefiniteLength(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Value
	}{
		{"indef array", "9f010203ff", []Value{uint64(1), uint64(2), uint64(3)}},
		{"indef string", "7f657374726561646d696e67ff", "streaming"},
		{"indef byte string", "5f42010243030405ff", []byte{0x01, 0x02, 0x03, 0x04, 0x05}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := hex.DecodeString(tt.input)
			got, err := Unmarshal(data)
			if err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONRoundTrip(t *testing.T) {
	m := NewMap()
	m.Add("name", "test")
	m.Add("value", int64(42))
	m.Add("active", true)
	m.Add("data", []byte{0x01, 0x02, 0x03})

	jsonData, err := ToJSON(m)
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	value, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	cborData, err := Marshal(value)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if len(cborData) == 0 {
		t.Errorf("expected non-empty CBOR data")
	}
}

func TestMapKeyOrderPreserved(t *testing.T) {
	m := NewMap()
	m.Add("z", uint64(1))
	m.Add("a", uint64(2))
	m.Add("m", uint64(3))

	jsonData, err := ToJSON(m)
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	expected := `{"z":1,"a":2,"m":3}`
	if string(jsonData) != expected {
		t.Errorf("Map key order not preserved.\nGot:      %s\nExpected: %s", string(jsonData), expected)
	}
}

func TestNestedMapKeyOrder(t *testing.T) {
	inner := NewMap()
	inner.Add("bbb", uint64(2))
	inner.Add("aaa", uint64(1))

	outer := NewMap()
	outer.Add("first", "value1")
	outer.Add("inner", inner)
	outer.Add("last", uint64(99))

	jsonData, err := ToJSON(outer)
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	expected := `{"first":"value1","inner":{"bbb":2,"aaa":1},"last":99}`
	if string(jsonData) != expected {
		t.Errorf("Nested map key order not preserved.\nGot:      %s\nExpected: %s", string(jsonData), expected)
	}
}

func TestMapWithIntegerKeys(t *testing.T) {
	m := NewMap()
	m.Add(int64(3), "three")
	m.Add(int64(1), "one")
	m.Add(int64(2), "two")

	jsonData, err := ToJSON(m)
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	expected := `{"3":"three","1":"one","2":"two"}`
	if string(jsonData) != expected {
		t.Errorf("Map with integer keys not ordered correctly.\nGot:      %s\nExpected: %s", string(jsonData), expected)
	}
}

func TestEmptyMapAndArray(t *testing.T) {
	emptyMap := NewMap()
	jsonMap, err := ToJSON(emptyMap)
	if err != nil {
		t.Fatalf("ToJSON(empty map) error = %v", err)
	}
	if string(jsonMap) != "{}" {
		t.Errorf("Empty map should be {}, got: %s", string(jsonMap))
	}

	emptyArr := []Value{}
	jsonArr, err := ToJSON(emptyArr)
	if err != nil {
		t.Fatalf("ToJSON(empty array) error = %v", err)
	}
	if string(jsonArr) != "[]" {
		t.Errorf("Empty array should be [], got: %s", string(jsonArr))
	}
}

package delta

import (
	"math"
	"reflect"
	"testing"
)

func TestEmptyArray(t *testing.T) {
	intEncoded := EncodeInt64([]int64{})
	if intEncoded.Base != 0 || len(intEncoded.Deltas) != 0 {
		t.Errorf("Int empty encode failed")
	}
	decoded := DecodeInt(intEncoded)
	if len(decoded) != 0 {
		t.Errorf("Int empty decode failed")
	}

	floatEncoded := EncodeFloat64([]float64{})
	if floatEncoded.Base != 0.0 || len(floatEncoded.Deltas) != 0 {
		t.Errorf("Float empty encode failed")
	}
	floatDecoded := DecodeFloat(floatEncoded)
	if len(floatDecoded) != 0 {
		t.Errorf("Float empty decode failed")
	}
}

func TestSingleElement(t *testing.T) {
	intEncoded := EncodeInt64([]int64{42})
	if intEncoded.Base != 42 || len(intEncoded.Deltas) != 0 {
		t.Errorf("Int single encode failed")
	}
	decoded := DecodeInt(intEncoded)
	if !reflect.DeepEqual(decoded, []int64{42}) {
		t.Errorf("Int single decode failed")
	}

	floatEncoded := EncodeFloat64([]float64{3.14})
	if floatEncoded.Base != 3.14 || len(floatEncoded.Deltas) != 0 {
		t.Errorf("Float single encode failed")
	}
	floatDecoded := DecodeFloat(floatEncoded)
	if len(floatDecoded) != 1 || math.Abs(floatDecoded[0]-3.14) > 1e-10 {
		t.Errorf("Float single decode failed")
	}
}

func TestNormalIntArray(t *testing.T) {
	original := []int64{10, 12, 15, 14, 18, 18, 20}
	encoded := EncodeInt64(original)

	if encoded.Base != 10 {
		t.Errorf("Base should be 10, got %d", encoded.Base)
	}

	expectedDeltas := []int64{2, 3, -1, 4, 0, 2}
	if !reflect.DeepEqual(encoded.Deltas, expectedDeltas) {
		t.Errorf("Deltas expected %v, got %v", expectedDeltas, encoded.Deltas)
	}

	decoded := DecodeInt(encoded)
	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("Decode failed")
	}
}

func TestNormalFloatArray(t *testing.T) {
	original := []float64{1.5, 2.5, 4.0, 3.5, 5.5}
	encoded := EncodeFloat64(original)

	if math.Abs(encoded.Base-1.5) > 1e-10 {
		t.Errorf("Base should be 1.5, got %f", encoded.Base)
	}

	expectedDeltas := []float64{1.0, 1.5, -0.5, 2.0}
	for i, d := range encoded.Deltas {
		if math.Abs(d-expectedDeltas[i]) > 1e-10 {
			t.Errorf("Delta mismatch at %d", i)
		}
	}

	decoded := DecodeFloat(encoded)
	for i, v := range decoded {
		if math.Abs(v-original[i]) > 1e-10 {
			t.Errorf("Decode mismatch at %d", i)
		}
	}
}

func TestIntOverflow(t *testing.T) {
	original := []int64{math.MaxInt32, math.MaxInt32 + 1000, math.MinInt32}
	encoded := EncodeInt64(original)

	if encoded.Base != math.MaxInt32 {
		t.Errorf("Base should be %d, got %d", math.MaxInt32, encoded.Base)
	}

	if encoded.Deltas[0] != 1000 {
		t.Errorf("First delta should be 1000, got %d", encoded.Deltas[0])
	}

	decoded := DecodeInt(encoded)
	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("Overflow decode failed")
	}
}

func TestFloatPrecision(t *testing.T) {
	original := make([]float64, 1000)
	for i := range original {
		original[i] = float64(i) * 0.1
	}

	encoded := EncodeFloat64(original)
	decoded := DecodeFloat(encoded)

	for i, v := range decoded {
		if math.Abs(v-original[i]) > 1e-10 {
			t.Errorf("Precision error at %d", i)
		}
	}
}

func TestStats(t *testing.T) {
	intData := []int64{10, 10, 12, 12, 12, 15}
	intStats := StatsInt(intData)

	if intStats.AbsoluteSum != 0+2+0+0+3 {
		t.Errorf("Int absolute sum expected 5, got %d", intStats.AbsoluteSum)
	}
	if intStats.MaxAbsDelta != 3 {
		t.Errorf("Int max abs delta expected 3, got %d", intStats.MaxAbsDelta)
	}
	if intStats.ZeroCount != 3 {
		t.Errorf("Int zero count expected 3, got %d", intStats.ZeroCount)
	}
	if math.Abs(intStats.ZeroRatio-0.6) > 1e-10 {
		t.Errorf("Int zero ratio expected 0.6, got %f", intStats.ZeroRatio)
	}

	floatData := []float64{1.0, 1.0 + 1e-10, 2.0, 2.0}
	floatStats := StatsFloat(floatData)

	if floatStats.ZeroCount != 2 {
		t.Errorf("Float zero count expected 2, got %d (epsilon 1e-9 test)", floatStats.ZeroCount)
	}

	floatData2 := []float64{1.0, 1.0 + 1e-8, 2.0, 2.0}
	floatStats2 := StatsFloat(floatData2)
	if floatStats2.ZeroCount != 1 {
		t.Errorf("Float zero count expected 1 for 1e-8, got %d", floatStats2.ZeroCount)
	}
}

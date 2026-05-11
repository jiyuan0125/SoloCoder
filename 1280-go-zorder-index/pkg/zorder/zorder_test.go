package zorder

import (
	"testing"
)

func TestEncode2D(t *testing.T) {
	testCases := []struct {
		name     string
		point    Point2D
		expected Code
	}{
		{"origin", Point2D{X: 0, Y: 0}, 0},
		{"(1,0)", Point2D{X: 1, Y: 0}, 1},
		{"(0,1)", Point2D{X: 0, Y: 1}, 2},
		{"(1,1)", Point2D{X: 1, Y: 1}, 3},
		{"(2,3)", Point2D{X: 2, Y: 3}, 0b1110},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotBit := Encode2DBit(tc.point)
			if gotBit != tc.expected {
				t.Errorf("Encode2DBit(%v) = %d, want %d", tc.point, gotBit, tc.expected)
			}

			gotLookup := Encode2DLookup(tc.point)
			if gotLookup != tc.expected {
				t.Errorf("Encode2DLookup(%v) = %d, want %d", tc.point, gotLookup, tc.expected)
			}
		})
	}
}

func TestDecode2D(t *testing.T) {
	testCases := []struct {
		name     string
		code     Code
		expected Point2D
	}{
		{"origin", 0, Point2D{X: 0, Y: 0}},
		{"code1", 1, Point2D{X: 1, Y: 0}},
		{"code2", 2, Point2D{X: 0, Y: 1}},
		{"code3", 3, Point2D{X: 1, Y: 1}},
		{"code14", 0b1110, Point2D{X: 2, Y: 3}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotBit := Decode2DBit(tc.code)
			if gotBit != tc.expected {
				t.Errorf("Decode2DBit(%d) = %v, want %v", tc.code, gotBit, tc.expected)
			}

			gotLookup := Decode2DLookup(tc.code)
			if gotLookup != tc.expected {
				t.Errorf("Decode2DLookup(%d) = %v, want %v", tc.code, gotLookup, tc.expected)
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	testPoints := []Point2D{
		{X: 0, Y: 0},
		{X: 12345, Y: 67890},
		{X: 0xFFFFFFFF, Y: 0xFFFFFFFF},
		{X: 0x12345678, Y: 0x9ABCDEF0},
	}

	for _, p := range testPoints {
		codeBit := Encode2DBit(p)
		decodedBit := Decode2DBit(codeBit)
		if decodedBit != p {
			t.Errorf("Bit method round-trip failed: %v -> %d -> %v", p, codeBit, decodedBit)
		}

		codeLookup := Encode2DLookup(p)
		decodedLookup := Decode2DLookup(codeLookup)
		if decodedLookup != p {
			t.Errorf("Lookup method round-trip failed: %v -> %d -> %v", p, codeLookup, decodedLookup)
		}

		if codeBit != codeLookup {
			t.Errorf("Encoding mismatch: Bit=%d, Lookup=%d", codeBit, codeLookup)
		}
	}
}

func TestEncode3D(t *testing.T) {
	testCases := []struct {
		name  string
		point Point3D
	}{
		{"origin", Point3D{X: 0, Y: 0, Z: 0}},
		{"small", Point3D{X: 1, Y: 2, Z: 3}},
		{"max", Point3D{X: max3DCoord, Y: max3DCoord, Z: max3DCoord}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code, err := Encode3D(tc.point)
			if err != nil {
				t.Fatalf("Encode3D(%v) failed: %v", tc.point, err)
			}

			decoded := Decode3D(code)
			if decoded != tc.point {
				t.Errorf("3D round-trip failed: %v -> %d -> %v", tc.point, code, decoded)
			}
		})
	}
}

func TestEncode3DOverflow(t *testing.T) {
	overflowPoints := []Point3D{
		{X: max3DCoord + 1, Y: 0, Z: 0},
		{X: 0, Y: max3DCoord + 1, Z: 0},
		{X: 0, Y: 0, Z: max3DCoord + 1},
		{X: 0xFFFFFFFF, Y: 0, Z: 0},
	}

	for _, p := range overflowPoints {
		_, err := Encode3D(p)
		if err == nil {
			t.Errorf("Expected error for overflow point %v, but got nil", p)
		}
	}
}

func TestQueryRanges(t *testing.T) {
	testCases := []struct {
		name string
		rect Rect2D
	}{
		{"aligned power of 2", Rect2D{Min: Point2D{X: 0, Y: 0}, Max: Point2D{X: 255, Y: 255}}},
		{"small rect", Rect2D{Min: Point2D{X: 0, Y: 0}, Max: Point2D{X: 3, Y: 3}}},
		{"crossing boundary", Rect2D{Min: Point2D{X: 7, Y: 7}, Max: Point2D{X: 8, Y: 8}}},
		{"single point", Rect2D{Min: Point2D{X: 100, Y: 100}, Max: Point2D{X: 100, Y: 100}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ranges := QueryRangesOptimized(tc.rect)
			if len(ranges) == 0 {
				t.Errorf("QueryRanges returned empty ranges for rect %v", tc.rect)
			}

			for _, r := range ranges {
				if r.Start > r.End {
					t.Errorf("Invalid range: Start=%d > End=%d", r.Start, r.End)
				}
			}
		})
	}
}

func TestQueryRangesAligned(t *testing.T) {
	alignedRect := Rect2D{Min: Point2D{X: 0, Y: 0}, Max: Point2D{X: 255, Y: 255}}
	ranges := QueryRangesOptimized(alignedRect)
	if len(ranges) != 1 {
		t.Errorf("Expected 1 range for aligned power-of-2 rect, got %d", len(ranges))
	}
}

func BenchmarkEncode2DBit(b *testing.B) {
	point := Point2D{X: 0x12345678, Y: 0x9ABCDEF0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode2DBit(point)
	}
}

func BenchmarkEncode2DLookup(b *testing.B) {
	point := Point2D{X: 0x12345678, Y: 0x9ABCDEF0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode2DLookup(point)
	}
}

func BenchmarkDecode2DBit(b *testing.B) {
	code := Encode2DBit(Point2D{X: 0x12345678, Y: 0x9ABCDEF0})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode2DBit(code)
	}
}

func BenchmarkDecode2DLookup(b *testing.B) {
	code := Encode2DBit(Point2D{X: 0x12345678, Y: 0x9ABCDEF0})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode2DLookup(code)
	}
}

func BenchmarkEncode3D(b *testing.B) {
	point := Point3D{X: 0x1FFFFF, Y: 0x1FFFFF, Z: 0x1FFFFF}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode3D(point)
	}
}

func BenchmarkDecode3D(b *testing.B) {
	code, _ := Encode3D(Point3D{X: 0x1FFFFF, Y: 0x1FFFFF, Z: 0x1FFFFF})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode3D(code)
	}
}

func BenchmarkQueryRangesSmall(b *testing.B) {
	rect := Rect2D{Min: Point2D{X: 0, Y: 0}, Max: Point2D{X: 100, Y: 100}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QueryRangesOptimized(rect)
	}
}

func BenchmarkQueryRangesLarge(b *testing.B) {
	rect := Rect2D{Min: Point2D{X: 1000, Y: 1000}, Max: Point2D{X: 10000, Y: 10000}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QueryRangesOptimized(rect)
	}
}

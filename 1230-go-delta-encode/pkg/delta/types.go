package delta

import "math"

const FloatEpsilon = 1e-9

type IntEncoded struct {
	Base   int64
	Deltas []int64
}

type FloatEncoded struct {
	Base   float64
	Deltas []float64
}

type IntStats struct {
	AbsoluteSum   int64
	MaxAbsDelta   int64
	ZeroRatio     float64
	TotalCount    int
	ZeroCount     int
}

type FloatStats struct {
	AbsoluteSum   float64
	MaxAbsDelta   float64
	ZeroRatio     float64
	TotalCount    int
	ZeroCount     int
}

func nearlyZero(x float64) bool {
	return math.Abs(x) < FloatEpsilon
}

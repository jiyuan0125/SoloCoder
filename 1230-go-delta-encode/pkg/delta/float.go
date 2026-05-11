package delta

import "math"

func EncodeFloat64(arr []float64) FloatEncoded {
	if len(arr) == 0 {
		return FloatEncoded{Base: 0.0, Deltas: []float64{}}
	}
	if len(arr) == 1 {
		return FloatEncoded{Base: arr[0], Deltas: []float64{}}
	}

	base := arr[0]
	deltas := make([]float64, 0, len(arr)-1)
	prev := base

	for i := 1; i < len(arr); i++ {
		deltas = append(deltas, arr[i]-prev)
		prev = arr[i]
	}

	return FloatEncoded{Base: base, Deltas: deltas}
}

func EncodeFloat32(arr []float32) FloatEncoded {
	converted := make([]float64, len(arr))
	for i, v := range arr {
		converted[i] = float64(v)
	}
	return EncodeFloat64(converted)
}

func DecodeFloat(encoded FloatEncoded) []float64 {
	if len(encoded.Deltas) == 0 {
		if encoded.Base == 0.0 {
			return []float64{}
		}
		return []float64{encoded.Base}
	}

	result := make([]float64, 0, len(encoded.Deltas)+1)
	current := encoded.Base
	result = append(result, current)

	for _, delta := range encoded.Deltas {
		current = current + delta
		result = append(result, current)
	}

	return result
}

func StatsFloat(arr []float64) FloatStats {
	if len(arr) <= 1 {
		return FloatStats{
			AbsoluteSum: 0.0,
			MaxAbsDelta: 0.0,
			ZeroRatio:   0.0,
			TotalCount:  len(arr),
			ZeroCount:   0,
		}
	}

	encoded := EncodeFloat64(arr)
	return statsFromFloatDeltas(encoded.Deltas)
}

func statsFromFloatDeltas(deltas []float64) FloatStats {
	if len(deltas) == 0 {
		return FloatStats{
			AbsoluteSum: 0.0,
			MaxAbsDelta: 0.0,
			ZeroRatio:   0.0,
			TotalCount:  0,
			ZeroCount:   0,
		}
	}

	var absSum float64
	var maxAbs float64
	zeroCount := 0

	for _, d := range deltas {
		absD := math.Abs(d)
		absSum += absD
		if absD > maxAbs {
			maxAbs = absD
		}
		if nearlyZero(d) {
			zeroCount++
		}
	}

	return FloatStats{
		AbsoluteSum: absSum,
		MaxAbsDelta: maxAbs,
		ZeroRatio:   float64(zeroCount) / float64(len(deltas)),
		TotalCount:  len(deltas),
		ZeroCount:   zeroCount,
	}
}

package delta

func EncodeInt64(arr []int64) IntEncoded {
	if len(arr) == 0 {
		return IntEncoded{Base: 0, Deltas: []int64{}}
	}
	if len(arr) == 1 {
		return IntEncoded{Base: arr[0], Deltas: []int64{}}
	}

	base := arr[0]
	deltas := make([]int64, 0, len(arr)-1)
	prev := base

	for i := 1; i < len(arr); i++ {
		deltas = append(deltas, arr[i]-prev)
		prev = arr[i]
	}

	return IntEncoded{Base: base, Deltas: deltas}
}

func EncodeInt32(arr []int32) IntEncoded {
	converted := make([]int64, len(arr))
	for i, v := range arr {
		converted[i] = int64(v)
	}
	return EncodeInt64(converted)
}

func EncodeInt(arr []int) IntEncoded {
	converted := make([]int64, len(arr))
	for i, v := range arr {
		converted[i] = int64(v)
	}
	return EncodeInt64(converted)
}

func DecodeInt(encoded IntEncoded) []int64 {
	if len(encoded.Deltas) == 0 {
		if encoded.Base == 0 {
			return []int64{}
		}
		return []int64{encoded.Base}
	}

	result := make([]int64, 0, len(encoded.Deltas)+1)
	current := encoded.Base
	result = append(result, current)

	for _, delta := range encoded.Deltas {
		current = current + delta
		result = append(result, current)
	}

	return result
}

func StatsInt(arr []int64) IntStats {
	if len(arr) <= 1 {
		return IntStats{
			AbsoluteSum: 0,
			MaxAbsDelta: 0,
			ZeroRatio:   0.0,
			TotalCount:  len(arr),
			ZeroCount:   0,
		}
	}

	encoded := EncodeInt64(arr)
	return statsFromIntDeltas(encoded.Deltas)
}

func statsFromIntDeltas(deltas []int64) IntStats {
	if len(deltas) == 0 {
		return IntStats{
			AbsoluteSum: 0,
			MaxAbsDelta: 0,
			ZeroRatio:   0.0,
			TotalCount:  0,
			ZeroCount:   0,
		}
	}

	var absSum int64
	var maxAbs int64
	zeroCount := 0

	for _, d := range deltas {
		absD := absInt64(d)
		absSum += absD
		if absD > maxAbs {
			maxAbs = absD
		}
		if d == 0 {
			zeroCount++
		}
	}

	return IntStats{
		AbsoluteSum: absSum,
		MaxAbsDelta: maxAbs,
		ZeroRatio:   float64(zeroCount) / float64(len(deltas)),
		TotalCount:  len(deltas),
		ZeroCount:   zeroCount,
	}
}

func absInt64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

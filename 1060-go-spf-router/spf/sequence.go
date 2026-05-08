package spf

import "math"

const (
	halfUint32 = uint32(math.MaxUint32)/2 + 1
)

func IsNewer(current, candidate uint32) bool {
	diff := candidate - current
	if diff == 0 {
		return false
	}
	return diff < halfUint32
}

func IsNewerOrEqual(current, candidate uint32) bool {
	diff := candidate - current
	return diff < halfUint32
}

func IncrementSeq(seq uint32) uint32 {
	return seq + 1
}

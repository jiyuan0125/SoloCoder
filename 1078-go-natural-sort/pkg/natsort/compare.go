package natsort

import (
	"strings"
	"unicode"
)

func Compare(a, b string, opts Options) int {
	segsA := buildSegments(a)
	segsB := buildSegments(b)

	lenA := len(segsA)
	lenB := len(segsB)
	maxLen := lenA
	if lenB > maxLen {
		maxLen = lenB
	}

	for i := 0; i < maxLen; i++ {
		if i >= lenA {
			if opts.Ascending {
				return -1
			}
			return 1
		}
		if i >= lenB {
			if opts.Ascending {
				return 1
			}
			return -1
		}

		cmp := compareSegment(&segsA[i], &segsB[i], opts)
		if cmp != 0 {
			if opts.Ascending {
				return cmp
			}
			return -cmp
		}
	}

	cmp := compareRawString(a, b, opts)
	if opts.Ascending {
		return cmp
	}
	return -cmp
}

func compareSegment(sa, sb *segment, opts Options) int {
	if sa.kind != sb.kind {
		return compareText(sa.text, sb.text, opts)
	}

	switch sa.kind {
	case segmentText:
		return compareText(sa.text, sb.text, opts)
	case segmentInteger:
		return compareInteger(sa, sb, opts)
	case segmentFloat:
		return compareFloat(sa, sb, opts)
	}
	return compareText(sa.text, sb.text, opts)
}

func compareInteger(sa, sb *segment, opts Options) int {
	aval := sa.intValue
	bval := sb.intValue

	if sa.isNegative {
		aval = -aval
	}
	if sb.isNegative {
		bval = -bval
	}

	if aval < bval {
		return -1
	}
	if aval > bval {
		return 1
	}

	if !opts.IgnoreLeadingZeros {
		if sa.leadingZeros > sb.leadingZeros {
			return -1
		}
		if sa.leadingZeros < sb.leadingZeros {
			return 1
		}
	}

	return compareText(sa.text, sb.text, opts)
}

func compareFloat(sa, sb *segment, opts Options) int {
	aInt := sa.intPart
	bInt := sb.intPart

	if sa.isNegative {
		aInt = -aInt
	}
	if sb.isNegative {
		bInt = -bInt
	}

	if aInt < bInt {
		return -1
	}
	if aInt > bInt {
		return 1
	}

	cmp := compareFracPart(sa.fracPart, sb.fracPart)
	if cmp != 0 {
		return cmp
	}

	if !opts.IgnoreLeadingZeros {
		if sa.leadingZeros > sb.leadingZeros {
			return -1
		}
		if sa.leadingZeros < sb.leadingZeros {
			return 1
		}
	}

	return compareText(sa.text, sb.text, opts)
}

func compareFracPart(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	la := len(ra)
	lb := len(rb)
	minLen := la
	if lb < minLen {
		minLen = lb
	}

	for i := 0; i < minLen; i++ {
		if ra[i] < rb[i] {
			return -1
		}
		if ra[i] > rb[i] {
			return 1
		}
	}

	if la < lb {
		return -1
	}
	if la > lb {
		return 1
	}
	return 0
}

func compareText(a, b string, opts Options) int {
	if opts.CaseSensitive {
		return strings.Compare(a, b)
	}

	ra := []rune(a)
	rb := []rune(b)
	la := len(ra)
	lb := len(rb)
	minLen := la
	if lb < minLen {
		minLen = lb
	}

	for i := 0; i < minLen; i++ {
		ca := unicode.ToLower(ra[i])
		cb := unicode.ToLower(rb[i])
		if ca < cb {
			return -1
		}
		if ca > cb {
			return 1
		}
	}

	if la < lb {
		return -1
	}
	if la > lb {
		return 1
	}

	for i := 0; i < la; i++ {
		ca := ra[i]
		cb := rb[i]
		if ca != cb {
			if unicode.IsLower(ca) && unicode.IsUpper(cb) {
				return -1
			}
			if unicode.IsUpper(ca) && unicode.IsLower(cb) {
				return 1
			}
			if ca < cb {
				return -1
			}
			return 1
		}
	}

	return 0
}

func compareRawString(a, b string, opts Options) int {
	return compareText(a, b, opts)
}

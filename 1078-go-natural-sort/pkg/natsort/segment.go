package natsort

import (
	"strings"
	"unicode"
)

type segmentKind int

const (
	segmentText segmentKind = iota
	segmentInteger
	segmentFloat
)

type segment struct {
	kind     segmentKind
	text     string
	intValue int64
	intPart  int64
	fracPart string
	leadingZeros int
	isNegative bool
}

func isFullWidthDigit(r rune) bool {
	return r >= '０' && r <= '９'
}

func fullWidthToHalfWidth(r rune) rune {
	if isFullWidthDigit(r) {
		return r - '０' + '0'
	}
	return r
}

func isDigit(r rune) bool {
	return unicode.IsDigit(r) || isFullWidthDigit(r)
}

func isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

func buildSegments(s string) []segment {
	var segs []segment
	runes := []rune(s)
	n := len(runes)
	i := 0

	for i < n {
		if isDigit(runes[i]) {
			start := i
			hasDot := false
			dotIndex := -1
			negIndex := -1

			if i > 0 && runes[i-1] == '-' {
				negIndex = i - 1
			}

			for i < n && (isDigit(runes[i]) || runes[i] == '.') {
				if runes[i] == '.' {
					if hasDot {
						break
					}
					hasDot = true
					dotIndex = i
				}
				i++
			}

			if hasDot {
				if dotIndex > start && dotIndex < i-1 {
						if isDigit(runes[dotIndex-1]) && isDigit(runes[dotIndex+1]) {
							intPartRunes := runes[start:dotIndex]
							fracPartRunes := runes[dotIndex+1 : i]
							isNeg := false

							leadingZeros := 0
							actualStart := start
							for actualStart < dotIndex {
								c := fullWidthToHalfWidth(runes[actualStart])
								if c == '0' {
									leadingZeros++
									actualStart++
								} else {
									break
								}
							}
							if actualStart == dotIndex {
								actualStart = dotIndex - 1
							}

							if negIndex >= 0 {
								isNeg = true
								segs = addTextSegment(segs, runes, negIndex, negIndex+1)
							}

							text := string(runes[start:i])
							intText := convertDigits(intPartRunes)
							fracText := convertDigits(fracPartRunes)
							var intVal int64
							parseInt(intText, &intVal)
							segs = append(segs, segment{
								kind:         segmentFloat,
								text:         text,
								intPart:      intVal,
								fracPart:     fracText,
								leadingZeros: leadingZeros,
								isNegative:   isNeg,
							})
							continue
						}
					}
			}

			leadingZeros := 0
			actualStart := start
			for actualStart < i {
				c := fullWidthToHalfWidth(runes[actualStart])
				if c == '0' {
					leadingZeros++
					actualStart++
				} else {
					break
				}
			}
			if actualStart == i {
				actualStart = i - 1
			}

			text := string(runes[start:i])
			valText := convertDigits(runes[start:i])
			var val int64
			var valFlt float64
			parseInt(valText, &val)

			isNeg := false
			if negIndex >= 0 {
				isNeg = true
				segs = addTextSegment(segs, runes, negIndex, negIndex+1)
			}

			segs = append(segs, segment{
				kind:         segmentInteger,
				text:         text,
				intValue:     val,
				fltValue:     valFlt,
				leadingZeros: leadingZeros,
				isNegative:   isNeg,
			})
			continue
		}

		if isWhitespace(runes[i]) {
			start := i
			for i < n && isWhitespace(runes[i]) {
				i++
			}
			segs = addTextSegment(segs, runes, start, i)
			continue
		}

		start := i
		for i < n && !isDigit(runes[i]) && !isWhitespace(runes[i]) {
			i++
		}
		segs = addTextSegment(segs, runes, start, i)
	}

	return segs
}

func addTextSegment(segs []segment, runes []rune, start, end int) []segment {
	text := string(runes[start:end])
	segs = append(segs, segment{
		kind: segmentText,
		text: text,
	})
	return segs
}

func convertDigits(rs []rune) string {
	var sb strings.Builder
	for _, r := range rs {
		sb.WriteRune(fullWidthToHalfWidth(r))
	}
	return sb.String()
}

func parseInt(s string, out *int64) {
	var val int64
	neg := false
	i := 0
	n := len(s)
	if n > 0 && s[0] == '-' {
		neg = true
		i = 1
	}
	for i < n {
		if s[i] >= '0' && s[i] <= '9' {
			val = val*10 + int64(s[i]-'0')
		}
		i++
	}
	if neg {
		val = -val
	}
	*out = val
}

func parseFloat(s string, outFlt *float64, outInt *int64) (bool, float64) {
	var intPart int64
	var fracPart int64
	fracLen := 0
	neg := false
	i := 0
	n := len(s)
	if n > 0 && s[0] == '-' {
		neg = true
		i = 1
	}
	dotSeen := false
	for i < n {
		if s[i] == '.' {
			dotSeen = true
			i++
			continue
		}
		if s[i] >= '0' && s[i] <= '9' {
			if !dotSeen {
				intPart = intPart*10 + int64(s[i]-'0')
			} else {
				fracPart = fracPart*10 + int64(s[i]-'0')
				fracLen++
			}
		}
		i++
	}
	var flt float64
	if fracLen > 0 {
		div := 1.0
		for j := 0; j < fracLen; j++ {
			div *= 10
		}
		flt = float64(intPart) + float64(fracPart)/div
	} else {
		flt = float64(intPart)
	}
	if neg {
		flt = -flt
		intPart = -intPart
	}
	*outFlt = flt
	*outInt = intPart
	return dotSeen, flt
}

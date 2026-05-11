package naturalsort

import (
	"math/big"
	"sort"
	"strings"
	"unicode"
)

type Options struct {
	CaseSensitive  bool
	Descending     bool
	KeepLeadingZeros bool
}

func DefaultOptions() Options {
	return Options{
		CaseSensitive:    false,
		Descending:       false,
		KeepLeadingZeros: false,
	}
}

type segment struct {
	kind   segmentKind
	text   string
	number *big.Rat
	leadingZeros int
	isFloat bool
}

type segmentKind int

const (
	segmentText segmentKind = iota
	segmentNumber
	segmentWhitespace
)

func isDigit(r rune) bool {
	if unicode.IsDigit(r) {
		return true
	}
	if r >= '０' && r <= '９' {
		return true
	}
	return false
}

func toHalfwidth(r rune) rune {
	if r >= '０' && r <= '９' {
		return r - '０' + '0'
	}
	return r
}

func parseNumber(s string) (*big.Rat, bool, int) {
	original := s
	leadingZeros := 0
	hasNegative := false
	
	if len(s) > 0 && s[0] == '-' {
		hasNegative = true
		s = s[1:]
	}
	
	isFloat := strings.Contains(s, ".")
	
	for i := 0; i < len(s) && s[i] == '0'; i++ {
		leadingZeros++
	}
	
	var normalized strings.Builder
	if hasNegative {
		normalized.WriteByte('-')
	}
	for _, r := range s {
		if r == '.' {
			normalized.WriteRune(r)
			continue
		}
		normalized.WriteRune(toHalfwidth(r))
	}
	normalizedStr := normalized.String()
	
	if normalizedStr == "-" {
		return nil, false, 0
	}
	
	num, ok := new(big.Rat).SetString(normalizedStr)
	if !ok {
		return nil, false, 0
	}
	
	_ = original
	return num, isFloat, leadingZeros
}

func isLikelyVersion(segments []segment, idx int) bool {
	if idx < 0 || idx >= len(segments) {
		return false
	}
	
	current := segments[idx]
	if current.kind != segmentNumber || current.isFloat {
		return false
	}
	
	hasPrev := idx > 0
	hasNext := idx+1 < len(segments)
	
	checkAdjacent := func(adjIdx int) bool {
		if adjIdx < 0 || adjIdx >= len(segments) {
			return false
		}
		adj := segments[adjIdx]
		if adj.kind != segmentText {
			return false
		}
		return adj.text == "."
	}
	
	if hasPrev && checkAdjacent(idx-1) {
		if hasNext && checkAdjacent(idx+1) {
			prevPrevIdx := idx - 2
			nextNextIdx := idx + 2
			hasPrevPrevNumber := prevPrevIdx >= 0 && segments[prevPrevIdx].kind == segmentNumber
			hasNextNextNumber := nextNextIdx < len(segments) && segments[nextNextIdx].kind == segmentNumber
			if hasPrevPrevNumber || hasNextNextNumber {
				return true
			}
		}
	}
	
	return false
}

func isFloatingPointContext(runes []rune, i int) bool {
	if i <= 0 || i >= len(runes) {
		return false
	}
	
	if !isDigit(runes[i-1]) {
		return false
	}
	
	if i+1 >= len(runes) || !isDigit(runes[i+1]) {
		return false
	}
	
	beforeDot := i - 1
	for beforeDot >= 0 && isDigit(runes[beforeDot]) {
		beforeDot--
	}
	
	if beforeDot >= 0 && runes[beforeDot] == '.' {
		return false
	}
	
	afterDot := i + 1
	for afterDot < len(runes) && isDigit(runes[afterDot]) {
		afterDot++
	}
	
	if afterDot < len(runes) && runes[afterDot] == '.' {
		return false
	}
	
	return true
}

func segmentize(s string) []segment {
	var segments []segment
	runes := []rune(s)
	n := len(runes)
	
	if n == 0 {
		return segments
	}
	
	i := 0
	for i < n {
		r := runes[i]
		
		if unicode.IsSpace(r) {
			start := i
			for i < n && unicode.IsSpace(runes[i]) {
				i++
			}
			segments = append(segments, segment{
				kind: segmentWhitespace,
				text: string(runes[start:i]),
			})
			continue
		}
		
		if r == '-' && i+1 < n && isDigit(runes[i+1]) {
			start := i
			i++
			for i < n && isDigit(runes[i]) {
				i++
			}
			if i < n && runes[i] == '.' && isFloatingPointContext(runes, i) {
				i++
				for i < n && isDigit(runes[i]) {
					i++
				}
			}
			numText := string(runes[start:i])
			if num, isFloat, leadingZeros := parseNumber(numText); num != nil {
				segments = append(segments, segment{
					kind:         segmentNumber,
					text:         numText,
					number:       num,
					leadingZeros: leadingZeros,
					isFloat:      isFloat,
				})
			} else {
				segments = append(segments, segment{
					kind: segmentText,
					text: numText,
				})
			}
			continue
		}
		
		if isDigit(r) {
			start := i
			for i < n && isDigit(runes[i]) {
				i++
			}
			if i < n && runes[i] == '.' && isFloatingPointContext(runes, i) {
				i++
				for i < n && isDigit(runes[i]) {
					i++
				}
			}
			numText := string(runes[start:i])
			if num, isFloat, leadingZeros := parseNumber(numText); num != nil {
				segments = append(segments, segment{
					kind:         segmentNumber,
					text:         numText,
					number:       num,
					leadingZeros: leadingZeros,
					isFloat:      isFloat,
				})
			} else {
				segments = append(segments, segment{
					kind: segmentText,
					text: numText,
				})
			}
			continue
		}
		
		start := i
		for i < n {
			curr := runes[i]
			if unicode.IsSpace(curr) || isDigit(curr) {
				break
			}
			if curr == '-' && i+1 < n && isDigit(runes[i+1]) {
				break
			}
			i++
		}
		segments = append(segments, segment{
			kind: segmentText,
			text: string(runes[start:i]),
		})
	}
	
	return segments
}

func compareSegments(a, b []segment, opts Options) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	
	for i := 0; i < minLen; i++ {
		sa := a[i]
		sb := b[i]
		
		if sa.kind == segmentWhitespace && sb.kind == segmentWhitespace {
			continue
		}
		
		if sa.kind == segmentWhitespace {
			return -1
		}
		if sb.kind == segmentWhitespace {
			return 1
		}
		
		if sa.kind == segmentNumber && sb.kind == segmentNumber {
			aIsVersion := isLikelyVersion(a, i)
			bIsVersion := isLikelyVersion(b, i)
			
			if sa.number == nil && sb.number == nil {
				cmp := strings.Compare(sa.text, sb.text)
				if cmp != 0 {
					return cmp
				}
				continue
			}
			if sa.number == nil {
				return -1
			}
			if sb.number == nil {
				return 1
			}
			
			cmp := sa.number.Cmp(sb.number)
			if cmp != 0 {
				return cmp
			}
			
			if opts.KeepLeadingZeros {
				if sa.leadingZeros != sb.leadingZeros {
					return sb.leadingZeros - sa.leadingZeros
				}
			}
			
			_ = aIsVersion
			_ = bIsVersion
			
			continue
		}
		
		if sa.kind == segmentNumber {
			return -1
		}
		if sb.kind == segmentNumber {
			return 1
		}
		
		if opts.CaseSensitive {
			cmp := strings.Compare(sa.text, sb.text)
			if cmp != 0 {
				return cmp
			}
		} else {
			cmp := strings.Compare(strings.ToLower(sa.text), strings.ToLower(sb.text))
			if cmp != 0 {
				return cmp
			}
			for j := 0; j < len(sa.text) && j < len(sb.text); j++ {
				ca := rune(sa.text[j])
				cb := rune(sb.text[j])
				if ca == cb {
					continue
				}
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
			if len(sa.text) != len(sb.text) {
				if len(sa.text) < len(sb.text) {
					return -1
				}
				return 1
			}
		}
	}
	
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

func Compare(a, b string, opts Options) int {
	segA := segmentize(a)
	segB := segmentize(b)
	result := compareSegments(segA, segB, opts)
	if opts.Descending {
		return -result
	}
	return result
}

func Sort(strings []string, opts Options) []string {
	result := make([]string, len(strings))
	copy(result, strings)
	
	indices := make([]int, len(result))
	for i := range indices {
		indices[i] = i
	}
	
	sort.SliceStable(indices, func(i, j int) bool {
		cmp := Compare(result[indices[i]], result[indices[j]], opts)
		if cmp == 0 {
			return indices[i] < indices[j]
		}
		return cmp < 0
	})
	
	sorted := make([]string, len(result))
	for i, idx := range indices {
		sorted[i] = result[idx]
	}
	
	return sorted
}

func SimpleSort(strings []string) []string {
	return Sort(strings, DefaultOptions())
}

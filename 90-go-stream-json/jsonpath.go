package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type SegmentType int

const (
	SegRoot SegmentType = iota
	SegField
	SegIndex
	SegWildcard
)

type PathSegment struct {
	Type  SegmentType
	Value string
	Index int
}

type JSONPath struct {
	Segments []PathSegment
}

func ParseJSONPath(path string) (*JSONPath, error) {
	if !strings.HasPrefix(path, "$.") && path != "$" {
		return nil, errors.New("JSONPath must start with $")
	}

	var segments []PathSegment
	segments = append(segments, PathSegment{Type: SegRoot})

	if path == "$" {
		return &JSONPath{Segments: segments}, nil
	}

	remainder := path[2:]

	for remainder != "" {
		if remainder[0] == '[' {
			closeIdx := strings.Index(remainder, "]")
			if closeIdx == -1 {
				return nil, errors.New("unmatched [ in JSONPath")
			}

			selector := remainder[1:closeIdx]
			remainder = remainder[closeIdx+1:]

			if selector == "*" {
				segments = append(segments, PathSegment{Type: SegWildcard})
			} else if strings.HasPrefix(selector, "'") && strings.HasSuffix(selector, "'") {
				segments = append(segments, PathSegment{Type: SegField, Value: selector[1 : len(selector)-1]})
			} else if strings.HasPrefix(selector, "\"") && strings.HasSuffix(selector, "\"") {
				segments = append(segments, PathSegment{Type: SegField, Value: selector[1 : len(selector)-1]})
			} else {
				index, err := strconv.Atoi(selector)
				if err != nil {
					return nil, fmt.Errorf("invalid array index: %s", selector)
				}
				segments = append(segments, PathSegment{Type: SegIndex, Index: index})
			}

			if strings.HasPrefix(remainder, ".") {
				remainder = remainder[1:]
			}
		} else {
			dotIdx := strings.Index(remainder, ".")
			bracketIdx := strings.Index(remainder, "[")

			var nextDot int
			if dotIdx == -1 && bracketIdx == -1 {
				nextDot = len(remainder)
			} else if dotIdx == -1 {
				nextDot = bracketIdx
			} else if bracketIdx == -1 {
				nextDot = dotIdx
			} else {
				nextDot = min(dotIdx, bracketIdx)
			}

			field := remainder[:nextDot]
			remainder = remainder[nextDot:]

			if field != "" {
				segments = append(segments, PathSegment{Type: SegField, Value: field})
			}

			if strings.HasPrefix(remainder, ".") {
				remainder = remainder[1:]
			}
		}
	}

	return &JSONPath{Segments: segments}, nil
}

type PathTracker struct {
	segments      []PathSegment
	currentDepth  int
	currentPath   []interface{}
	arrayIndices  []int
	inMatch       bool
	matchLevel    int
	captureValues bool
}

func NewPathTracker(jp *JSONPath) *PathTracker {
	return &PathTracker{
		segments:     jp.Segments,
		currentDepth: 0,
		currentPath:  make([]interface{}, 0),
		arrayIndices: make([]int, 0),
		inMatch:      false,
		matchLevel:   -1,
	}
}

func (pt *PathTracker) EnterObject() {
	pt.currentDepth++
}

func (pt *PathTracker) ExitObject() {
	if pt.matchLevel == pt.currentDepth {
		pt.matchLevel = -1
		pt.inMatch = false
	}
	pt.currentDepth--
	if len(pt.currentPath) > 0 {
		pt.currentPath = pt.currentPath[:len(pt.currentPath)-1]
	}
}

func (pt *PathTracker) EnterArray() {
	pt.currentDepth++
	pt.arrayIndices = append(pt.arrayIndices, -1)
}

func (pt *PathTracker) ExitArray() {
	if pt.matchLevel == pt.currentDepth {
		pt.matchLevel = -1
		pt.inMatch = false
	}
	pt.currentDepth--
	if len(pt.arrayIndices) > 0 {
		pt.arrayIndices = pt.arrayIndices[:len(pt.arrayIndices)-1]
	}
	if len(pt.currentPath) > 0 {
		pt.currentPath = pt.currentPath[:len(pt.currentPath)-1]
	}
}

func (pt *PathTracker) NextArrayElement() {
	if len(pt.arrayIndices) > 0 {
		pt.arrayIndices[len(pt.arrayIndices)-1]++
	}
}

func (pt *PathTracker) SetField(name string) {
	pt.currentPath = append(pt.currentPath, name)
}

func (pt *PathTracker) IsMatching() bool {
	if len(pt.segments) == 1 {
		return true
	}

	pathIdx := 1
	segCount := len(pt.segments)

	for i := 0; i < len(pt.currentPath); i++ {
		if pathIdx >= segCount {
			break
		}

		seg := pt.segments[pathIdx]
		elem := pt.currentPath[i]

		switch seg.Type {
		case SegField:
			if name, ok := elem.(string); ok {
				if name == seg.Value {
					pathIdx++
				} else {
					return false
				}
			} else {
				return false
			}
		case SegIndex:
			if idx, ok := elem.(int); ok {
				if idx == seg.Index {
					pathIdx++
				} else {
					return false
				}
			} else {
				return false
			}
		case SegWildcard:
			if _, ok := elem.(int); ok {
				pathIdx++
			} else {
				return false
			}
		default:
			return false
		}
	}

	if pathIdx >= segCount {
		pt.inMatch = true
		pt.matchLevel = pt.currentDepth
		return true
	}

	return false
}

func (pt *PathTracker) SetArrayIndex() {
	if len(pt.arrayIndices) > 0 {
		idx := pt.arrayIndices[len(pt.arrayIndices)-1]
		pt.currentPath = append(pt.currentPath, idx)
	}
}

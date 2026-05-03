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

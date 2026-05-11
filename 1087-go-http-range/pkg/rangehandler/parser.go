package rangehandler

import (
	"strconv"
	"strings"
)

func ParseRangeHeader(header string, size int64) ([]Range, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return nil, nil
	}

	parts := strings.SplitN(header, "=", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidRangeFormat
	}

	unit := strings.TrimSpace(parts[0])
	if unit != "bytes" {
		return nil, ErrInvalidRangeUnit
	}

	rangesPart := strings.TrimSpace(parts[1])
	if rangesPart == "" {
		return nil, ErrInvalidRangeFormat
	}

	rangeStrings := strings.Split(rangesPart, ",")
	var ranges []Range

	for _, rs := range rangeStrings {
		rs = strings.TrimSpace(rs)
		if rs == "" {
			continue
		}

		r, err := parseSingleRange(rs, size)
		if err != nil {
			return nil, err
		}

		ranges = append(ranges, r)
	}

	if len(ranges) == 0 {
		return nil, ErrInvalidRangeFormat
	}

	return ranges, nil
}

func parseSingleRange(rs string, size int64) (Range, error) {
	parts := strings.SplitN(rs, "-", 2)
	if len(parts) != 2 {
		return Range{}, ErrInvalidRangeFormat
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	var start, end int64

	if startStr == "" {
		if endStr == "" {
			return Range{}, ErrInvalidRangeFormat
		}
		suffixLength, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil || suffixLength <= 0 {
			return Range{}, ErrInvalidRangeFormat
		}

		if suffixLength > size {
			suffixLength = size
		}

		start = size - suffixLength
		end = size - 1
	} else {
		s, err := strconv.ParseInt(startStr, 10, 64)
		if err != nil || s < 0 {
			return Range{}, ErrInvalidRangeFormat
		}
		start = s

		if endStr == "" {
			end = size - 1
		} else {
			e, err := strconv.ParseInt(endStr, 10, 64)
			if err != nil || e < 0 {
				return Range{}, ErrInvalidRangeFormat
			}
			end = e
		}
	}

	if start >= size {
		return Range{}, ErrRangeUnsatisfiable
	}

	if end >= size {
		end = size - 1
	}

	if start > end {
		return Range{}, ErrInvalidRangeFormat
	}

	return Range{Start: start, End: end}, nil
}

package ipparser

import (
	"go-ip-parser/pkg/common"
	"strings"
)

func parseIPv6(s string) (*Address, error) {
	if s == "::" {
		return &Address{
			version: common.IPv6,
			ipv6:    [16]byte{},
		}, nil
	}

	if s == "::1" {
		return &Address{
			version: common.IPv6,
			ipv6:    [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		}, nil
	}

	var ipv4Suffix string
	hasIPv4Suffix := false
	workingStr := s

	idx := strings.LastIndex(s, ":")
	if idx > 0 && idx < len(s)-1 {
		after := s[idx+1:]
		if strings.Contains(after, ".") {
			ipv4Suffix = after
			workingStr = s[:idx]
			hasIPv4Suffix = true
		}
	}

	doubleColon := strings.Index(workingStr, "::")
	doubleColonCount := strings.Count(workingStr, "::")

	if doubleColonCount > 1 {
		return nil, &ParseError{Message: "multiple '::' in IPv6 address"}
	}

	var groups [8]uint16

	isIPv4Mapped := false
	isIPv4Compatible := false

	if doubleColon != -1 {
		leftPart := workingStr[:doubleColon]
		rightPart := workingStr[doubleColon+2:]

		var leftGroups []uint16
		if leftPart != "" {
			for _, g := range strings.Split(leftPart, ":") {
				if g == "" {
					continue
				}
				val, err := parseHexGroup(g)
				if err != nil {
					return nil, err
				}
				leftGroups = append(leftGroups, val)
			}
		}

		var rightGroups []uint16
		if rightPart != "" {
			for _, g := range strings.Split(rightPart, ":") {
				if g == "" {
					continue
				}
				val, err := parseHexGroup(g)
				if err != nil {
					return nil, err
				}
				rightGroups = append(rightGroups, val)
			}
		}

		totalGroups := len(leftGroups) + len(rightGroups)
		requiredGroups := 8
		if hasIPv4Suffix {
			requiredGroups = 6
		}

		if totalGroups > requiredGroups {
			return nil, &ParseError{Message: "too many groups in IPv6 address"}
		}

		zerosToAdd := requiredGroups - totalGroups

		groupIdx := 0
		for _, g := range leftGroups {
			groups[groupIdx] = g
			groupIdx++
		}

		for i := 0; i < zerosToAdd; i++ {
			groups[groupIdx] = 0
			groupIdx++
		}

		for _, g := range rightGroups {
			groups[groupIdx] = g
			groupIdx++
		}

		if hasIPv4Suffix {
			ipv4Addr, err := parseIPv4(ipv4Suffix)
			if err != nil {
				return nil, err
			}
			octets := ipv4Addr.IPv4Octets()

			isIPv4Mapped = groups[5] == 0xffff
			isIPv4Compatible = groups[0] == 0x0000 && groups[1] == 0x0000 && groups[2] == 0x0000 &&
				groups[3] == 0x0000 && groups[4] == 0x0000 && groups[5] == 0x0000

			groups[6] = uint16(octets[0])<<8 | uint16(octets[1])
			groups[7] = uint16(octets[2])<<8 | uint16(octets[3])
		}
	} else {
		parts := strings.Split(workingStr, ":")
		expectedGroups := 8
		if hasIPv4Suffix {
			expectedGroups = 6
		}

		if len(parts) > expectedGroups {
			return nil, &ParseError{Message: "too many groups in IPv6 address"}
		}

		for i, g := range parts {
			if g == "" && i != 0 && i != len(parts)-1 {
				return nil, &ParseError{Message: "empty group in IPv6 address"}
			}
			if g == "" {
				continue
			}
			val, err := parseHexGroup(g)
			if err != nil {
				return nil, err
			}
			groups[i] = val
		}

		if hasIPv4Suffix {
			ipv4Addr, err := parseIPv4(ipv4Suffix)
			if err != nil {
				return nil, err
			}
			octets := ipv4Addr.IPv4Octets()

			isIPv4Mapped = groups[5] == 0xffff
			isIPv4Compatible = groups[0] == 0x0000 && groups[1] == 0x0000 && groups[2] == 0x0000 &&
				groups[3] == 0x0000 && groups[4] == 0x0000 && groups[5] == 0x0000

			groups[6] = uint16(octets[0])<<8 | uint16(octets[1])
			groups[7] = uint16(octets[2])<<8 | uint16(octets[3])
		}
	}

	var ipv6 [16]byte
	for i := 0; i < 8; i++ {
		ipv6[i*2] = byte(groups[i] >> 8)
		ipv6[i*2+1] = byte(groups[i] & 0xff)
	}

	return &Address{
		version:          common.IPv6,
		ipv6:             ipv6,
		isIPv4Mapped:     isIPv4Mapped,
		isIPv4Compatible: isIPv4Compatible,
	}, nil
}

func looksLikeIPv6(s string) bool {
	colons := countColons(s)
	if colons < 2 {
		return false
	}
	if s == "::" || s == "::1" {
		return true
	}
	if strings.HasPrefix(s, ":") || strings.HasSuffix(s, ":") {
		return true
	}
	if strings.Contains(s, "::") {
		return true
	}
	if colons >= 7 {
		return true
	}
	return false
}

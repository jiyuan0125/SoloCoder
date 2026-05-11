package ipparser

import (
	"go-ip-parser/pkg/common"
	"strings"
)

func parseIPv4(s string) (*Address, error) {
	parts := strings.Split(s, ".")

	if len(parts) == 0 || len(parts) > 4 {
		return nil, &ParseError{Message: "invalid IPv4 address format"}
	}

	var octets [4]byte

	if len(parts) == 1 {
		val, err := parseOctet(parts[0])
		if err != nil {
			return nil, err
		}
		if val > 0xFFFFFFFF {
			return nil, &ParseError{Message: "IPv4 address value out of range"}
		}
		octets[0] = byte((val >> 24) & 0xFF)
		octets[1] = byte((val >> 16) & 0xFF)
		octets[2] = byte((val >> 8) & 0xFF)
		octets[3] = byte(val & 0xFF)
	} else if len(parts) == 2 {
		a, err := parseOctet(parts[0])
		if err != nil {
			return nil, err
		}
		if a > 0xFF {
			return nil, &ParseError{Message: "first octet out of range"}
		}
		rest, err := parseOctet(parts[1])
		if err != nil {
			return nil, err
		}
		if rest > 0xFFFFFF {
			return nil, &ParseError{Message: "address part out of range"}
		}
		octets[0] = byte(a)
		octets[1] = byte((rest >> 16) & 0xFF)
		octets[2] = byte((rest >> 8) & 0xFF)
		octets[3] = byte(rest & 0xFF)
	} else if len(parts) == 3 {
		a, err := parseOctet(parts[0])
		if err != nil {
			return nil, err
		}
		if a > 0xFF {
			return nil, &ParseError{Message: "first octet out of range"}
		}
		b, err := parseOctet(parts[1])
		if err != nil {
			return nil, err
		}
		if b > 0xFF {
			return nil, &ParseError{Message: "second octet out of range"}
		}
		rest, err := parseOctet(parts[2])
		if err != nil {
			return nil, err
		}
		if rest > 0xFFFF {
			return nil, &ParseError{Message: "address part out of range"}
		}
		octets[0] = byte(a)
		octets[1] = byte(b)
		octets[2] = byte((rest >> 8) & 0xFF)
		octets[3] = byte(rest & 0xFF)
	} else {
		for i := 0; i < 4; i++ {
			val, err := parseOctet(parts[i])
			if err != nil {
				return nil, err
			}
			if val > 0xFF {
				return nil, &ParseError{Message: "octet out of range"}
			}
			octets[i] = byte(val)
		}
	}

	return &Address{
		version: common.IPv4,
		ipv4:    octets,
	}, nil
}

func looksLikeIPv4(s string) bool {
	dots := countDots(s)
	if dots > 3 {
		return false
	}

	if dots == 0 {
		for i := 0; i < len(s); i++ {
			c := s[i]
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || (c == 'x' || c == 'X')) {
				return false
			}
		}
		_, err := parseOctet(s)
		return err == nil
	}

	return true
}

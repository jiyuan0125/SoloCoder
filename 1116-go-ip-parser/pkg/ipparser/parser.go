package ipparser

import (
	"strings"
)

func Parse(ip string) (*Address, error) {
	ip = strings.TrimSpace(ip)

	if ip == "" {
		return nil, &ParseError{Message: "empty IP address"}
	}

	if looksLikeIPv6(ip) {
		return parseIPv6(ip)
	}

	if looksLikeIPv4(ip) {
		return parseIPv4(ip)
	}

	return nil, &ParseError{Message: "unrecognized IP address format"}
}

func Format(ip string, compact bool) (string, error) {
	addr, err := Parse(ip)
	if err != nil {
		return "", err
	}

	if compact {
		return addr.Canonical(), nil
	}
	return addr.Standard(), nil
}

func Classify(ip string) (*Address, error) {
	return Parse(ip)
}

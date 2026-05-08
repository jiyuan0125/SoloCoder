package cidr

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func Parse(cidr string) (*CIDR, error) {
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid CIDR format: %s", cidr)
	}

	ipStr := parts[0]
	prefixStr := parts[1]

	prefix, err := strconv.Atoi(prefixStr)
	if err != nil {
		return nil, fmt.Errorf("invalid prefix: %s", prefixStr)
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	var version IPVersion
	var maxPrefix int

	if ip.To4() != nil {
		version = IPv4
		maxPrefix = 32
		ip = ip.To4()
	} else {
		version = IPv6
		maxPrefix = 128
	}

	if prefix < 0 || prefix > maxPrefix {
		return nil, fmt.Errorf("invalid prefix length %d for %s, must be 0-%d", prefix, version, maxPrefix)
	}

	result := &CIDR{
		IP:      ip,
		Prefix:  prefix,
		Version: version,
	}

	network := result.Network()
	if !ip.Equal(network) {
		return nil, fmt.Errorf("IP address %s is not properly aligned with prefix /%d, expected network address %s", ipStr, prefix, network.String())
	}

	return result, nil
}

func MustParse(cidr string) *CIDR {
	result, err := Parse(cidr)
	if err != nil {
		panic(err)
	}
	return result
}

func ParseIP(ipStr string) (net.IP, IPVersion, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, 0, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	var version IPVersion
	if ip.To4() != nil {
		version = IPv4
		ip = ip.To4()
	} else {
		version = IPv6
	}

	return ip, version, nil
}

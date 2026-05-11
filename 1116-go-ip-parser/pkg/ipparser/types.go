package ipparser

import "go-ip-parser/pkg/common"

type Address struct {
	version common.IPVersion
	ipv4    [4]byte
	ipv6    [16]byte
	isIPv4Mapped bool
	isIPv4Compatible bool
}

func (a *Address) Version() common.IPVersion {
	return a.version
}

func (a *Address) IsIPv4() bool {
	return a.version == common.IPv4
}

func (a *Address) IsIPv6() bool {
	return a.version == common.IPv6
}

func (a *Address) IPv4Octets() [4]byte {
	return a.ipv4
}

func (a *Address) IPv6Bytes() [16]byte {
	return a.ipv6
}

func (a *Address) IPv6Groups() [8]uint16 {
	var groups [8]uint16
	for i := 0; i < 8; i++ {
		groups[i] = uint16(a.ipv6[i*2])<<8 | uint16(a.ipv6[i*2+1])
	}
	return groups
}

func (a *Address) IsIPv4Mapped() bool {
	return a.isIPv4Mapped
}

func (a *Address) IsIPv4Compatible() bool {
	return a.isIPv4Compatible
}

func (a *Address) Equal(other *Address) bool {
	if a.version != other.version {
		return false
	}
	if a.version == common.IPv4 {
		return a.ipv4 == other.ipv4
	}
	return a.ipv6 == other.ipv6
}

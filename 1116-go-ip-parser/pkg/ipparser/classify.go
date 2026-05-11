package ipparser

import (
	"go-ip-parser/pkg/common"
)

func (a *Address) GetIPv4Class() common.IPv4Class {
	if !a.IsIPv4() {
		return common.ClassUnknown
	}

	octets := a.IPv4Octets()
	first := octets[0]

	if first&0x80 == 0 {
		return common.ClassA
	} else if first&0xC0 == 0x80 {
		return common.ClassB
	} else if first&0xE0 == 0xC0 {
		return common.ClassC
	} else if first&0xF0 == 0xE0 {
		return common.ClassD
	} else {
		return common.ClassE
	}
}

func (a *Address) IsLoopback() bool {
	if a.IsIPv4() {
		octets := a.IPv4Octets()
		return octets[0] == 127
	}

	if a.IsIPv6() {
		return a.ipv6 == [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	}
	return false
}

func (a *Address) IsPrivate() bool {
	if a.IsIPv4() {
		octets := a.IPv4Octets()

		if octets[0] == 10 {
			return true
		}

		if octets[0] == 172 && octets[1] >= 16 && octets[1] <= 31 {
			return true
		}

		if octets[0] == 192 && octets[1] == 168 {
			return true
		}

		if octets[0] == 169 && octets[1] == 254 {
			return true
		}

		return false
	}

	if a.IsIPv6() {
		if a.IsLinkLocal() {
			return true
		}
		if a.IsUniqueLocal() {
			return true
		}
		if a.IsLoopback() {
			return true
		}
		return false
	}
	return false
}

func (a *Address) IsMulticast() bool {
	if a.IsIPv4() {
		octets := a.IPv4Octets()
		return octets[0] >= 224 && octets[0] <= 239
	}

	if a.IsIPv6() {
		return (a.ipv6[0] & 0xFF) == 0xFF
	}
	return false
}

func (a *Address) IsLinkLocal() bool {
	if a.IsIPv4() {
		octets := a.IPv4Octets()
		return octets[0] == 169 && octets[1] == 254
	}

	if a.IsIPv6() {
		return (a.ipv6[0] == 0xFE) && ((a.ipv6[1] & 0xC0) == 0x80)
	}
	return false
}

func (a *Address) IsUniqueLocal() bool {
	if !a.IsIPv6() {
		return false
	}
	return (a.ipv6[0] == 0xFC) || (a.ipv6[0] == 0xFD)
}

func (a *Address) IsUnspecified() bool {
	if a.IsIPv4() {
		return a.IPv4Octets() == [4]byte{0, 0, 0, 0}
	}

	if a.IsIPv6() {
		return a.ipv6 == [16]byte{}
	}
	return false
}

func (a *Address) GetIPv6Type() common.IPv6Type {
	if !a.IsIPv6() {
		return ""
	}

	if a.IsUnspecified() {
		return common.IPv6TypeUnspecified
	}

	if a.IsLoopback() {
		return common.IPv6TypeLoopback
	}

	if a.isIPv4Mapped {
		return common.IPv6TypeIPv4Mapped
	}

	if a.isIPv4Compatible {
		return common.IPv6TypeIPv4Compatible
	}

	if a.IsLinkLocal() {
		return common.IPv6TypeLinkLocal
	}

	if a.IsUniqueLocal() {
		return common.IPv6TypeUniqueLocal
	}

	if a.IsMulticast() {
		return common.IPv6TypeMulticast
	}

	return common.IPv6TypeGlobal
}

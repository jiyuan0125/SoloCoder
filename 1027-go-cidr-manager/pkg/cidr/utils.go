package cidr

import (
	"encoding/binary"
	"net"
)

func ipToUint32(ip net.IP) uint32 {
	if len(ip) == 16 {
		ip = ip[12:16]
	}
	return binary.BigEndian.Uint32(ip)
}

func uint32ToIP(n uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, n)
	return ip
}

func compareIP(a, b net.IP) int {
	an := ipToUint32(a)
	bn := ipToUint32(b)
	if an < bn {
		return -1
	} else if an > bn {
		return 1
	}
	return 0
}

func nextNetwork(c *CIDR) net.IP {
	start := ipToUint32(c.Network())
	size := uint32(c.Size())
	return uint32ToIP(start + size)
}

func (list CIDRList) Sort() {
	n := len(list)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if compareIP(list[j].Network(), list[j+1].Network()) > 0 {
				list[j], list[j+1] = list[j+1], list[j]
			}
		}
	}
}

func (c *CIDR) canMergeWith(other *CIDR) bool {
	if c.Version != other.Version {
		return false
	}
	if c.Prefix != other.Prefix {
		return false
	}

	cStart := ipToUint32(c.Network())
	otherStart := ipToUint32(other.Network())
	size := uint32(c.Size())

	if cStart > otherStart {
		cStart, otherStart = otherStart, cStart
	}

	if otherStart-cStart != size {
		return false
	}

	xor := cStart ^ otherStart
	if xor != size {
		return false
	}

	if cStart&size != 0 {
		return false
	}

	return true
}

func (c *CIDR) mergeWith(other *CIDR) (*CIDR, error) {
	if !c.canMergeWith(other) {
		return nil, nil
	}

	cStart := ipToUint32(c.Network())
	otherStart := ipToUint32(other.Network())

	var newStart uint32
	if cStart < otherStart {
		newStart = cStart
	} else {
		newStart = otherStart
	}

	newPrefix := c.Prefix - 1
	newIP := uint32ToIP(newStart)

	result := &CIDR{
		IP:      newIP,
		Prefix:  newPrefix,
		Version: c.Version,
	}

	return result, nil
}

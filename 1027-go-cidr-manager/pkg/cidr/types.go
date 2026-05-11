package cidr

import (
	"fmt"
	"math"
	"net"
)

type IPVersion int

const (
	IPv4 IPVersion = 4
	IPv6 IPVersion = 6
)

type CIDR struct {
	IP      net.IP
	Prefix  int
	Version IPVersion
}

type CIDRList []*CIDR

type Pool struct {
	Available CIDRList
	Used      CIDRList
}

func (c *CIDR) String() string {
	return fmt.Sprintf("%s/%d", c.IP.String(), c.Prefix)
}

func (c *CIDR) Size() uint64 {
	if c.Version == IPv4 {
		hostBits := 32 - c.Prefix
		if hostBits < 0 {
			return 0
		}
		return uint64(math.Pow(2, float64(hostBits)))
	}
	hostBits := 128 - c.Prefix
	if hostBits < 0 {
		return 0
	}
	return uint64(math.Pow(2, float64(hostBits)))
}

func (c *CIDR) Network() net.IP {
	mask := net.CIDRMask(c.Prefix, 32)
	if c.Version == IPv6 {
		mask = net.CIDRMask(c.Prefix, 128)
	}
	return c.IP.Mask(mask)
}

func (c *CIDR) Broadcast() net.IP {
	if c.Prefix == 32 || c.Prefix == 128 {
		return c.IP
	}

	bcast := make(net.IP, len(c.IP))
	copy(bcast, c.IP)
	mask := net.CIDRMask(c.Prefix, 32)
	if c.Version == IPv6 {
		mask = net.CIDRMask(c.Prefix, 128)
	}
	for i := range mask {
		bcast[i] = c.IP[i] | ^mask[i]
	}
	return bcast
}

func (c *CIDR) FirstUsable() net.IP {
	if c.Prefix == 32 || c.Prefix == 128 {
		return c.IP
	}
	if c.Prefix == 31 && c.Version == IPv4 {
		return c.Network()
	}
	if c.Prefix == 127 && c.Version == IPv6 {
		return c.Network()
	}

	network := c.Network()
	result := make(net.IP, len(network))
	copy(result, network)
	if c.Version == IPv4 {
		result[3]++
	} else {
		result[15]++
	}
	return result
}

func (c *CIDR) LastUsable() net.IP {
	if c.Prefix == 32 || c.Prefix == 128 {
		return c.IP
	}
	if c.Prefix == 31 && c.Version == IPv4 {
		return c.Broadcast()
	}
	if c.Prefix == 127 && c.Version == IPv6 {
		return c.Broadcast()
	}

	bcast := c.Broadcast()
	result := make(net.IP, len(bcast))
	copy(result, bcast)
	if c.Version == IPv4 {
		result[3]--
	} else {
		result[15]--
	}
	return result
}

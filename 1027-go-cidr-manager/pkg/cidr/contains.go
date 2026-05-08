package cidr

import (
	"net"
)

func (c *CIDR) Contains(ip net.IP) bool {
	network := c.Network()
	mask := net.CIDRMask(c.Prefix, 32)
	if c.Version == IPv6 {
		mask = net.CIDRMask(c.Prefix, 128)
	}
	maskedIP := ip.Mask(mask)
	return network.Equal(maskedIP)
}

func (c *CIDR) ContainsCIDR(other *CIDR) bool {
	if c.Version != other.Version {
		return false
	}
	if other.Prefix < c.Prefix {
		return false
	}
	return c.Contains(other.Network())
}

func (c *CIDR) Overlaps(other *CIDR) bool {
	if c.Version != other.Version {
		return false
	}
	return c.Contains(other.Network()) || other.Contains(c.Network())
}

func FindContainingCIDRs(ipStr string, cidrs CIDRList) (CIDRList, error) {
	ip, _, err := ParseIP(ipStr)
	if err != nil {
		return nil, err
	}

	var result CIDRList
	for _, c := range cidrs {
		if c.Contains(ip) {
			result = append(result, c)
		}
	}
	return result, nil
}

package cidr

import (
	"fmt"
)

func NewPool() *Pool {
	return &Pool{
		Available: make(CIDRList, 0),
		Used:      make(CIDRList, 0),
	}
}

func (p *Pool) AddAvailable(cidr *CIDR) {
	p.Available = append(p.Available, cidr)
	p.Available.Sort()
}

func (p *Pool) RemoveAvailable(cidr *CIDR) {
	for i, c := range p.Available {
		if c.String() == cidr.String() {
			p.Available = append(p.Available[:i], p.Available[i+1:]...)
			break
		}
	}
}

func (p *Pool) AddUsed(cidr *CIDR) {
	p.Used = append(p.Used, cidr)
	p.Used.Sort()
}

func (p *Pool) RemoveUsed(cidr *CIDR) {
	for i, c := range p.Used {
		if c.String() == cidr.String() {
			p.Used = append(p.Used[:i], p.Used[i+1:]...)
			break
		}
	}
}

func (p *Pool) Exclude(used *CIDR) error {
	var newAvailable CIDRList

	for _, available := range p.Available {
		parts := splitCIDRByExclusion(available, used)
		newAvailable = append(newAvailable, parts...)
	}

	p.Available = newAvailable
	p.Available.Sort()
	p.AddUsed(used)
	return nil
}

func splitCIDRByExclusion(available, exclude *CIDR) CIDRList {
	if !available.Overlaps(exclude) {
		return CIDRList{available}
	}

	if exclude.ContainsCIDR(available) {
		return CIDRList{}
	}

	var result CIDRList
	splitAndExcludeRecursive(available, exclude, &result)
	return result
}

func splitAndExcludeRecursive(current, exclude *CIDR, result *CIDRList) {
	if !current.Overlaps(exclude) {
		*result = append(*result, current)
		return
	}

	if exclude.ContainsCIDR(current) {
		return
	}

	if current.Prefix >= 32 {
		return
	}

	subnets := splitCIDRIntoTwo(current)
	splitAndExcludeRecursive(subnets[0], exclude, result)
	splitAndExcludeRecursive(subnets[1], exclude, result)
}

func splitCIDRIntoTwo(c *CIDR) [2]*CIDR {
	newPrefix := c.Prefix + 1
	start := ipToUint32(c.Network())
	halfSize := uint32(1) << (32 - newPrefix)

	return [2]*CIDR{
		{
			IP:      uint32ToIP(start),
			Prefix:  newPrefix,
			Version: c.Version,
		},
		{
			IP:      uint32ToIP(start + halfSize),
			Prefix:  newPrefix,
			Version: c.Version,
		},
	}
}

func createCIDRFromRange(start, end uint32) (*CIDR, error) {
	if start > end {
		return nil, fmt.Errorf("invalid range")
	}

	length := end - start + 1

	var prefix int
	for prefix = 32; prefix >= 0; prefix-- {
		size := uint32(1) << (32 - prefix)
		if size == length && (start&(size-1)) == 0 {
			break
		}
	}

	if prefix < 0 {
		return nil, fmt.Errorf("cannot create CIDR from range")
	}

	return &CIDR{
		IP:      uint32ToIP(start),
		Prefix:  prefix,
		Version: IPv4,
	}, nil
}

func (p *Pool) Allocate(requestedPrefix int) (*CIDR, error) {
	if requestedPrefix < 0 || requestedPrefix > 32 {
		return nil, fmt.Errorf("invalid prefix length: %d", requestedPrefix)
	}

	for i, available := range p.Available {
		if available.Prefix > requestedPrefix {
			continue
		}

		if available.Prefix == requestedPrefix {
			allocated := available
			p.Available = append(p.Available[:i], p.Available[i+1:]...)
			p.AddUsed(allocated)
			return allocated, nil
		}

		if available.Prefix < requestedPrefix {
			subnets := splitCIDR(available, requestedPrefix)
			if len(subnets) == 0 {
				continue
			}

			allocated := subnets[0]
			p.Available = append(p.Available[:i], p.Available[i+1:]...)
			if len(subnets) > 1 {
				p.Available = append(p.Available, subnets[1:]...)
				p.Available.Sort()
			}
			p.AddUsed(allocated)
			return allocated, nil
		}
	}

	return nil, fmt.Errorf("no available CIDR block for prefix /%d", requestedPrefix)
}

func splitCIDR(c *CIDR, targetPrefix int) CIDRList {
	if targetPrefix < c.Prefix {
		return CIDRList{c}
	}

	if targetPrefix == c.Prefix {
		return CIDRList{c}
	}

	start := ipToUint32(c.Network())
	size := uint32(1) << (32 - targetPrefix)
	numSubnets := 1 << (targetPrefix - c.Prefix)

	var result CIDRList
	for i := 0; i < numSubnets; i++ {
		subnetStart := start + uint32(i)*size
		result = append(result, &CIDR{
			IP:      uint32ToIP(subnetStart),
			Prefix:  targetPrefix,
			Version: c.Version,
		})
	}

	return result
}

func (p *Pool) Release(released *CIDR) error {
	found := false
	for i, used := range p.Used {
		if used.String() == released.String() {
			p.Used = append(p.Used[:i], p.Used[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("CIDR %s not found in used pool", released.String())
	}

	p.Available = append(p.Available, released)
	p.Available.Sort()
	p.Available = Merge(p.Available)

	return nil
}

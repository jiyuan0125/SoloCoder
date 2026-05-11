package ipparser

import (
	"fmt"
	"strings"
)

func formatIPv4(addr *Address) string {
	octets := addr.IPv4Octets()
	return fmt.Sprintf("%d.%d.%d.%d", octets[0], octets[1], octets[2], octets[3])
}

func formatIPv6Standard(addr *Address) string {
	groups := addr.IPv6Groups()
	return fmt.Sprintf("%x:%x:%x:%x:%x:%x:%x:%x",
		groups[0], groups[1], groups[2], groups[3],
		groups[4], groups[5], groups[6], groups[7])
}

func formatIPv6Canonical(addr *Address) string {
	if addr.isIPv4Mapped {
		octets := addr.IPv4OctetsFromIPv6()
		return fmt.Sprintf("::ffff:%d.%d.%d.%d", octets[0], octets[1], octets[2], octets[3])
	}

	if addr.isIPv4Compatible {
		octets := addr.IPv4OctetsFromIPv6()
		return fmt.Sprintf("::%d.%d.%d.%d", octets[0], octets[1], octets[2], octets[3])
	}

	groups := addr.IPv6Groups()

	bestStart := -1
	bestLen := 0

	for i := 0; i < 8; i++ {
		if groups[i] == 0 {
			j := i
			for j < 8 && groups[j] == 0 {
				j++
			}
			runLen := j - i
			if runLen > bestLen && runLen >= 2 {
				bestStart = i
				bestLen = runLen
			}
			i = j
		}
	}

	if bestLen >= 2 {
		var parts []string

		for i := 0; i < bestStart; i++ {
			parts = append(parts, fmt.Sprintf("%x", groups[i]))
		}

		parts = append(parts, "")

		for i := bestStart + bestLen; i < 8; i++ {
			parts = append(parts, fmt.Sprintf("%x", groups[i]))
		}

		result := strings.Join(parts, ":")
		if result == "" || result == ":" {
			return "::"
		}
		if strings.HasPrefix(result, ":") {
			return ":" + result
		}
		if strings.HasSuffix(result, ":") {
			return result + ":"
		}
		return result
	}

	var parts []string
	for _, g := range groups {
		parts = append(parts, fmt.Sprintf("%x", g))
	}
	return strings.Join(parts, ":")
}

func (a *Address) IPv4OctetsFromIPv6() [4]byte {
	var octets [4]byte
	octets[0] = a.ipv6[12]
	octets[1] = a.ipv6[13]
	octets[2] = a.ipv6[14]
	octets[3] = a.ipv6[15]
	return octets
}

func (a *Address) Standard() string {
	if a.version == "" {
		return ""
	}
	if a.IsIPv4() {
		return formatIPv4(a)
	}
	return formatIPv6Standard(a)
}

func (a *Address) Canonical() string {
	if a.version == "" {
		return ""
	}
	if a.IsIPv4() {
		return formatIPv4(a)
	}
	return formatIPv6Canonical(a)
}

func (a *Address) String() string {
	return a.Canonical()
}

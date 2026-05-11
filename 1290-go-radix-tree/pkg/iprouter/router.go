package iprouter

import (
	"encoding/binary"
	"net"

	"github.com/example/radix-router/pkg/radix"
)

type IPRoute struct {
	CIDR     string
	Nexthop  string
	Network  net.IPNet
	PrefixLen int
	IsIPv6   bool
}

type IPMatchResult struct {
	Route  IPRoute
	Found  bool
	MatchedCIDR string
}

type Router struct {
	v4Tree *radix.Tree
	v6Tree *radix.Tree
}

func NewRouter() *Router {
	return &Router{
		v4Tree: radix.NewTree(),
		v6Tree: radix.NewTree(),
	}
}

func (r *Router) AddRoute(cidr, nexthop string) error {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	route := IPRoute{
		CIDR:     cidr,
		Nexthop:  nexthop,
		Network:  *network,
		PrefixLen: 0,
	}

	if network.IP.To4() != nil {
		route.IsIPv6 = false
		bits, _ := network.Mask.Size()
		key := ipToBitString(network.IP.To4(), bits)
		r.v4Tree.Insert(key, route)
	} else {
		route.IsIPv6 = true
		bits, _ := network.Mask.Size()
		key := ipToBitString(network.IP.To16(), bits)
		r.v6Tree.Insert(key, route)
	}

	return nil
}

func (r *Router) DeleteRoute(cidr string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	bits, _ := network.Mask.Size()
	if network.IP.To4() != nil {
		key := ipToBitString(network.IP.To4(), bits)
		return r.v4Tree.Delete(key)
	} else {
		key := ipToBitString(network.IP.To16(), bits)
		return r.v6Tree.Delete(key)
	}
}

func (r *Router) Match(ipStr string) IPMatchResult {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return IPMatchResult{Found: false}
	}

	var tree *radix.Tree
	var fullKey string
	if ip.To4() != nil {
		tree = r.v4Tree
		fullKey = ipToBitString(ip.To4(), 32)
	} else {
		tree = r.v6Tree
		fullKey = ipToBitString(ip.To16(), 128)
	}

	var bestRoute IPRoute
	var bestCIDR string
	var bestLen int
	var found bool

	for i := len(fullKey); i >= 0; i-- {
		prefix := fullKey[:i]
		value, _, foundNow := tree.Search(prefix)
		if foundNow {
			if route, ok := value.(IPRoute); ok {
				prefixLen := len(prefix)
				if prefixLen >= bestLen {
					bestLen = prefixLen
					bestRoute = route
					bestCIDR = route.CIDR
					found = true
				}
			}
		}
	}

	if found {
		return IPMatchResult{
			Route:        bestRoute,
			Found:        true,
			MatchedCIDR:  bestCIDR,
		}
	}

	return IPMatchResult{Found: false}
}

func (r *Router) List() []IPRoute {
	var routes []IPRoute

	for _, key := range r.v4Tree.List() {
		if value, _, found := r.v4Tree.Search(key); found {
			if route, ok := value.(IPRoute); ok {
				routes = append(routes, route)
			}
		}
	}

	for _, key := range r.v6Tree.List() {
		if value, _, found := r.v6Tree.Search(key); found {
			if route, ok := value.(IPRoute); ok {
				routes = append(routes, route)
			}
		}
	}

	return routes
}

func (r *Router) Stats() map[string]radix.Stats {
	stats := make(map[string]radix.Stats)
	stats["ipv4"] = r.v4Tree.Stats()
	stats["ipv6"] = r.v6Tree.Stats()
	return stats
}

func ipToBitString(ip []byte, bits int) string {
	result := make([]byte, 0, bits)

	for _, b := range ip {
		for i := 7; i >= 0; i-- {
			if len(result) >= bits {
				break
			}
			if (b>>i)&1 == 1 {
				result = append(result, '1')
			} else {
				result = append(result, '0')
			}
		}
	}

	return string(result)
}

func ipToUint32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

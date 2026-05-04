package iplocation

import "fmt"

type Location struct {
	Country string
	Province string
	City string
}

func (l Location) String() string {
	return fmt.Sprintf("%s %s %s", l.Country, l.Province, l.City)
}

type IPRange struct {
	StartIP uint32
	EndIP uint32
	Location Location
}

func (r IPRange) Size() uint32 {
	return r.EndIP - r.StartIP + 1
}

type IPType int

const (
	IPTypeInvalid IPType = iota
	IPTypeA
	IPTypeB
	IPTypeC
	IPTypeD
	IPTypeE
	IPTypeSpecial
)

func (t IPType) String() string {
	switch t {
	case IPTypeA:
		return "A类"
	case IPTypeB:
		return "B类"
	case IPTypeC:
		return "C类"
	case IPTypeD:
		return "D类"
	case IPTypeE:
		return "E类"
	case IPTypeSpecial:
		return "特殊地址"
	default:
		return "无效"
	}
}

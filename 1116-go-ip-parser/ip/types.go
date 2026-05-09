package ip

type IPVersion int

const (
	IPv4 IPVersion = iota
	IPv6
)

type IPv6MappingType int

const (
	IPv6MappingNone IPv6MappingType = iota
	IPv6MappingIPv4
	IPv6MappingCompatible
)

type IP interface {
	Version() IPVersion
	ToBytes() []byte
	ToInt64() uint32
	ToInt128() [2]uint64
	IsLoopback() bool
	IsPrivate() bool
	IsMulticast() bool
	IsBroadcast() bool
	IsLinkLocal() bool
	IsAnycast() bool
	IsUnspecified() bool
	String() string
}

type IPv4Address struct {
	bytes [4]byte
}

type IPv6Address struct {
	bytes [16]byte
	mappingType IPv6MappingType
}

type IPv4Class int

const (
	IPv4ClassA IPv4Class = iota
	IPv4ClassB
	IPv4ClassC
	IPv4ClassD
	IPv4ClassE
	IPv4ClassUnknown
)

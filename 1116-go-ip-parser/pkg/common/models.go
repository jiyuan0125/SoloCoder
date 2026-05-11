package common

type IPVersion string

const (
	IPv4 IPVersion = "IPv4"
	IPv6 IPVersion = "IPv6"
)

type IPv4Class string

const (
	ClassA   IPv4Class = "A"
	ClassB   IPv4Class = "B"
	ClassC   IPv4Class = "C"
	ClassD   IPv4Class = "D"
	ClassE   IPv4Class = "E"
	ClassUnknown IPv4Class = "Unknown"
)

type IPv6Type string

const (
	IPv6TypeUnspecified   IPv6Type = "Unspecified"
	IPv6TypeLoopback      IPv6Type = "Loopback"
	IPv6TypeLinkLocal     IPv6Type = "LinkLocal"
	IPv6TypeUniqueLocal   IPv6Type = "UniqueLocal"
	IPv6TypeMulticast     IPv6Type = "Multicast"
	IPv6TypeGlobal        IPv6Type = "Global"
	IPv6TypeIPv4Mapped    IPv6Type = "IPv4Mapped"
	IPv6TypeIPv4Compatible IPv6Type = "IPv4Compatible"
)

type ParseRequest struct {
	IP string `json:"ip"`
}

type ParseResponse struct {
	Success      bool       `json:"success"`
	Error        string     `json:"error,omitempty"`
	Original     string     `json:"original,omitempty"`
	Version      IPVersion  `json:"version,omitempty"`
	Standard     string     `json:"standard,omitempty"`
	Canonical    string     `json:"canonical,omitempty"`
	IsLoopback   bool       `json:"is_loopback,omitempty"`
	IsPrivate    bool       `json:"is_private,omitempty"`
	IsMulticast  bool       `json:"is_multicast,omitempty"`
	IsLinkLocal  bool       `json:"is_link_local,omitempty"`
	IsUnspecified bool      `json:"is_unspecified,omitempty"`
	IPv4Class    *IPv4Class `json:"ipv4_class,omitempty"`
	IPv6Type     *IPv6Type  `json:"ipv6_type,omitempty"`
	IPv4Octets   [4]uint8   `json:"ipv4_octets,omitempty"`
	IPv6Groups   [8]uint16  `json:"ipv6_groups,omitempty"`
}

type FormatRequest struct {
	IP      string `json:"ip"`
	Compact bool   `json:"compact,omitempty"`
}

type FormatResponse struct {
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Original string `json:"original,omitempty"`
	Formatted string `json:"formatted,omitempty"`
	Compact  string `json:"compact,omitempty"`
}

type ClassifyRequest struct {
	IP string `json:"ip"`
}

type ClassifyResponse struct {
	Success      bool       `json:"success"`
	Error        string     `json:"error,omitempty"`
	IP           string     `json:"ip,omitempty"`
	Version      IPVersion  `json:"version,omitempty"`
	IPv4Class    *IPv4Class `json:"ipv4_class,omitempty"`
	IPv6Type     *IPv6Type  `json:"ipv6_type,omitempty"`
	IsLoopback   bool       `json:"is_loopback,omitempty"`
	IsPrivate    bool       `json:"is_private,omitempty"`
	IsMulticast  bool       `json:"is_multicast,omitempty"`
	IsLinkLocal  bool       `json:"is_link_local,omitempty"`
	IsUnspecified bool      `json:"is_unspecified,omitempty"`
	Description  string     `json:"description,omitempty"`
}

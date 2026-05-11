package api

import (
	"time"
)

type ParseRequest struct {
	Hex string `json:"hex"`
}

type OptionInfo struct {
	Code   byte   `json:"code"`
	Length byte   `json:"length"`
	Value  string `json:"value"`
}

type PacketInfo struct {
	Op          byte         `json:"op"`
	OpName      string       `json:"op_name"`
	Htype       byte         `json:"htype"`
	Hlen        byte         `json:"hlen"`
	Hops        byte         `json:"hops"`
	Xid         uint32       `json:"xid"`
	Secs        uint16       `json:"secs"`
	Flags       uint16       `json:"flags"`
	Broadcast   bool         `json:"broadcast"`
	Ciaddr      string       `json:"ciaddr"`
	Yiaddr      string       `json:"yiaddr"`
	Siaddr      string       `json:"siaddr"`
	Giaddr      string       `json:"giaddr"`
	Chaddr      string       `json:"chaddr"`
	MessageType byte         `json:"message_type,omitempty"`
	MessageName string       `json:"message_name,omitempty"`
	Options     []OptionInfo `json:"options"`
}

type ParseResponse struct {
	Success bool       `json:"success"`
	Error   string     `json:"error,omitempty"`
	Packet  PacketInfo `json:"packet,omitempty"`
}

type BuildRequest struct {
	Op            byte              `json:"op"`
	Htype         byte              `json:"htype,omitempty"`
	Hlen          byte              `json:"hlen,omitempty"`
	Hops          byte              `json:"hops,omitempty"`
	Xid           uint32            `json:"xid,omitempty"`
	Secs          uint16            `json:"secs,omitempty"`
	Flags         uint16            `json:"flags,omitempty"`
	Broadcast     bool              `json:"broadcast,omitempty"`
	Ciaddr        string            `json:"ciaddr,omitempty"`
	Yiaddr        string            `json:"yiaddr,omitempty"`
	Siaddr        string            `json:"siaddr,omitempty"`
	Giaddr        string            `json:"giaddr,omitempty"`
	Chaddr        string            `json:"chaddr"`
	MessageType   byte              `json:"message_type"`
	ServerIP      string            `json:"server_ip,omitempty"`
	LeaseTime     uint32            `json:"lease_time,omitempty"`
	RequestedIP   string            `json:"requested_ip,omitempty"`
	SubnetMask    string            `json:"subnet_mask,omitempty"`
	Routers       []string          `json:"routers,omitempty"`
	DNSServers    []string          `json:"dns_servers,omitempty"`
	HostName      string            `json:"host_name,omitempty"`
	CustomOptions map[byte][]byte   `json:"custom_options,omitempty"`
}

type BuildResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Hex     string `json:"hex,omitempty"`
	Bytes   []byte `json:"bytes,omitempty"`
}

type IPRangeRequest struct {
	StartIP string `json:"start_ip"`
	EndIP   string `json:"end_ip"`
}

type IPRangeResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type PoolStatusResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	Total     int    `json:"total"`
	Available int    `json:"available"`
	InUse     int    `json:"in_use"`
}

type LeaseInfo struct {
	MAC         string    `json:"mac"`
	IP          string    `json:"ip"`
	StartTime   time.Time `json:"start_time"`
	Duration    uint32    `json:"duration"`
	ExpireTime  time.Time `json:"expire_time"`
	T1Time      time.Time `json:"t1_time"`
	T2Time      time.Time `json:"t2_time"`
	ServerIP    string    `json:"server_ip"`
	IsExpired   bool      `json:"is_expired"`
	IsT1Reached bool      `json:"is_t1_reached"`
	IsT2Reached bool      `json:"is_t2_reached"`
}

type LeaseListResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Leases  []LeaseInfo `json:"leases,omitempty"`
}

type LeaseResponse struct {
	Success bool      `json:"success"`
	Error   string    `json:"error,omitempty"`
	Lease   LeaseInfo `json:"lease,omitempty"`
}

type DORARequest struct {
	MAC string `json:"mac"`
}

type DORAResponse struct {
	Success    bool       `json:"success"`
	Error      string     `json:"error,omitempty"`
	AssignedIP string     `json:"assigned_ip,omitempty"`
	Lease      *LeaseInfo `json:"lease,omitempty"`
}

type ServerConfigRequest struct {
	ServerIP   string   `json:"server_ip"`
	SubnetMask string   `json:"subnet_mask"`
	Routers    []string `json:"routers,omitempty"`
	DNSServers []string `json:"dns_servers,omitempty"`
	LeaseTime  uint32   `json:"lease_time,omitempty"`
}

type ServerConfigResponse struct {
	Success bool `json:"success"`
	Error   string `json:"error,omitempty"`
}

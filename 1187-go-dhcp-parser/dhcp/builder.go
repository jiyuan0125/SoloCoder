package dhcp

import (
	"crypto/rand"
	"encoding/binary"
	"net"
)

type Builder struct {
	p *Packet
}

func NewBuilder(op byte) *Builder {
	return &Builder{
		p: NewPacket(op),
	}
}

func (b *Builder) WithXid(xid uint32) *Builder {
	b.p.Xid = xid
	return b
}

func (b *Builder) WithRandomXid() *Builder {
	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	b.p.Xid = binary.BigEndian.Uint32(buf)
	return b
}

func (b *Builder) WithHtype(htype byte) *Builder {
	b.p.Htype = htype
	return b
}

func (b *Builder) WithHlen(hlen byte) *Builder {
	b.p.Hlen = hlen
	return b
}

func (b *Builder) WithHops(hops byte) *Builder {
	b.p.Hops = hops
	return b
}

func (b *Builder) WithSecs(secs uint16) *Builder {
	b.p.Secs = secs
	return b
}

func (b *Builder) WithBroadcast(broadcast bool) *Builder {
	b.p.SetBroadcast(broadcast)
	return b
}

func (b *Builder) WithCiaddr(ip net.IP) *Builder {
	b.p.Ciaddr = ip
	return b
}

func (b *Builder) WithYiaddr(ip net.IP) *Builder {
	b.p.Yiaddr = ip
	return b
}

func (b *Builder) WithSiaddr(ip net.IP) *Builder {
	b.p.Siaddr = ip
	return b
}

func (b *Builder) WithGiaddr(ip net.IP) *Builder {
	b.p.Giaddr = ip
	return b
}

func (b *Builder) WithChaddr(mac net.HardwareAddr) *Builder {
	b.p.Chaddr = mac
	return b
}

func (b *Builder) WithSname(sname []byte) *Builder {
	b.p.Sname = sname
	return b
}

func (b *Builder) WithFile(file []byte) *Builder {
	b.p.File = file
	return b
}

func (b *Builder) WithMessageType(msgType byte) *Builder {
	b.p.Options.SetMessageType(msgType)
	return b
}

func (b *Builder) WithServerIdentifier(ip net.IP) *Builder {
	b.p.Options.SetServerIdentifier(ip)
	return b
}

func (b *Builder) WithLeaseTime(seconds uint32) *Builder {
	b.p.Options.SetLeaseTime(seconds)
	return b
}

func (b *Builder) WithRequestedIP(ip net.IP) *Builder {
	b.p.Options.SetRequestedIP(ip)
	return b
}

func (b *Builder) WithSubnetMask(mask net.IPMask) *Builder {
	b.p.Options.SetSubnetMask(mask)
	return b
}

func (b *Builder) WithRouters(ips []net.IP) *Builder {
	b.p.Options.SetRouters(ips)
	return b
}

func (b *Builder) WithDNSServers(ips []net.IP) *Builder {
	b.p.Options.SetDNSServers(ips)
	return b
}

func (b *Builder) WithHostName(name string) *Builder {
	b.p.Options.SetHostName(name)
	return b
}

func (b *Builder) WithOption(code byte, data []byte) *Builder {
	b.p.Options.Add(code, data)
	return b
}

func (b *Builder) Build() *Packet {
	return b.p
}

func BuildDiscover(mac net.HardwareAddr) *Packet {
	return NewBuilder(OpBootRequest).
		WithRandomXid().
		WithHtype(HtypeEthernet).
		WithHlen(HlenEthernet).
		WithBroadcast(true).
		WithChaddr(mac).
		WithMessageType(MsgDiscover).
		Build()
}

func BuildOffer(discover *Packet, offeredIP net.IP, serverIP net.IP, leaseTime uint32, mask net.IPMask, routers []net.IP, dns []net.IP) *Packet {
	builder := NewBuilder(OpBootReply).
		WithXid(discover.Xid).
		WithHtype(HtypeEthernet).
		WithHlen(HlenEthernet).
		WithChaddr(discover.Chaddr).
		WithYiaddr(offeredIP).
		WithSiaddr(serverIP).
		WithBroadcast(discover.IsBroadcast()).
		WithMessageType(MsgOffer).
		WithServerIdentifier(serverIP).
		WithLeaseTime(leaseTime).
		WithSubnetMask(mask)
	if len(routers) > 0 {
		builder.WithRouters(routers)
	}
	if len(dns) > 0 {
		builder.WithDNSServers(dns)
	}
	return builder.Build()
}

func BuildRequestFromDiscover(discover *Packet, offeredIP net.IP, serverIP net.IP) *Packet {
	return NewBuilder(OpBootRequest).
		WithXid(discover.Xid).
		WithHtype(HtypeEthernet).
		WithHlen(HlenEthernet).
		WithBroadcast(true).
		WithChaddr(discover.Chaddr).
		WithMessageType(MsgRequest).
		WithRequestedIP(offeredIP).
		WithServerIdentifier(serverIP).
		Build()
}

func BuildRequestForRenewal(mac net.HardwareAddr, ciaddr net.IP, xid uint32) *Packet {
	return NewBuilder(OpBootRequest).
		WithXid(xid).
		WithHtype(HtypeEthernet).
		WithHlen(HlenEthernet).
		WithChaddr(mac).
		WithCiaddr(ciaddr).
		WithMessageType(MsgRequest).
		Build()
}

func BuildAck(request *Packet, assignedIP net.IP, serverIP net.IP, leaseTime uint32, mask net.IPMask, routers []net.IP, dns []net.IP) *Packet {
	builder := NewBuilder(OpBootReply).
		WithXid(request.Xid).
		WithHtype(HtypeEthernet).
		WithHlen(HlenEthernet).
		WithChaddr(request.Chaddr).
		WithYiaddr(assignedIP).
		WithSiaddr(serverIP).
		WithBroadcast(request.IsBroadcast()).
		WithMessageType(MsgAck).
		WithServerIdentifier(serverIP).
		WithLeaseTime(leaseTime).
		WithSubnetMask(mask)
	if len(routers) > 0 {
		builder.WithRouters(routers)
	}
	if len(dns) > 0 {
		builder.WithDNSServers(dns)
	}
	return builder.Build()
}

func BuildNak(request *Packet, serverIP net.IP) *Packet {
	return NewBuilder(OpBootReply).
		WithXid(request.Xid).
		WithHtype(HtypeEthernet).
		WithHlen(HlenEthernet).
		WithChaddr(request.Chaddr).
		WithBroadcast(request.IsBroadcast()).
		WithMessageType(MsgNak).
		WithServerIdentifier(serverIP).
		Build()
}

func BuildRelease(mac net.HardwareAddr, ciaddr net.IP, serverIP net.IP) *Packet {
	return NewBuilder(OpBootRequest).
		WithRandomXid().
		WithHtype(HtypeEthernet).
		WithHlen(HlenEthernet).
		WithChaddr(mac).
		WithCiaddr(ciaddr).
		WithMessageType(MsgRelease).
		WithServerIdentifier(serverIP).
		Build()
}

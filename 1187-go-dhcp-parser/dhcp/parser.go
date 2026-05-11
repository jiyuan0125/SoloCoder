package dhcp

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	OpBootRequest byte = 1
	OpBootReply        = 2
	HtypeEthernet      = 1
	HlenEthernet       = 6
	FixedHeaderSize    = 236
)

const BroadcastFlag uint16 = 0x8000

type Packet struct {
	Op      byte
	Htype   byte
	Hlen    byte
	Hops    byte
	Xid     uint32
	Secs    uint16
	Flags   uint16
	Ciaddr  net.IP
	Yiaddr  net.IP
	Siaddr  net.IP
	Giaddr  net.IP
	Chaddr  net.HardwareAddr
	Sname   []byte
	File    []byte
	Options *Options
}

func NewPacket(op byte) *Packet {
	return &Packet{
		Op:      op,
		Htype:   HtypeEthernet,
		Hlen:    HlenEthernet,
		Ciaddr:  make(net.IP, 4),
		Yiaddr:  make(net.IP, 4),
		Siaddr:  make(net.IP, 4),
		Giaddr:  make(net.IP, 4),
		Chaddr:  make(net.HardwareAddr, 6),
		Sname:   make([]byte, 64),
		File:    make([]byte, 128),
		Options: NewOptions(),
	}
}

func Parse(data []byte) (*Packet, error) {
	if len(data) < FixedHeaderSize {
		return nil, fmt.Errorf("packet too short: %d bytes, need at least %d", len(data), FixedHeaderSize)
	}
	p := &Packet{
		Op:    data[0],
		Htype: data[1],
		Hlen:  data[2],
		Hops:  data[3],
		Xid:   binary.BigEndian.Uint32(data[4:8]),
		Secs:  binary.BigEndian.Uint16(data[8:10]),
		Flags: binary.BigEndian.Uint16(data[10:12]),
	}
	p.Ciaddr = make(net.IP, 4)
	copy(p.Ciaddr, data[12:16])
	p.Yiaddr = make(net.IP, 4)
	copy(p.Yiaddr, data[16:20])
	p.Siaddr = make(net.IP, 4)
	copy(p.Siaddr, data[20:24])
	p.Giaddr = make(net.IP, 4)
	copy(p.Giaddr, data[24:28])

	chaddrLen := int(p.Hlen)
	if chaddrLen > 16 {
		chaddrLen = 16
	}
	if chaddrLen > 0 {
		p.Chaddr = make(net.HardwareAddr, chaddrLen)
		copy(p.Chaddr, data[28:28+chaddrLen])
	} else {
		p.Chaddr = make(net.HardwareAddr, 0)
	}

	p.Sname = make([]byte, 64)
	copy(p.Sname, data[44:108])
	p.File = make([]byte, 128)
	copy(p.File, data[108:236])

	if len(data) > FixedHeaderSize {
		opts, err := ParseOptions(data[FixedHeaderSize:])
		if err != nil {
			return nil, fmt.Errorf("failed to parse options: %w", err)
		}
		p.Options = opts
	} else {
		p.Options = NewOptions()
	}
	return p, nil
}

func (p *Packet) Bytes() []byte {
	buf := make([]byte, FixedHeaderSize)
	buf[0] = p.Op
	buf[1] = p.Htype
	buf[2] = p.Hlen
	buf[3] = p.Hops
	binary.BigEndian.PutUint32(buf[4:8], p.Xid)
	binary.BigEndian.PutUint16(buf[8:10], p.Secs)
	binary.BigEndian.PutUint16(buf[10:12], p.Flags)

	copy(buf[12:16], p.Ciaddr.To4())
	copy(buf[16:20], p.Yiaddr.To4())
	copy(buf[20:24], p.Siaddr.To4())
	copy(buf[24:28], p.Giaddr.To4())

	if p.Hlen > 0 && len(p.Chaddr) >= int(p.Hlen) {
		copy(buf[28:28+int(p.Hlen)], p.Chaddr)
	}

	if len(p.Sname) > 0 {
		snameLen := len(p.Sname)
		if snameLen > 64 {
			snameLen = 64
		}
		copy(buf[44:44+snameLen], p.Sname[:snameLen])
	}
	if len(p.File) > 0 {
		fileLen := len(p.File)
		if fileLen > 128 {
			fileLen = 128
		}
		copy(buf[108:108+fileLen], p.File[:fileLen])
	}

	optsBytes := p.Options.Bytes()
	buf = append(buf, optsBytes...)
	return buf
}

func (p *Packet) IsBroadcast() bool {
	return (p.Flags & BroadcastFlag) != 0
}

func (p *Packet) SetBroadcast(broadcast bool) {
	if broadcast {
		p.Flags |= BroadcastFlag
	} else {
		p.Flags &= ^BroadcastFlag
	}
}

func (p *Packet) MessageType() (byte, error) {
	return p.Options.MessageType()
}

func OpToString(op byte) string {
	switch op {
	case OpBootRequest:
		return "BOOTREQUEST"
	case OpBootReply:
		return "BOOTREPLY"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", op)
	}
}

func MessageTypeToString(msgType byte) string {
	switch msgType {
	case MsgDiscover:
		return "DISCOVER"
	case MsgOffer:
		return "OFFER"
	case MsgRequest:
		return "REQUEST"
	case MsgDecline:
		return "DECLINE"
	case MsgAck:
		return "ACK"
	case MsgNak:
		return "NAK"
	case MsgRelease:
		return "RELEASE"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", msgType)
	}
}

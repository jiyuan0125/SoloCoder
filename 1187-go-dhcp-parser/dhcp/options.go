package dhcp

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	MagicCookie     uint32 = 0x63825363
	OptionPad       byte   = 0
	OptionSubnetMask        = 1
	OptionRouter            = 3
	OptionDNS               = 6
	OptionHostName          = 12
	OptionRequestedIP       = 50
	OptionLeaseTime         = 51
	OptionMsgType           = 53
	OptionServerID          = 54
	OptionRelayAgent        = 82
	OptionEnd               = 255
)

const (
	MsgDiscover byte = 1
	MsgOffer         = 2
	MsgRequest       = 3
	MsgDecline       = 4
	MsgAck           = 5
	MsgNak           = 6
	MsgRelease       = 7
)

type Option struct {
	Code   byte
	Length byte
	Data   []byte
}

type SubOption struct {
	Code   byte
	Length byte
	Data   []byte
}

type Options struct {
	raw       []Option
	subAgents []Option
}

func NewOptions() *Options {
	return &Options{
		raw: make([]Option, 0),
	}
}

func ParseOptions(data []byte) (*Options, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("options too short")
	}
	cookie := binary.BigEndian.Uint32(data[0:4])
	if cookie != MagicCookie {
		return nil, fmt.Errorf("invalid magic cookie: 0x%x", cookie)
	}
	opts := NewOptions()
	idx := 4
	for idx < len(data) {
		code := data[idx]
		if code == OptionEnd {
			break
		}
		if code == OptionPad {
			idx++
			continue
		}
		if idx+1 >= len(data) {
			break
		}
		length := data[idx+1]
		start := idx + 2
		end := start + int(length)
		if end > len(data) {
			break
		}
		optData := make([]byte, length)
		copy(optData, data[start:end])
		opt := Option{
			Code:   code,
			Length: length,
			Data:   optData,
		}
		opts.raw = append(opts.raw, opt)
		if code == OptionRelayAgent {
			opts.subAgents = append(opts.subAgents, opt)
		}
		idx = end
	}
	return opts, nil
}

func (o *Options) Bytes() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf[0:4], MagicCookie)
	for _, opt := range o.raw {
		if opt.Code == OptionPad || opt.Code == OptionEnd {
			continue
		}
		buf = append(buf, opt.Code)
		buf = append(buf, opt.Length)
		buf = append(buf, opt.Data...)
	}
	buf = append(buf, OptionEnd)
	return buf
}

func (o *Options) Add(code byte, data []byte) {
	opt := Option{
		Code:   code,
		Length: byte(len(data)),
		Data:   data,
	}
	o.raw = append(o.raw, opt)
}

func (o *Options) Get(code byte) []Option {
	result := make([]Option, 0)
	for _, opt := range o.raw {
		if opt.Code == code {
			result = append(result, opt)
		}
	}
	return result
}

func (o *Options) GetFirst(code byte) *Option {
	for i := range o.raw {
		if o.raw[i].Code == code {
			return &o.raw[i]
		}
	}
	return nil
}

func (o *Options) MessageType() (byte, error) {
	opt := o.GetFirst(OptionMsgType)
	if opt == nil || opt.Length < 1 {
		return 0, fmt.Errorf("message type option not found")
	}
	return opt.Data[0], nil
}

func (o *Options) SetMessageType(msgType byte) {
	o.Add(OptionMsgType, []byte{msgType})
}

func (o *Options) ServerIdentifier() (net.IP, error) {
	opt := o.GetFirst(OptionServerID)
	if opt == nil || opt.Length != 4 {
		return nil, fmt.Errorf("server identifier option not found")
	}
	return net.IP(opt.Data), nil
}

func (o *Options) SetServerIdentifier(ip net.IP) {
	o.Add(OptionServerID, ip.To4())
}

func (o *Options) LeaseTime() (uint32, error) {
	opt := o.GetFirst(OptionLeaseTime)
	if opt == nil || opt.Length != 4 {
		return 0, fmt.Errorf("lease time option not found")
	}
	return binary.BigEndian.Uint32(opt.Data), nil
}

func (o *Options) SetLeaseTime(seconds uint32) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, seconds)
	o.Add(OptionLeaseTime, buf)
}

func (o *Options) RequestedIP() (net.IP, error) {
	opt := o.GetFirst(OptionRequestedIP)
	if opt == nil || opt.Length != 4 {
		return nil, fmt.Errorf("requested IP option not found")
	}
	return net.IP(opt.Data), nil
}

func (o *Options) SetRequestedIP(ip net.IP) {
	o.Add(OptionRequestedIP, ip.To4())
}

func (o *Options) SubnetMask() (net.IPMask, error) {
	opt := o.GetFirst(OptionSubnetMask)
	if opt == nil || opt.Length != 4 {
		return nil, fmt.Errorf("subnet mask option not found")
	}
	return net.IPMask(opt.Data), nil
}

func (o *Options) SetSubnetMask(mask net.IPMask) {
	o.Add(OptionSubnetMask, []byte(mask))
}

func (o *Options) Routers() ([]net.IP, error) {
	opts := o.Get(OptionRouter)
	ips := make([]net.IP, 0)
	for _, opt := range opts {
		if opt.Length%4 != 0 {
			continue
		}
		for i := 0; i < int(opt.Length); i += 4 {
			ips = append(ips, net.IP(opt.Data[i:i+4]))
		}
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("router option not found")
	}
	return ips, nil
}

func (o *Options) SetRouters(ips []net.IP) {
	buf := make([]byte, 0, len(ips)*4)
	for _, ip := range ips {
		buf = append(buf, ip.To4()...)
	}
	o.Add(OptionRouter, buf)
}

func (o *Options) DNSServers() ([]net.IP, error) {
	opts := o.Get(OptionDNS)
	ips := make([]net.IP, 0)
	for _, opt := range opts {
		if opt.Length%4 != 0 {
			continue
		}
		for i := 0; i < int(opt.Length); i += 4 {
			ips = append(ips, net.IP(opt.Data[i:i+4]))
		}
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("DNS server option not found")
	}
	return ips, nil
}

func (o *Options) SetDNSServers(ips []net.IP) {
	buf := make([]byte, 0, len(ips)*4)
	for _, ip := range ips {
		buf = append(buf, ip.To4()...)
	}
	o.Add(OptionDNS, buf)
}

func (o *Options) HostName() (string, error) {
	opt := o.GetFirst(OptionHostName)
	if opt == nil {
		return "", fmt.Errorf("host name option not found")
	}
	return string(opt.Data), nil
}

func (o *Options) SetHostName(name string) {
	o.Add(OptionHostName, []byte(name))
}

func (o *Options) All() []Option {
	return o.raw
}

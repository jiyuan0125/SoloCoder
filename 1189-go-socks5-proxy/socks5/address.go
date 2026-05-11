package socks5

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

type Address struct {
	Type byte
	IP   net.IP
	Port uint16
	Name string
}

func ReadAddress(r io.Reader) (*Address, error) {
	addrType := make([]byte, 1)
	if _, err := io.ReadFull(r, addrType); err != nil {
		return nil, err
	}
	return ReadAddressWithType(r, addrType[0])
}

func ReadAddressWithType(r io.Reader, addrType byte) (*Address, error) {
	addr := &Address{Type: addrType}

	switch addr.Type {
	case AddrTypeIPv4:
		ip := make([]byte, 4)
		if _, err := io.ReadFull(r, ip); err != nil {
			return nil, err
		}
		addr.IP = net.IP(ip)

	case AddrTypeDomain:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(r, lenBuf); err != nil {
			return nil, err
		}
		nameLen := int(lenBuf[0])
		name := make([]byte, nameLen)
		if _, err := io.ReadFull(r, name); err != nil {
			return nil, err
		}
		addr.Name = string(name)

	case AddrTypeIPv6:
		ip := make([]byte, 16)
		if _, err := io.ReadFull(r, ip); err != nil {
			return nil, err
		}
		addr.IP = net.IP(ip)

	default:
		return nil, fmt.Errorf("unsupported address type: %d", addr.Type)
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(r, portBuf); err != nil {
		return nil, err
	}
	addr.Port = binary.BigEndian.Uint16(portBuf)

	return addr, nil
}

func WriteAddress(w io.Writer, addr *Address) error {
	if _, err := w.Write([]byte{addr.Type}); err != nil {
		return err
	}

	switch addr.Type {
	case AddrTypeIPv4:
		ipv4 := addr.IP.To4()
		if ipv4 == nil {
			return fmt.Errorf("invalid IPv4 address")
		}
		if _, err := w.Write(ipv4); err != nil {
			return err
		}

	case AddrTypeDomain:
		if len(addr.Name) > 255 {
			return fmt.Errorf("domain name too long")
		}
		if _, err := w.Write([]byte{byte(len(addr.Name))}); err != nil {
			return err
		}
		if _, err := w.Write([]byte(addr.Name)); err != nil {
			return err
		}

	case AddrTypeIPv6:
		ipv6 := addr.IP.To16()
		if ipv6 == nil {
			return fmt.Errorf("invalid IPv6 address")
		}
		if _, err := w.Write(ipv6); err != nil {
			return err
		}
	}

	portBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(portBuf, addr.Port)
	_, err := w.Write(portBuf)
	return err
}

func (a *Address) String() string {
	var host string
	switch a.Type {
	case AddrTypeIPv4, AddrTypeIPv6:
		host = a.IP.String()
	case AddrTypeDomain:
		host = a.Name
	}
	return net.JoinHostPort(host, strconv.Itoa(int(a.Port)))
}

func (a *Address) Resolve() (string, error) {
	switch a.Type {
	case AddrTypeIPv4, AddrTypeIPv6:
		return a.String(), nil
	case AddrTypeDomain:
		addr, err := net.ResolveTCPAddr("tcp", a.String())
		if err != nil {
			return "", err
		}
		return addr.String(), nil
	}
	return "", fmt.Errorf("unsupported address type")
}

func AddressFromAddr(addr net.Addr) (*Address, error) {
	switch addr := addr.(type) {
	case *net.TCPAddr:
		ip := addr.IP
		if ip.To4() != nil {
			return &Address{
				Type: AddrTypeIPv4,
				IP:   ip.To4(),
				Port: uint16(addr.Port),
			}, nil
		}
		return &Address{
			Type: AddrTypeIPv6,
			IP:   ip,
			Port: uint16(addr.Port),
		}, nil

	case *net.UDPAddr:
		ip := addr.IP
		if ip.To4() != nil {
			return &Address{
				Type: AddrTypeIPv4,
				IP:   ip.To4(),
				Port: uint16(addr.Port),
			}, nil
		}
		return &Address{
			Type: AddrTypeIPv6,
			IP:   ip,
			Port: uint16(addr.Port),
		}, nil
	}
	return nil, fmt.Errorf("unsupported address type")
}

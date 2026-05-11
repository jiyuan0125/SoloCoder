package ping

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"time"
)

type RawSocket struct {
	fd int
	ttl int
}

func NewRawSocket(ttl int) (*RawSocket, error) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		return nil, fmt.Errorf("failed to create raw socket: %w (note: this operation requires root privileges)", err)
	}

	rs := &RawSocket{
		fd:  fd,
		ttl: ttl,
	}

	if err := rs.SetTTL(ttl); err != nil {
		syscall.Close(fd)
		return nil, err
	}

	return rs, nil
}

func (rs *RawSocket) SetTTL(ttl int) error {
	rs.ttl = ttl
	err := syscall.SetsockoptInt(rs.fd, syscall.IPPROTO_IP, syscall.IP_TTL, ttl)
	if err != nil {
		return fmt.Errorf("failed to set TTL: %w", err)
	}
	return nil
}

func (rs *RawSocket) Close() error {
	return syscall.Close(rs.fd)
}

func (rs *RawSocket) SetReadTimeout(timeout time.Duration) error {
	tv := syscall.Timeval{
		Sec:  int64(timeout / time.Second),
		Usec: int64((timeout % time.Second) / time.Microsecond),
	}
	return syscall.SetsockoptTimeval(rs.fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &tv)
}

func (rs *RawSocket) SendTo(ip net.IP, data []byte) error {
	destAddr := syscall.SockaddrInet4{
		Port: 0,
		Addr: [4]byte{ip[0], ip[1], ip[2], ip[3]},
	}
	return syscall.Sendto(rs.fd, data, 0, &destAddr)
}

func (rs *RawSocket) RecvFrom(buf []byte) (int, net.IP, error) {
	n, from, err := syscall.Recvfrom(rs.fd, buf, 0)
	if err != nil {
		return 0, nil, err
	}

	var srcIP net.IP
	if sa, ok := from.(*syscall.SockaddrInet4); ok {
		srcIP = net.IP{sa.Addr[0], sa.Addr[1], sa.Addr[2], sa.Addr[3]}
	}

	return n, srcIP, nil
}

func ResolveTarget(target string) (net.IP, error) {
	addrs, err := net.LookupIP(target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target: %w", err)
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			return ipv4, nil
		}
	}

	return nil, fmt.Errorf("no IPv4 address found for target: %s", target)
}

func GetProcessID() uint16 {
	return uint16(os.Getpid() & 0xffff)
}

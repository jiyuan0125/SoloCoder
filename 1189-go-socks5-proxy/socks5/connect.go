package socks5

import (
	"io"
	"net"
	"sync"
	"time"
)

type Connection struct {
	ClientAddr net.Addr
	TargetAddr string
	StartTime  time.Time
	conn       net.Conn
}

func handleConnect(conn net.Conn, targetAddr *Address) (*Connection, error) {
	dst, err := net.DialTimeout("tcp", targetAddr.String(), 10*time.Second)
	if err != nil {
		return nil, err
	}

	bindAddr, err := AddressFromAddr(dst.LocalAddr())
	if err != nil {
		dst.Close()
		return nil, err
	}

	reply := buildReply(ReplySuccess, bindAddr)
	if _, err := conn.Write(reply); err != nil {
		dst.Close()
		return nil, err
	}

	c := &Connection{
		ClientAddr: conn.RemoteAddr(),
		TargetAddr: targetAddr.String(),
		StartTime:  time.Now(),
		conn:       conn,
	}

	go bidirectionalCopy(conn, dst)

	return c, nil
}

func bidirectionalCopy(client, target net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(target, client)
		if tcp, ok := target.(*net.TCPConn); ok {
			tcp.CloseWrite()
		}
	}()

	go func() {
		defer wg.Done()
		io.Copy(client, target)
		if tcp, ok := client.(*net.TCPConn); ok {
			tcp.CloseWrite()
		}
	}()

	wg.Wait()
	client.Close()
	target.Close()
}

func buildReply(rep byte, addr *Address) []byte {
	buf := make([]byte, 4)
	buf[0] = Version
	buf[1] = rep
	buf[2] = RSV
	buf[3] = addr.Type

	switch addr.Type {
	case AddrTypeIPv4:
		buf = append(buf, addr.IP.To4()...)
	case AddrTypeIPv6:
		buf = append(buf, addr.IP.To16()...)
	case AddrTypeDomain:
		buf = append(buf, byte(len(addr.Name)))
		buf = append(buf, addr.Name...)
	}

	portBytes := make([]byte, 2)
	portBytes[0] = byte(addr.Port >> 8)
	portBytes[1] = byte(addr.Port)
	buf = append(buf, portBytes...)

	return buf
}

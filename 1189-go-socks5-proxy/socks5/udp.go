package socks5

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"
)

type UDPAssociation struct {
	ClientAddr    net.Addr
	RelayAddr     net.Addr
	StartTime     time.Time
	clientConn    net.Conn
	udpConn       *net.UDPConn
	clientUDPAddr *net.UDPAddr
	closed        bool
	mu            sync.Mutex
}

func handleUDPAssociate(clientConn net.Conn, clientAddr *Address) (*UDPAssociation, error) {
	localAddr := clientConn.LocalAddr().(*net.TCPAddr)
	udpAddr := &net.UDPAddr{IP: localAddr.IP, Port: 0}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	bindAddr, err := AddressFromAddr(udpConn.LocalAddr())
	if err != nil {
		udpConn.Close()
		return nil, err
	}

	reply := buildReply(ReplySuccess, bindAddr)
	if _, err := clientConn.Write(reply); err != nil {
		udpConn.Close()
		return nil, err
	}

	assoc := &UDPAssociation{
		ClientAddr: clientConn.RemoteAddr(),
		RelayAddr:  udpConn.LocalAddr(),
		StartTime:  time.Now(),
		clientConn: clientConn,
		udpConn:    udpConn,
	}

	go assoc.handleUDP()
	go assoc.watchTCPClose()

	return assoc, nil
}

func (a *UDPAssociation) handleUDP() {
	buf := make([]byte, 65535)
	for {
		n, srcAddr, err := a.udpConn.ReadFromUDP(buf)
		if err != nil {
			break
		}

		a.mu.Lock()
		if a.closed {
			a.mu.Unlock()
			break
		}

		if a.clientUDPAddr == nil {
			a.clientUDPAddr = srcAddr
		} else if a.clientUDPAddr.String() != srcAddr.String() {
			a.mu.Unlock()
			continue
		}
		a.mu.Unlock()

		go a.processUDPDatagram(buf[:n], srcAddr)
	}
}

func (a *UDPAssociation) processUDPDatagram(data []byte, clientAddr *net.UDPAddr) {
	if len(data) < 10 {
		return
	}

	if data[2] != 0 {
		return
	}

	reader := bytes.NewReader(data[3:])
	targetAddr, err := ReadAddress(reader)
	if err != nil {
		return
	}

	payload := data[3+reader.Size()-int64(reader.Len()):]

	targetUDPAddr, err := net.ResolveUDPAddr("udp", targetAddr.String())
	if err != nil {
		return
	}

	targetConn, err := net.DialUDP("udp", nil, targetUDPAddr)
	if err != nil {
		return
	}
	defer targetConn.Close()

	if _, err := targetConn.Write(payload); err != nil {
		return
	}

	targetConn.SetReadDeadline(time.Now().Add(30 * time.Second))
	respBuf := make([]byte, 65535)
	n, _, err := targetConn.ReadFromUDP(respBuf)
	if err != nil {
		return
	}

	respAddr, err := AddressFromAddr(targetConn.RemoteAddr())
	if err != nil {
		return
	}

	response := buildUDPDatagram(respAddr, respBuf[:n])
	a.udpConn.WriteToUDP(response, clientAddr)
}

func buildUDPDatagram(addr *Address, data []byte) []byte {
	buf := make([]byte, 0, 10+len(data))
	buf = append(buf, 0x00, 0x00, 0x00)

	switch addr.Type {
	case AddrTypeIPv4:
		buf = append(buf, AddrTypeIPv4)
		buf = append(buf, addr.IP.To4()...)
	case AddrTypeIPv6:
		buf = append(buf, AddrTypeIPv6)
		buf = append(buf, addr.IP.To16()...)
	case AddrTypeDomain:
		buf = append(buf, AddrTypeDomain)
		buf = append(buf, byte(len(addr.Name)))
		buf = append(buf, addr.Name...)
	}

	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, addr.Port)
	buf = append(buf, portBytes...)
	buf = append(buf, data...)

	return buf
}

func (a *UDPAssociation) watchTCPClose() {
	buf := make([]byte, 1)
	_, err := io.ReadFull(a.clientConn, buf)
	if err != nil {
		a.Close()
	}
}

func (a *UDPAssociation) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return
	}
	a.closed = true
	if a.udpConn != nil {
		a.udpConn.Close()
	}
	if a.clientConn != nil {
		a.clientConn.Close()
	}
}

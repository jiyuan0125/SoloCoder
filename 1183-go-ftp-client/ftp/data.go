package ftp

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type dataConn struct {
	net.Conn
}

var pasvRegex = regexp.MustCompile(`\((\d+),(\d+),(\d+),(\d+),(\d+),(\d+)\)`)

func (c *Client) openDataConnection() (*dataConn, error) {
	if c.mode == ModePassive {
		return c.openPassiveConnection()
	}
	return c.openActiveConnection()
}

func (c *Client) openPassiveConnection() (*dataConn, error) {
	resp, err := c.sendCommand("PASV")
	if err != nil {
		return nil, fmt.Errorf("PASV command failed: %w", err)
	}
	if resp.Code/100 != 2 {
		return nil, fmt.Errorf("PASV failed: %d %s", resp.Code, resp.Msg)
	}

	ip, port, err := parsePasvResponse(resp.Msg)
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", addr, defaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to data port %s: %w", addr, err)
	}

	return &dataConn{Conn: conn}, nil
}

func parsePasvResponse(msg string) (string, int, error) {
	matches := pasvRegex.FindStringSubmatch(msg)
	if matches == nil || len(matches) != 7 {
		return "", 0, fmt.Errorf("invalid PASV response: %s", msg)
	}

	ip := fmt.Sprintf("%s.%s.%s.%s", matches[1], matches[2], matches[3], matches[4])
	p1, _ := strconv.Atoi(matches[5])
	p2, _ := strconv.Atoi(matches[6])
	port := p1*256 + p2

	return ip, port, nil
}

func (c *Client) openActiveConnection() (*dataConn, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	localAddr := listener.Addr().(*net.TCPAddr)
	port := localAddr.Port

	controlAddr := c.controlConn.LocalAddr().(*net.TCPAddr)
	ip := controlAddr.IP

	portArg, err := formatPortArgument(ip, port)
	if err != nil {
		listener.Close()
		return nil, err
	}

	resp, err := c.sendCommand("PORT %s", portArg)
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("PORT command failed: %w", err)
	}
	if resp.Code/100 != 2 {
		listener.Close()
		return nil, fmt.Errorf("PORT failed: %d %s", resp.Code, resp.Msg)
	}

	errCh := make(chan error, 1)
	connCh := make(chan net.Conn, 1)

	go func() {
		conn, err := listener.Accept()
		listener.Close()
		if err != nil {
			errCh <- err
			return
		}
		connCh <- conn
	}()

	select {
	case conn := <-connCh:
		return &dataConn{Conn: conn}, nil
	case err := <-errCh:
		return nil, fmt.Errorf("failed to accept data connection: %w", err)
	case <-time.After(defaultTimeout):
		return nil, fmt.Errorf("timeout waiting for data connection")
	}
}

func formatPortArgument(ip net.IP, port int) (string, error) {
	if ip == nil || ip.IsUnspecified() {
		return "", fmt.Errorf("invalid IP address")
	}

	ip = ip.To4()
	if ip == nil {
		return "", fmt.Errorf("IPv6 not supported for active mode")
	}

	p1 := port >> 8
	p2 := port & 0xff

	return fmt.Sprintf("%d,%d,%d,%d,%d,%d",
		ip[0], ip[1], ip[2], ip[3], p1, p2), nil
}

func (dc *dataConn) Close() error {
	if dc.Conn != nil {
		return dc.Conn.Close()
	}
	return nil
}

func (c *Client) readTransferComplete() error {
	resp, err := c.readResponse()
	if err != nil {
		return err
	}
	if resp.Code/100 != 2 {
		return fmt.Errorf("transfer failed: %d %s", resp.Code, resp.Msg)
	}
	return nil
}

func escapePath(path string) string {
	return strings.ReplaceAll(path, "\"", "\"\"")
}

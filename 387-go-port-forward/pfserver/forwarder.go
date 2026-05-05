package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Forwarder struct {
	localPort    int
	remoteAddr   string
	timeout      time.Duration
	bufferSize   int
	verbose      bool
	listener     net.Listener
	NewConnCh    chan *ConnInfo
	ConnCloseCh  chan *ConnCloseInfo
	shutdownCh   chan struct{}
	wg           sync.WaitGroup
	conns        map[uint64]*net.TCPConn
	connsMu      sync.RWMutex
}

type ConnInfo struct {
	ID         uint64
	ClientAddr string
	RemoteAddr string
	IDReady    chan struct{}
}

type ConnCloseInfo struct {
	ID            uint64
	BytesSent     int64
	BytesReceived int64
	Duration      time.Duration
}

func NewForwarder(localPort int, remoteAddr string, timeout time.Duration, bufferSize int, verbose bool) *Forwarder {
	return &Forwarder{
		localPort:   localPort,
		remoteAddr:  remoteAddr,
		timeout:     timeout,
		bufferSize:  bufferSize,
		verbose:     verbose,
		NewConnCh:   make(chan *ConnInfo, 10),
		ConnCloseCh: make(chan *ConnCloseInfo, 10),
		shutdownCh:  make(chan struct{}),
		conns:       make(map[uint64]*net.TCPConn),
	}
}

func (f *Forwarder) Start() error {
	addr := fmt.Sprintf(":%d", f.localPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %v", f.localPort, err)
	}
	
	f.listener = ln
	f.wg.Add(1)
	go f.acceptLoop()
	
	logf("Forwarder started: :%d -> %s", f.localPort, f.remoteAddr)
	return nil
}

func (f *Forwarder) acceptLoop() {
	defer f.wg.Done()
	
	for {
		select {
		case <-f.shutdownCh:
			return
		default:
		}
		
		conn, err := f.listener.Accept()
		if err != nil {
			select {
			case <-f.shutdownCh:
				return
			default:
				logf("Accept error on port %d: %v", f.localPort, err)
				continue
			}
		}
		
		f.wg.Add(1)
		go f.handleConnection(conn)
	}
}

func (f *Forwarder) handleConnection(clientConn net.Conn) {
	defer f.wg.Done()
	defer clientConn.Close()
	
	startTime := time.Now()
	clientAddr := clientConn.RemoteAddr().String()
	
	logf("New connection from %s to port %d", clientAddr, f.localPort)
	
	connInfo := &ConnInfo{
		ClientAddr: clientAddr,
		RemoteAddr: f.remoteAddr,
		IDReady:    make(chan struct{}),
	}
	
	select {
	case f.NewConnCh <- connInfo:
	case <-f.shutdownCh:
		logf("Connection from %s rejected: shutting down", clientAddr)
		return
	}
	
	select {
	case <-connInfo.IDReady:
	case <-f.shutdownCh:
		logf("Connection from %s cancelled during ID assignment", clientAddr)
		return
	}
	
	connID := connInfo.ID
	
	remoteConn, err := net.DialTimeout("tcp", f.remoteAddr, 5*time.Second)
	if err != nil {
		logf("Failed to connect to %s: %v", f.remoteAddr, err)
		clientConn.Write([]byte(fmt.Sprintf("Connection failed: %v\r\n", err)))
		f.sendCloseInfo(connID, 0, 0, time.Since(startTime))
		return
	}
	defer remoteConn.Close()
	
	f.registerConn(connID, clientConn, remoteConn)
	defer f.unregisterConn(connID)
	
	logf("Connected %s <-> %s", clientAddr, f.remoteAddr)
	
	var bytesSent, bytesReceived int64
	var wg sync.WaitGroup
	
	clientTCP, _ := clientConn.(*net.TCPConn)
	remoteTCP, _ := remoteConn.(*net.TCPConn)
	
	wg.Add(2)
	
	go func() {
		defer wg.Done()
		n, err := f.copyWithTimeout(clientConn, remoteConn, f.timeout, f.bufferSize)
		bytesReceived += n
		if err != nil && err != io.EOF {
			logf("Error copying client->remote: %v", err)
		}
		if remoteTCP != nil {
			remoteTCP.CloseWrite()
		}
	}()
	
	go func() {
		defer wg.Done()
		n, err := f.copyWithTimeout(remoteConn, clientConn, f.timeout, f.bufferSize)
		bytesSent += n
		if err != nil && err != io.EOF {
			logf("Error copying remote->client: %v", err)
		}
		if clientTCP != nil {
			clientTCP.CloseWrite()
		}
	}()
	
	wg.Wait()
	
	duration := time.Since(startTime)
	f.sendCloseInfo(connID, bytesSent, bytesReceived, duration)
	
	logf("Connection closed %s <-> %s: sent %d, received %d, duration %v",
		clientAddr, f.remoteAddr, bytesSent, bytesReceived, duration)
}

func (f *Forwarder) copyWithTimeout(dst, src net.Conn, timeout time.Duration, bufferSize int) (int64, error) {
	buf := make([]byte, bufferSize)
	var total int64
	
	for {
		select {
		case <-f.shutdownCh:
			return total, nil
		default:
		}
		
		src.SetReadDeadline(time.Now().Add(timeout))
		n, err := src.Read(buf)
		if n > 0 {
			dst.SetWriteDeadline(time.Now().Add(timeout))
			written, writeErr := dst.Write(buf[:n])
			if written > 0 {
				total += int64(written)
				if f.verbose {
					logf("Forwarded %d bytes", written)
				}
			}
			if writeErr != nil {
				return total, writeErr
			}
		}
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logf("Connection idle timeout")
				return total, err
			}
			return total, err
		}
	}
}

func (f *Forwarder) registerConn(connID uint64, clientConn, remoteConn net.Conn) {
	f.connsMu.Lock()
	defer f.connsMu.Unlock()
	
	if tcpConn, ok := clientConn.(*net.TCPConn); ok {
		f.conns[connID] = tcpConn
	}
	if tcpConn, ok := remoteConn.(*net.TCPConn); ok {
		f.conns[connID+1000000] = tcpConn
	}
}

func (f *Forwarder) unregisterConn(connID uint64) {
	f.connsMu.Lock()
	defer f.connsMu.Unlock()
	
	delete(f.conns, connID)
	delete(f.conns, connID+1000000)
}

func (f *Forwarder) sendCloseInfo(connID uint64, bytesSent, bytesReceived int64, duration time.Duration) {
	closeInfo := &ConnCloseInfo{
		ID:            connID,
		BytesSent:     bytesSent,
		BytesReceived: bytesReceived,
		Duration:      duration,
	}
	
	select {
	case f.ConnCloseCh <- closeInfo:
	case <-f.shutdownCh:
	}
}

func (f *Forwarder) Stop() {
	select {
	case <-f.shutdownCh:
		return
	default:
		close(f.shutdownCh)
	}
	
	if f.listener != nil {
		f.listener.Close()
	}
	
	f.connsMu.RLock()
	for _, conn := range f.conns {
		conn.Close()
	}
	f.connsMu.RUnlock()
	
	f.wg.Wait()
	logf("Forwarder stopped on port %d", f.localPort)
}

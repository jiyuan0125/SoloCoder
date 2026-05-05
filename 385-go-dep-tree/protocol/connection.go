package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	DefaultPort    = 9876
	DefaultAddress = "localhost:9876"
	MaxMessageSize = 64 * 1024 * 1024
	ReadTimeout    = 30 * time.Second
	WriteTimeout   = 30 * time.Second
)

func WriteMessage(conn net.Conn, data []byte) error {
	conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
	
	msgLen := uint32(len(data))
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, msgLen)
	
	_, err := conn.Write(lenBuf)
	if err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}
	
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}
	
	return nil
}

func ReadMessage(conn net.Conn) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(ReadTimeout))
	
	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lenBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read message length: %w", err)
	}
	
	msgLen := binary.BigEndian.Uint32(lenBuf)
	if msgLen > MaxMessageSize {
		return nil, fmt.Errorf("message too large: %d bytes (max %d)", msgLen, MaxMessageSize)
	}
	
	data := make([]byte, msgLen)
	_, err = io.ReadFull(conn, data)
	if err != nil {
		return nil, fmt.Errorf("failed to read message data: %w", err)
	}
	
	return data, nil
}

func ConnectToServer(addr string) (net.Conn, error) {
	if addr == "" {
		addr = DefaultAddress
	}
	
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	
	return conn, nil
}

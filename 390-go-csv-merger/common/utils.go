package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
)

func SendMessage(conn net.Conn, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	length := len(data)
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length >> 24)
	lengthBytes[1] = byte(length >> 16)
	lengthBytes[2] = byte(length >> 8)
	lengthBytes[3] = byte(length)

	if _, err := conn.Write(lengthBytes); err != nil {
		return fmt.Errorf("failed to send message length: %w", err)
	}

	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to send message data: %w", err)
	}

	return nil
}

func ReceiveMessage(conn net.Conn, target interface{}) error {
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBytes); err != nil {
		return fmt.Errorf("failed to read message length: %w", err)
	}

	length := int(lengthBytes[0])<<24 | int(lengthBytes[1])<<16 | int(lengthBytes[2])<<8 | int(lengthBytes[3])

	if length > MaxMessageSize {
		return fmt.Errorf("message too large: %d bytes (max: %d)", length, MaxMessageSize)
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return fmt.Errorf("failed to read message data: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return nil
}

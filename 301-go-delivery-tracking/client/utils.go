package main

import (
	"fmt"
	"net"
	"time"
)

func connectToServer(addr string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("连接服务器超时: %w", err)
	}
	return conn, nil
}

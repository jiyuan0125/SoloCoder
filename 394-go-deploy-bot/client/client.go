package main

import (
	"deploybot/common"
	"fmt"
	"net"
	"time"
)

type Client struct {
	conn   net.Conn
	bc     *common.BufferedConn
	addr   string
}

func NewClient(addr string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("连接服务端失败: %v", err)
	}

	return &Client{
		conn: conn,
		bc:   common.NewBufferedConn(conn),
		addr: addr,
	}, nil
}

func (c *Client) Send(req *common.DeployRequest) (*common.DeployResponse, error) {
	if err := c.bc.SendRequest(req); err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}

	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	resp, err := c.bc.ReceiveResponse()
	if err != nil {
		return nil, fmt.Errorf("接收响应失败: %v", err)
	}

	return resp, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

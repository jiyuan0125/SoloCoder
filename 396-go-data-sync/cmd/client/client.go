package main

import (
	"bufio"
	"fmt"
	"net"

	"go-data-sync/pkg/protocol"
)

type Client struct {
	conn net.Conn
	host string
	port string
}

func NewClient(host, port string) (*Client, error) {
	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	
	return &Client{
		conn: conn,
		host: host,
		port: port,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Sync(sourceFile, targetFile, keyColumn string, apply bool) (*protocol.Response, error) {
	req := &protocol.Request{
		Command:    protocol.CmdSync,
		SourceFile: sourceFile,
		TargetFile: targetFile,
		KeyColumn:  keyColumn,
		Apply:      apply,
	}
	
	return c.sendRequest(req)
}

func (c *Client) Status() (*protocol.Response, error) {
	req := &protocol.Request{
		Command: protocol.CmdStatus,
	}
	
	return c.sendRequest(req)
}

func (c *Client) History() (*protocol.Response, error) {
	req := &protocol.Request{
		Command: protocol.CmdHistory,
	}
	
	return c.sendRequest(req)
}

func (c *Client) sendRequest(req *protocol.Request) (*protocol.Response, error) {
	writer := bufio.NewWriter(c.conn)
	if err := protocol.Encode(writer, req); err != nil {
		return nil, err
	}
	if err := writer.Flush(); err != nil {
		return nil, err
	}
	
	var resp protocol.Response
	reader := bufio.NewReader(c.conn)
	if err := protocol.Decode(reader, &resp); err != nil {
		return nil, err
	}
	
	return &resp, nil
}

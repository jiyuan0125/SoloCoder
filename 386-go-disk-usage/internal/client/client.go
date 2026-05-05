package client

import (
	"fmt"
	"net"
	"time"
	
	"disk-usage/internal/protocol"
)

type Client struct {
	serverAddr string
	conn       net.Conn
}

func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
	}
}

func (c *Client) Connect() error {
	conn, err := net.DialTimeout("tcp", c.serverAddr, 5*time.Second)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) SendScanRequest(req *protocol.ScanRequest) (*protocol.ScanResult, error) {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
		defer c.Close()
	}
	
	payload, err := protocol.EncodeScanRequest(req)
	if err != nil {
		return nil, err
	}
	
	msg := &protocol.Message{
		Type:    protocol.MsgTypeScanRequest,
		Payload: payload,
	}
	
	if err := protocol.EncodeMessage(msg, c.conn); err != nil {
		return nil, err
	}
	
	respMsg, err := protocol.DecodeMessage(c.conn)
	if err != nil {
		return nil, err
	}
	
	if respMsg.Type == protocol.MsgTypeError {
		return nil, fmt.Errorf("%s", respMsg.Payload)
	}
	
	return protocol.DecodeScanResult(respMsg.Payload)
}

func (c *Client) SendCompareRequest(req *protocol.ScanRequest) (*protocol.CompareResult, error) {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
		defer c.Close()
	}
	
	payload, err := protocol.EncodeScanRequest(req)
	if err != nil {
		return nil, err
	}
	
	msg := &protocol.Message{
		Type:    protocol.MsgTypeCompareRequest,
		Payload: payload,
	}
	
	if err := protocol.EncodeMessage(msg, c.conn); err != nil {
		return nil, err
	}
	
	respMsg, err := protocol.DecodeMessage(c.conn)
	if err != nil {
		return nil, err
	}
	
	if respMsg.Type == protocol.MsgTypeError {
		return nil, fmt.Errorf("%s", respMsg.Payload)
	}
	
	if respMsg.Type == protocol.MsgTypeScanResponse {
		scanResult, err := protocol.DecodeScanResult(respMsg.Payload)
		if err != nil {
			return nil, err
		}
		return &protocol.CompareResult{
			OldScan:    scanResult,
			NewScan:    scanResult,
			GrowthDirs: make([]protocol.GrowthDir, 0),
		}, nil
	}
	
	return protocol.DecodeCompareResult(respMsg.Payload)
}

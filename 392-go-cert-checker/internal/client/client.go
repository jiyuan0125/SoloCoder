package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"cert-checker/internal/protocol"
)

const defaultTimeout = 10 * time.Second

type Client struct {
	serverAddr string
}

func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
	}
}

func (c *Client) sendRequest(req protocol.Request) (*protocol.Response, error) {
	conn, err := net.DialTimeout("tcp", c.serverAddr, defaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("连接服务端失败: %v", err)
	}
	defer conn.Close()

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	if _, err := conn.Write(reqBytes); err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}

	if _, err := conn.Write([]byte("\n")); err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		if scanner.Err() != nil {
			return nil, fmt.Errorf("读取响应失败: %v", scanner.Err())
		}
		return nil, fmt.Errorf("服务端未返回响应")
	}

	var resp protocol.Response
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &resp, nil
}

func (c *Client) AddDomain(domain string) error {
	payload := protocol.AddDomainPayload{
		Domain: domain,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req := protocol.Request{
		Type:    protocol.RequestTypeAddDomain,
		Payload: payloadBytes,
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	if resp.Type == protocol.ResponseTypeError {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func (c *Client) RemoveDomain(domain string) error {
	payload := protocol.RemoveDomainPayload{
		Domain: domain,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req := protocol.Request{
		Type:    protocol.RequestTypeRemoveDomain,
		Payload: payloadBytes,
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	if resp.Type == protocol.ResponseTypeError {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func (c *Client) ListDomains() (*protocol.ListDomainsResponse, error) {
	req := protocol.Request{
		Type: protocol.RequestTypeListDomains,
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.Type == protocol.ResponseTypeError {
		return nil, fmt.Errorf(resp.Message)
	}

	var result protocol.ListDomainsResponse
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &result, nil
}

func (c *Client) CheckDomain(domain string, warnDays int) (protocol.CertInfo, error) {
	payload := protocol.CheckDomainPayload{
		Domain:   domain,
		WarnDays: warnDays,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return protocol.CertInfo{}, err
	}

	req := protocol.Request{
		Type:    protocol.RequestTypeCheckDomain,
		Payload: payloadBytes,
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return protocol.CertInfo{}, err
	}

	if resp.Type == protocol.ResponseTypeError {
		return protocol.CertInfo{}, fmt.Errorf(resp.Message)
	}

	var result protocol.CertInfo
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
		return protocol.CertInfo{}, fmt.Errorf("解析响应失败: %v", err)
	}

	return result, nil
}

func (c *Client) CheckAll(warnDays int) ([]protocol.CertInfo, error) {
	payload := protocol.CheckAllPayload{
		WarnDays: warnDays,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req := protocol.Request{
		Type:    protocol.RequestTypeCheckAll,
		Payload: payloadBytes,
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.Type == protocol.ResponseTypeError {
		return nil, fmt.Errorf(resp.Message)
	}

	var result protocol.CheckAllResponse
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return result.Results, nil
}

func (c *Client) GetStatus(domain string) (*protocol.DomainStatus, error) {
	payload := protocol.AddDomainPayload{
		Domain: domain,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req := protocol.Request{
		Type:    protocol.RequestTypeGetStatus,
		Payload: payloadBytes,
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.Type == protocol.ResponseTypeError {
		return nil, fmt.Errorf(resp.Message)
	}

	var result protocol.GetStatusResponse
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &result.Status, nil
}

func (c *Client) GetHistory(domain string, limit int) ([]protocol.CertInfo, error) {
	payload := protocol.GetHistoryPayload{
		Domain: domain,
		Limit:  limit,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req := protocol.Request{
		Type:    protocol.RequestTypeGetHistory,
		Payload: payloadBytes,
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	if resp.Type == protocol.ResponseTypeError {
		return nil, fmt.Errorf(resp.Message)
	}

	var result protocol.GetHistoryResponse
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return result.History, nil
}

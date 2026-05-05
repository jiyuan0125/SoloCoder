package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"github.com/notify-relay/internal/protocol"
)

type Client struct {
	addr string
}

func NewClient(addr string) *Client {
	return &Client{addr: addr}
}

func (c *Client) sendRequest(req *protocol.Request) (*protocol.Response, error) {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return nil, fmt.Errorf("无法连接到服务: %v", err)
	}
	defer conn.Close()

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	if _, err := conn.Write(append(reqBytes, '\n')); err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("读取响应失败: %v", err)
		}
		return nil, fmt.Errorf("服务未返回响应")
	}

	var resp protocol.Response
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &resp, nil
}

func (c *Client) AddWatch(filePath string) error {
	payload, _ := json.Marshal(protocol.AddWatchRequest{FilePath: filePath})
	req := &protocol.Request{
		Type:    protocol.MessageTypeAddWatch,
		Payload: payload,
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("添加监控失败: %s", resp.Error)
	}

	fmt.Printf("已添加监控文件: %s\n", filePath)
	return nil
}

func (c *Client) RemoveWatch(filePath string) error {
	payload, _ := json.Marshal(protocol.RemoveWatchRequest{FilePath: filePath})
	req := &protocol.Request{
		Type:    protocol.MessageTypeRemoveWatch,
		Payload: payload,
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("移除监控失败: %s", resp.Error)
	}

	fmt.Printf("已移除监控文件: %s\n", filePath)
	return nil
}

func (c *Client) ListWatches() error {
	req := &protocol.Request{
		Type: protocol.MessageTypeListWatches,
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("获取监控列表失败: %s", resp.Error)
	}

	var listResp protocol.ListWatchesResponse
	if err := json.Unmarshal(resp.Payload, &listResp); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if len(listResp.Files) == 0 {
		fmt.Println("当前没有监控的文件")
		return nil
	}

	fmt.Println("监控的文件列表:")
	fmt.Println("--------------------------------------------------")
	for i, file := range listResp.Files {
		fmt.Printf("%d. 文件路径: %s\n", i+1, file.FilePath)
		fmt.Printf("   当前偏移量: %d\n", file.CurrentOffset)
		fmt.Printf("   已解析告警: %d\n", file.AlertsParsed)
		fmt.Println("--------------------------------------------------")
	}

	return nil
}

func (c *Client) GetStatus() error {
	req := &protocol.Request{
		Type: protocol.MessageTypeGetStatus,
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("获取状态失败: %s", resp.Error)
	}

	var statusResp protocol.StatusResponse
	if err := json.Unmarshal(resp.Payload, &statusResp); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	fmt.Println("服务状态:")
	fmt.Println("--------------------------------------------------")
	fmt.Printf("状态: %s\n", statusResp.ServerStatus)
	fmt.Printf("运行时间: %s\n", statusResp.Uptime)
	fmt.Printf("总告警数: %d\n", statusResp.TotalAlerts)
	fmt.Println("渠道统计:")
	for ch, count := range statusResp.ChannelStats {
		fmt.Printf("  - %s: %d\n", ch, count)
	}
	fmt.Println("--------------------------------------------------")

	return nil
}

func (c *Client) GetForwardLogs() error {
	req := &protocol.Request{
		Type: protocol.MessageTypeGetForwardLogs,
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("获取转发日志失败: %s", resp.Error)
	}

	var logsResp protocol.ForwardLogsResponse
	if err := json.Unmarshal(resp.Payload, &logsResp); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if len(logsResp.Logs) == 0 {
		fmt.Println("当前没有转发日志")
		return nil
	}

	fmt.Println("转发日志 (最近100条):")
	fmt.Println("--------------------------------------------------")
	for _, log := range logsResp.Logs {
		fmt.Printf("时间: %s\n", log.Timestamp)
		fmt.Printf("级别: %s\n", log.Level)
		fmt.Printf("消息: %s\n", log.Message)
		fmt.Printf("渠道: %s\n", log.Channel)
		fmt.Printf("状态: %s\n", log.Status)
		if log.Error != "" {
			fmt.Printf("错误: %s\n", log.Error)
		}
		fmt.Println("--------------------------------------------------")
	}

	return nil
}

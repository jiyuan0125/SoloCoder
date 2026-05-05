package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-report-gen/protocol"
)

type Client struct {
	serverAddr string
	conn       net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
}

func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
	}
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		return fmt.Errorf("连接服务端失败: %w", err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)

	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) SendRequest(req *protocol.Request) (*protocol.Response, error) {
	data, err := protocol.SerializeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	c.writer.WriteString(string(data) + "END\n")
	c.writer.Flush()

	respData, err := c.readResponse()
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	resp, err := protocol.DeserializeResponse(respData)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return resp, nil
}

func (c *Client) readResponse() ([]byte, error) {
	var data []byte
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasSuffix(line, "END") {
			data = append(data, []byte(line[:len(line)-3])...)
			break
		}

		data = append(data, []byte(line)...)
	}

	return data, nil
}

func (c *Client) GenerateReport(repoPath, since, author, format string) (*protocol.Report, error) {
	req := &protocol.Request{
		Type: protocol.RequestTypeGenerateReport,
		Generate: &protocol.GenerateRequest{
			RepoPath: repoPath,
			Since:    since,
			Author:   author,
			Format:   format,
		},
	}

	resp, err := c.SendRequest(req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Message)
	}

	return resp.Report, nil
}

func (c *Client) ListHistory(limit int) ([]protocol.HistoryEntry, error) {
	req := &protocol.Request{
		Type: protocol.RequestTypeListHistory,
		ListHistory: &protocol.ListHistoryRequest{
			Limit: limit,
		},
	}

	resp, err := c.SendRequest(req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Message)
	}

	return resp.History, nil
}

func (c *Client) GetReport(reportID string) (*protocol.Report, error) {
	req := &protocol.Request{
		Type: protocol.RequestTypeGetReport,
		GetReport: &protocol.GetReportRequest{
			ReportID: reportID,
		},
	}

	resp, err := c.SendRequest(req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Message)
	}

	return resp.Report, nil
}

func PrintReport(report *protocol.Report, format string) {
	if format == "markdown" && report.ContentMarkdown != "" {
		fmt.Println(report.ContentMarkdown)
	} else {
		fmt.Println(report.Content)
	}
}

func SaveReportToFile(report *protocol.Report, format, outputPath string) (string, error) {
	var content string
	if format == "markdown" && report.ContentMarkdown != "" {
		content = report.ContentMarkdown
	} else {
		content = report.Content
	}

	var filePath string
	if outputPath != "" {
		filePath = outputPath
	} else {
		timestamp := time.Now().Format("20060102_150405")
		var ext string
		if format == "markdown" {
			ext = "md"
		} else {
			ext = "txt"
		}
		fileName := fmt.Sprintf("report_%s.%s", timestamp, ext)

		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "."
		}
		outputDir := filepath.Join(homeDir, ".report_history")

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return "", fmt.Errorf("创建输出目录失败: %w", err)
		}

		filePath = filepath.Join(outputDir, fileName)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("保存报告到文件失败: %w", err)
	}

	return filePath, nil
}

func PrintHistory(history []protocol.HistoryEntry) {
	if len(history) == 0 {
		fmt.Println("暂无历史报告")
		return
	}

	fmt.Println("========================================")
	fmt.Println("           历史报告列表")
	fmt.Println("========================================")

	for i, entry := range history {
		fmt.Printf("\n报告 %d:\n", i+1)
		fmt.Printf("  ID: %s\n", entry.ID)
		fmt.Printf("  仓库: %s\n", entry.RepoPath)
		fmt.Printf("  生成时间: %s\n", entry.GeneratedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  提交数: %d\n", entry.CommitCount)
	}

	fmt.Println("\n========================================")
}

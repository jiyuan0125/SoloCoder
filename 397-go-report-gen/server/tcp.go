package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"

	"go-report-gen/protocol"
)

const (
	defaultPort = "8765"
	bufferSize  = 4096
)

type Server struct {
	port     string
	history  *HistoryManager
	listener net.Listener
}

func NewServer(port string, history *HistoryManager) *Server {
	if port == "" {
		port = defaultPort
	}
	return &Server{
		port:    port,
		history: history,
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("启动服务器失败: %w", err)
	}
	s.listener = listener

	fmt.Printf("报告生成服务已启动，监听端口: %s\n", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("接受连接失败: %v\n", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		data, err := readMessage(reader)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("读取消息失败: %v\n", err)
			}
			return
		}

		req, err := protocol.DeserializeRequest(data)
		if err != nil {
			sendError(writer, fmt.Sprintf("解析请求失败: %v", err))
			continue
		}

		resp := s.processRequest(req)
		sendResponse(writer, resp)
	}
}

func (s *Server) processRequest(req *protocol.Request) *protocol.Response {
	switch req.Type {
	case protocol.RequestTypeGenerateReport:
		return s.handleGenerateReport(req.Generate)
	case protocol.RequestTypeListHistory:
		return s.handleListHistory(req.ListHistory)
	case protocol.RequestTypeGetReport:
		return s.handleGetReport(req.GetReport)
	default:
		return &protocol.Response{
			Type:    protocol.MessageTypeError,
			Success: false,
			Message: fmt.Sprintf("未知的请求类型: %s", req.Type),
		}
	}
}

func (s *Server) handleGenerateReport(req *protocol.GenerateRequest) *protocol.Response {
	if req == nil {
		return &protocol.Response{
			Type:    protocol.MessageTypeError,
			Success: false,
			Message: "生成报告请求参数为空",
		}
	}

	if req.RepoPath == "" {
		return &protocol.Response{
			Type:    protocol.MessageTypeError,
			Success: false,
			Message: "仓库路径不能为空",
		}
	}

	if req.Since == "" {
		req.Since = "1 week ago"
	}

	report, err := GenerateReport(req.RepoPath, req.Since, req.Author, req.Format)
	if err != nil {
		return &protocol.Response{
			Type:    protocol.MessageTypeError,
			Success: false,
			Message: err.Error(),
		}
	}

	if err := s.history.AddReport(report); err != nil {
		fmt.Printf("保存报告到历史记录失败: %v\n", err)
	}

	return &protocol.Response{
		Type:    protocol.MessageTypeResponse,
		Success: true,
		Report:  report,
	}
}

func (s *Server) handleListHistory(req *protocol.ListHistoryRequest) *protocol.Response {
	limit := 0
	if req != nil {
		limit = req.Limit
	}

	history := s.history.ListHistory(limit)

	return &protocol.Response{
		Type:    protocol.MessageTypeResponse,
		Success: true,
		History: history,
	}
}

func (s *Server) handleGetReport(req *protocol.GetReportRequest) *protocol.Response {
	if req == nil || req.ReportID == "" {
		return &protocol.Response{
			Type:    protocol.MessageTypeError,
			Success: false,
			Message: "报告ID不能为空",
		}
	}

	report, exists := s.history.GetReport(req.ReportID)
	if !exists {
		return &protocol.Response{
			Type:    protocol.MessageTypeError,
			Success: false,
			Message: fmt.Sprintf("未找到报告: %s", req.ReportID),
		}
	}

	return &protocol.Response{
		Type:    protocol.MessageTypeResponse,
		Success: true,
		Report:  report,
	}
}

func readMessage(reader *bufio.Reader) ([]byte, error) {
	var data []byte
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "LEN:") {
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

func sendResponse(writer *bufio.Writer, resp *protocol.Response) {
	data, err := protocol.SerializeResponse(resp)
	if err != nil {
		sendError(writer, fmt.Sprintf("序列化响应失败: %v", err))
		return
	}

	writer.WriteString(string(data) + "END\n")
	writer.Flush()
}

func sendError(writer *bufio.Writer, message string) {
	resp := &protocol.Response{
		Type:    protocol.MessageTypeError,
		Success: false,
		Message: message,
	}
	sendResponse(writer, resp)
}

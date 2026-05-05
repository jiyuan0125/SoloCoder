package server

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"sync"

	"cert-checker/internal/protocol"
)

type Server struct {
	manager    *DomainManager
	handler    *RequestHandler
	listener   net.Listener
	wg         sync.WaitGroup
	stopChan   chan struct{}
	once       sync.Once
}

func NewServer() *Server {
	manager := NewDomainManager()
	handler := NewRequestHandler(manager)

	return &Server{
		manager:  manager,
		handler:  handler,
		stopChan: make(chan struct{}),
	}
}

func (s *Server) Start(addr string) error {
	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		select {
		case <-s.stopChan:
			return nil
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.stopChan:
					return nil
				default:
					log.Printf("接受连接失败: %v", err)
					continue
				}
			}

			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

func (s *Server) Stop() {
	s.once.Do(func() {
		close(s.stopChan)
		if s.listener != nil {
			s.listener.Close()
		}
	})
	s.wg.Wait()
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	log.Printf("新连接来自: %s", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)
	writer := bufio.NewWriter(conn)

	for {
		select {
		case <-s.stopChan:
			return
		default:
			if !scanner.Scan() {
				if scanner.Err() != nil {
					log.Printf("读取失败: %v", scanner.Err())
				}
				return
			}

			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var req protocol.Request
			if err := json.Unmarshal(line, &req); err != nil {
				log.Printf("解析请求失败: %v", err)
				s.sendError(writer, "无效的请求格式")
				continue
			}

			resp := s.handler.HandleRequest(req)

			respBytes, err := json.Marshal(resp)
			if err != nil {
				log.Printf("序列化响应失败: %v", err)
				s.sendError(writer, "内部错误")
				continue
			}

			if _, err := writer.Write(respBytes); err != nil {
				log.Printf("写入响应失败: %v", err)
				return
			}

			if _, err := writer.WriteString("\n"); err != nil {
				log.Printf("写入换行失败: %v", err)
				return
			}

			if err := writer.Flush(); err != nil {
				log.Printf("刷新失败: %v", err)
				return
			}
		}
	}
}

func (s *Server) sendError(writer *bufio.Writer, message string) {
	resp := protocol.Response{
		Type:    protocol.ResponseTypeError,
		Message: message,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return
	}

	writer.Write(respBytes)
	writer.WriteString("\n")
	writer.Flush()
}

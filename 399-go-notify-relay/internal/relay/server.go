package relay

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/notify-relay/internal/protocol"
)

type Server struct {
	addr         string
	listener     net.Listener
	offsetStore  *OffsetStore
	forwarder    *Forwarder
	watchers     map[string]*FileWatcher
	watcherMu    sync.RWMutex
	startTime    time.Time
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

func NewServer(addr string, dataPath string) (*Server, error) {
	offsetStore, err := NewOffsetStore(dataPath)
	if err != nil {
		return nil, err
	}

	forwarder := NewForwarder()

	return &Server{
		addr:        addr,
		offsetStore: offsetStore,
		forwarder:   forwarder,
		watchers:    make(map[string]*FileWatcher),
		startTime:   time.Now(),
		stopChan:    make(chan struct{}),
	}, nil
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = listener

	log.Printf("服务启动，监听地址: %s", s.addr)

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

func (s *Server) Stop() {
	close(s.stopChan)
	
	if s.listener != nil {
		s.listener.Close()
	}

	s.watcherMu.Lock()
	for _, watcher := range s.watchers {
		watcher.Stop()
	}
	s.watcherMu.Unlock()

	s.wg.Wait()
	log.Println("服务已停止")
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopChan:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.stopChan:
					return
				default:
					log.Printf("接受连接错误: %v", err)
					continue
				}
			}
			s.wg.Add(1)
			go s.handleConn(conn)
		}
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var req protocol.Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			s.sendResponse(conn, false, "无效的请求格式", nil)
			continue
		}

		resp := s.processRequest(&req)
		respBytes, err := json.Marshal(resp)
		if err != nil {
			log.Printf("序列化响应错误: %v", err)
			continue
		}

		conn.Write(append(respBytes, '\n'))
	}
}

func (s *Server) processRequest(req *protocol.Request) *protocol.Response {
	switch req.Type {
	case protocol.MessageTypeAddWatch:
		return s.handleAddWatch(req)
	case protocol.MessageTypeRemoveWatch:
		return s.handleRemoveWatch(req)
	case protocol.MessageTypeListWatches:
		return s.handleListWatches()
	case protocol.MessageTypeGetStatus:
		return s.handleGetStatus()
	case protocol.MessageTypeGetForwardLogs:
		return s.handleGetForwardLogs()
	default:
		return &protocol.Response{
			Type:    protocol.MessageTypeResponse,
			Success: false,
			Error:   fmt.Sprintf("未知的消息类型: %s", req.Type),
		}
	}
}

func (s *Server) handleAddWatch(req *protocol.Request) *protocol.Response {
	var addReq protocol.AddWatchRequest
	if err := json.Unmarshal(req.Payload, &addReq); err != nil {
		return &protocol.Response{
			Type:    protocol.MessageTypeResponse,
			Success: false,
			Error:   "无效的添加监控请求",
		}
	}

	s.watcherMu.Lock()
	defer s.watcherMu.Unlock()

	if _, exists := s.watchers[addReq.FilePath]; exists {
		return &protocol.Response{
			Type:    protocol.MessageTypeResponse,
			Success: true,
		}
	}

	watcher := NewFileWatcher(addReq.FilePath, s.offsetStore, s.forwarder)
	s.watchers[addReq.FilePath] = watcher
	watcher.Start()

	log.Printf("已添加监控文件: %s", addReq.FilePath)

	return &protocol.Response{
		Type:    protocol.MessageTypeResponse,
		Success: true,
	}
}

func (s *Server) handleRemoveWatch(req *protocol.Request) *protocol.Response {
	var removeReq protocol.RemoveWatchRequest
	if err := json.Unmarshal(req.Payload, &removeReq); err != nil {
		return &protocol.Response{
			Type:    protocol.MessageTypeResponse,
			Success: false,
			Error:   "无效的移除监控请求",
		}
	}

	s.watcherMu.Lock()
	defer s.watcherMu.Unlock()

	if watcher, exists := s.watchers[removeReq.FilePath]; exists {
		watcher.Stop()
		delete(s.watchers, removeReq.FilePath)
		log.Printf("已移除监控文件: %s", removeReq.FilePath)
	}

	return &protocol.Response{
		Type:    protocol.MessageTypeResponse,
		Success: true,
	}
}

func (s *Server) handleListWatches() *protocol.Response {
	s.watcherMu.RLock()
	defer s.watcherMu.RUnlock()

	files := make([]protocol.WatchedFileInfo, 0, len(s.watchers))
	for _, watcher := range s.watchers {
		files = append(files, protocol.WatchedFileInfo{
			FilePath:       watcher.GetFilePath(),
			CurrentOffset: watcher.GetCurrentOffset(),
			AlertsParsed: watcher.GetAlertsParsed(),
		})
	}

	respPayload := protocol.ListWatchesResponse{Files: files}
	payloadBytes, _ := json.Marshal(respPayload)

	return &protocol.Response{
		Type:    protocol.MessageTypeResponse,
		Success: true,
		Payload: payloadBytes,
	}
}

func (s *Server) handleGetStatus() *protocol.Response {
	uptime := time.Since(s.startTime).String()

	respPayload := protocol.StatusResponse{
		ServerStatus: "running",
		Uptime:       uptime,
		TotalAlerts:  s.forwarder.GetTotalAlerts(),
		ChannelStats: s.forwarder.GetChannelStats(),
	}

	payloadBytes, _ := json.Marshal(respPayload)

	return &protocol.Response{
		Type:    protocol.MessageTypeResponse,
		Success: true,
		Payload: payloadBytes,
	}
}

func (s *Server) handleGetForwardLogs() *protocol.Response {
	logs := s.forwarder.GetForwardLogs()

	respPayload := protocol.ForwardLogsResponse{Logs: logs}
	payloadBytes, _ := json.Marshal(respPayload)

	return &protocol.Response{
		Type:    protocol.MessageTypeResponse,
		Success: true,
		Payload: payloadBytes,
	}
}

func (s *Server) sendResponse(conn net.Conn, success bool, errMsg string, payload interface{}) {
	resp := &protocol.Response{
		Type:    protocol.MessageTypeResponse,
		Success: success,
		Error:   errMsg,
	}

	if payload != nil {
		payloadBytes, _ := json.Marshal(payload)
		resp.Payload = payloadBytes
	}

	respBytes, _ := json.Marshal(resp)
	conn.Write(append(respBytes, '\n'))
}

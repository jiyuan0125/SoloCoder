package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	
	"disk-usage/internal/protocol"
)

type TCPServer struct {
	addr       string
	listener   net.Listener
	stopCh     chan struct{}
	lastScans  map[string]*protocol.ScanResult
	mu         sync.RWMutex
}

func NewTCPServer(addr string) *TCPServer {
	return &TCPServer{
		addr:      addr,
		stopCh:    make(chan struct{}),
		lastScans: make(map[string]*protocol.ScanResult),
	}
}

func (s *TCPServer) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	
	log.Printf("磁盘使用分析服务已启动，监听地址: %s", s.addr)
	
	for {
		select {
		case <-s.stopCh:
			return nil
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				log.Printf("接受连接失败: %v", err)
				continue
			}
			
			go s.handleConnection(conn)
		}
	}
}

func (s *TCPServer) Stop() {
	close(s.stopCh)
	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	log.Printf("接受连接来自: %s", conn.RemoteAddr())
	
	for {
		select {
		case <-s.stopCh:
			return
		default:
			msg, err := protocol.DecodeMessage(conn)
			if err != nil {
				if err.Error() != "EOF" {
					log.Printf("解码消息失败: %v", err)
				}
				return
			}
			
			response, err := s.processMessage(msg)
			if err != nil {
				errMsg := &protocol.Message{
					Type:    protocol.MsgTypeError,
					Payload: err.Error(),
				}
				protocol.EncodeMessage(errMsg, conn)
				continue
			}
			
			err = protocol.EncodeMessage(response, conn)
			if err != nil {
				log.Printf("编码响应失败: %v", err)
				return
			}
		}
	}
}

func (s *TCPServer) processMessage(msg *protocol.Message) (*protocol.Message, error) {
	switch msg.Type {
	case protocol.MsgTypeScanRequest:
		return s.handleScanRequest(msg.Payload)
	case protocol.MsgTypeCompareRequest:
		return s.handleCompareRequest(msg.Payload)
	default:
		return nil, fmt.Errorf("未知消息类型: %d", msg.Type)
	}
}

func (s *TCPServer) handleScanRequest(payload string) (*protocol.Message, error) {
	req, err := protocol.DecodeScanRequest(payload)
	if err != nil {
		return nil, err
	}
	
	maxDepth := req.MaxDepth
	if maxDepth == 0 {
		maxDepth = 10
	}
	
	scanner := NewScanner(maxDepth, req.MinSize, req.All)
	
	result, err := scanner.Scan(req.Path)
	if err != nil {
		return nil, err
	}
	
	if req.Top > 0 && len(result.Directories) > req.Top {
		result.Directories = result.Directories[:req.Top]
	}
	
	s.mu.Lock()
	s.lastScans[req.Path] = result
	s.mu.Unlock()
	
	payload, err = protocol.EncodeScanResult(result)
	if err != nil {
		return nil, err
	}
	
	return &protocol.Message{
		Type:    protocol.MsgTypeScanResponse,
		Payload: payload,
	}, nil
}

func (s *TCPServer) handleCompareRequest(payload string) (*protocol.Message, error) {
	req, err := protocol.DecodeScanRequest(payload)
	if err != nil {
		return nil, err
	}
	
	s.mu.RLock()
	oldScan := s.lastScans[req.Path]
	s.mu.RUnlock()
	
	if oldScan == nil {
		return s.handleScanRequest(payload)
	}
	
	maxDepth := req.MaxDepth
	if maxDepth == 0 {
		maxDepth = 10
	}
	
	scanner := NewScanner(maxDepth, req.MinSize, req.All)
	
	newResult, err := scanner.Scan(req.Path)
	if err != nil {
		return nil, err
	}
	
	compareResult := s.compareScans(oldScan, newResult)
	
	s.mu.Lock()
	s.lastScans[req.Path] = newResult
	s.mu.Unlock()
	
	payload, err = protocol.EncodeCompareResult(compareResult)
	if err != nil {
		return nil, err
	}
	
	return &protocol.Message{
		Type:    protocol.MsgTypeCompareResponse,
		Payload: payload,
	}, nil
}

func (s *TCPServer) compareScans(oldScan, newScan *protocol.ScanResult) *protocol.CompareResult {
	oldMap := make(map[string]protocol.DirInfo)
	for _, dir := range oldScan.Directories {
		oldMap[dir.Path] = dir
	}
	
	growthDirs := make([]protocol.GrowthDir, 0)
	
	for _, newDir := range newScan.Directories {
		if oldDir, exists := oldMap[newDir.Path]; exists {
			if newDir.Size > oldDir.Size {
				growthDirs = append(growthDirs, protocol.GrowthDir{
					Path:     newDir.Path,
					OldSize:  oldDir.Size,
					NewSize:  newDir.Size,
					Growth:   newDir.Size - oldDir.Size,
					OldFiles: oldDir.Files,
					NewFiles: newDir.Files,
				})
			}
		} else {
			growthDirs = append(growthDirs, protocol.GrowthDir{
				Path:     newDir.Path,
				OldSize:  0,
				NewSize:  newDir.Size,
				Growth:   newDir.Size,
				OldFiles: 0,
				NewFiles: newDir.Files,
			})
		}
	}
	
	return &protocol.CompareResult{
		OldScan:    oldScan,
		NewScan:    newScan,
		GrowthDirs: growthDirs,
	}
}

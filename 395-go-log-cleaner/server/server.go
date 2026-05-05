package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"

	"go-log-cleaner/common"
)

type Server struct {
	host     string
	port     string
	listener net.Listener
	tasks    map[string]*common.CleanTaskResult
	history  []common.CleanHistoryRecord
	mu       sync.RWMutex
}

func NewServer(host, port string) *Server {
	return &Server{
		host:    host,
		port:    port,
		tasks:   make(map[string]*common.CleanTaskResult),
		history: make([]common.CleanHistoryRecord, 0),
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	s.listener = listener

	fmt.Printf("Log cleaner server started on %s\n", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accepting connection: %v\n", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		s.sendErrorResponse(conn, "failed to read request: "+err.Error())
		return
	}

	var req common.Request
	if err := json.Unmarshal(data, &req); err != nil {
		s.sendErrorResponse(conn, "failed to parse request: "+err.Error())
		return
	}

	s.handleRequest(conn, &req)
}

func (s *Server) handleRequest(conn net.Conn, req *common.Request) {
	switch req.Type {
	case common.RequestTypeSubmitTask:
		s.handleSubmitTask(conn, req.Payload)
	case common.RequestTypeGetHistory:
		s.handleGetHistory(conn)
	case common.RequestTypeGetTask:
		s.handleGetTask(conn, req.Payload)
	case common.RequestTypeListTasks:
		s.handleListTasks(conn)
	case common.RequestTypeHealthCheck:
		s.handleHealthCheck(conn)
	default:
		s.sendErrorResponse(conn, "unknown request type")
	}
}

func (s *Server) handleSubmitTask(conn net.Conn, payload json.RawMessage) {
	var taskReq common.CleanTaskRequest
	if err := json.Unmarshal(payload, &taskReq); err != nil {
		s.sendErrorResponse(conn, "failed to parse task request: "+err.Error())
		return
	}

	taskID := common.GenerateTaskID()
	result := &common.CleanTaskResult{
		TaskID: taskID,
		Status: common.TaskStatusPending,
	}

	s.mu.Lock()
	s.tasks[taskID] = result
	s.mu.Unlock()

	go s.executeTask(taskID, &taskReq)

	resp := common.Response{
		Status:  common.ResponseStatusSuccess,
		Payload: common.SubmitTaskResponse{TaskID: taskID},
	}

	s.sendResponse(conn, &resp)
}

func (s *Server) executeTask(taskID string, taskReq *common.CleanTaskRequest) {
	s.mu.Lock()
	result := s.tasks[taskID]
	s.mu.Unlock()

	if result == nil {
		return
	}

	cleaner := NewCleaner(*taskReq)
	execResult, err := cleaner.Execute()
	if err != nil {
		result.Status = common.TaskStatusFailed
		result.Error = err.Error()
	} else {
		result.Status = execResult.Status
		result.FilesToDelete = execResult.FilesToDelete
		result.FilesDeleted = execResult.FilesDeleted
		result.TotalSizeToDelete = execResult.TotalSizeToDelete
		result.TotalSizeDeleted = execResult.TotalSizeDeleted
		result.FilesSkipped = execResult.FilesSkipped
		result.StartTime = execResult.StartTime
		result.EndTime = execResult.EndTime
		result.Error = execResult.Error
	}

	s.mu.Lock()
	s.tasks[taskID] = result
	s.history = append(s.history, common.CleanHistoryRecord{
		TaskID:       taskID,
		TargetDir:    taskReq.TargetDir,
		Mode:         taskReq.Mode,
		Status:       result.Status,
		FilesDeleted: len(result.FilesDeleted),
		SizeDeleted:  result.TotalSizeDeleted,
		ExecTime:     result.StartTime,
	})
	s.mu.Unlock()
}

func (s *Server) handleGetHistory(conn net.Conn) {
	s.mu.RLock()
	history := make([]common.CleanHistoryRecord, len(s.history))
	copy(history, s.history)
	s.mu.RUnlock()

	resp := common.Response{
		Status:  common.ResponseStatusSuccess,
		Payload: common.GetHistoryResponse{Records: history},
	}

	s.sendResponse(conn, &resp)
}

func (s *Server) handleGetTask(conn net.Conn, payload json.RawMessage) {
	var taskReq struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(payload, &taskReq); err != nil {
		s.sendErrorResponse(conn, "failed to parse task id: "+err.Error())
		return
	}

	s.mu.RLock()
	result, exists := s.tasks[taskReq.TaskID]
	s.mu.RUnlock()

	if !exists {
		s.sendErrorResponse(conn, "task not found")
		return
	}

	resp := common.Response{
		Status:  common.ResponseStatusSuccess,
		Payload: common.GetTaskResponse{Result: *result},
	}

	s.sendResponse(conn, &resp)
}

func (s *Server) handleListTasks(conn net.Conn) {
	s.mu.RLock()
	tasks := make([]common.CleanTaskResult, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, *task)
	}
	s.mu.RUnlock()

	resp := common.Response{
		Status:  common.ResponseStatusSuccess,
		Payload: common.ListTasksResponse{Tasks: tasks},
	}

	s.sendResponse(conn, &resp)
}

func (s *Server) handleHealthCheck(conn net.Conn) {
	resp := common.Response{
		Status:  common.ResponseStatusSuccess,
		Message: "OK",
	}

	s.sendResponse(conn, &resp)
}

func (s *Server) sendResponse(conn net.Conn, resp *common.Response) {
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal response: %v\n", err)
		return
	}

	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send response: %v\n", err)
	}
}

func (s *Server) sendErrorResponse(conn net.Conn, message string) {
	resp := common.Response{
		Status:  common.ResponseStatusError,
		Message: message,
	}

	s.sendResponse(conn, &resp)
}

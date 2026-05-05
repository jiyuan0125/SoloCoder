package tcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"scheduler/internal/cron"
	"scheduler/internal/protocol"
	"scheduler/internal/scheduler"
)

type TCPServer struct {
	sched    *scheduler.Scheduler
	addr     string
	listener net.Listener
	wg       sync.WaitGroup
	stopChan chan struct{}
}

func NewTCPServer(addr string, s *scheduler.Scheduler) *TCPServer {
	return &TCPServer{
		sched:    s,
		addr:     addr,
		stopChan: make(chan struct{}),
	}
}

func (t *TCPServer) Start() error {
	var err error
	t.listener, err = net.Listen("tcp", t.addr)
	if err != nil {
		return err
	}

	go t.accept()
	return nil
}

func (t *TCPServer) accept() {
	for {
		select {
		case <-t.stopChan:
			return
		default:
			conn, err := t.listener.Accept()
			if err != nil {
				select {
				case <-t.stopChan:
					return
				default:
					continue
				}
			}
			t.wg.Add(1)
			go t.handleConn(conn)
		}
	}
}

func (t *TCPServer) handleConn(conn net.Conn) {
	defer t.wg.Done()
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		select {
		case <-t.stopChan:
			return
		default:
		}

		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var req protocol.Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			t.sendResponse(conn, protocol.Response{
				Success: false,
				Error:   "invalid request format: " + err.Error(),
			})
			continue
		}

		resp := t.handleRequest(&req)
		t.sendResponse(conn, resp)
	}
}

func (t *TCPServer) sendResponse(conn net.Conn, resp protocol.Response) {
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	data = append(data, '\n')
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	conn.Write(data)
}

func (t *TCPServer) handleRequest(req *protocol.Request) protocol.Response {
	switch req.Operation {
	case protocol.OpPing:
		return protocol.Response{Success: true}
	case protocol.OpListTasks:
		return t.handleListTasks()
	case protocol.OpGetTask:
		return t.handleGetTask(req)
	case protocol.OpAddTask:
		return t.handleAddTask(req)
	case protocol.OpDeleteTask:
		return t.handleDeleteTask(req)
	case protocol.OpListExecutions:
		return t.handleListExecutions(req)
	default:
		return protocol.Response{
			Success: false,
			Error:   fmt.Sprintf("unknown operation: %s", req.Operation),
		}
	}
}

func (t *TCPServer) handleListTasks() protocol.Response {
	tasks := t.sched.ListTaskStatuses()
	data := map[string]interface{}{
		"tasks": tasks,
	}
	return protocol.Response{
		Success: true,
		Data:    data,
	}
}

func (t *TCPServer) handleGetTask(req *protocol.Request) protocol.Response {
	name, ok := req.Data["name"].(string)
	if !ok || name == "" {
		return protocol.Response{
			Success: false,
			Error:   "task name is required",
		}
	}

	status, err := t.sched.GetTaskStatus(name)
	if err != nil {
		if os.IsNotExist(err) {
			return protocol.Response{
				Success: false,
				Error:   "task not found",
			}
		}
		return protocol.Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	data := map[string]interface{}{
		"task": status,
	}
	return protocol.Response{
		Success: true,
		Data:    data,
	}
}

func (t *TCPServer) handleAddTask(req *protocol.Request) protocol.Response {
	name, _ := req.Data["name"].(string)
	command, _ := req.Data["command"].(string)
	cronExpr, _ := req.Data["cron_expr"].(string)

	if name == "" {
		return protocol.Response{
			Success: false,
			Error:   "task name is required",
		}
	}
	if command == "" {
		return protocol.Response{
			Success: false,
			Error:   "command is required",
		}
	}
	if cronExpr == "" {
		return protocol.Response{
			Success: false,
			Error:   "cron expression is required",
		}
	}

	if _, err := cron.ParseCron(cronExpr); err != nil {
		return protocol.Response{
			Success: false,
			Error:   "invalid cron expression: " + err.Error(),
		}
	}

	cfg := protocol.TaskConfig{
		Name:     name,
		Command:  command,
		CronExpr: cronExpr,
		Disabled: false,
	}

	if timeout, ok := req.Data["timeout"].(float64); ok {
		cfg.Timeout = time.Duration(timeout) * time.Second
	}
	if maxRetry, ok := req.Data["max_retry"].(float64); ok {
		val := int(maxRetry)
		cfg.MaxRetry = &val
	}
	if retryInterval, ok := req.Data["retry_interval"].(float64); ok {
		cfg.RetryInterval = time.Duration(retryInterval) * time.Second
	}

	if err := t.sched.AddTask(cfg); err != nil {
		if os.IsExist(err) {
			return protocol.Response{
				Success: false,
				Error:   "task already exists",
			}
		}
		return protocol.Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return protocol.Response{Success: true}
}

func (t *TCPServer) handleDeleteTask(req *protocol.Request) protocol.Response {
	name, ok := req.Data["name"].(string)
	if !ok || name == "" {
		return protocol.Response{
			Success: false,
			Error:   "task name is required",
		}
	}

	if err := t.sched.DeleteTask(name); err != nil {
		if os.IsNotExist(err) {
			return protocol.Response{
				Success: false,
				Error:   "task not found",
			}
		}
		return protocol.Response{
			Success: false,
			Error:   err.Error(),
		}
	}

	return protocol.Response{Success: true}
}

func (t *TCPServer) handleListExecutions(req *protocol.Request) protocol.Response {
	var executions []protocol.ExecutionRecord

	if taskName, ok := req.Data["task_name"].(string); ok && taskName != "" {
		executions = t.sched.GetExecutions(taskName)
	} else {
		executions = t.sched.GetAllExecutions()
	}

	data := map[string]interface{}{
		"executions": executions,
	}
	return protocol.Response{
		Success: true,
		Data:    data,
	}
}

func (t *TCPServer) Stop(ctx context.Context) error {
	close(t.stopChan)
	t.listener.Close()

	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

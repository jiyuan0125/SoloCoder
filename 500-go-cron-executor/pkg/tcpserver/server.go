package tcpserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"cron-executor/pkg/common"
	"cron-executor/pkg/model"
	"cron-executor/pkg/server"
)

type TCPServer struct {
	port   int
	server *server.Server
	ln     net.Listener
	stopCh chan struct{}
}

func NewTCPServer(port int, srv *server.Server) *TCPServer {
	return &TCPServer{
		port:   port,
		server: srv,
		stopCh: make(chan struct{}),
	}
}

func (ts *TCPServer) Start() error {
	addr := fmt.Sprintf(":%d", ts.port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	ts.ln = ln

	go ts.acceptLoop()
	return nil
}

func (ts *TCPServer) Stop() {
	close(ts.stopCh)
	if ts.ln != nil {
		ts.ln.Close()
	}
}

func (ts *TCPServer) acceptLoop() {
	for {
		select {
		case <-ts.stopCh:
			return
		default:
		}

		conn, err := ts.ln.Accept()
		if err != nil {
			select {
			case <-ts.stopCh:
				return
			default:
			}
			continue
		}

		go ts.handleConn(conn)
	}
}

func (ts *TCPServer) handleConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() {
				continue
			}
			return
		}

		var msg common.Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			ts.sendError(writer, "invalid message format")
			continue
		}

		response := ts.handleMessage(&msg)

		respBytes, _ := json.Marshal(response)
		writer.Write(respBytes)
		writer.WriteByte('\n')
		writer.Flush()
	}
}

func (ts *TCPServer) handleMessage(msg *common.Message) *common.Message {
	switch msg.Type {
	case common.MsgTypePing:
		return &common.Message{Type: common.MsgTypePong}

	case common.MsgTypeListTasks:
		return ts.handleListTasks()

	case common.MsgTypeTriggerTask:
		return ts.handleTriggerTask(msg.Payload)

	case common.MsgTypeGetStatus:
		return ts.handleGetStatus(msg.Payload)

	case common.MsgTypeGetExecutions:
		return ts.handleGetExecutions(msg.Payload)

	case common.MsgTypeGetStats:
		return ts.handleGetStats()

	default:
		return ts.createErrorResponse("unknown message type")
	}
}

func (ts *TCPServer) handleListTasks() *common.Message {
	tasks := ts.server.ListTasks()

	var taskInfos []common.TaskInfo
	for _, t := range tasks {
		ti := common.TaskInfo{
			ID:           t.ID,
			Name:         t.Name,
			CronExpr:     t.CronExpr,
			Command:      t.Command,
			Timeout:      t.Timeout,
			MaxRetries:   t.MaxRetries,
			Status:       string(t.Status),
			Dependencies: t.Dependencies,
		}
		if !t.LastRun.IsZero() {
			ti.LastRun = t.LastRun.Format(time.RFC3339)
		}
		if !t.NextRun.IsZero() {
			ti.NextRun = t.NextRun.Format(time.RFC3339)
		}
		taskInfos = append(taskInfos, ti)
	}

	resp := common.TaskListResponse{Tasks: taskInfos}
	payload, _ := json.Marshal(resp)

	return &common.Message{
		Type:    common.MsgTypeTaskList,
		Payload: payload,
	}
}

func (ts *TCPServer) handleTriggerTask(payload json.RawMessage) *common.Message {
	var req common.TriggerTaskRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return ts.createErrorResponse("invalid trigger request")
	}

	execID, err := ts.server.TriggerTaskManually(req.TaskName)
	if err != nil {
		resp := common.TriggerResultResponse{
			Success: false,
			Message: err.Error(),
		}
		respBytes, _ := json.Marshal(resp)
		return &common.Message{
			Type:    common.MsgTypeTriggerResult,
			Payload: respBytes,
		}
	}

	resp := common.TriggerResultResponse{
		Success: true,
		Message: "task triggered",
		ExecID:  execID,
	}
	respBytes, _ := json.Marshal(resp)
	return &common.Message{
		Type:    common.MsgTypeTriggerResult,
		Payload: respBytes,
	}
}

func (ts *TCPServer) handleGetStatus(payload json.RawMessage) *common.Message {
	var req common.GetStatusRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return ts.createErrorResponse("invalid status request")
	}

	task := ts.server.GetTaskByName(req.TaskName)
	if task == nil {
		return ts.createErrorResponse("task not found")
	}

	ti := common.TaskInfo{
		ID:           task.ID,
		Name:         task.Name,
		CronExpr:     task.CronExpr,
		Command:      task.Command,
		Timeout:      task.Timeout,
		MaxRetries:   task.MaxRetries,
		Status:       string(task.Status),
		Dependencies: task.Dependencies,
	}
	if !task.LastRun.IsZero() {
		ti.LastRun = task.LastRun.Format(time.RFC3339)
	}
	if !task.NextRun.IsZero() {
		ti.NextRun = task.NextRun.Format(time.RFC3339)
	}

	resp := common.TaskStatusResponse{
		TaskInfo:     ti,
		ActiveExecID: task.ActiveExecID,
	}
	respBytes, _ := json.Marshal(resp)

	return &common.Message{
		Type:    common.MsgTypeTaskStatus,
		Payload: respBytes,
	}
}

func (ts *TCPServer) handleGetExecutions(payload json.RawMessage) *common.Message {
	var req common.GetExecutionsRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		req.TaskName = ""
		req.Limit = 20
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	executions := ts.server.GetExecutions(req.TaskName, req.Limit)

	var execInfos []common.ExecutionInfo
	for _, e := range executions {
		ei := common.ExecutionInfo{
			ID:              e.ID,
			TaskID:          e.TaskID,
			TaskName:        e.TaskName,
			StartTime:       e.StartTime.Format(time.RFC3339),
			ExitCode:        e.ExitCode,
			Status:          string(e.Status),
			RetryCount:      e.RetryCount,
			IsTimeout:       e.IsTimeout,
			IsManualTrigger: e.IsManualTrigger,
		}
		if !e.EndTime.IsZero() {
			ei.EndTime = e.EndTime.Format(time.RFC3339)
			ei.Duration = e.Duration.String()
		}
		execInfos = append(execInfos, ei)
	}

	resp := common.ExecutionListResponse{Executions: execInfos}
	respBytes, _ := json.Marshal(resp)

	return &common.Message{
		Type:    common.MsgTypeExecutionList,
		Payload: respBytes,
	}
}

func (ts *TCPServer) handleGetStats() *common.Message {
	stats := ts.server.GetStats()

	var recentFailures []common.ExecutionInfo
	for _, e := range stats.RecentFailures {
		ei := common.ExecutionInfo{
			ID:              e.ID,
			TaskID:          e.TaskID,
			TaskName:        e.TaskName,
			StartTime:       e.StartTime.Format(time.RFC3339),
			ExitCode:        e.ExitCode,
			Status:          string(e.Status),
			RetryCount:      e.RetryCount,
			IsTimeout:       e.IsTimeout,
			IsManualTrigger: e.IsManualTrigger,
		}
		if !e.EndTime.IsZero() {
			ei.EndTime = e.EndTime.Format(time.RFC3339)
		}
		recentFailures = append(recentFailures, ei)
	}

	resp := common.StatsResponse{
		TotalTasks:       stats.TotalTasks,
		RunningTasks:     stats.RunningTasks,
		TotalExecutions:  stats.TotalExecutions,
		SuccessRate:      stats.SuccessRate,
		AvgDuration:      stats.AvgDuration,
		RecentFailures:   recentFailures,
	}

	runningTasks := 0
	for _, t := range ts.server.ListTasks() {
		if t.Status == model.TaskStatusRunning {
			runningTasks++
		}
	}
	resp.RunningTasks = runningTasks
	resp.TotalTasks = len(ts.server.ListTasks())

	respBytes, _ := json.Marshal(resp)

	return &common.Message{
		Type:    common.MsgTypeStatsResult,
		Payload: respBytes,
	}
}

func (ts *TCPServer) createErrorResponse(message string) *common.Message {
	errResp := common.ErrorResponse{Message: message}
	payload, _ := json.Marshal(errResp)
	return &common.Message{
		Type:    common.MsgTypeError,
		Payload: payload,
	}
}

func (ts *TCPServer) sendError(writer *bufio.Writer, message string) {
	resp := ts.createErrorResponse(message)
	respBytes, _ := json.Marshal(resp)
	writer.Write(respBytes)
	writer.WriteByte('\n')
	writer.Flush()
}

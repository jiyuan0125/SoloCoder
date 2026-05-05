package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"go-log-merger/protocol"
)

func handleConnection(conn net.Conn, taskManager *TaskManager) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("读取消息失败: %v", err)
			}
			return
		}

		msg, err := protocol.DecodeMessage([]byte(line))
		if err != nil {
			sendError(writer, protocol.ErrInvalidPayload)
			continue
		}

		response, err := processMessage(msg, taskManager)
		if err != nil {
			sendError(writer, err)
			continue
		}

		if response != nil {
			data, err := response.Encode()
			if err != nil {
				log.Printf("编码响应失败: %v", err)
				continue
			}
			if _, err := writer.WriteString(string(data) + "\n"); err != nil {
				log.Printf("写入响应失败: %v", err)
				return
			}
			writer.Flush()
		}
	}
}

func processMessage(msg *protocol.Message, taskManager *TaskManager) (*protocol.Message, error) {
	switch msg.Type {
	case protocol.MsgTypeSubmitTask:
		return handleSubmitTask(msg, taskManager)
	case protocol.MsgTypeGetStatus:
		return handleGetStatus(msg, taskManager)
	default:
		return nil, protocol.ErrInvalidMessageType
	}
}

func handleSubmitTask(msg *protocol.Message, taskManager *TaskManager) (*protocol.Message, error) {
	var req protocol.SubmitTaskRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return nil, protocol.ErrInvalidPayload
	}

	if len(req.InputFiles) < 2 {
		return nil, fmt.Errorf("至少需要2个输入文件")
	}

	if req.TimeFormat == "" {
		req.TimeFormat = protocol.DefaultTimeFormat
	}

	task, err := taskManager.SubmitTask(&req)
	if err != nil {
		return nil, err
	}

	resp := &protocol.SubmitTaskResponse{
		TaskID:  task.ID,
		Status:  task.Status,
		Message: "任务已提交",
	}

	return protocol.NewMessage(protocol.MsgTypeTaskSubmitted, resp)
}

func handleGetStatus(msg *protocol.Message, taskManager *TaskManager) (*protocol.Message, error) {
	var req protocol.GetStatusRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return nil, protocol.ErrInvalidPayload
	}

	task, exists := taskManager.GetTask(req.TaskID)
	if !exists {
		return nil, protocol.ErrTaskNotFound
	}

	resp := &protocol.GetStatusResponse{
		TaskID:          task.ID,
		Status:          task.Status,
		TotalFiles:      task.TotalFiles,
		ProcessedLines:  task.ProcessedLines,
		CurrentProgress: task.Progress,
		OutputFile:      task.OutputFile,
	}

	if task.Error != nil {
		resp.ErrorMessage = task.Error.Error()
	}

	return protocol.NewMessage(protocol.MsgTypeStatusResponse, resp)
}

func sendError(writer *bufio.Writer, err error) {
	errResp := &protocol.ErrorResponse{
		Message: err.Error(),
	}
	msg, _ := protocol.NewMessage(protocol.MsgTypeError, errResp)
	data, _ := msg.Encode()
	writer.WriteString(string(data) + "\n")
	writer.Flush()
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"go-log-merger/protocol"
)

func submitTask(serverAddr string, inputFiles []string, outputFile string, timeFormat string, resume bool) (string, error) {
	conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if err != nil {
		return "", fmt.Errorf("连接服务端失败: %v", err)
	}
	defer conn.Close()

	req := &protocol.SubmitTaskRequest{
		InputFiles: inputFiles,
		OutputFile: outputFile,
		TimeFormat: timeFormat,
		Resume:     resume,
	}

	msg, err := protocol.NewMessage(protocol.MsgTypeSubmitTask, req)
	if err != nil {
		return "", fmt.Errorf("构建消息失败: %v", err)
	}

	data, err := msg.Encode()
	if err != nil {
		return "", fmt.Errorf("编码消息失败: %v", err)
	}

	if _, err := conn.Write(append(data, '\n')); err != nil {
		return "", fmt.Errorf("发送消息失败: %v", err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	respMsg, err := protocol.DecodeMessage([]byte(line))
	if err != nil {
		return "", fmt.Errorf("解码响应失败: %v", err)
	}

	if respMsg.Type == protocol.MsgTypeError {
		var errResp protocol.ErrorResponse
		if err := json.Unmarshal(respMsg.Payload, &errResp); err != nil {
			return "", fmt.Errorf("服务端返回错误")
		}
		return "", fmt.Errorf(errResp.Message)
	}

	if respMsg.Type != protocol.MsgTypeTaskSubmitted {
		return "", fmt.Errorf("意外的响应类型: %s", respMsg.Type)
	}

	var taskResp protocol.SubmitTaskResponse
	if err := json.Unmarshal(respMsg.Payload, &taskResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	return taskResp.TaskID, nil
}

func showTaskStatus(serverAddr string, taskID string) error {
	conn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("连接服务端失败: %v", err)
	}
	defer conn.Close()

	req := &protocol.GetStatusRequest{
		TaskID: taskID,
	}

	msg, err := protocol.NewMessage(protocol.MsgTypeGetStatus, req)
	if err != nil {
		return fmt.Errorf("构建消息失败: %v", err)
	}

	data, err := msg.Encode()
	if err != nil {
		return fmt.Errorf("编码消息失败: %v", err)
	}

	if _, err := conn.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	respMsg, err := protocol.DecodeMessage([]byte(line))
	if err != nil {
		return fmt.Errorf("解码响应失败: %v", err)
	}

	if respMsg.Type == protocol.MsgTypeError {
		var errResp protocol.ErrorResponse
		if err := json.Unmarshal(respMsg.Payload, &errResp); err != nil {
			return fmt.Errorf("服务端返回错误")
		}
		return fmt.Errorf(errResp.Message)
	}

	if respMsg.Type != protocol.MsgTypeStatusResponse {
		return fmt.Errorf("意外的响应类型: %s", respMsg.Type)
	}

	var statusResp protocol.GetStatusResponse
	if err := json.Unmarshal(respMsg.Payload, &statusResp); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	printStatus(&statusResp)
	return nil
}

func printStatus(status *protocol.GetStatusResponse) {
	fmt.Println("=== 任务状态 ===")
	fmt.Printf("任务ID: %s\n", status.TaskID)
	fmt.Printf("状态: %s\n", getStatusText(status.Status))
	fmt.Printf("总文件数: %d\n", status.TotalFiles)
	fmt.Printf("已处理行数: %d\n", status.ProcessedLines)
	fmt.Printf("进度: %.2f%%\n", status.CurrentProgress)

	if status.OutputFile != "" {
		fmt.Printf("输出文件: %s\n", status.OutputFile)
	}

	if status.ErrorMessage != "" {
		fmt.Printf("错误: %s\n", status.ErrorMessage)
	}
}

func getStatusText(status protocol.TaskStatus) string {
	switch status {
	case protocol.TaskStatusPending:
		return "等待中"
	case protocol.TaskStatusRunning:
		return "运行中"
	case protocol.TaskStatusCompleted:
		return "已完成"
	case protocol.TaskStatusFailed:
		return "失败"
	default:
		return string(status)
	}
}

package main

import (
	"deploybot/common"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	addr        string
	listener    net.Listener
	wg          sync.WaitGroup
	stopChan    chan struct{}
	running     atomic.Bool
	deployMgr   *DeployManager
	historyMgr  *HistoryManager
}

func NewServer(addr string) *Server {
	return &Server{
		addr:       addr,
		stopChan:   make(chan struct{}),
		deployMgr:  NewDeployManager(),
		historyMgr: NewHistoryManager(),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.listener = ln
	s.running.Store(true)
	Info("服务监听: %s", s.addr)

	for {
		select {
		case <-s.stopChan:
			return nil
		default:
		}

		ln.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))
		conn, err := ln.Accept()

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if s.running.Load() {
				Error("接受连接失败: %v", err)
			}
			return err
		}

		Info("新连接: %s", conn.RemoteAddr())
		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

func (s *Server) Stop() {
	if !s.running.Swap(false) {
		return
	}

	close(s.stopChan)
	if s.listener != nil {
		s.listener.Close()
	}
	Info("等待所有连接处理完成...")
	s.wg.Wait()
	s.deployMgr.Stop()
	Info("所有连接已处理完成")
}

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	bc := common.NewBufferedConn(conn)

	for {
		select {
		case <-s.stopChan:
			return
		default:
		}

		req, err := bc.ReceiveRequest()
		if err != nil {
			if err.Error() != "EOF" {
				Warn("接收请求失败: %v", err)
			}
			return
		}

		Info("收到命令: %s", req.Command)
		resp := s.handleCommand(req)

		if err := bc.SendResponse(resp); err != nil {
			Error("发送响应失败: %v", err)
			return
		}
	}
}

func (s *Server) handleCommand(req *common.DeployRequest) *common.DeployResponse {
	switch req.Command {
	case common.CmdDeploy:
		return s.handleDeploy(req)
	case common.CmdRollback:
		return s.handleRollback(req)
	case common.CmdStatus:
		return s.handleStatus(req)
	case common.CmdHistory:
		return s.handleHistory(req)
	default:
		return &common.DeployResponse{
			Success: false,
			Message: fmt.Sprintf("未知命令: %s", req.Command),
		}
	}
}

func (s *Server) handleDeploy(req *common.DeployRequest) *common.DeployResponse {
	if req.Config == nil {
		return &common.DeployResponse{
			Success: false,
			Message: "缺少部署配置",
		}
	}

	deployment := &common.Deployment{
		ID:         generateDeployID(),
		Status:     common.StatusQueued,
		StartTime:  time.Now(),
		Config:     req.Config,
		IsRollback: false,
	}

	s.historyMgr.Add(deployment)

	go func() {
		s.deployMgr.Execute(deployment)
		s.historyMgr.Update(deployment)
	}()

	return &common.DeployResponse{
		Success:  true,
		Message:  "部署任务已提交",
		DeployID: deployment.ID,
		Status:   deployment.Status,
	}
}

func (s *Server) handleRollback(req *common.DeployRequest) *common.DeployResponse {
	if req.BackupPath == "" && req.Config == nil {
		return &common.DeployResponse{
			Success: false,
			Message: "回滚需要指定备份路径或配置",
		}
	}

	deployment := &common.Deployment{
		ID:         generateDeployID(),
		Status:     common.StatusQueued,
		StartTime:  time.Now(),
		Config:     req.Config,
		BackupPath: req.BackupPath,
		IsRollback: true,
	}

	s.historyMgr.Add(deployment)

	go func() {
		s.deployMgr.ExecuteRollback(deployment)
		s.historyMgr.Update(deployment)
	}()

	return &common.DeployResponse{
		Success:  true,
		Message:  "回滚任务已提交",
		DeployID: deployment.ID,
		Status:   deployment.Status,
	}
}

func (s *Server) handleStatus(req *common.DeployRequest) *common.DeployResponse {
	if req.Config != nil && req.Config.BinaryPath != "" {
		return &common.DeployResponse{
			Success: true,
			Message: "服务状态检查",
		}
	}

	if current := s.deployMgr.Current(); current != nil {
		return &common.DeployResponse{
			Success:     true,
			Message:     "当前正在进行的部署",
			DeployID:    current.ID,
			Status:      current.Status,
			Deployments: []*common.Deployment{current},
		}
	}

	return &common.DeployResponse{
		Success: true,
		Message: "当前没有正在进行的部署",
		Status:  common.StatusQueued,
	}
}

func (s *Server) handleHistory(req *common.DeployRequest) *common.DeployResponse {
	history := s.historyMgr.List()

	if len(history) == 0 {
		return &common.DeployResponse{
			Success: true,
			Message: "暂无部署历史",
		}
	}

	return &common.DeployResponse{
		Success:     true,
		Message:     fmt.Sprintf("共 %d 条历史记录", len(history)),
		Deployments: history,
	}
}

func generateDeployID() string {
	return fmt.Sprintf("deploy-%s", time.Now().Format("20060102-150405"))
}

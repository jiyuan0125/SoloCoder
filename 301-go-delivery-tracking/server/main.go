package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go-delivery-tracking/common"
)

const (
	defaultPort = "8080"
)

type Server struct {
	store  *DeliveryStore
	addr   string
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(addr string) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		store:  NewDeliveryStore(),
		addr:   addr,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Server) Run() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("监听失败: %w", err)
	}
	defer ln.Close()

	log.Printf("服务器已启动，监听端口: %s", s.addr)

	go s.handleSignals()

	for {
		select {
		case <-s.ctx.Done():
			log.Println("服务器正在关闭...")
			ln.Close()
			s.wg.Wait()
			log.Println("服务器已优雅关闭")
			return nil
		default:
			ln.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))
			conn, err := ln.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				if s.ctx.Err() != nil {
					continue
				}
				log.Printf("接受连接失败: %v", err)
				continue
			}
			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

func (s *Server) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-s.ctx.Done():
		return
	case sig := <-sigChan:
		log.Printf("收到信号: %v，开始优雅关闭", sig)
		s.cancel()
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	log.Printf("新连接: %s", remoteAddr)

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("连接关闭（服务器关闭）: %s", remoteAddr)
			return
		default:
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			msg, err := common.ReadMessage(conn)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Printf("读取消息失败: %v", err)
				return
			}
			if msg == nil {
				log.Printf("连接关闭: %s", remoteAddr)
				return
			}

			s.handleMessage(conn, msg)
		}
	}
}

func (s *Server) handleMessage(conn net.Conn, msg *common.Message) {
	switch msg.Type {
	case common.MsgTypeReport:
		s.handleReport(conn, msg)
	case common.MsgTypeQuery:
		s.handleQuery(conn, msg)
	default:
		log.Printf("未知消息类型: %s", msg.Type)
		respMsg, _ := common.BuildResponseMessage(false, fmt.Sprintf("未知消息类型: %s", msg.Type), nil)
		common.WriteMessage(conn, respMsg)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	addr := fmt.Sprintf(":%s", port)
	server := NewServer(addr)

	if err := server.Run(); err != nil {
		log.Fatalf("服务器运行失败: %v", err)
	}
}

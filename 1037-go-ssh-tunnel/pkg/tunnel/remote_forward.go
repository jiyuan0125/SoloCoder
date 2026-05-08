package tunnel

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"ssh-tunnel/pkg/common"
)

type RemoteForward struct {
	config *common.TunnelConfig
	conn   *SSHConnection
	listener net.Listener
	ctx    context.Context
	cancel context.CancelFunc
	logger *log.Logger
	mu     sync.Mutex
	wg     sync.WaitGroup
}

func NewRemoteForward(cfg *common.TunnelConfig, conn *SSHConnection) *RemoteForward {
	ctx, cancel := context.WithCancel(context.Background())
	return &RemoteForward{
		config: cfg,
		conn:   conn,
		ctx:    ctx,
		cancel: cancel,
		logger: log.New(os.Stdout, "[remote-forward] ", log.LstdFlags),
	}
}

func (rf *RemoteForward) Start() error {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	remoteAddr := net.JoinHostPort(rf.config.LocalAddress, strconv.Itoa(rf.config.LocalPort))
	
	listener, err := rf.conn.Listen("tcp", remoteAddr)
	if err != nil {
		return err
	}
	
	rf.listener = listener
	rf.logger.Printf("远程转发监听: %s -> %s:%d", remoteAddr, rf.config.RemoteAddress, rf.config.RemotePort)

	rf.wg.Add(1)
	go rf.acceptLoop()

	return nil
}

func (rf *RemoteForward) Stop() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	rf.cancel()

	if rf.listener != nil {
		_ = rf.listener.Close()
		rf.listener = nil
	}
	
	rf.wg.Wait()
	rf.logger.Printf("远程转发已停止")
}

func (rf *RemoteForward) acceptLoop() {
	defer rf.wg.Done()

	for {
		select {
		case <-rf.ctx.Done():
			return
		default:
		}

		remoteConn, err := rf.listener.Accept()
		if err != nil {
			select {
			case <-rf.ctx.Done():
				return
			default:
				rf.logger.Printf("接受远程连接失败: %v", err)
				continue
			}
		}

		rf.logger.Printf("接受远程连接: %s", remoteConn.RemoteAddr())
		rf.wg.Add(1)
		go rf.handleConnection(remoteConn)
	}
}

func (rf *RemoteForward) handleConnection(remoteConn net.Conn) {
	defer rf.wg.Done()
	defer remoteConn.Close()

	localAddr := net.JoinHostPort(rf.config.RemoteAddress, strconv.Itoa(rf.config.RemotePort))
	
	localConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		rf.logger.Printf("连接本地目标失败: %v", err)
		return
	}
	defer localConn.Close()

	rf.logger.Printf("建立转发: %s <-> %s", remoteConn.RemoteAddr(), localAddr)

	done := make(chan struct{}, 2)

	go func() {
		io.Copy(localConn, remoteConn)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(remoteConn, localConn)
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-rf.ctx.Done():
	}
}

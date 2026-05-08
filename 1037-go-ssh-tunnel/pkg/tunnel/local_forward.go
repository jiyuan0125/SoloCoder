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

type LocalForward struct {
	config *common.TunnelConfig
	conn   *SSHConnection
	listener net.Listener
	ctx    context.Context
	cancel context.CancelFunc
	logger *log.Logger
	mu     sync.Mutex
	wg     sync.WaitGroup
}

func NewLocalForward(cfg *common.TunnelConfig, conn *SSHConnection) *LocalForward {
	ctx, cancel := context.WithCancel(context.Background())
	return &LocalForward{
		config: cfg,
		conn:   conn,
		ctx:    ctx,
		cancel: cancel,
		logger: log.New(os.Stdout, "[local-forward] ", log.LstdFlags),
	}
}

func (lf *LocalForward) Start() error {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	localAddr := net.JoinHostPort(lf.config.LocalAddress, strconv.Itoa(lf.config.LocalPort))
	
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return err
	}
	
	lf.listener = listener
	lf.logger.Printf("本地转发监听: %s -> %s:%d", localAddr, lf.config.RemoteAddress, lf.config.RemotePort)

	lf.wg.Add(1)
	go lf.acceptLoop()

	return nil
}

func (lf *LocalForward) Stop() {
	lf.mu.Lock()
	defer lf.mu.Unlock()

	lf.cancel()

	if lf.listener != nil {
		_ = lf.listener.Close()
		lf.listener = nil
	}
	
	lf.wg.Wait()
	lf.logger.Printf("本地转发已停止")
}

func (lf *LocalForward) acceptLoop() {
	defer lf.wg.Done()

	for {
		select {
		case <-lf.ctx.Done():
			return
		default:
		}

		localConn, err := lf.listener.Accept()
		if err != nil {
			select {
			case <-lf.ctx.Done():
				return
			default:
				lf.logger.Printf("接受连接失败: %v", err)
				continue
			}
		}

		lf.logger.Printf("接受本地连接: %s", localConn.RemoteAddr())
		lf.wg.Add(1)
		go lf.handleConnection(localConn)
	}
}

func (lf *LocalForward) handleConnection(localConn net.Conn) {
	defer lf.wg.Done()
	defer localConn.Close()

	remoteAddr := net.JoinHostPort(lf.config.RemoteAddress, strconv.Itoa(lf.config.RemotePort))
	
	remoteConn, err := lf.conn.Dial("tcp", remoteAddr)
	if err != nil {
		lf.logger.Printf("连接远程目标失败: %v", err)
		return
	}
	defer remoteConn.Close()

	lf.logger.Printf("建立转发: %s <-> %s", localConn.RemoteAddr(), remoteAddr)

	done := make(chan struct{}, 2)

	go func() {
		io.Copy(remoteConn, localConn)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(localConn, remoteConn)
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-lf.ctx.Done():
	}
}

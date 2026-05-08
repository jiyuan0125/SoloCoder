package tunnel

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"ssh-tunnel/pkg/common"

	"golang.org/x/crypto/ssh"
)

type SSHConnection struct {
	config     *common.TunnelConfig
	client     *ssh.Client
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	activeConn map[net.Conn]struct{}
	connMu     sync.Mutex
	logger     *log.Logger
}

func NewSSHConnection(cfg *common.TunnelConfig) *SSHConnection {
	ctx, cancel := context.WithCancel(context.Background())
	return &SSHConnection{
		config:     cfg,
		ctx:        ctx,
		cancel:     cancel,
		activeConn: make(map[net.Conn]struct{}),
		logger:     log.New(os.Stdout, "[ssh-conn] ", log.LstdFlags),
	}
}

func (sc *SSHConnection) buildSSHConfig() (*ssh.ClientConfig, error) {
	var authMethod ssh.AuthMethod

	switch sc.config.AuthMethod {
	case common.AuthMethodPassword:
		authMethod = ssh.Password(sc.config.Password)
	case common.AuthMethodKey:
		key, err := os.ReadFile(sc.config.KeyFilePath)
		if err != nil {
			return nil, err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
		authMethod = ssh.PublicKeys(signer)
	default:
		return nil, fmt.Errorf("unsupported auth method: %s", sc.config.AuthMethod)
	}

	return &ssh.ClientConfig{
		User: sc.config.SSHUser,
		Auth: []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

func (sc *SSHConnection) Connect() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.client != nil {
		_ = sc.client.Close()
		sc.client = nil
	}

	cfg, err := sc.buildSSHConfig()
	if err != nil {
		return err
	}

	addr := sc.config.SSHServer
	if sc.config.SSHPort != 22 && sc.config.SSHPort != 0 {
		addr = net.JoinHostPort(sc.config.SSHServer, strconv.Itoa(sc.config.SSHPort))
	} else {
		addr = net.JoinHostPort(sc.config.SSHServer, "22")
	}

	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return err
	}

	sc.client = client
	sc.startKeepalive()
	sc.logger.Printf("SSH连接建立成功: %s", addr)
	return nil
}

func (sc *SSHConnection) Close() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.cancel()

	sc.connMu.Lock()
	for conn := range sc.activeConn {
		_ = conn.Close()
		delete(sc.activeConn, conn)
	}
	sc.connMu.Unlock()

	if sc.client != nil {
		err := sc.client.Close()
		sc.client = nil
		return err
	}
	return nil
}

func (sc *SSHConnection) IsConnected() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.client != nil
}

func (sc *SSHConnection) startKeepalive() {
	if sc.config.KeepaliveInterval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(time.Duration(sc.config.KeepaliveInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-sc.ctx.Done():
				return
			case <-ticker.C:
				sc.mu.RLock()
				client := sc.client
				sc.mu.RUnlock()

				if client == nil {
					return
				}

				_, _, err := client.SendRequest("keepalive@golang.org", true, nil)
				if err != nil {
					sc.logger.Printf("SSH Keepalive发送失败: %v", err)
					return
				}
			}
		}
	}()
}

func (sc *SSHConnection) Dial(network, addr string) (net.Conn, error) {
	sc.mu.RLock()
	client := sc.client
	sc.mu.RUnlock()

	if client == nil {
		return nil, ErrConnectionNotEstablished
	}

	conn, err := client.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	sc.connMu.Lock()
	sc.activeConn[conn] = struct{}{}
	sc.connMu.Unlock()

	go sc.trackConnection(conn)

	return conn, nil
}

func (sc *SSHConnection) trackConnection(conn net.Conn) {
	select {
	case <-sc.ctx.Done():
		_ = conn.Close()
	}
	sc.connMu.Lock()
	delete(sc.activeConn, conn)
	sc.connMu.Unlock()
}

func (sc *SSHConnection) Listen(network, addr string) (net.Listener, error) {
	sc.mu.RLock()
	client := sc.client
	sc.mu.RUnlock()

	if client == nil {
		return nil, ErrConnectionNotEstablished
	}

	return client.Listen(network, addr)
}

func (sc *SSHConnection) CloseActiveConnections() {
	sc.connMu.Lock()
	defer sc.connMu.Unlock()

	for conn := range sc.activeConn {
		_ = conn.Close()
		delete(sc.activeConn, conn)
	}
}

func copyAndClose(dst, src net.Conn) {
	defer dst.Close()
	defer src.Close()
	io.Copy(dst, src)
}

var ErrConnectionNotEstablished = fmt.Errorf("SSH connection not established")

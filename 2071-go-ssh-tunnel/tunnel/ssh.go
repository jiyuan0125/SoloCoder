package tunnel

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
	"ssh-tunnel-manager/models"
)

type Stats struct {
	BytesUp   int64
	BytesDown int64
}

type SSHConnection struct {
	config *models.TunnelConfig
	client *ssh.Client
	listener net.Listener
	stats *Stats
}

func NewSSHConnection(config *models.TunnelConfig) *SSHConnection {
	return &SSHConnection{
		config: config,
		stats:  &Stats{},
	}
}

func (s *SSHConnection) BuildSSHConfig() (*ssh.ClientConfig, error) {
	var authMethod ssh.AuthMethod
	var err error

	switch s.config.AuthType {
	case models.AuthTypePassword:
		authMethod = ssh.Password(s.config.AuthData)
	case models.AuthTypeKey:
		signer, parseErr := ssh.ParsePrivateKey([]byte(s.config.AuthData))
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", parseErr)
		}
		authMethod = ssh.PublicKeys(signer)
	default:
		return nil, fmt.Errorf("invalid auth type: %s", s.config.AuthType)
	}

	_ = err
	return &ssh.ClientConfig{
		User: s.config.SSHUser,
		Auth: []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout: 30 * time.Second,
	}, nil
}

func (s *SSHConnection) Connect(ctx context.Context) error {
	sshConfig, err := s.BuildSSHConfig()
	if err != nil {
		return err
	}

	client, err := ssh.Dial("tcp", s.config.SSHServer, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	s.client = client
	return nil
}

func (s *SSHConnection) Close() error {
	var errs []error
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.client != nil {
		if err := s.client.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing connection: %v", errs)
	}
	return nil
}

func (s *SSHConnection) GetStats() *Stats {
	return s.stats
}

func (s *SSHConnection) copyData(dst io.Writer, src io.Reader, counter *int64, done chan<- error) {
	n, err := io.Copy(dst, src)
	*counter += n
	done <- err
}

func (s *SSHConnection) handleConnection(local net.Conn) {
	defer local.Close()

	var remote net.Conn
	var err error

	switch s.config.Type {
	case models.TunnelTypeLocal:
		remote, err = s.client.Dial("tcp", fmt.Sprintf("%s:%d", s.config.RemoteHost, s.config.RemotePort))
	case models.TunnelTypeRemote:
		remote, err = s.client.Dial("tcp", fmt.Sprintf("%s:%d", s.config.RemoteHost, s.config.RemotePort))
	}

	if err != nil {
		return
	}
	defer remote.Close()

	errUp := make(chan error, 1)
	errDown := make(chan error, 1)

	go s.copyData(remote, local, &s.stats.BytesUp, errUp)
	go s.copyData(local, remote, &s.stats.BytesDown, errDown)

	<-errUp
	<-errDown
}

func (s *SSHConnection) Start(ctx context.Context) error {
	if s.client == nil {
		if err := s.Connect(ctx); err != nil {
			return err
		}
	}

	localAddr := fmt.Sprintf("0.0.0.0:%d", s.config.LocalPort)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on local port %d: %w", s.config.LocalPort, err)
	}

	s.listener = listener

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				go s.handleConnection(conn)
			}
		}
	}()

	return nil
}

func IsPortInUse(port int) (bool, int, error) {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err == nil {
		listener.Close()
		return false, 0, nil
	}
	return true, 0, nil
}

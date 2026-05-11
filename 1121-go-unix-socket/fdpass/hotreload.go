package fdpass

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

const (
	EnvListenFD       = "UNIXSOCKET_LISTEN_FD"
	EnvListenAddr     = "UNIXSOCKET_LISTEN_ADDR"
	EnvIsChild        = "UNIXSOCKET_IS_CHILD"
	EnvSocketPairFD   = "UNIXSOCKET_SOCKETPAIR_FD"
)

type HotReloadManager struct {
	listener        net.Listener
	listenAddr      string
	isPrimary       bool
	restartCount    int
	restartHistory  []RestartRecord
	activeRequests  int64
	shutdownTimeout time.Duration
	onRestart       func(oldPID, newPID int, success bool)
}

type RestartRecord struct {
	OldPID     int
	NewPID     int
	Timestamp  time.Time
	Successful bool
}

type HotReloadConfig struct {
	ShutdownTimeout time.Duration
	OnRestart       func(oldPID, newPID int, success bool)
}

func NewHotReloadManager(cfg HotReloadConfig) *HotReloadManager {
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 30 * time.Second
	}
	return &HotReloadManager{
		isPrimary:       os.Getenv(EnvIsChild) != "1",
		shutdownTimeout: cfg.ShutdownTimeout,
		onRestart:       cfg.OnRestart,
	}
}

func (m *HotReloadManager) IsPrimary() bool {
	return m.isPrimary
}

func (m *HotReloadManager) RestartCount() int {
	return m.restartCount
}

func (m *HotReloadManager) RestartHistory() []RestartRecord {
	return m.restartHistory
}

func (m *HotReloadManager) ListenAddr() string {
	return m.listenAddr
}

func (m *HotReloadManager) Start(addr string) (net.Listener, error) {
	m.listenAddr = addr

	if fdStr := os.Getenv(EnvListenFD); fdStr != "" {
		fd, err := strconv.Atoi(fdStr)
		if err != nil {
			return nil, fmt.Errorf("invalid listen fd: %w", err)
		}

		f := os.NewFile(uintptr(fd), "listener")
		if f == nil {
			return nil, fmt.Errorf("failed to create file from fd %d", fd)
		}
		defer f.Close()

		ln, err := net.FileListener(f)
		if err != nil {
			return nil, fmt.Errorf("failed to create listener from fd: %w", err)
		}
		m.listener = ln
		return ln, nil
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	m.listener = ln
	return ln, nil
}

func (m *HotReloadManager) Restart() (int, error) {
	if m.listener == nil {
		return -1, fmt.Errorf("no listener to restart")
	}

	lnFile, err := m.getListenerFile()
	if err != nil {
		return -1, err
	}
	defer lnFile.Close()

	parentFD, childFD, err := CreateSocketPair()
	if err != nil {
		return -1, fmt.Errorf("failed to create socket pair: %w", err)
	}

	exe, err := os.Executable()
	if err != nil {
		CloseFD(parentFD)
		CloseFD(childFD)
		return -1, fmt.Errorf("failed to get executable path: %w", err)
	}

	args := os.Args[1:]

	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(),
		EnvIsChild+"=1",
		EnvListenAddr+"="+m.listenAddr,
		EnvSocketPairFD+"="+strconv.Itoa(3),
	)

	cmd.ExtraFiles = []*os.File{
		os.NewFile(uintptr(childFD), "socketpair-child"),
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	oldPID := os.Getpid()

	if err := cmd.Start(); err != nil {
		CloseFD(parentFD)
		CloseFD(childFD)
		return -1, fmt.Errorf("failed to start child process: %w", err)
	}

	newPID := cmd.Process.Pid

	CloseFD(childFD)

	if err := SendFD(parentFD, int(lnFile.Fd())); err != nil {
		CloseFD(parentFD)
		m.addHistory(oldPID, newPID, false)
		return -1, fmt.Errorf("failed to send fd to child: %w", err)
	}

	ackCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if !waitForAck(ackCtx, parentFD) {
		CloseFD(parentFD)
		m.addHistory(oldPID, newPID, false)
		return -1, fmt.Errorf("timeout waiting for child ack")
	}

	CloseFD(parentFD)

	if err := m.listener.Close(); err != nil {
	}

	m.addHistory(oldPID, newPID, true)
	if m.onRestart != nil {
		m.onRestart(oldPID, newPID, true)
	}

	return newPID, nil
}

func (m *HotReloadManager) ChildHandshake() (net.Listener, error) {
	spFDStr := os.Getenv(EnvSocketPairFD)
	if spFDStr == "" {
		return nil, fmt.Errorf("no socketpair fd env")
	}

	spFD, err := strconv.Atoi(spFDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid socketpair fd: %w", err)
	}

	receivedFD, err := ReceiveFD(spFD)
	if err != nil {
		CloseFD(spFD)
		return nil, fmt.Errorf("failed to receive fd: %w", err)
	}

	if err := sendAck(spFD); err != nil {
		CloseFD(spFD)
		CloseFD(receivedFD)
		return nil, fmt.Errorf("failed to send ack: %w", err)
	}

	CloseFD(spFD)

	f := os.NewFile(uintptr(receivedFD), "listener")
	if f == nil {
		CloseFD(receivedFD)
		return nil, fmt.Errorf("failed to create file from received fd")
	}
	defer f.Close()

	ln, err := net.FileListener(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener from received fd: %w", err)
	}

	m.listener = ln
	return ln, nil
}

func (m *HotReloadManager) getListenerFile() (*os.File, error) {
	type fileListener interface {
		File() (*os.File, error)
	}

	fl, ok := m.listener.(fileListener)
	if !ok {
		return nil, fmt.Errorf("listener does not support File()")
	}
	return fl.File()
}

func (m *HotReloadManager) addHistory(oldPID, newPID int, success bool) {
	m.restartCount++
	m.restartHistory = append(m.restartHistory, RestartRecord{
		OldPID:     oldPID,
		NewPID:     newPID,
		Timestamp:  time.Now(),
		Successful: success,
	})
}

func waitForAck(ctx context.Context, fd int) bool {
	done := make(chan struct{})
	go func() {
		buf := make([]byte, 1)
		_, _ = syscall.Read(fd, buf)
		close(done)
	}()

	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}

func sendAck(fd int) error {
	buf := []byte{'A'}
	_, err := syscall.Write(fd, buf)
	return err
}

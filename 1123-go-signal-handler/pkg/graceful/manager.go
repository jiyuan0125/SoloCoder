package graceful

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/example/graceful-shutdown/pkg/api"
)

const (
	signalChanSize = 32
)

type cleanupCallback struct {
	name     string
	order    int
	callback func(ctx context.Context) error
	info     api.CallbackInfo
}

type Manager struct {
	mu sync.RWMutex

	state     api.ShutdownState
	callbacks map[string]*cleanupCallback
	ordered   []*cleanupCallback

	signalChan   chan os.Signal
	ctx          context.Context
	cancel       context.CancelFunc
	shutdownDone chan struct{}

	timeout time.Duration
	logger  func(string, ...interface{})

	startTime time.Time
}

type Option func(*Manager)

func WithTimeout(timeout time.Duration) Option {
	return func(m *Manager) {
		m.timeout = timeout
	}
}

func WithLogger(logger func(string, ...interface{})) Option {
	return func(m *Manager) {
		m.logger = logger
	}
}

func New(opts ...Option) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		state:        api.StateIdle,
		callbacks:    make(map[string]*cleanupCallback),
		ordered:      []*cleanupCallback{},
		signalChan:   make(chan os.Signal, signalChanSize),
		ctx:          ctx,
		cancel:       cancel,
		shutdownDone: make(chan struct{}),
		timeout:      30 * time.Second,
		logger:       func(string, ...interface{}) {},
	}

	for _, opt := range opts {
		opt(m)
	}

	signal.Notify(m.signalChan, syscall.SIGINT, syscall.SIGTERM)

	go m.watchSignals()

	return m
}

func (m *Manager) ResetSigpipeForChild(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Pdeathsig = syscall.SIGPIPE
	m.logger("Reset SIGPIPE for child process: %s", cmd.Path)
}

func (m *Manager) Register(name string, order int, callback func(ctx context.Context) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state != api.StateIdle {
		return fmt.Errorf("cannot register callback: shutdown already in progress")
	}

	if name == "" {
		return fmt.Errorf("callback name cannot be empty")
	}

	if _, exists := m.callbacks[name]; exists {
		return fmt.Errorf("callback '%s' already registered", name)
	}

	cb := &cleanupCallback{
		name:     name,
		order:    order,
		callback: callback,
		info: api.CallbackInfo{
			Name:   name,
			Order:  order,
			Status: api.StatusPending,
		},
	}

	m.callbacks[name] = cb
	m.ordered = append(m.ordered, cb)

	sort.Slice(m.ordered, func(i, j int) bool {
		if m.ordered[i].order == m.ordered[j].order {
			return m.ordered[i].name < m.ordered[j].name
		}
		return m.ordered[i].order < m.ordered[j].order
	})

	m.logger("Registered cleanup callback: %s (order: %d)", name, order)
	return nil
}

func (m *Manager) RegisterDynamic(name string, order int) error {
	return m.Register(name, order, func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			m.logger("Dynamic callback '%s' executed", name)
			return nil
		}
	})
}

func (m *Manager) watchSignals() {
	for {
		select {
		case sig, ok := <-m.signalChan:
			if !ok {
				return
			}

			m.mu.RLock()
			currentState := m.state
			m.mu.RUnlock()

			switch currentState {
			case api.StateIdle:
				m.logger("Received signal: %v, starting graceful shutdown", sig)
				go m.Shutdown()

			case api.StateGraceful:
				m.logger("Received second signal: %v, force exit", sig)
				m.forceExit()
				return
			}

		case <-m.ctx.Done():
			return
		}
	}
}

func (m *Manager) Shutdown() {
	m.mu.Lock()
	if m.state != api.StateIdle {
		m.mu.Unlock()
		return
	}

	m.state = api.StateGraceful
	m.startTime = time.Now()
	m.mu.Unlock()

	m.logger("=== Starting graceful shutdown ===")
	m.cancel()

	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	m.runCallbacks(ctx)

	close(m.shutdownDone)
}

func (m *Manager) runCallbacks(ctx context.Context) {
	m.mu.RLock()
	callbacks := make([]*cleanupCallback, len(m.ordered))
	copy(callbacks, m.ordered)
	m.mu.RUnlock()

	for _, cb := range callbacks {
		select {
		case <-ctx.Done():
			m.logger("Shutdown timeout reached, remaining callbacks skipped")
			m.markRemainingSkipped(callbacks)
			return
		default:
			m.runCallback(ctx, cb)
		}
	}

	m.mu.Lock()
	m.state = api.StateCompleted
	m.mu.Unlock()
	m.logger("=== Graceful shutdown completed ===")
}

func (m *Manager) runCallback(ctx context.Context, cb *cleanupCallback) {
	m.mu.Lock()
	cb.info.Status = api.StatusRunning
	cb.info.StartTime = time.Now().Format(time.RFC3339)
	startTime := time.Now()
	m.mu.Unlock()

	m.logger("Executing callback: %s", cb.name)

	errChan := make(chan error, 1)
	go func() {
		errChan <- cb.callback(ctx)
	}()

	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-errChan:
	}

	duration := time.Since(startTime)

	m.mu.Lock()
	cb.info.EndTime = time.Now().Format(time.RFC3339)
	cb.info.Duration = duration.String()

	if err != nil {
		cb.info.Status = api.StatusFailed
		cb.info.Error = err.Error()
		m.logger("Callback '%s' failed after %s: %v", cb.name, duration, err)
	} else {
		cb.info.Status = api.StatusCompleted
		m.logger("Callback '%s' completed in %s", cb.name, duration)
	}
	m.mu.Unlock()
}

func (m *Manager) markRemainingSkipped(callbacks []*cleanupCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cb := range callbacks {
		if cb.info.Status == api.StatusPending {
			cb.info.Status = api.StatusSkipped
			cb.info.EndTime = time.Now().Format(time.RFC3339)
			m.logger("Callback '%s' skipped due to timeout", cb.name)
		}
	}

	m.state = api.StateCompleted
}

func (m *Manager) forceExit() {
	m.mu.Lock()
	m.state = api.StateForceExit
	m.mu.Unlock()

	m.logger("=== Force exit triggered ===")
	os.Exit(1)
}

func (m *Manager) ForceExit() {
	m.forceExit()
}

func (m *Manager) GetState() api.ShutdownState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

func (m *Manager) GetCallbacksInfo() []api.CallbackInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]api.CallbackInfo, 0, len(m.ordered))
	for _, cb := range m.ordered {
		result = append(result, cb.info)
	}
	return result
}

func (m *Manager) Status() api.StatusResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()

	resp := api.StatusResponse{
		State:     m.state,
		Callbacks: make([]api.CallbackInfo, 0, len(m.ordered)),
	}

	for _, cb := range m.ordered {
		resp.Callbacks = append(resp.Callbacks, cb.info)
	}

	if m.state == api.StateGraceful || m.state == api.StateCompleted {
		resp.StartTime = m.startTime.Format(time.RFC3339)
	}

	switch m.state {
	case api.StateIdle:
		resp.Message = "服务运行中，等待信号"
	case api.StateGraceful:
		resp.Message = "优雅退出进行中"
	case api.StateCompleted:
		resp.Message = "优雅退出完成"
	case api.StateForceExit:
		resp.Message = "强制退出"
	}

	return resp
}

func (m *Manager) Done() <-chan struct{} {
	return m.shutdownDone
}

func (m *Manager) Context() context.Context {
	return m.ctx
}

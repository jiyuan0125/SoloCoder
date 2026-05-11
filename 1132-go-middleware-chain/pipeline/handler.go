package pipeline

import (
	"fmt"
	"time"
)

type Handler interface {
	ID() string
	Name() string
	Process(ctx *Context, next func())
}

type Middleware interface {
	Handler
	IsAsync() bool
	Timeout() time.Duration
}

type BaseHandler struct {
	id          string
	name        string
	panicHistory []string
}

func (h *BaseHandler) ID() string {
	return h.id
}

func (h *BaseHandler) Name() string {
	return h.name
}

func (h *BaseHandler) AddPanic(err string) {
	h.panicHistory = append(h.panicHistory, err)
}

func (h *BaseHandler) PanicHistory() []string {
	return h.panicHistory
}

type SimpleHandler struct {
	BaseHandler
	handlerFunc func(ctx *Context, next func())
}

func NewSimpleHandler(id, name string, f func(ctx *Context, next func())) *SimpleHandler {
	if id == "" {
		id = fmt.Sprintf("handler_%d", time.Now().UnixNano())
	}
	return &SimpleHandler{
		BaseHandler: BaseHandler{
			id:          id,
			name:        name,
			panicHistory: make([]string, 0),
		},
		handlerFunc: f,
	}
}

func (h *SimpleHandler) Process(ctx *Context, next func()) {
	if h.handlerFunc != nil {
		h.handlerFunc(ctx, next)
	} else {
		next()
	}
}

type SyncMiddleware struct {
	BaseHandler
	beforeFunc func(ctx *Context)
	afterFunc  func(ctx *Context)
}

func NewSyncMiddleware(id, name string, before, after func(ctx *Context)) *SyncMiddleware {
	if id == "" {
		id = fmt.Sprintf("handler_%d", time.Now().UnixNano())
	}
	return &SyncMiddleware{
		BaseHandler: BaseHandler{
			id:          id,
			name:        name,
			panicHistory: make([]string, 0),
		},
		beforeFunc: before,
		afterFunc:  after,
	}
}

func (m *SyncMiddleware) IsAsync() bool {
	return false
}

func (m *SyncMiddleware) Timeout() time.Duration {
	return 0
}

func (m *SyncMiddleware) Process(ctx *Context, next func()) {
	if m.beforeFunc != nil {
		m.beforeFunc(ctx)
	}
	next()
	if m.afterFunc != nil {
		m.afterFunc(ctx)
	}
}

type AsyncMiddleware struct {
	BaseHandler
	timeout    time.Duration
	workFunc   func(ctx *Context) error
}

func NewAsyncMiddleware(id, name string, timeoutMs int64, work func(ctx *Context) error) *AsyncMiddleware {
	if id == "" {
		id = fmt.Sprintf("handler_%d", time.Now().UnixNano())
	}
	return &AsyncMiddleware{
		BaseHandler: BaseHandler{
			id:          id,
			name:        name,
			panicHistory: make([]string, 0),
		},
		timeout:  time.Duration(timeoutMs) * time.Millisecond,
		workFunc: work,
	}
}

func (m *AsyncMiddleware) IsAsync() bool {
	return true
}

func (m *AsyncMiddleware) Timeout() time.Duration {
	return m.timeout
}

func (m *AsyncMiddleware) Process(ctx *Context, next func()) {
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Sprintf("async middleware panic: %v", r)
				m.AddPanic(err)
				ctx.SetMetadata("async_middleware_error_"+m.ID(), err)
				done <- fmt.Errorf(err)
			}
		}()
		if m.workFunc != nil {
			done <- m.workFunc(ctx)
		} else {
			done <- nil
		}
	}()

	select {
	case err := <-done:
		if err != nil {
			ctx.SetMetadata("async_middleware_error_"+m.ID(), err.Error())
		}
	case <-time.After(m.timeout):
		ctx.SetMetadata("async_middleware_timeout_"+m.ID(), "true")
		ctx.SetMetadata("async_middleware_timeout_reason_"+m.ID(), "execution exceeded timeout")
	}

	next()
}

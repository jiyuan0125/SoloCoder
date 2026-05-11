package pipeline

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type ExecutionTrace struct {
	RequestID    string
	Pipeline     string
	StartTime    int64
	EndTime      int64
	DurationMs   int64
	Status       string
	HandlerTraces []HandlerTrace
	Response     *Response
	Error        string
}

type HandlerTrace struct {
	HandlerID   string
	HandlerName string
	Status      string
	StartTime   int64
	EndTime     int64
	DurationMs  int64
	Output      string
	Error       string
	Panic       bool
}

type Manager struct {
	pipelines   map[string]*Pipeline
	handlers    map[string]Handler
	middlewares map[string]Middleware
	traces      map[string][]*ExecutionTrace
	mu          sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		pipelines:   make(map[string]*Pipeline),
		handlers:    make(map[string]Handler),
		middlewares: make(map[string]Middleware),
		traces:      make(map[string][]*ExecutionTrace),
	}
}

func (m *Manager) CreatePipeline(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.pipelines[name]; exists {
		return errors.New("pipeline already exists: " + name)
	}
	m.pipelines[name] = NewPipeline(name)
	return nil
}

func (m *Manager) GetPipeline(name string) (*Pipeline, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.pipelines[name]
	return p, ok
}

func (m *Manager) ListPipelines() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.pipelines))
	for name := range m.pipelines {
		names = append(names, name)
	}
	return names
}

func (m *Manager) DeletePipeline(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.pipelines[name]; !exists {
		return errors.New("pipeline not found: " + name)
	}
	delete(m.pipelines, name)
	return nil
}

func (m *Manager) RegisterHandler(h Handler) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.handlers[h.ID()]; exists {
		return errors.New("handler ID already exists: " + h.ID())
	}
	m.handlers[h.ID()] = h
	return nil
}

func (m *Manager) GetHandler(id string) (Handler, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h, ok := m.handlers[id]
	return h, ok
}

func (m *Manager) ListHandlers() []Handler {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]Handler, 0, len(m.handlers))
	for _, h := range m.handlers {
		list = append(list, h)
	}
	return list
}

func (m *Manager) DeleteHandler(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.pipelines {
		ids := p.GetHandlerIDs()
		for _, pid := range ids {
			if pid == id {
				return errors.New("handler is still used by pipeline")
			}
		}
	}
	if _, exists := m.handlers[id]; !exists {
		return errors.New("handler not found: " + id)
	}
	delete(m.handlers, id)
	return nil
}

func (m *Manager) RegisterMiddleware(mw Middleware) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.middlewares[mw.ID()]; exists {
		return errors.New("middleware ID already exists: " + mw.ID())
	}
	m.middlewares[mw.ID()] = mw
	m.handlers[mw.ID()] = mw
	return nil
}

func (m *Manager) GetMiddleware(id string) (Middleware, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mw, ok := m.middlewares[id]
	return mw, ok
}

func (m *Manager) ListMiddlewares() []Middleware {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]Middleware, 0, len(m.middlewares))
	for _, mw := range m.middlewares {
		list = append(list, mw)
	}
	return list
}

func (m *Manager) AddHandlerToPipeline(pipelineName, handlerID string, position int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.pipelines[pipelineName]
	if !exists {
		return errors.New("pipeline not found: " + pipelineName)
	}
	h, exists := m.handlers[handlerID]
	if !exists {
		return errors.New("handler not found: " + handlerID)
	}
	if position < 0 {
		p.AddHandler(h)
	} else {
		p.InsertHandlerAt(h, position)
	}
	return nil
}

func (m *Manager) RemoveHandlerFromPipeline(pipelineName, handlerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.pipelines[pipelineName]
	if !exists {
		return errors.New("pipeline not found: " + pipelineName)
	}
	if !p.RemoveHandler(handlerID) {
		return errors.New("handler not found in pipeline: " + handlerID)
	}
	return nil
}

func (m *Manager) ReorderPipelineHandlers(pipelineName string, handlerIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.pipelines[pipelineName]
	if !exists {
		return errors.New("pipeline not found: " + pipelineName)
	}
	if !p.ReorderHandlers(handlerIDs) {
		return errors.New("failed to reorder handlers")
	}
	return nil
}

func (m *Manager) ExecutePipeline(pipelineName string, ctx *Context) (*commonExecutionResult, error) {
	m.mu.RLock()
	p, exists := m.pipelines[pipelineName]
	m.mu.RUnlock()
	if !exists {
		return nil, errors.New("pipeline not found: " + pipelineName)
	}

	startTime := time.Now().UnixNano() / 1e6
	result := p.Execute(ctx)
	endTime := time.Now().UnixNano() / 1e6

	trace := &ExecutionTrace{
		RequestID:    ctx.RequestID,
		Pipeline:     pipelineName,
		StartTime:    startTime,
		EndTime:      endTime,
		DurationMs:   result.DurationMs,
		Status:       result.Status,
		HandlerTraces: make([]HandlerTrace, 0, len(result.Records)),
		Response:     ctx.Response,
		Error:        result.Error,
	}

	for _, r := range result.Records {
		trace.HandlerTraces = append(trace.HandlerTraces, HandlerTrace{
			HandlerID:   r.HandlerID,
			HandlerName: r.HandlerName,
			Status:      r.Status,
			StartTime:   r.StartTime,
			EndTime:     r.EndTime,
			DurationMs:  r.DurationMs,
			Output:      r.Output,
			Error:       r.Error,
			Panic:       r.Panic,
		})
	}

	m.mu.Lock()
	if m.traces == nil {
		m.traces = make(map[string][]*ExecutionTrace)
	}
	m.traces[ctx.RequestID] = append(m.traces[ctx.RequestID], trace)
	m.mu.Unlock()

	return &commonExecutionResult{
		RequestID:    ctx.RequestID,
		Pipeline:     pipelineName,
		Status:       result.Status,
		StartTime:    startTime,
		EndTime:      endTime,
		DurationMs:   result.DurationMs,
		HandlerResults: trace.HandlerTraces,
		Response:     ctx.Response,
		Error:        result.Error,
	}, nil
}

func (m *Manager) GetTraces(requestID string, startTime, endTime int64) []*ExecutionTrace {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if requestID != "" {
		if traces, ok := m.traces[requestID]; ok {
			return filterTracesByTime(traces, startTime, endTime)
		}
		return nil
	}
	allTraces := make([]*ExecutionTrace, 0)
	for _, traces := range m.traces {
		allTraces = append(allTraces, filterTracesByTime(traces, startTime, endTime)...)
	}
	return allTraces
}

func filterTracesByTime(traces []*ExecutionTrace, startTime, endTime int64) []*ExecutionTrace {
	if startTime == 0 && endTime == 0 {
		return traces
	}
	filtered := make([]*ExecutionTrace, 0)
	for _, t := range traces {
		if (startTime == 0 || t.StartTime >= startTime) && (endTime == 0 || t.EndTime <= endTime) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func (m *Manager) GetHandlerPanicHistory(id string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h, ok := m.handlers[id]
	if !ok {
		return nil
	}
	if bh, ok := h.(interface{ PanicHistory() []string }); ok {
		return bh.PanicHistory()
	}
	return nil
}

func (m *Manager) GenerateHandlerID() string {
	return fmt.Sprintf("handler_%d", time.Now().UnixNano())
}

type commonExecutionResult struct {
	RequestID      string
	Pipeline       string
	Status         string
	StartTime      int64
	EndTime        int64
	DurationMs     int64
	HandlerResults []HandlerTrace
	Response       *Response
	Error          string
}

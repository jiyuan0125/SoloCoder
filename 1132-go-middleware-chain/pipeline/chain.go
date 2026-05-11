package pipeline

import (
	"fmt"
	"time"
)

type HandlerExecutionRecord struct {
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

type ChainExecutionResult struct {
	Status       string
	StartTime    int64
	EndTime      int64
	DurationMs   int64
	Records      []HandlerExecutionRecord
	Error        string
	ConsecutivePanics int
}

type Chain struct {
	handlers []Handler
}

func NewChain() *Chain {
	return &Chain{
		handlers: make([]Handler, 0),
	}
}

func (c *Chain) AddHandler(h Handler) {
	c.handlers = append(c.handlers, h)
}

func (c *Chain) InsertHandlerAt(h Handler, position int) {
	if position <= 0 {
		c.handlers = append([]Handler{h}, c.handlers...)
		return
	}
	if position >= len(c.handlers) {
		c.handlers = append(c.handlers, h)
		return
	}
	c.handlers = append(c.handlers[:position], append([]Handler{h}, c.handlers[position:]...)...)
}

func (c *Chain) RemoveHandler(id string) bool {
	for i, h := range c.handlers {
		if h.ID() == id {
			c.handlers = append(c.handlers[:i], c.handlers[i+1:]...)
			return true
		}
	}
	return false
}

func (c *Chain) ReorderHandlers(ids []string) bool {
	newHandlers := make([]Handler, 0, len(ids))
	existing := make(map[string]Handler)
	for _, h := range c.handlers {
		existing[h.ID()] = h
	}
	for _, id := range ids {
		if h, ok := existing[id]; ok {
			newHandlers = append(newHandlers, h)
		} else {
			return false
		}
	}
	if len(newHandlers) != len(c.handlers) {
		return false
	}
	c.handlers = newHandlers
	return true
}

func (c *Chain) GetHandlerIDs() []string {
	ids := make([]string, 0, len(c.handlers))
	for _, h := range c.handlers {
		ids = append(ids, h.ID())
	}
	return ids
}

func (c *Chain) Execute(ctx *Context) *ChainExecutionResult {
	result := &ChainExecutionResult{
		Status:       "success",
		StartTime:    time.Now().UnixNano() / 1e6,
		Records:      make([]HandlerExecutionRecord, 0),
		ConsecutivePanics: 0,
	}

	index := 0
	consecutivePanics := 0

	var next func()
	next = func() {
		if index >= len(c.handlers) {
			return
		}
		if ctx.Terminated {
			return
		}
		if consecutivePanics >= 3 {
			result.Status = "failed"
			result.Error = "consecutive panics limit reached (3)"
			return
		}

		h := c.handlers[index]
		index++

		record := HandlerExecutionRecord{
			HandlerID:   h.ID(),
			HandlerName: h.Name(),
			Status:      "success",
			StartTime:   time.Now().UnixNano() / 1e6,
		}

		handlerDone := make(chan struct{})
		panicOccurred := false
		var panicErr interface{}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					panicOccurred = true
					panicErr = r
				}
				close(handlerDone)
			}()
			h.Process(ctx, next)
		}()

		<-handlerDone

		record.EndTime = time.Now().UnixNano() / 1e6
		record.DurationMs = record.EndTime - record.StartTime

		if panicOccurred {
			consecutivePanics++
			record.Status = "failed"
			record.Panic = true
			errStr := fmt.Sprintf("panic: %v", panicErr)
			record.Error = errStr
			if bh, ok := h.(interface{ AddPanic(string) }); ok {
				bh.AddPanic(errStr)
			}
			ctx.PanicCount++
			if consecutivePanics < 3 {
				next()
			}
		} else {
			consecutivePanics = 0
		}

		result.Records = append(result.Records, record)
	}

	next()

	result.EndTime = time.Now().UnixNano() / 1e6
	result.DurationMs = result.EndTime - result.StartTime
	result.ConsecutivePanics = consecutivePanics

	if ctx.Terminated {
		result.Status = "terminated"
		if ctx.TerminateReason != "" {
			result.Error = ctx.TerminateReason
		}
	}

	return result
}

package pipeline

import "sync"

type Context struct {
	RequestID  string
	Metadata   map[string]string
	Header     map[string]string
	Body       string
	Response   *Response
	Data       map[string]interface{}
	Terminated bool
	TerminateReason string
	PanicCount int
	mu         sync.RWMutex
}

type Response struct {
	StatusCode int
	Header     map[string]string
	Body       string
}

func NewContext() *Context {
	return &Context{
		Metadata: make(map[string]string),
		Header:   make(map[string]string),
		Body:     "",
		Response: &Response{
			StatusCode: 200,
			Header:     make(map[string]string),
			Body:       "",
		},
		Data:       make(map[string]interface{}),
		Terminated: false,
		PanicCount: 0,
	}
}

func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data[key] = value
}

func (c *Context) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.Data[key]
	return val, ok
}

func (c *Context) GetString(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.Data[key]
	if !ok {
		return ""
	}
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}

func (c *Context) SetMetadata(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Metadata[key] = value
}

func (c *Context) GetMetadata(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.Metadata[key]
	return val, ok
}

func (c *Context) Terminate(reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Terminated = true
	c.TerminateReason = reason
}

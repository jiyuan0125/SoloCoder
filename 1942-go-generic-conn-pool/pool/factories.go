package pool

import (
	"errors"
	"net/http"
	"sync"
)

type FactoryRegistry struct {
	mu       sync.RWMutex
	factories map[string]Factory
}

func NewFactoryRegistry() *FactoryRegistry {
	registry := &FactoryRegistry{
		factories: make(map[string]Factory),
	}
	registry.RegisterDefaults()
	return registry
}

func (r *FactoryRegistry) Register(name string, factory Factory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

func (r *FactoryRegistry) Get(name string) (Factory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	factory, exists := r.factories[name]
	return factory, exists
}

func (r *FactoryRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

func (r *FactoryRegistry) RegisterDefaults() {
	r.Register("http_client", NewHTTPClient)
	r.Register("noop", NewNoOpConn)
}

type HTTPClientConn struct {
	Client *http.Client
	closed bool
}

func NewHTTPClient() (Connection, error) {
	return &HTTPClientConn{
		Client: &http.Client{
			Timeout: 30 * 1000000000,
		},
	}, nil
}

func (c *HTTPClientConn) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	if c.Client != nil {
		c.Client.CloseIdleConnections()
	}
	return nil
}

type NoOpConn struct {
	closed bool
}

func NewNoOpConn() (Connection, error) {
	return &NoOpConn{}, nil
}

func (c *NoOpConn) Close() error {
	if c.closed {
		return errors.New("already closed")
	}
	c.closed = true
	return nil
}

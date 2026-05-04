package circuitbreaker

import (
	"fmt"
	"sync"
)

type Registry struct {
	mu       sync.RWMutex
	circuits map[string]CircuitBreaker
}

func NewRegistry() *Registry {
	return &Registry{
		circuits: make(map[string]CircuitBreaker),
	}
}

func (r *Registry) Register(name string, config ...Config) (CircuitBreaker, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.circuits[name]; exists {
		return nil, fmt.Errorf("circuit breaker %q already registered", name)
	}

	cb, err := New(name, config...)
	if err != nil {
		return nil, err
	}

	r.circuits[name] = cb
	return cb, nil
}

func (r *Registry) RegisterFromFile(name, path string, config ...Config) (CircuitBreaker, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.circuits[name]; exists {
		return nil, fmt.Errorf("circuit breaker %q already registered", name)
	}

	cb, err := NewFromFile(name, path, config...)
	if err != nil {
		return nil, err
	}

	r.circuits[name] = cb
	return cb, nil
}

func (r *Registry) Get(name string) (CircuitBreaker, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cb, exists := r.circuits[name]
	return cb, exists
}

func (r *Registry) GetOrCreate(name string, config ...Config) (CircuitBreaker, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cb, exists := r.circuits[name]; exists {
		return cb, nil
	}

	cb, err := New(name, config...)
	if err != nil {
		return nil, err
	}

	r.circuits[name] = cb
	return cb, nil
}

func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.circuits, name)
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.circuits))
	for name := range r.circuits {
		names = append(names, name)
	}
	return names
}

func (r *Registry) ResetAll() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, cb := range r.circuits {
		cb.Reset()
	}
}

func (r *Registry) SaveAll(dir string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, cb := range r.circuits {
		path := fmt.Sprintf("%s/%s.json", dir, name)
		if err := cb.SaveToFile(path); err != nil {
			return fmt.Errorf("failed to save circuit %q: %w", name, err)
		}
	}

	return nil
}

func (r *Registry) Execute(name string, fn func() error) error {
	r.mu.RLock()
	cb, exists := r.circuits[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("circuit breaker %q not found", name)
	}

	return cb.Execute(fn)
}

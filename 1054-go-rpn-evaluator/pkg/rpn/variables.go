package rpn

import (
	"sync"
)

type VariableStore struct {
	mu    sync.RWMutex
	vars  map[string]float64
}

func NewVariableStore() *VariableStore {
	return &VariableStore{
		vars: make(map[string]float64),
	}
}

func (vs *VariableStore) Set(name string, value float64) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	vs.vars[name] = value
}

func (vs *VariableStore) Get(name string) (float64, bool) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	v, ok := vs.vars[name]
	return v, ok
}

func (vs *VariableStore) GetAll() map[string]float64 {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	result := make(map[string]float64, len(vs.vars))
	for k, v := range vs.vars {
		result[k] = v
	}
	return result
}

func (vs *VariableStore) Delete(name string) bool {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	if _, ok := vs.vars[name]; ok {
		delete(vs.vars, name)
		return true
	}
	return false
}

func (vs *VariableStore) Clear() {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	vs.vars = make(map[string]float64)
}

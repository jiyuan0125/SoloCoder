package registry

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type Resolver func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error)

type FieldDef struct {
	Name    string
	Type    string
	Args    map[string]string
	Resolve Resolver
}

type TypeDef struct {
	Name   string
	Fields map[string]*FieldDef
}

type Registry struct {
	mu          sync.RWMutex
	types       map[string]*TypeDef
	queryType   string
	mutationType string
	callCounts  map[string]int
	nplus1Log   []string
}

func New() *Registry {
	return &Registry{
		types:      make(map[string]*TypeDef),
		queryType:  "Query",
		callCounts: make(map[string]int),
		nplus1Log:  []string{},
	}
}

func (r *Registry) SetQueryType(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.queryType = name
}

func (r *Registry) GetQueryType() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.queryType
}

func (r *Registry) SetMutationType(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mutationType = name
}

func (r *Registry) GetMutationType() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mutationType
}

func (r *Registry) RegisterType(name string) *TypeDef {
	r.mu.Lock()
	defer r.mu.Unlock()
	if td, ok := r.types[name]; ok {
		return td
	}
	td := &TypeDef{
		Name:   name,
		Fields: make(map[string]*FieldDef),
	}
	r.types[name] = td
	return td
}

func (r *Registry) GetType(name string) (*TypeDef, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	td, ok := r.types[name]
	return td, ok
}

func (r *Registry) Types() map[string]*TypeDef {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]*TypeDef)
	for k, v := range r.types {
		result[k] = v
	}
	return result
}

func (t *TypeDef) RegisterField(name string, fieldType string, args map[string]string, resolve Resolver) *FieldDef {
	if args == nil {
		args = make(map[string]string)
	}
	fd := &FieldDef{
		Name:    name,
		Type:    fieldType,
		Args:    args,
		Resolve: resolve,
	}
	t.Fields[name] = fd
	return fd
}

func (r *Registry) GetResolver(typeName, fieldName string) (Resolver, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	td, ok := r.types[typeName]
	if !ok {
		return nil, false
	}
	fd, ok := td.Fields[fieldName]
	if !ok {
		return nil, false
	}
	return fd.Resolve, true
}

func (r *Registry) GetFieldType(typeName, fieldName string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	td, ok := r.types[typeName]
	if !ok {
		return "", false
	}
	fd, ok := td.Fields[fieldName]
	if !ok {
		return "", false
	}
	return fd.Type, true
}

func (r *Registry) GetFieldArgs(typeName, fieldName string) (map[string]string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	td, ok := r.types[typeName]
	if !ok {
		return nil, false
	}
	fd, ok := td.Fields[fieldName]
	if !ok {
		return nil, false
	}
	return fd.Args, true
}

func (r *Registry) RecordResolverCall(typeName, fieldName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := typeName + "." + fieldName
	r.callCounts[key]++
	if r.callCounts[key] > 1 {
		r.nplus1Log = append(r.nplus1Log,
			fmt.Sprintf("N+1 warning: resolver %s.%s called %d times", typeName, fieldName, r.callCounts[key]))
	}
}

func (r *Registry) ResetCallCounts() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.callCounts = make(map[string]int)
	r.nplus1Log = []string{}
}

func (r *Registry) NPlus1Warnings() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]string, len(r.nplus1Log))
	copy(result, r.nplus1Log)
	return result
}

func (r *Registry) CheckCyclicReferences() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for typeName := range r.types {
		visited := make(map[string]bool)
		if r.hasCycle(typeName, visited, []string{}) {
			return fmt.Errorf("cyclic type reference detected for type: %s", typeName)
		}
	}
	return nil
}

func (r *Registry) hasCycle(typeName string, visited map[string]bool, path []string) bool {
	if visited[typeName] {
		for i := range path {
			if path[i] == typeName {
				return true
			}
		}
		return false
	}
	visited[typeName] = true
	path = append(path, typeName)
	td, ok := r.types[typeName]
	if !ok {
		return false
	}
	for _, fd := range td.Fields {
		fieldType := cleanType(fd.Type)
		if _, ok := r.types[fieldType]; ok {
			if r.hasCycle(fieldType, visited, path) {
				return true
			}
		}
	}
	return false
}

func cleanType(t string) string {
	t = strings.TrimSuffix(t, "!")
	t = strings.TrimPrefix(t, "[")
	t = strings.TrimSuffix(t, "!")
	t = strings.TrimSuffix(t, "]")
	t = strings.TrimSuffix(t, "!")
	return t
}

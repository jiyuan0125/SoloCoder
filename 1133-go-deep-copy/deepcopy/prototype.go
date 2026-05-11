package deepcopy

import (
	"sync"
)

type Prototype interface {
	Clone() (interface{}, error)
}

type prototypeRegistry struct {
	sync.RWMutex
	prototypes map[string]interface{}
}

var registry = &prototypeRegistry{
	prototypes: make(map[string]interface{}),
}

func Register(name string, prototype interface{}) {
	registry.Lock()
	defer registry.Unlock()
	registry.prototypes[name] = prototype
}

func Unregister(name string) {
	registry.Lock()
	defer registry.Unlock()
	delete(registry.prototypes, name)
}

func GetPrototype(name string) (interface{}, bool) {
	registry.RLock()
	defer registry.RUnlock()
	proto, exists := registry.prototypes[name]
	return proto, exists
}

func ListPrototypes() []string {
	registry.RLock()
	defer registry.RUnlock()
	names := make([]string, 0, len(registry.prototypes))
	for name := range registry.prototypes {
		names = append(names, name)
	}
	return names
}

func Clone(name string) (interface{}, error) {
	proto, exists := GetPrototype(name)
	if !exists {
		return nil, ErrPrototypeNotFound
	}

	if p, ok := proto.(Prototype); ok {
		return p.Clone()
	}

	return DeepCopy(proto)
}

type DeepCopyPrototype struct {
	Value interface{}
}

func (d *DeepCopyPrototype) Clone() (interface{}, error) {
	return DeepCopy(d.Value)
}

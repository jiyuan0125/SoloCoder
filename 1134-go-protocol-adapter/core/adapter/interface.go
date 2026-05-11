package adapter

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
)

type Adapter interface {
	Name() string
	CanHandle(format string) bool
	Decode(data []byte, v interface{}) error
	Encode(v interface{}) ([]byte, error)
	Priority() int
}

type AdapterRegistry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
	byFormat map[string][]Adapter
}

func NewRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		adapters: make(map[string]Adapter),
		byFormat: make(map[string][]Adapter),
	}
}

func (r *AdapterRegistry) Register(adapter Adapter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := adapter.Name()
	if _, exists := r.adapters[name]; exists {
		return fmt.Errorf("adapter %s already exists", name)
	}

	r.adapters[name] = adapter

	for format := range getAllFormats(adapter) {
		r.byFormat[format] = append(r.byFormat[format], adapter)
		sort.Slice(r.byFormat[format], func(i, j int) bool {
			return r.byFormat[format][i].Priority() > r.byFormat[format][j].Priority()
		})
	}

	return nil
}

func (r *AdapterRegistry) List() []AdapterInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]AdapterInfo, 0, len(r.adapters))
	for name, adapter := range r.adapters {
		infos = append(infos, AdapterInfo{
			Name:     name,
			Priority: adapter.Priority(),
		})
	}
	return infos
}

func (r *AdapterRegistry) Get(name string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, ok := r.adapters[name]
	return adapter, ok
}

func (r *AdapterRegistry) Find(format string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapters, ok := r.byFormat[format]
	if !ok || len(adapters) == 0 {
		return nil, false
	}
	return adapters[0], true
}

func (r *AdapterRegistry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.adapters[name]
	if !exists {
		return false
	}

	delete(r.adapters, name)

	for format, adapters := range r.byFormat {
		for i, a := range adapters {
			if a.Name() == name {
				r.byFormat[format] = append(adapters[:i], adapters[i+1:]...)
				break
			}
		}
	}

	return true
}

type AdapterInfo struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

func getAllFormats(adapter Adapter) map[string]struct{} {
	formats := make(map[string]struct{})
	for _, format := range []string{"json", "xml", "csv", "yaml", "binary", "custom"} {
		if adapter.CanHandle(format) {
			formats[format] = struct{}{}
		}
	}
	return formats
}

type TypeConverter struct{}

func NewTypeConverter() *TypeConverter {
	return &TypeConverter{}
}

func (c *TypeConverter) Convert(value string, targetType reflect.Type) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.String:
		return reflect.ValueOf(value), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var result int64
		_, err := fmt.Sscanf(value, "%d", &result)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to %s: %w", value, targetType.Kind(), err)
		}
		return reflect.ValueOf(result).Convert(targetType), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var result uint64
		_, err := fmt.Sscanf(value, "%d", &result)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to %s: %w", value, targetType.Kind(), err)
		}
		return reflect.ValueOf(result).Convert(targetType), nil
	case reflect.Float32, reflect.Float64:
		var result float64
		_, err := fmt.Sscanf(value, "%f", &result)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to %s: %w", value, targetType.Kind(), err)
		}
		return reflect.ValueOf(result).Convert(targetType), nil
	case reflect.Bool:
		switch value {
		case "true", "1", "yes":
			return reflect.ValueOf(true), nil
		case "false", "0", "no":
			return reflect.ValueOf(false), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert '%s' to bool", value)
		}
	case reflect.Struct:
		if targetType == reflect.TypeOf(reflect.Value{}) {
			return reflect.Value{}, fmt.Errorf("cannot convert to reflect.Value")
		}
		return reflect.Zero(targetType), nil
	default:
		return reflect.Zero(targetType), fmt.Errorf("unsupported type: %s", targetType.Kind())
	}
}

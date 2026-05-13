package metrics

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type MetricType string

const (
	TypeCounter   MetricType = "counter"
	TypeGauge     MetricType = "gauge"
	TypeHistogram MetricType = "histogram"
)

var (
	ErrMetricNotFound     = errors.New("metric not found")
	ErrMetricTypeMismatch = errors.New("metric type mismatch")
	ErrLabelSeriesLimit   = errors.New("label series limit exceeded")
	ErrInvalidValue       = errors.New("invalid metric value")
)

const DefaultMaxLabelSeries = 1000

type Labels map[string]string

type MetricDescriptor struct {
	Name        string
	Type        MetricType
	LabelSeries map[string]struct{}
	MaxSeries   int
}

func NewMetricDescriptor(name string, typ MetricType) *MetricDescriptor {
	return &MetricDescriptor{
		Name:        name,
		Type:        typ,
		LabelSeries: make(map[string]struct{}),
		MaxSeries:   DefaultMaxLabelSeries,
	}
}

func LabelsKey(labels Labels) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(k)
		builder.WriteByte('=')
		builder.WriteString(labels[k])
	}
	return builder.String()
}

type Registry struct {
	metrics map[string]*MetricDescriptor
	mu      sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]*MetricDescriptor),
	}
}

func (r *Registry) Register(name string, typ MetricType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.metrics[name]; ok {
		if existing.Type != typ {
			return ErrMetricTypeMismatch
		}
		return nil
	}

	r.metrics[name] = NewMetricDescriptor(name, typ)
	return nil
}

func (r *Registry) GetOrRegister(name string, defaultType MetricType) *MetricDescriptor {
	r.mu.Lock()
	defer r.mu.Unlock()

	if metric, ok := r.metrics[name]; ok {
		return metric
	}

	metric := NewMetricDescriptor(name, defaultType)
	r.metrics[name] = metric
	return metric
}

func (r *Registry) Get(name string) (*MetricDescriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metric, ok := r.metrics[name]
	return metric, ok
}

func (r *Registry) List() []*MetricDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*MetricDescriptor, 0, len(r.metrics))
	for _, m := range r.metrics {
		list = append(list, m)
	}
	return list
}

func (r *Registry) AddLabelSeries(name string, labelKey string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	metric, ok := r.metrics[name]
	if !ok {
		return fmt.Errorf("%w: %s", ErrMetricNotFound, name)
	}

	if _, exists := metric.LabelSeries[labelKey]; exists {
		return nil
	}

	if len(metric.LabelSeries) >= metric.MaxSeries {
		return ErrLabelSeriesLimit
	}

	metric.LabelSeries[labelKey] = struct{}{}
	return nil
}

func (r *Registry) MetricCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.metrics)
}

func (r *Registry) SeriesCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	total := 0
	for _, m := range r.metrics {
		total += len(m.LabelSeries)
	}
	return total
}

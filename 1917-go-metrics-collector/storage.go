package main

import (
	"regexp"
	"sort"
	"sync"
)

type MetricType string

const (
	Counter MetricType = "counter"
	Gauge   MetricType = "gauge"
	Timer   MetricType = "timer"
)

var validMetricName = regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)

func IsValidMetricName(name string) bool {
	return validMetricName.MatchString(name)
}

type Tags map[string]string

type TimerValue struct {
	Values []float64
}

func NewTimerValue() *TimerValue {
	return &TimerValue{
		Values: make([]float64, 0),
	}
}

func (t *TimerValue) Add(v float64) {
	t.Values = append(t.Values, v)
}

func (t *TimerValue) Merge(other *TimerValue) {
	t.Values = append(t.Values, other.Values...)
}

func (t *TimerValue) Quantile(q float64) float64 {
	if len(t.Values) == 0 {
		return 0
	}
	sorted := make([]float64, len(t.Values))
	copy(sorted, t.Values)
	sort.Float64s(sorted)

	idx := int(q * float64(len(sorted)-1))
	if idx < 0 {
		idx = 0
	} else if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func (t *TimerValue) P50() float64 {
	return t.Quantile(0.5)
}

func (t *TimerValue) P90() float64 {
	return t.Quantile(0.9)
}

func (t *TimerValue) P99() float64 {
	return t.Quantile(0.99)
}

func (t *TimerValue) Max() float64 {
	if len(t.Values) == 0 {
		return 0
	}
	sorted := make([]float64, len(t.Values))
	copy(sorted, t.Values)
	sort.Float64s(sorted)
	return sorted[len(sorted)-1]
}

func (t *TimerValue) Min() float64 {
	if len(t.Values) == 0 {
		return 0
	}
	sorted := make([]float64, len(t.Values))
	copy(sorted, t.Values)
	sort.Float64s(sorted)
	return sorted[0]
}

type Metric struct {
	Name     string
	Type     MetricType
	Counter  map[string]float64
	Gauge    map[string]float64
	Timer    map[string]*TimerValue
	TagKeys  []string
	Lock     sync.RWMutex
}

func NewMetric(name string, metricType MetricType) *Metric {
	m := &Metric{
		Name:    name,
		Type:    metricType,
		Counter: make(map[string]float64),
		Gauge:   make(map[string]float64),
		Timer:   make(map[string]*TimerValue),
		TagKeys: make([]string, 0),
	}
	return m
}

func (m *Metric) tagsToKey(tags Tags) string {
	if len(tags) == 0 {
		return ""
	}
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := ""
	for i, k := range keys {
		if i > 0 {
			result += ","
		}
		result += k + "=" + tags[k]
	}
	return result
}

func (m *Metric) Record(tags Tags, value float64, newType MetricType) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	if newType != m.Type {
		m.Type = newType
	}

	key := m.tagsToKey(tags)

	switch m.Type {
	case Counter:
		m.Counter[key] += value
	case Gauge:
		m.Gauge[key] = value
	case Timer:
		tv, ok := m.Timer[key]
		if !ok {
			tv = NewTimerValue()
			m.Timer[key] = tv
		}
		tv.Add(value)
	}
}

type MetricStore struct {
	metrics map[string]*Metric
	lock    sync.RWMutex
}

func NewMetricStore() *MetricStore {
	return &MetricStore{
		metrics: make(map[string]*Metric),
	}
}

func (s *MetricStore) GetOrCreate(name string, metricType MetricType) *Metric {
	s.lock.Lock()
	defer s.lock.Unlock()

	m, ok := s.metrics[name]
	if !ok {
		m = NewMetric(name, metricType)
		s.metrics[name] = m
	}
	return m
}

func (s *MetricStore) Get(name string) *Metric {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.metrics[name]
}

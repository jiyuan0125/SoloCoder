package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type MetricType string

const (
	TypeCounter   MetricType = "counter"
	TypeGauge     MetricType = "gauge"
	TypeHistogram MetricType = "histogram"
)

type Labels map[string]string

func (l Labels) Key() string {
	if len(l) == 0 {
		return ""
	}
	keys := make([]string, 0, len(l))
	for k := range l {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(l))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, l[k]))
	}
	return strings.Join(parts, ",")
}

func (l Labels) Match(filter Labels) bool {
	for k, v := range filter {
		if lv, ok := l[k]; !ok || lv != v {
			return false
		}
	}
	return true
}

type Bucket struct {
	Start    time.Time
	Count    int64
	Sum      float64
	Min      float64
	Max      float64
	Values   []float64
}

func NewBucket(start time.Time) *Bucket {
	return &Bucket{
		Start:  start,
		Min:    0,
		Max:    0,
		Values: make([]float64, 0),
	}
}

func (b *Bucket) Add(v float64) {
	if b.Count == 0 {
		b.Min = v
		b.Max = v
	} else {
		if v < b.Min {
			b.Min = v
		}
		if v > b.Max {
			b.Max = v
		}
	}
	b.Count++
	b.Sum += v
	b.Values = append(b.Values, v)
}

func (b *Bucket) Avg() float64 {
	if b.Count == 0 {
		return 0
	}
	return b.Sum / float64(b.Count)
}

type Series struct {
	mu      sync.RWMutex
	buckets []*Bucket
	maxAge  time.Duration
}

func NewSeries(maxAge time.Duration) *Series {
	return &Series{
		buckets: make([]*Bucket, 0),
		maxAge:  maxAge,
	}
}

func (s *Series) record(t time.Time, v float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	truncated := t.Truncate(time.Minute)
	var target *Bucket

	if len(s.buckets) > 0 {
		last := s.buckets[len(s.buckets)-1]
		if last.Start.Equal(truncated) {
			target = last
		}
	}

	if target == nil {
		target = NewBucket(truncated)
		s.buckets = append(s.buckets, target)
	}

	target.Add(v)
	s.cleanupLocked(t)
}

func (s *Series) set(t time.Time, v float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	truncated := t.Truncate(time.Minute)
	var target *Bucket

	if len(s.buckets) > 0 {
		last := s.buckets[len(s.buckets)-1]
		if last.Start.Equal(truncated) {
			target = last
		}
	}

	if target == nil {
		target = NewBucket(truncated)
		s.buckets = append(s.buckets, target)
	}

	target.Count = 1
	target.Sum = v
	target.Min = v
	target.Max = v
	target.Values = []float64{v}
	s.cleanupLocked(t)
}

func (s *Series) cleanupLocked(now time.Time) {
	cutoff := now.Add(-s.maxAge)
	startIdx := 0
	for i, b := range s.buckets {
		if b.Start.After(cutoff) {
			startIdx = i
			break
		}
		if i == len(s.buckets)-1 {
			startIdx = len(s.buckets)
		}
	}
	if startIdx > 0 {
		s.buckets = s.buckets[startIdx:]
	}
}

func (s *Series) Query(start, end time.Time) []*Bucket {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Bucket, 0)
	for _, b := range s.buckets {
		if !b.Start.Before(start) && !b.Start.After(end) {
			result = append(result, b)
		}
	}
	return result
}

type Metric struct {
	Name     string
	Type     MetricType
	Help     string
	Labels   Labels
	series   map[string]*Series
	mu       sync.RWMutex
	maxAge   time.Duration
}

func NewMetric(name string, typ MetricType, help string, maxAge time.Duration) *Metric {
	return &Metric{
		Name:   name,
		Type:   typ,
		Help:   help,
		Labels: make(Labels),
		series: make(map[string]*Series),
		maxAge: maxAge,
	}
}

func (m *Metric) getOrCreateSeries(labels Labels) *Series {
	key := labels.Key()

	m.mu.RLock()
	s, ok := m.series[key]
	m.mu.RUnlock()

	if ok {
		return s
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok = m.series[key]; ok {
		return s
	}

	s = NewSeries(m.maxAge)
	m.series[key] = s
	return s
}

func (m *Metric) Inc(labels Labels) {
	m.getOrCreateSeries(labels).record(time.Now(), 1)
}

func (m *Metric) Add(v float64, labels Labels) {
	m.getOrCreateSeries(labels).record(time.Now(), v)
}

func (m *Metric) Set(v float64, labels Labels) {
	m.getOrCreateSeries(labels).set(time.Now(), v)
}

func (m *Metric) Observe(v float64, labels Labels) {
	m.getOrCreateSeries(labels).record(time.Now(), v)
}

func (m *Metric) Query(filter Labels, start, end time.Time) map[string][]*Bucket {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]*Bucket)
	for key, series := range m.series {
		if len(filter) > 0 {
			labels := parseLabelsKey(key)
			if !labels.Match(filter) {
				continue
			}
		}
		result[key] = series.Query(start, end)
	}
	return result
}

func (m *Metric) ListSeries() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(m.series))
	for key := range m.series {
		result = append(result, map[string]interface{}{
			"labels": parseLabelsKey(key),
		})
	}
	return result
}

func parseLabelsKey(key string) Labels {
	labels := make(Labels)
	if key == "" {
		return labels
	}
	parts := strings.Split(key, ",")
	for _, p := range parts {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 {
			labels[kv[0]] = kv[1]
		}
	}
	return labels
}

type Registry struct {
	metrics map[string]*Metric
	mu      sync.RWMutex
	maxAge  time.Duration
}

func NewRegistry(maxAge time.Duration) *Registry {
	return &Registry{
		metrics: make(map[string]*Metric),
		maxAge:  maxAge,
	}
}

func (r *Registry) Register(name string, typ MetricType, help string) *Metric {
	r.mu.Lock()
	defer r.mu.Unlock()

	if m, ok := r.metrics[name]; ok {
		return m
	}

	m := NewMetric(name, typ, help, r.maxAge)
	r.metrics[name] = m
	return m
}

func (r *Registry) Get(name string) (*Metric, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, ok := r.metrics[name]
	return m, ok
}

func (r *Registry) List() []*Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Metric, 0, len(r.metrics))
	for _, m := range r.metrics {
		result = append(result, m)
	}
	return result
}

var DefaultRegistry = NewRegistry(7 * 24 * time.Hour)

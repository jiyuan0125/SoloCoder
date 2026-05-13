package metrics

import (
	"math"
	"sort"
	"sync"
)

var DefaultBuckets = []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, math.Inf(1)}

type CounterData struct {
	Value float64
}

type GaugeData struct {
	Value float64
}

type HistogramData struct {
	Count   int
	Sum     float64
	Buckets map[float64]int
	Values  []float64
}

func NewHistogramData(buckets []float64) *HistogramData {
	bucketMap := make(map[float64]int)
	for _, b := range buckets {
		bucketMap[b] = 0
	}
	return &HistogramData{
		Buckets: bucketMap,
		Values:  make([]float64, 0),
	}
}

func (h *HistogramData) Observe(value float64) {
	h.Count++
	h.Sum += value
	for bucket := range h.Buckets {
		if value <= bucket {
			h.Buckets[bucket]++
		}
	}
	h.Values = append(h.Values, value)
}

func (h *HistogramData) Percentile(p float64) float64 {
	if len(h.Values) == 0 {
		return 0
	}
	sorted := make([]float64, len(h.Values))
	copy(sorted, h.Values)
	sort.Float64s(sorted)
	index := int(float64(len(sorted)-1) * p)
	return sorted[index]
}

type SeriesKey struct {
	MetricName string
	LabelKey   string
}

type Store struct {
	counters   map[SeriesKey]*CounterData
	gauges     map[SeriesKey]*GaugeData
	histograms map[SeriesKey]*HistogramData
	mu         sync.RWMutex
	buckets    map[string][]float64
}

func NewStore() *Store {
	return &Store{
		counters:   make(map[SeriesKey]*CounterData),
		gauges:     make(map[SeriesKey]*GaugeData),
		histograms: make(map[SeriesKey]*HistogramData),
		buckets:    make(map[string][]float64),
	}
}

func (s *Store) SetBuckets(metricName string, buckets []float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if buckets == nil {
		delete(s.buckets, metricName)
		return
	}
	s.buckets[metricName] = make([]float64, len(buckets))
	copy(s.buckets[metricName], buckets)
}

func (s *Store) GetBuckets(metricName string) []float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if buckets, ok := s.buckets[metricName]; ok {
		return buckets
	}
	return DefaultBuckets
}

func (s *Store) IncrementCounter(metricName string, labelKey string, value float64) error {
	if value < 0 {
		return ErrInvalidValue
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := SeriesKey{MetricName: metricName, LabelKey: labelKey}
	if counter, ok := s.counters[key]; ok {
		counter.Value += value
	} else {
		s.counters[key] = &CounterData{Value: value}
	}
	return nil
}

func (s *Store) SetGauge(metricName string, labelKey string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := SeriesKey{MetricName: metricName, LabelKey: labelKey}
	s.gauges[key] = &GaugeData{Value: value}
}

func (s *Store) getBucketsNoLock(metricName string) []float64 {
	if buckets, ok := s.buckets[metricName]; ok {
		return buckets
	}
	return DefaultBuckets
}

func (s *Store) ObserveHistogram(metricName string, labelKey string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := SeriesKey{MetricName: metricName, LabelKey: labelKey}
	if hist, ok := s.histograms[key]; ok {
		hist.Observe(value)
	} else {
		buckets := s.getBucketsNoLock(metricName)
		hist = NewHistogramData(buckets)
		hist.Observe(value)
		s.histograms[key] = hist
	}
}

func (s *Store) GetCounter(metricName string, labelKey string) (*CounterData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	counter, ok := s.counters[SeriesKey{MetricName: metricName, LabelKey: labelKey}]
	return counter, ok
}

func (s *Store) GetGauge(metricName string, labelKey string) (*GaugeData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	gauge, ok := s.gauges[SeriesKey{MetricName: metricName, LabelKey: labelKey}]
	return gauge, ok
}

func (s *Store) GetHistogram(metricName string, labelKey string) (*HistogramData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	hist, ok := s.histograms[SeriesKey{MetricName: metricName, LabelKey: labelKey}]
	return hist, ok
}

func (s *Store) GetCounterSeries(metricName string) map[string]*CounterData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]*CounterData)
	for k, v := range s.counters {
		if k.MetricName == metricName {
			result[k.LabelKey] = v
		}
	}
	return result
}

func (s *Store) GetGaugeSeries(metricName string) map[string]*GaugeData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]*GaugeData)
	for k, v := range s.gauges {
		if k.MetricName == metricName {
			result[k.LabelKey] = v
		}
	}
	return result
}

func (s *Store) GetHistogramSeries(metricName string) map[string]*HistogramData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]*HistogramData)
	for k, v := range s.histograms {
		if k.MetricName == metricName {
			result[k.LabelKey] = v
		}
	}
	return result
}

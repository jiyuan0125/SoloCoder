package server

import (
	"metric-aggregator/pkg/common"
	"sort"
	"strconv"
	"sync"
	"time"
)

type rawDataPoint struct {
	Timestamp int64
	Value     float64
}

type aggregateBucket struct {
	Timestamp int64
	Count     int
	Sum       float64
	Min       float64
	Max       float64
}

type metricData struct {
	Name           string
	RawPoints      []rawDataPoint
	Aggregated1m   map[int64]*aggregateBucket
	Aggregated5m   map[int64]*aggregateBucket
	Aggregated1h   map[int64]*aggregateBucket
	Outliers       []rawDataPoint
	mu             sync.RWMutex
	CreatedAt      int64
}

type cacheEntry struct {
	Result    common.AggregationResult
	ExpiresAt int64
}

type Store struct {
	metrics map[string]*metricData
	cache   map[string]cacheEntry
	mu      sync.RWMutex
	stopCh  chan struct{}
}

func NewStore() *Store {
	s := &Store{
		metrics: make(map[string]*metricData),
		cache:   make(map[string]cacheEntry),
		stopCh:  make(chan struct{}),
	}
	go s.cleanupLoop()
	return s
}

func (s *Store) Stop() {
	close(s.stopCh)
}

func (s *Store) MetricExists(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.metrics[name]
	return exists
}

func (s *Store) CreateMetric(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.metrics[name]; exists {
		return &MetricError{Code: common.ErrMetricExists, Metric: name}
	}
	
	s.metrics[name] = &metricData{
		Name:         name,
		RawPoints:    make([]rawDataPoint, 0),
		Aggregated1m: make(map[int64]*aggregateBucket),
		Aggregated5m: make(map[int64]*aggregateBucket),
		Aggregated1h: make(map[int64]*aggregateBucket),
		Outliers:     make([]rawDataPoint, 0),
		CreatedAt:    time.Now().Unix(),
	}
	return nil
}

func (s *Store) DeleteMetric(name string) error {
	s.mu.Lock()
	
	if _, exists := s.metrics[name]; !exists {
		s.mu.Unlock()
		return &MetricError{Code: common.ErrMetricNotFound, Metric: name}
	}
	
	delete(s.metrics, name)
	s.mu.Unlock()
	
	s.invalidateCacheForMetric(name)
	return nil
}

func (s *Store) ReportPoint(point common.DataPoint) (bool, int64) {
	ts := common.TruncateTimestamp(point.Timestamp)
	value := point.Value
	
	s.mu.RLock()
	m, exists := s.metrics[point.Metric]
	s.mu.RUnlock()
	
	if !exists {
		return false, ts
	}
	
	if !common.ValidValue(value) {
		m.mu.Lock()
		m.Outliers = append(m.Outliers, rawDataPoint{Timestamp: ts, Value: value})
		m.mu.Unlock()
		return true, ts
	}
	
	m.mu.Lock()
	
	m.RawPoints = append(m.RawPoints, rawDataPoint{Timestamp: ts, Value: value})
	
	s.aggregatePoint(m, ts, value)
	
	m.mu.Unlock()
	
	s.invalidateCacheForMetric(point.Metric)
	
	return true, ts
}

func (s *Store) aggregatePoint(m *metricData, ts int64, value float64) {
	granularities := []struct {
		g        common.Granularity
		bucket   *map[int64]*aggregateBucket
	}{
		{common.Granularity1Min, &m.Aggregated1m},
		{common.Granularity5Min, &m.Aggregated5m},
		{common.Granularity1Hour, &m.Aggregated1h},
	}
	
	for _, g := range granularities {
		alignedTs := common.AlignToGranularity(ts, g.g)
		bucket := *g.bucket
		
		if b, exists := bucket[alignedTs]; exists {
			b.Count++
			b.Sum += value
			if value < b.Min {
				b.Min = value
			}
			if value > b.Max {
				b.Max = value
			}
		} else {
			bucket[alignedTs] = &aggregateBucket{
				Timestamp: alignedTs,
				Count:     1,
				Sum:       value,
				Min:       value,
				Max:       value,
			}
		}
	}
}

func (s *Store) Query(req common.QueryRequest) ([]common.AggregationResult, error) {
	if len(req.Metrics) == 0 {
		return nil, &MetricError{Code: common.ErrEmptyMetrics}
	}
	
	if req.Start >= req.End {
		return nil, &MetricError{Code: common.ErrInvalidTimeRange}
	}
	
	if !common.ValidGranularity(req.Granularity) {
		return nil, &MetricError{Code: common.ErrInvalidGranularity}
	}
	
	if !common.ValidAggregation(req.Aggregation) {
		return nil, &MetricError{Code: common.ErrInvalidAggregation}
	}
	
	results := make([]common.AggregationResult, 0, len(req.Metrics))
	
	for _, metricName := range req.Metrics {
		result, err := s.querySingle(metricName, req.Start, req.End, req.Granularity, req.Aggregation)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	
	return results, nil
}

func (s *Store) querySingle(
	metricName string,
	start, end int64,
	g common.Granularity,
	aggType common.AggregationType,
) (common.AggregationResult, error) {
	cacheKey := s.buildCacheKey(metricName, start, end, g, aggType)
	if cached, exists := s.getFromCache(cacheKey); exists {
		return cached, nil
	}
	
	s.mu.RLock()
	m, exists := s.metrics[metricName]
	s.mu.RUnlock()
	
	if !exists {
		return common.AggregationResult{}, &MetricError{Code: common.ErrMetricNotFound, Metric: metricName}
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var buckets map[int64]*aggregateBucket
	switch g {
	case common.Granularity1Min:
		buckets = m.Aggregated1m
	case common.Granularity5Min:
		buckets = m.Aggregated5m
	case common.Granularity1Hour:
		buckets = m.Aggregated1h
	}
	
	alignedStart := common.AlignToGranularity(start, g)
	alignedEnd := common.AlignToGranularity(end, g)
	interval := s.getInterval(g)
	
	var previousValue float64 = 0
	var points []common.AggregatedPoint
	
	for ts := alignedStart; ts < alignedEnd; ts += interval {
		if bucket, exists := buckets[ts]; exists {
			var value float64
			switch aggType {
			case common.AggregationAvg:
				value = bucket.Sum / float64(bucket.Count)
			case common.AggregationSum:
				value = bucket.Sum
			case common.AggregationMin:
				value = bucket.Min
			case common.AggregationMax:
				value = bucket.Max
			}
			
			points = append(points, common.AggregatedPoint{
				Timestamp: ts,
				Value:     value,
				Count:     bucket.Count,
				IsFilled:  false,
			})
			previousValue = value
		} else {
			points = append(points, common.AggregatedPoint{
				Timestamp: ts,
				Value:     previousValue,
				Count:     0,
				IsFilled:  true,
			})
		}
	}
	
	result := common.AggregationResult{
		Metric:      metricName,
		Start:       alignedStart,
		End:         alignedEnd,
		Granularity: g,
		Points:      points,
	}
	
	s.setCache(cacheKey, result)
	return result, nil
}

func (s *Store) getInterval(g common.Granularity) int64 {
	switch g {
	case common.Granularity1Min:
		return 60
	case common.Granularity5Min:
		return 300
	case common.Granularity1Hour:
		return 3600
	}
	return 60
}

func (s *Store) buildCacheKey(
	metric string,
	start, end int64,
	g common.Granularity,
	aggType common.AggregationType,
) string {
	return metric + "_" + string(g) + "_" + string(aggType) + "_" +
		strconv.FormatInt(start, 10) + "_" + strconv.FormatInt(end, 10)
}

func (s *Store) getFromCache(key string) (common.AggregationResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if entry, exists := s.cache[key]; exists {
		if time.Now().Unix() < entry.ExpiresAt {
			return entry.Result, true
		}
		delete(s.cache, key)
	}
	return common.AggregationResult{}, false
}

func (s *Store) setCache(key string, result common.AggregationResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.cache[key] = cacheEntry{
		Result:    result,
		ExpiresAt: time.Now().Unix() + 300,
	}
}

func (s *Store) invalidateCacheForMetric(metric string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for key := range s.cache {
		if len(key) >= len(metric) && key[:len(metric)] == metric {
			delete(s.cache, key)
		}
	}
}

func (s *Store) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.stopCh:
			return
		}
	}
}

func (s *Store) cleanup() {
	cutoff := time.Now().Unix() - 3600
	
	s.mu.RLock()
	for _, m := range s.metrics {
		m.mu.Lock()
		
		filtered := make([]rawDataPoint, 0, len(m.RawPoints))
		for _, p := range m.RawPoints {
			if p.Timestamp >= cutoff {
				filtered = append(filtered, p)
			}
		}
		m.RawPoints = filtered
		
		m.mu.Unlock()
	}
	s.mu.RUnlock()
}

type MetricError struct {
	Code   common.ErrorCode
	Metric string
}

func (e *MetricError) Error() string {
	if e.Metric != "" {
		return e.Code.Message() + ": " + e.Metric
	}
	return e.Code.Message()
}

func (e *MetricError) ToResponse() common.ErrorResponse {
	return common.ErrorResponse{
		Code:    e.Code,
		Message: e.Error(),
	}
}

func (s *Store) ListMetrics() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	metrics := make([]string, 0, len(s.metrics))
	for name := range s.metrics {
		metrics = append(metrics, name)
	}
	sort.Strings(metrics)
	return metrics
}

func (s *Store) GetMetricInfo(name string) (map[string]interface{}, error) {
	s.mu.RLock()
	m, exists := s.metrics[name]
	s.mu.RUnlock()
	
	if !exists {
		return nil, &MetricError{Code: common.ErrMetricNotFound, Metric: name}
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	info := map[string]interface{}{
		"name":           name,
		"raw_points":     len(m.RawPoints),
		"outliers":       len(m.Outliers),
		"created_at":     m.CreatedAt,
		"aggregated_1m":  len(m.Aggregated1m),
		"aggregated_5m":  len(m.Aggregated5m),
		"aggregated_1h":  len(m.Aggregated1h),
	}
	
	return info, nil
}

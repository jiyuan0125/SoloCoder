package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type MetricType string

const (
	TypeCounter   MetricType = "counter"
	TypeGauge     MetricType = "gauge"
	TypeHistogram MetricType = "histogram"
)

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Labels []Label

func (l Labels) Key() string {
	sorted := make(Labels, len(l))
	copy(sorted, l)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Key < sorted[j].Key
	})
	var builder strings.Builder
	for i, lb := range sorted {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(lb.Key)
		builder.WriteString("=")
		builder.WriteString(lb.Value)
	}
	return builder.String()
}

func (l Labels) ToMap() map[string]string {
	m := make(map[string]string)
	for _, lb := range l {
		m[lb.Key] = lb.Value
	}
	return m
}

func LabelsFromMap(m map[string]string) Labels {
	labels := make(Labels, 0, len(m))
	for k, v := range m {
		labels = append(labels, Label{Key: k, Value: v})
	}
	sort.Slice(labels, func(i, j int) bool {
		return labels[i].Key < labels[j].Key
	})
	return labels
}

func LabelsEqual(a, b Labels) bool {
	return a.Key() == b.Key()
}

type TimeSeriesPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type HistogramState struct {
	Buckets map[float64]uint64
	Sum     float64
	Count   uint64
	Values  []float64
}

func newHistogramState() *HistogramState {
	return &HistogramState{
		Buckets: map[float64]uint64{
			0.005: 0, 0.01: 0, 0.025: 0, 0.05: 0, 0.1: 0,
			0.25: 0, 0.5: 0, 1.0: 0, 2.5: 0, 5.0: 0, 10.0: 0,
		},
		Values: make([]float64, 0),
	}
}

type MetricFamily struct {
	Name string
	Type MetricType
}

type MetricSeries struct {
	Family    MetricFamily
	Labels    Labels
	Value     float64
	Histogram *HistogramState
	Points    []TimeSeriesPoint
	mu        sync.RWMutex
}

func newMetricSeries(family MetricFamily, labels Labels) *MetricSeries {
	ms := &MetricSeries{
		Family: family,
		Labels: labels,
		Points: make([]TimeSeriesPoint, 0),
	}
	if family.Type == TypeHistogram {
		ms.Histogram = newHistogramState()
	}
	return ms
}

func (ms *MetricSeries) addCounter(timestamp int64, value float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Value += value
	ms.appendPoint(timestamp, ms.Value)
}

func (ms *MetricSeries) setGauge(timestamp int64, value float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Value = value
	ms.appendPoint(timestamp, value)
}

func (ms *MetricSeries) observeHistogram(timestamp int64, value float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Histogram.Count++
	ms.Histogram.Sum += value
	ms.Histogram.Values = append(ms.Histogram.Values, value)
	for bound := range ms.Histogram.Buckets {
		if value <= bound {
			ms.Histogram.Buckets[bound]++
		}
	}
	ms.Histogram.Buckets[1e9] = ms.Histogram.Count
	ms.appendPoint(timestamp, value)
}

func (ms *MetricSeries) appendPoint(timestamp int64, value float64) {
	ms.Points = append(ms.Points, TimeSeriesPoint{Timestamp: timestamp, Value: value})
	if len(ms.Points) > 10000 {
		ms.Points = ms.Points[len(ms.Points)-10000:]
	}
}

func (ms *MetricSeries) latestValue() (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if ms.Family.Type == TypeHistogram && ms.Histogram != nil {
		return float64(ms.Histogram.Count), ms.Histogram.Count > 0
	}
	if len(ms.Points) == 0 {
		return 0, false
	}
	return ms.Points[len(ms.Points)-1].Value, true
}

func (ms *MetricSeries) getPoints(start, end int64) []TimeSeriesPoint {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if len(ms.Points) == 0 {
		return nil
	}
	result := make([]TimeSeriesPoint, 0)
	for _, p := range ms.Points {
		if p.Timestamp >= start && p.Timestamp <= end {
			result = append(result, p)
		}
	}
	return result
}

type MetricStore struct {
	series map[string]*MetricSeries
	mu     sync.RWMutex
}

func newMetricStore() *MetricStore {
	return &MetricStore{series: make(map[string]*MetricSeries)}
}

func seriesKey(name, labelKey string) string {
	if labelKey == "" {
		return name
	}
	return name + "{" + labelKey + "}"
}

func (s *MetricStore) getOrCreateSeries(family MetricFamily, labels Labels) *MetricSeries {
	key := seriesKey(family.Name, labels.Key())
	s.mu.RLock()
	if ms, ok := s.series[key]; ok {
		s.mu.RUnlock()
		return ms
	}
	s.mu.RUnlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	if ms, ok := s.series[key]; ok {
		return ms
	}
	ms := newMetricSeries(family, labels)
	s.series[key] = ms
	return ms
}

func (s *MetricStore) findSeriesByName(name string) []*MetricSeries {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*MetricSeries, 0)
	for k, ms := range s.series {
		if strings.HasPrefix(k, name+"{") || k == name {
			result = append(result, ms)
		}
	}
	return result
}

func (s *MetricStore) findSeriesByPrefix(prefix string) []*MetricSeries {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*MetricSeries, 0)
	for k, ms := range s.series {
		if strings.HasPrefix(k, prefix) {
			result = append(result, ms)
		}
	}
	return result
}

func (s *MetricStore) deleteByPrefix(prefix string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.series {
		if strings.HasPrefix(k, prefix) {
			delete(s.series, k)
		}
	}
}

func (s *MetricStore) listAll() []*MetricSeries {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*MetricSeries, 0, len(s.series))
	for _, ms := range s.series {
		result = append(result, ms)
	}
	return result
}

type CounterReport struct {
	Name      string  `json:"name"`
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
	Labels    []Label `json:"labels"`
}

type GaugeReport struct {
	Name      string  `json:"name"`
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
	Labels    []Label `json:"labels"`
}

type HistogramReport struct {
	Name      string  `json:"name"`
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
	Labels    []Label `json:"labels"`
}

type LatestValueResponse struct {
	Name      string             `json:"name"`
	Type      MetricType         `json:"type"`
	Labels    map[string]string  `json:"labels"`
	Value     *float64           `json:"value,omitempty"`
	Count     *uint64            `json:"count,omitempty"`
	Sum       *float64           `json:"sum,omitempty"`
	Quantiles map[string]float64 `json:"quantiles,omitempty"`
}

type TimeSeriesResponse struct {
	Name     string            `json:"name"`
	Type     MetricType        `json:"type"`
	Labels   map[string]string `json:"labels"`
	Data     []TimeSeriesPoint `json:"data"`
}

type AggregateResponse struct {
	Name      string             `json:"name"`
	Type      MetricType         `json:"type"`
	Aggregate map[string]float64 `json:"aggregate"`
}

var store = newMetricStore()

func getNowMillis() int64 {
	return time.Now().UnixMilli()
}

func reportCounter(c *gin.Context) {
	var req CounterReport
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if req.Timestamp == 0 {
		req.Timestamp = getNowMillis()
	}
	if req.Labels == nil {
		req.Labels = make([]Label, 0)
	}
	family := MetricFamily{Name: req.Name, Type: TypeCounter}
	ms := store.getOrCreateSeries(family, Labels(req.Labels))
	ms.addCounter(req.Timestamp, req.Value)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func reportGauge(c *gin.Context) {
	var req GaugeReport
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if req.Timestamp == 0 {
		req.Timestamp = getNowMillis()
	}
	if req.Labels == nil {
		req.Labels = make([]Label, 0)
	}
	family := MetricFamily{Name: req.Name, Type: TypeGauge}
	ms := store.getOrCreateSeries(family, Labels(req.Labels))
	ms.setGauge(req.Timestamp, req.Value)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func reportHistogram(c *gin.Context) {
	var req HistogramReport
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if req.Timestamp == 0 {
		req.Timestamp = getNowMillis()
	}
	if req.Labels == nil {
		req.Labels = make([]Label, 0)
	}
	family := MetricFamily{Name: req.Name, Type: TypeHistogram}
	ms := store.getOrCreateSeries(family, Labels(req.Labels))
	ms.observeHistogram(req.Timestamp, req.Value)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func parseTimeRange(c *gin.Context) (start, end int64, ok bool) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	if startStr == "" || endStr == "" {
		end = getNowMillis()
		start = end - 5*60*1000
		return start, end, true
	}
	var err error
	start, err = strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start timestamp"})
		return 0, 0, false
	}
	end, err = strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end timestamp"})
		return 0, 0, false
	}
	return start, end, true
}

func queryLatest(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	series := store.findSeriesByName(name)
	if len(series) == 0 {
		c.JSON(http.StatusOK, []LatestValueResponse{})
		return
	}
	results := make([]LatestValueResponse, 0, len(series))
	for _, ms := range series {
		resp := LatestValueResponse{
			Name:   ms.Family.Name,
			Type:   ms.Family.Type,
			Labels: ms.Labels.ToMap(),
		}
		if ms.Family.Type == TypeHistogram {
			ms.mu.RLock()
			if ms.Histogram != nil {
				count := ms.Histogram.Count
				sum := ms.Histogram.Sum
				resp.Count = &count
				resp.Sum = &sum
				resp.Quantiles = calculateQuantiles(ms.Histogram.Values)
			}
			ms.mu.RUnlock()
		} else {
			if val, ok := ms.latestValue(); ok {
				resp.Value = &val
			}
		}
		results = append(results, resp)
	}
	c.JSON(http.StatusOK, results)
}

func calculateQuantiles(values []float64) map[string]float64 {
	if len(values) == 0 {
		return nil
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	quantile := func(q float64) float64 {
		n := len(sorted)
		if n == 0 {
			return 0
		}
		pos := float64(n-1) * q
		lower := int(pos)
		upper := lower
		if upper < n-1 {
			upper++
		}
		frac := pos - float64(lower)
		return sorted[lower]*(1-frac) + sorted[upper]*frac
	}
	return map[string]float64{
		"p50": quantile(0.50),
		"p95": quantile(0.95),
		"p99": quantile(0.99),
	}
}

func queryTimeSeries(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	start, end, ok := parseTimeRange(c)
	if !ok {
		return
	}
	series := store.findSeriesByName(name)
	if len(series) == 0 {
		c.JSON(http.StatusOK, []TimeSeriesResponse{})
		return
	}
	results := make([]TimeSeriesResponse, 0, len(series))
	for _, ms := range series {
		resp := TimeSeriesResponse{
			Name:   ms.Family.Name,
			Type:   ms.Family.Type,
			Labels: ms.Labels.ToMap(),
			Data:   ms.getPoints(start, end),
		}
		if resp.Data == nil {
			resp.Data = []TimeSeriesPoint{}
		}
		results = append(results, resp)
	}
	c.JSON(http.StatusOK, results)
}

func queryAggregate(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	aggLabel := c.Query("by")
	if aggLabel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "by is required"})
		return
	}
	series := store.findSeriesByName(name)
	if len(series) == 0 {
		c.JSON(http.StatusOK, []AggregateResponse{})
		return
	}
	metricType := series[0].Family.Type
	groups := make(map[string][]float64)
	for _, ms := range series {
		labelMap := ms.Labels.ToMap()
		key := labelMap[aggLabel]
		var val float64
		if metricType == TypeHistogram {
			ms.mu.RLock()
			val = float64(ms.Histogram.Count)
			ms.mu.RUnlock()
		} else {
			if v, ok := ms.latestValue(); ok {
				val = v
			}
		}
		groups[key] = append(groups[key], val)
	}
	result := AggregateResponse{
		Name:      name,
		Type:      metricType,
		Aggregate: make(map[string]float64),
	}
	for key, vals := range groups {
		sum := 0.0
		for _, v := range vals {
			sum += v
		}
		result.Aggregate[key] = sum
	}
	c.JSON(http.StatusOK, result)
}

func deleteGroup(c *gin.Context) {
	prefix := c.Query("prefix")
	if prefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prefix is required"})
		return
	}
	store.deleteByPrefix(prefix)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func listGroups(c *gin.Context) {
	prefix := c.Query("prefix")
	series := store.findSeriesByPrefix(prefix)
	type GroupInfo struct {
		Name   string            `json:"name"`
		Type   MetricType        `json:"type"`
		Labels map[string]string `json:"labels"`
	}
	seen := make(map[string]bool)
	groups := make([]GroupInfo, 0)
	for _, ms := range series {
		key := ms.Family.Name + "|" + string(ms.Family.Type)
		if !seen[key] {
			seen[key] = true
			groups = append(groups, GroupInfo{
				Name:   ms.Family.Name,
				Type:   ms.Family.Type,
				Labels: ms.Labels.ToMap(),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"prefix": prefix, "metrics": groups})
}

func prometheusFormat(c *gin.Context) {
	all := store.listAll()
	grouped := make(map[string][]*MetricSeries)
	for _, ms := range all {
		key := ms.Family.Name + "|" + string(ms.Family.Type)
		grouped[key] = append(grouped[key], ms)
	}
	var builder strings.Builder
	for _, seriesList := range grouped {
		if len(seriesList) == 0 {
			continue
		}
		ms := seriesList[0]
		name := ms.Family.Name
		typ := ms.Family.Type
		builder.WriteString(fmt.Sprintf("# HELP %s %s\n", name, name))
		builder.WriteString(fmt.Sprintf("# TYPE %s %s\n", name, typ))
		for _, s := range seriesList {
			labelStr := formatLabels(s.Labels)
			switch typ {
			case TypeCounter:
				s.mu.RLock()
				builder.WriteString(fmt.Sprintf("%s%s %f\n", name, labelStr, s.Value))
				s.mu.RUnlock()
			case TypeGauge:
				s.mu.RLock()
				builder.WriteString(fmt.Sprintf("%s%s %f\n", name, labelStr, s.Value))
				s.mu.RUnlock()
			case TypeHistogram:
				s.mu.RLock()
				bucketLabels := make(map[string]string)
				for _, l := range s.Labels {
					bucketLabels[l.Key] = l.Value
				}
				bounds := make([]float64, 0, len(s.Histogram.Buckets))
				for b := range s.Histogram.Buckets {
					bounds = append(bounds, b)
				}
				sort.Float64s(bounds)
				for _, bound := range bounds {
					bucketLabels["le"] = fmt.Sprintf("%g", bound)
					bl := LabelsFromMap(bucketLabels)
					builder.WriteString(fmt.Sprintf("%s_bucket%s %d\n", name, formatLabels(bl), s.Histogram.Buckets[bound]))
				}
				builder.WriteString(fmt.Sprintf("%s_sum%s %f\n", name, labelStr, s.Histogram.Sum))
				builder.WriteString(fmt.Sprintf("%s_count%s %d\n", name, labelStr, s.Histogram.Count))
				s.mu.RUnlock()
			}
		}
	}
	c.Data(http.StatusOK, "text/plain; version=0.0.4", []byte(builder.String()))
}

func formatLabels(labels Labels) string {
	if len(labels) == 0 {
		return ""
	}
	parts := make([]string, 0, len(labels))
	for _, l := range labels {
		parts = append(parts, fmt.Sprintf("%s=%q", l.Key, l.Value))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r := gin.Default()

	api := r.Group("/api")
	{
		report := api.Group("/report")
		{
			report.POST("/counter", reportCounter)
			report.POST("/gauge", reportGauge)
			report.POST("/histogram", reportHistogram)
		}
		query := api.Group("/query")
		{
			query.GET("/latest", queryLatest)
			query.GET("/timeseries", queryTimeSeries)
			query.GET("/aggregate", queryAggregate)
		}
		api.GET("/groups", listGroups)
		api.DELETE("/groups", deleteGroup)
	}
	r.GET("/metrics", prometheusFormat)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = ctx

	_ = r.Run(":" + port)
}

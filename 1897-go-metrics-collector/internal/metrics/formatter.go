package metrics

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type MetricValue struct {
	Name        string
	Type        MetricType
	Labels      Labels
	LabelKey    string
	Counter     *CounterValue
	Gauge       *GaugeValue
	Histogram   *HistogramValue
}

type CounterValue struct {
	Value float64
}

type GaugeValue struct {
	Value float64
}

type HistogramBucketValue struct {
	UpperBound float64
	Count      int
}

type HistogramValue struct {
	Count   int
	Sum     float64
	P50     float64
	P90     float64
	P99     float64
	Buckets []HistogramBucketValue
}

func ParseLabelKey(labelKey string) Labels {
	if labelKey == "" {
		return Labels{}
	}
	result := make(Labels)
	pairs := strings.Split(labelKey, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func formatLabelValue(v string) string {
	v = strings.ReplaceAll(v, "\\", "\\\\")
	v = strings.ReplaceAll(v, "\"", "\\\"")
	v = strings.ReplaceAll(v, "\n", "\\n")
	return v
}

func formatLabelsPrometheus(labels Labels) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var builder strings.Builder
	builder.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(k)
		builder.WriteString("=\"")
		builder.WriteString(formatLabelValue(labels[k]))
		builder.WriteByte('"')
	}
	builder.WriteByte('}')
	return builder.String()
}

func formatFloat(f float64) string {
	if math.IsNaN(f) {
		return "NaN"
	}
	if math.IsInf(f, 1) {
		return "+Inf"
	}
	if math.IsInf(f, -1) {
		return "-Inf"
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func ToPrometheusFormat(registry *Registry, store *Store) string {
	metrics := registry.List()
	if len(metrics) == 0 {
		return ""
	}

	var builder strings.Builder

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})

	for _, desc := range metrics {
		switch desc.Type {
		case TypeCounter:
			builder.WriteString(fmt.Sprintf("# HELP %s Counter metric\n", desc.Name))
			builder.WriteString(fmt.Sprintf("# TYPE %s counter\n", desc.Name))
			series := store.GetCounterSeries(desc.Name)
			for labelKey, data := range series {
				labels := ParseLabelKey(labelKey)
				labelStr := formatLabelsPrometheus(labels)
				builder.WriteString(fmt.Sprintf("%s%s %s\n", desc.Name, labelStr, formatFloat(data.Value)))
			}
		case TypeGauge:
			builder.WriteString(fmt.Sprintf("# HELP %s Gauge metric\n", desc.Name))
			builder.WriteString(fmt.Sprintf("# TYPE %s gauge\n", desc.Name))
			series := store.GetGaugeSeries(desc.Name)
			for labelKey, data := range series {
				labels := ParseLabelKey(labelKey)
				labelStr := formatLabelsPrometheus(labels)
				builder.WriteString(fmt.Sprintf("%s%s %s\n", desc.Name, labelStr, formatFloat(data.Value)))
			}
		case TypeHistogram:
			builder.WriteString(fmt.Sprintf("# HELP %s Histogram metric\n", desc.Name))
			builder.WriteString(fmt.Sprintf("# TYPE %s histogram\n", desc.Name))
			series := store.GetHistogramSeries(desc.Name)
			for labelKey, data := range series {
				labels := ParseLabelKey(labelKey)
				buckets := store.GetBuckets(desc.Name)
				sortedBuckets := make([]float64, len(buckets))
				copy(sortedBuckets, buckets)
				sort.Float64s(sortedBuckets)

				for _, bucket := range sortedBuckets {
					bucketLabels := make(Labels, len(labels)+1)
					for k, v := range labels {
						bucketLabels[k] = v
					}
					bucketLabels["le"] = formatFloat(bucket)
					labelStr := formatLabelsPrometheus(bucketLabels)
					count := data.Buckets[bucket]
					builder.WriteString(fmt.Sprintf("%s_bucket%s %d\n", desc.Name, labelStr, count))
				}

				labelStr := formatLabelsPrometheus(labels)
				builder.WriteString(fmt.Sprintf("%s_sum%s %s\n", desc.Name, labelStr, formatFloat(data.Sum)))
				builder.WriteString(fmt.Sprintf("%s_count%s %d\n", desc.Name, labelStr, data.Count))
			}
		}
	}

	return builder.String()
}

func GetMetricValues(registry *Registry, store *Store, name string) ([]*MetricValue, bool) {
	desc, exists := registry.Get(name)
	if !exists {
		return nil, false
	}

	var values []*MetricValue

	switch desc.Type {
	case TypeCounter:
		series := store.GetCounterSeries(name)
		for labelKey, data := range series {
			values = append(values, &MetricValue{
				Name:     name,
				Type:     TypeCounter,
				Labels:   ParseLabelKey(labelKey),
				LabelKey: labelKey,
				Counter:  &CounterValue{Value: data.Value},
			})
		}
	case TypeGauge:
		series := store.GetGaugeSeries(name)
		for labelKey, data := range series {
			values = append(values, &MetricValue{
				Name:     name,
				Type:     TypeGauge,
				Labels:   ParseLabelKey(labelKey),
				LabelKey: labelKey,
				Gauge:    &GaugeValue{Value: data.Value},
			})
		}
	case TypeHistogram:
		series := store.GetHistogramSeries(name)
		for labelKey, data := range series {
			buckets := make([]HistogramBucketValue, 0, len(data.Buckets))
			for bound, count := range data.Buckets {
				buckets = append(buckets, HistogramBucketValue{
					UpperBound: bound,
					Count:      count,
				})
			}
			sort.Slice(buckets, func(i, j int) bool {
				return buckets[i].UpperBound < buckets[j].UpperBound
			})
			values = append(values, &MetricValue{
				Name:     name,
				Type:     TypeHistogram,
				Labels:   ParseLabelKey(labelKey),
				LabelKey: labelKey,
				Histogram: &HistogramValue{
					Count:   data.Count,
					Sum:     data.Sum,
					P50:     data.Percentile(0.50),
					P90:     data.Percentile(0.90),
					P99:     data.Percentile(0.99),
					Buckets: buckets,
				},
			})
		}
	}

	return values, true
}

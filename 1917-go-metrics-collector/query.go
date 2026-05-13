package main

import (
	"strings"
)

type QueryFilter map[string]string

type AggregatedMetric struct {
	Tags    Tags         `json:"tags,omitempty"`
	Value   float64      `json:"value,omitempty"`
	P50     float64      `json:"p50,omitempty"`
	P90     float64      `json:"p90,omitempty"`
	P99     float64      `json:"p99,omitempty"`
	Max     float64      `json:"max,omitempty"`
	Min     float64      `json:"min,omitempty"`
}

type QueryResult struct {
	Name     string             `json:"name"`
	Type     MetricType         `json:"type"`
	Metrics  []AggregatedMetric `json:"metrics"`
}

func keyToTags(key string) Tags {
	tags := make(Tags)
	if key == "" {
		return tags
	}
	parts := strings.Split(key, ",")
	for _, part := range parts {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			tags[kv[0]] = kv[1]
		}
	}
	return tags
}

func matchesFilter(tags Tags, filter QueryFilter) bool {
	for k, v := range filter {
		if tags[k] != v {
			return false
		}
	}
	return true
}

func (m *Metric) Query(filter QueryFilter, groupBy string) *QueryResult {
	m.Lock.RLock()
	defer m.Lock.RUnlock()

	result := &QueryResult{
		Name:    m.Name,
		Type:    m.Type,
		Metrics: make([]AggregatedMetric, 0),
	}

	if groupBy == "" {
		switch m.Type {
		case Counter:
			total := 0.0
			for key, val := range m.Counter {
				tags := keyToTags(key)
				if !matchesFilter(tags, filter) {
					continue
				}
				total += val
			}
			result.Metrics = append(result.Metrics, AggregatedMetric{Value: total})
		case Gauge:
			for key, val := range m.Gauge {
				tags := keyToTags(key)
				if !matchesFilter(tags, filter) {
					continue
				}
				result.Metrics = append(result.Metrics, AggregatedMetric{Tags: tags, Value: val})
			}
		case Timer:
			merged := NewTimerValue()
			for key, tv := range m.Timer {
				tags := keyToTags(key)
				if !matchesFilter(tags, filter) {
					continue
				}
				merged.Merge(tv)
			}
			if len(merged.Values) > 0 {
				result.Metrics = append(result.Metrics, AggregatedMetric{
					P50: merged.P50(),
					P90: merged.P90(),
					P99: merged.P99(),
					Max: merged.Max(),
					Min: merged.Min(),
				})
			}
		}
	} else {
		groups := make(map[string]*AggregatedMetric)

		switch m.Type {
		case Counter:
			for key, val := range m.Counter {
				tags := keyToTags(key)
				if !matchesFilter(tags, filter) {
					continue
				}
				groupVal := tags[groupBy]
				agg, ok := groups[groupVal]
				if !ok {
					agg = &AggregatedMetric{
						Tags:  Tags{groupBy: groupVal},
						Value: 0,
					}
					groups[groupVal] = agg
				}
				agg.Value += val
			}
		case Gauge:
			for key, val := range m.Gauge {
				tags := keyToTags(key)
				if !matchesFilter(tags, filter) {
					continue
				}
				groupVal := tags[groupBy]
				agg, ok := groups[groupVal]
				if !ok {
					agg = &AggregatedMetric{
						Tags:  Tags{groupBy: groupVal},
						Value: 0,
					}
					groups[groupVal] = agg
				}
				agg.Value = val
			}
		case Timer:
			groupTimers := make(map[string]*TimerValue)
			for key, tv := range m.Timer {
				tags := keyToTags(key)
				if !matchesFilter(tags, filter) {
					continue
				}
				groupVal := tags[groupBy]
				merged, ok := groupTimers[groupVal]
				if !ok {
					merged = NewTimerValue()
					groupTimers[groupVal] = merged
				}
				merged.Merge(tv)
			}

			for groupVal, merged := range groupTimers {
				if len(merged.Values) > 0 {
					groups[groupVal] = &AggregatedMetric{
						Tags: Tags{groupBy: groupVal},
						P50:  merged.P50(),
						P90:  merged.P90(),
						P99:  merged.P99(),
						Max:  merged.Max(),
						Min:  merged.Min(),
					}
				}
			}
		}

		for _, agg := range groups {
			result.Metrics = append(result.Metrics, *agg)
		}
	}

	return result
}

package internal

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Aggregator struct {
	config       *Config
	states       map[int64]map[int]*AggregationState
	minTime      time.Time
	maxTime      time.Time
	header       []string
	totalRows    int
	validRows    int
	invalidRows  int
}

func NewAggregator(config *Config) *Aggregator {
	return &Aggregator{
		config: config,
		states: make(map[int64]map[int]*AggregationState),
	}
}

func (a *Aggregator) SetHeader(header []string) {
	a.header = header
}

func (a *Aggregator) GetBucketStart(t time.Time) int64 {
	switch a.config.Granularity {
	case GranularityMinute:
		return t.Truncate(time.Minute).Unix()
	case GranularityHour:
		return t.Truncate(time.Hour).Unix()
	case GranularityDay:
		year, month, day := t.Date()
		return time.Date(year, month, day, 0, 0, 0, 0, t.Location()).Unix()
	default:
		return t.Unix()
	}
}

func (a *Aggregator) GetBucketEnd(bucketStart int64, loc *time.Location) int64 {
	start := time.Unix(bucketStart, 0).In(loc)
	switch a.config.Granularity {
	case GranularityMinute:
		return start.Add(time.Minute).Unix()
	case GranularityHour:
		return start.Add(time.Hour).Unix()
	case GranularityDay:
		return start.AddDate(0, 0, 1).Unix()
	default:
		return bucketStart
	}
}

func (a *Aggregator) AddRow(t time.Time, values []float64) {
	a.totalRows++
	a.validRows++

	if a.minTime.IsZero() || t.Before(a.minTime) {
		a.minTime = t
	}
	if a.maxTime.IsZero() || t.After(a.maxTime) {
		a.maxTime = t
	}

	bucketStart := a.GetBucketStart(t)

	if _, exists := a.states[bucketStart]; !exists {
		a.states[bucketStart] = make(map[int]*AggregationState)
	}

	for i := range a.config.Columns {
		if _, exists := a.states[bucketStart][i]; !exists {
			a.states[bucketStart][i] = NewAggregationState()
		}

		val := values[i]
		state := a.states[bucketStart][i]

		state.Count++
		state.Sum += val
		if val < state.Min {
			state.Min = val
		}
		if val > state.Max {
			state.Max = val
		}
	}
}

func (a *Aggregator) AddInvalidRow() {
	a.totalRows++
	a.invalidRows++
}

func (a *Aggregator) ComputeResult(colIdx int, state *AggregationState) string {
	method := a.config.Columns[colIdx].Method

	switch method {
	case AggAvg:
		if state.Count == 0 {
			return "null"
		}
		return strconv.FormatFloat(state.Sum/float64(state.Count), 'f', -1, 64)
	case AggMax:
		if state.Count == 0 {
			return "null"
		}
		return strconv.FormatFloat(state.Max, 'f', -1, 64)
	case AggMin:
		if state.Count == 0 {
			return "null"
		}
		return strconv.FormatFloat(state.Min, 'f', -1, 64)
	case AggSum:
		return strconv.FormatFloat(state.Sum, 'f', -1, 64)
	case AggCount:
		return strconv.Itoa(state.Count)
	default:
		return "null"
	}
}

func (a *Aggregator) GenerateBuckets(loc *time.Location) []int64 {
	var buckets []int64

	if a.minTime.IsZero() || a.maxTime.IsZero() {
		return buckets
	}

	startBucket := a.GetBucketStart(a.minTime)
	endBucket := a.GetBucketStart(a.maxTime)

	for bucket := startBucket; bucket <= endBucket; {
		buckets = append(buckets, bucket)

		bucketStart := time.Unix(bucket, 0).In(loc)
		switch a.config.Granularity {
		case GranularityMinute:
			bucket = bucketStart.Add(time.Minute).Unix()
		case GranularityHour:
			bucket = bucketStart.Add(time.Hour).Unix()
		case GranularityDay:
			bucket = bucketStart.AddDate(0, 0, 1).Unix()
		}
	}

	return buckets
}

func (a *Aggregator) WriteOutput(filename string, loc *time.Location) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %v", err)
	}
	defer file.Close()

	var headers []string
	headers = append(headers, "time")
	for _, col := range a.config.Columns {
		colName := col.Name
		if colName == "" && col.Index >= 0 && col.Index < len(a.header) {
			colName = a.header[col.Index]
		}
		if colName == "" {
			colName = fmt.Sprintf("col%d", col.Index)
		}
		headers = append(headers, fmt.Sprintf("%s_%s", colName, string(col.Method)))
	}

	file.WriteString(strings.Join(headers, ",") + "\n")

	if a.minTime.IsZero() || a.maxTime.IsZero() {
		return nil
	}

	buckets := a.GenerateBuckets(loc)
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i] < buckets[j]
	})

	for _, bucketStart := range buckets {
		row := make([]string, 0, len(a.config.Columns)+1)

		bucketTime := time.Unix(bucketStart, 0).In(loc)
		row = append(row, bucketTime.Format(time.RFC3339))

		states, exists := a.states[bucketStart]
		for i := range a.config.Columns {
			if exists {
				if state, ok := states[i]; ok {
					row = append(row, a.ComputeResult(i, state))
				} else {
					row = append(row, "null")
				}
			} else {
				row = append(row, "null")
			}
		}

		file.WriteString(strings.Join(row, ",") + "\n")
	}

	return nil
}

func (a *Aggregator) GetStats() map[string]interface{} {
	bucketCount := len(a.states)
	return map[string]interface{}{
		"total_rows":    a.totalRows,
		"valid_rows":    a.validRows,
		"invalid_rows":  a.invalidRows,
		"bucket_count":  bucketCount,
		"min_time":      a.minTime,
		"max_time":      a.maxTime,
	}
}

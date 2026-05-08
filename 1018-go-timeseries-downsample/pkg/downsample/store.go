package downsample

import (
	"sync"
	"time"

	"github.com/example/timeseries-downsample/pkg/common"
)

type Store struct {
	mu      sync.RWMutex
	raw     map[string][]common.DataPoint
	ohlc1m  map[string][]common.OHLC
	ohlc5m  map[string][]common.OHLC
	ohlc1h  map[string][]common.OHLC
}

func NewStore() *Store {
	return &Store{
		raw:    make(map[string][]common.DataPoint),
		ohlc1m: make(map[string][]common.OHLC),
		ohlc5m: make(map[string][]common.OHLC),
		ohlc1h: make(map[string][]common.OHLC),
	}
}

func (s *Store) Write(metric string, points []common.DataPoint) int {
	if len(points) == 0 {
		return 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.raw[metric] = append(s.raw[metric], points...)
	s.aggregate(metric, points)
	return len(points)
}

func (s *Store) aggregate(metric string, newPoints []common.DataPoint) {
	new1m := DownsampleRaw(newPoints, common.Window1Minute)
	s.ohlc1m[metric] = mergeOHLC(s.ohlc1m[metric], new1m)

	merged1m := s.ohlc1m[metric]
	s.ohlc5m[metric] = DownsampleOHLC(merged1m, common.Window5Minutes)

	merged5m := s.ohlc5m[metric]
	s.ohlc1h[metric] = DownsampleOHLC(merged5m, common.Window1Hour)
}

func mergeOHLC(existing, incoming []common.OHLC) []common.OHLC {
	byTime := make(map[time.Time]common.OHLC)
	for _, o := range existing {
		byTime[o.Timestamp] = o
	}
	for _, o := range incoming {
		if existingO, ok := byTime[o.Timestamp]; ok {
			byTime[o.Timestamp] = mergeSingleOHLC(existingO, o)
		} else {
			byTime[o.Timestamp] = o
		}
	}

	result := make([]common.OHLC, 0, len(byTime))
	for _, o := range byTime {
		result = append(result, o)
	}

	sortOHLC(result)
	return result
}

func mergeSingleOHLC(a, b common.OHLC) common.OHLC {
	if a.Open == nil {
		return b
	}
	if b.Open == nil {
		return a
	}

	agg := newOHLCAggregator(a.Timestamp)
	agg.add(a)
	agg.add(b)
	return agg.toOHLC()
}

func sortOHLC(ohlcs []common.OHLC) {
	for i := 0; i < len(ohlcs)-1; i++ {
		for j := i + 1; j < len(ohlcs); j++ {
			if ohlcs[j].Timestamp.Before(ohlcs[i].Timestamp) {
				ohlcs[i], ohlcs[j] = ohlcs[j], ohlcs[i]
			}
		}
	}
}

func (s *Store) Query(query common.DownsampleQuery) []common.OHLC {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var source []common.OHLC

	switch query.Window.Name {
	case "1m":
		source = s.ohlc1m[query.Metric]
	case "5m":
		source = s.ohlc5m[query.Metric]
	case "1h":
		source = s.ohlc1h[query.Metric]
	default:
		return []common.OHLC{}
	}

	filtered := filterByTime(source, query.From, query.To)
	return FillMissingWindows(filtered, query.From, query.To, query.Window)
}

func filterByTime(ohlcs []common.OHLC, from, to time.Time) []common.OHLC {
	var result []common.OHLC
	for _, o := range ohlcs {
		if (o.Timestamp.Equal(from) || o.Timestamp.After(from)) && o.Timestamp.Before(to) {
			result = append(result, o)
		}
	}
	return result
}

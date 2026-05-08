package downsample

import (
	"math"
	"sort"
	"time"

	"github.com/example/timeseries-downsample/pkg/common"
)

type rawAggregator struct {
	windowStart time.Time
	open        *float64
	high        *float64
	low         *float64
	close       *float64
	firstTime   *time.Time
	lastTime    *time.Time
	hasNaN      bool
	hasInf      bool
}

func newRawAggregator(ws time.Time) *rawAggregator {
	return &rawAggregator{windowStart: ws}
}

func (a *rawAggregator) add(p common.DataPoint) {
	v := p.Value
	ts := p.Timestamp

	if math.IsNaN(v) {
		a.hasNaN = true
	}
	if math.IsInf(v, 0) {
		a.hasInf = true
	}

	if a.firstTime == nil || ts.Before(*a.firstTime) {
		t := ts
		a.firstTime = &t
		v2 := v
		a.open = &v2
	}

	if a.lastTime == nil || !ts.Before(*a.lastTime) {
		t := ts
		a.lastTime = &t
		v2 := v
		a.close = &v2
	}

	if a.hasNaN {
		return
	}

	if a.high == nil || v > *a.high {
		v2 := v
		a.high = &v2
	}
	if a.low == nil || v < *a.low {
		v2 := v
		a.low = &v2
	}
}

func (a *rawAggregator) toOHLC() common.OHLC {
	if a.firstTime == nil {
		return common.OHLC{
			Timestamp: a.windowStart,
			Open:      nil,
			High:      nil,
			Low:       nil,
			Close:     nil,
			HasNaN:    false,
			HasInf:    false,
		}
	}

	ohlc := common.OHLC{
		Timestamp: a.windowStart,
		Open:      a.open,
		High:      a.high,
		Low:       a.low,
		Close:     a.close,
		HasNaN:    a.hasNaN,
		HasInf:    a.hasInf,
	}

	if a.hasNaN {
		nan := math.NaN()
		ohlc.High = &nan
		ohlc.Low = &nan
	}

	return ohlc
}

func DownsampleRaw(points []common.DataPoint, window common.WindowSize) []common.OHLC {
	if len(points) == 0 {
		return []common.OHLC{}
	}

	sorted := make([]common.DataPoint, len(points))
	copy(sorted, points)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	aggMap := make(map[time.Time]*rawAggregator)
	var orderedKeys []time.Time

	for _, p := range sorted {
		ws := AlignToWindow(p.Timestamp, window)
		agg, ok := aggMap[ws]
		if !ok {
			agg = newRawAggregator(ws)
			aggMap[ws] = agg
			orderedKeys = append(orderedKeys, ws)
		}
		agg.add(p)
	}

	sort.Slice(orderedKeys, func(i, j int) bool {
		return orderedKeys[i].Before(orderedKeys[j])
	})

	result := make([]common.OHLC, 0, len(orderedKeys))
	for _, k := range orderedKeys {
		result = append(result, aggMap[k].toOHLC())
	}

	return result
}

type ohlcAggregator struct {
	windowStart time.Time
	open        *float64
	high        *float64
	low         *float64
	close       *float64
	firstTime   *time.Time
	lastTime    *time.Time
	hasNaN      bool
	hasInf      bool
}

func newOHLCAggregator(ws time.Time) *ohlcAggregator {
	return &ohlcAggregator{windowStart: ws}
}

func (a *ohlcAggregator) add(o common.OHLC) {
	if o.Open == nil && o.High == nil && o.Low == nil && o.Close == nil {
		return
	}

	ts := o.Timestamp

	if o.HasNaN {
		a.hasNaN = true
	}
	if o.HasInf {
		a.hasInf = true
	}

	if a.firstTime == nil || ts.Before(*a.firstTime) {
		t := ts
		a.firstTime = &t
		if o.Open != nil {
			v := *o.Open
			a.open = &v
		}
	}

	if a.lastTime == nil || !ts.Before(*a.lastTime) {
		t := ts
		a.lastTime = &t
		if o.Close != nil {
			v := *o.Close
			a.close = &v
		}
	}

	if a.hasNaN {
		return
	}

	if o.High != nil {
		if a.high == nil || *o.High > *a.high {
			v := *o.High
			a.high = &v
		}
	}

	if o.Low != nil {
		if a.low == nil || *o.Low < *a.low {
			v := *o.Low
			a.low = &v
		}
	}
}

func (a *ohlcAggregator) toOHLC() common.OHLC {
	if a.firstTime == nil {
		return common.OHLC{
			Timestamp: a.windowStart,
			Open:      nil,
			High:      nil,
			Low:       nil,
			Close:     nil,
			HasNaN:    false,
			HasInf:    false,
		}
	}

	ohlc := common.OHLC{
		Timestamp: a.windowStart,
		Open:      a.open,
		High:      a.high,
		Low:       a.low,
		Close:     a.close,
		HasNaN:    a.hasNaN,
		HasInf:    a.hasInf,
	}

	if a.hasNaN {
		nan := math.NaN()
		ohlc.High = &nan
		ohlc.Low = &nan
	}

	return ohlc
}

func DownsampleOHLC(ohlcs []common.OHLC, window common.WindowSize) []common.OHLC {
	if len(ohlcs) == 0 {
		return []common.OHLC{}
	}

	sorted := make([]common.OHLC, len(ohlcs))
	copy(sorted, ohlcs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	aggMap := make(map[time.Time]*ohlcAggregator)
	var orderedKeys []time.Time

	for _, o := range sorted {
		ws := AlignToWindow(o.Timestamp, window)
		agg, ok := aggMap[ws]
		if !ok {
			agg = newOHLCAggregator(ws)
			aggMap[ws] = agg
			orderedKeys = append(orderedKeys, ws)
		}
		agg.add(o)
	}

	sort.Slice(orderedKeys, func(i, j int) bool {
		return orderedKeys[i].Before(orderedKeys[j])
	})

	result := make([]common.OHLC, 0, len(orderedKeys))
	for _, k := range orderedKeys {
		result = append(result, aggMap[k].toOHLC())
	}

	return result
}

func FillMissingWindows(result []common.OHLC, from, to time.Time, window common.WindowSize) []common.OHLC {
	fromAligned := AlignToWindow(from, window)
	toAligned := AlignToWindow(to, window)

	existing := make(map[time.Time]common.OHLC)
	for _, o := range result {
		existing[o.Timestamp] = o
	}

	var filled []common.OHLC
	for current := fromAligned; !current.After(toAligned); current = current.Add(window.Duration) {
		if o, ok := existing[current]; ok {
			filled = append(filled, o)
		} else {
			filled = append(filled, common.OHLC{
				Timestamp: current,
				Open:      nil,
				High:      nil,
				Low:       nil,
				Close:     nil,
				HasNaN:    false,
				HasInf:    false,
			})
		}
	}

	return filled
}

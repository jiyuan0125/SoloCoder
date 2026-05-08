package common

import (
	"encoding/json"
	"fmt"
	"time"
)

type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type OHLC struct {
	Timestamp time.Time  `json:"timestamp"`
	Open      *float64   `json:"open,omitempty"`
	High      *float64   `json:"high,omitempty"`
	Low       *float64   `json:"low,omitempty"`
	Close     *float64   `json:"close,omitempty"`
	HasNaN    bool       `json:"has_nan,omitempty"`
	HasInf    bool       `json:"has_inf,omitempty"`
}

type WindowSize struct {
	Duration time.Duration
	Name     string
}

var (
	Window1Minute  = WindowSize{Duration: time.Minute, Name: "1m"}
	Window5Minutes = WindowSize{Duration: 5 * time.Minute, Name: "5m"}
	Window1Hour    = WindowSize{Duration: time.Hour, Name: "1h"}
	Window1Day     = WindowSize{Duration: 24 * time.Hour, Name: "1d"}
)

var WindowsByName = map[string]WindowSize{
	"1m": Window1Minute,
	"5m": Window5Minutes,
	"1h": Window1Hour,
	"1d": Window1Day,
}

func ParseWindowSize(name string) (WindowSize, error) {
	if w, ok := WindowsByName[name]; ok {
		return w, nil
	}
	return WindowSize{}, fmt.Errorf("unknown window size: %s", name)
}

type WriteRequest struct {
	Metric string      `json:"metric"`
	Data   []DataPoint `json:"data"`
}

type WriteResponse struct {
	Success bool   `json:"success"`
	Count   int    `json:"count"`
	Message string `json:"message,omitempty"`
}

type DownsampleQuery struct {
	Metric string
	Window WindowSize
	From   time.Time
	To     time.Time
}

type DownsampleResponse struct {
	Success bool       `json:"success"`
	Data    []OHLC     `json:"data"`
	Metric  string     `json:"metric"`
	Window  string     `json:"window"`
	From    time.Time  `json:"from"`
	To      time.Time  `json:"to"`
	Message string     `json:"message,omitempty"`
}

type ConfigResponse struct {
	SupportedWindows []string `json:"supported_windows"`
}

func NewConfigResponse() ConfigResponse {
	windows := make([]string, 0, len(WindowsByName))
	for k := range WindowsByName {
		windows = append(windows, k)
	}
	return ConfigResponse{SupportedWindows: windows}
}

func (o OHLC) MarshalJSON() ([]byte, error) {
	type Alias OHLC
	if o.Open == nil && o.High == nil && o.Low == nil && o.Close == nil {
		return json.Marshal(&struct {
			Timestamp time.Time `json:"timestamp"`
			Open      *float64  `json:"open"`
			High      *float64  `json:"high"`
			Low       *float64  `json:"low"`
			Close     *float64  `json:"close"`
			HasNaN    bool      `json:"has_nan,omitempty"`
			HasInf    bool      `json:"has_inf,omitempty"`
			*Alias
		}{
			Timestamp: o.Timestamp,
			Open:      nil,
			High:      nil,
			Low:       nil,
			Close:     nil,
			Alias:     (*Alias)(&o),
		})
	}
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&o),
	})
}

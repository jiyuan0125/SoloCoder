package api

import (
	"encoding/json"
	"fmt"
	"math"
)

type Float64 float64

func (f Float64) MarshalJSON() ([]byte, error) {
	val := float64(f)
	if math.IsNaN(val) {
		return json.Marshal("NaN")
	}
	if math.IsInf(val, 1) {
		return json.Marshal("+Inf")
	}
	if math.IsInf(val, -1) {
		return json.Marshal("-Inf")
	}
	return json.Marshal(val)
}

func (f *Float64) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		switch s {
		case "NaN":
			*f = Float64(math.NaN())
			return nil
		case "+Inf", "Inf":
			*f = Float64(math.Inf(1))
			return nil
		case "-Inf":
			*f = Float64(math.Inf(-1))
			return nil
		}
	}
	var num float64
	if err := json.Unmarshal(data, &num); err != nil {
		return fmt.Errorf("invalid float64 value: %s", string(data))
	}
	*f = Float64(num)
	return nil
}

func ToFloat64Slice(arr []float64) []Float64 {
	result := make([]Float64, len(arr))
	for i, v := range arr {
		result[i] = Float64(v)
	}
	return result
}

type FitRequest struct {
	Simple   [][]float64 `json:"simple,omitempty"`
	Features [][]float64 `json:"features,omitempty"`
	Target   []float64   `json:"target,omitempty"`
}

type FitResponse struct {
	Success bool      `json:"success"`
	Error   string    `json:"error,omitempty"`
	Slope   Float64   `json:"slope"`
	Intercept Float64 `json:"intercept"`
	Coefficients []Float64 `json:"coefficients,omitempty"`
	R2      Float64   `json:"r2"`
}

type PredictRequest struct {
	SimpleX  *float64   `json:"simple_x,omitempty"`
	Features []float64 `json:"features,omitempty"`
}

type PredictResponse struct {
	Success bool    `json:"success"`
	Error   string  `json:"error,omitempty"`
	Value   Float64 `json:"value"`
}

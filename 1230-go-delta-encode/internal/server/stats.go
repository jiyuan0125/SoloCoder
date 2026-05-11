package server

import (
	"encoding/json"
	"net/http"

	"delta-encode/pkg/api"
	"delta-encode/pkg/delta"
)

type rawStatsRequest struct {
	DataType api.DataType    `json:"data_type"`
	Data     json.RawMessage `json:"data"`
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var raw rawStatsRequest
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	switch raw.DataType {
	case api.DataTypeInt:
		var arr []int64
		if len(raw.Data) > 0 {
			if err := json.Unmarshal(raw.Data, &arr); err != nil {
				writeError(w, http.StatusBadRequest, "Invalid integer array: "+err.Error())
				return
			}
		}
		stats := delta.StatsInt(arr)
		writeJSON(w, http.StatusOK, api.IntStatsResponse{
			AbsoluteSum: stats.AbsoluteSum,
			MaxAbsDelta: stats.MaxAbsDelta,
			ZeroRatio:   stats.ZeroRatio,
			TotalCount:  stats.TotalCount,
			ZeroCount:   stats.ZeroCount,
		})

	case api.DataTypeFloat:
		var arr []float64
		if len(raw.Data) > 0 {
			if err := json.Unmarshal(raw.Data, &arr); err != nil {
				writeError(w, http.StatusBadRequest, "Invalid float array: "+err.Error())
				return
			}
		}
		stats := delta.StatsFloat(arr)
		writeJSON(w, http.StatusOK, api.FloatStatsResponse{
			AbsoluteSum: stats.AbsoluteSum,
			MaxAbsDelta: stats.MaxAbsDelta,
			ZeroRatio:   stats.ZeroRatio,
			TotalCount:  stats.TotalCount,
			ZeroCount:   stats.ZeroCount,
		})

	default:
		writeError(w, http.StatusBadRequest, "Invalid data_type: must be 'int' or 'float'")
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/example/timeseries-downsample/pkg/common"
	"github.com/example/timeseries-downsample/pkg/downsample"
)

type handler struct {
	store *downsample.Store
}

func newHandler(store *downsample.Store) *handler {
	return &handler{store: store}
}

func (h *handler) writeData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.WriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.WriteResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if req.Metric == "" {
		writeJSON(w, http.StatusBadRequest, common.WriteResponse{
			Success: false,
			Message: "metric is required",
		})
		return
	}

	count := h.store.Write(req.Metric, req.Data)
	writeJSON(w, http.StatusOK, common.WriteResponse{
		Success: true,
		Count:   count,
	})
}

func (h *handler) queryDownsample(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metric := r.URL.Query().Get("metric")
	windowName := r.URL.Query().Get("window")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if metric == "" {
		writeJSON(w, http.StatusBadRequest, common.DownsampleResponse{
			Success: false,
			Message: "metric is required",
		})
		return
	}

	window, err := common.ParseWindowSize(windowName)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.DownsampleResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.DownsampleResponse{
			Success: false,
			Message: "invalid from time, expected RFC3339",
		})
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.DownsampleResponse{
			Success: false,
			Message: "invalid to time, expected RFC3339",
		})
		return
	}

	query := common.DownsampleQuery{
		Metric: metric,
		Window: window,
		From:   from,
		To:     to,
	}

	data := h.store.Query(query)
	writeJSON(w, http.StatusOK, common.DownsampleResponse{
		Success: true,
		Data:    data,
		Metric:  metric,
		Window:  window.Name,
		From:    from,
		To:      to,
	})
}

func (h *handler) getConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, common.NewConfigResponse())
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func main() {
	store := downsample.NewStore()
	h := newHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/ts/data", h.writeData)
	mux.HandleFunc("/ts/downsample", h.queryDownsample)
	mux.HandleFunc("/ts/config", h.getConfig)

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	fmt.Printf("server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

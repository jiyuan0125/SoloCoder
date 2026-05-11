package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"serial-system/internal/api"
	"serial-system/internal/serial"
)

func getPort() string {
	if envPort := os.Getenv("SERIAL_PORT"); envPort != "" {
		port, err := strconv.Atoi(envPort)
		if err == nil && port > 0 && port < 65536 {
			return fmt.Sprintf(":%d", port)
		}
	}

	var port int
	flag.IntVar(&port, "port", 8080, "server port (default 8080, or SERIAL_PORT env)")
	flag.Parse()

	return fmt.Sprintf(":%d", port)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

type handler struct {
	gen *serial.Generator
}

func (h *handler) next(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.NextResponse{Success: false, Error: "method not allowed"})
		return
	}
	var req api.NextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.NextResponse{Success: false, Error: "invalid request body"})
		return
	}
	if req.BizType == "" {
		writeJSON(w, http.StatusBadRequest, api.NextResponse{Success: false, Error: "biz_type is required"})
		return
	}
	serialStr, err := h.gen.Next(req.BizType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.NextResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, api.NextResponse{Success: true, Serial: serialStr})
}

func (h *handler) batch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.BatchResponse{Success: false, Error: "method not allowed"})
		return
	}
	var req api.BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.BatchResponse{Success: false, Error: "invalid request body"})
		return
	}
	if req.BizType == "" {
		writeJSON(w, http.StatusBadRequest, api.BatchResponse{Success: false, Error: "biz_type is required"})
		return
	}
	if req.Count <= 0 {
		writeJSON(w, http.StatusBadRequest, api.BatchResponse{Success: false, Error: "count must be positive"})
		return
	}
	start, end, date, err := h.gen.Batch(req.BizType, req.Count)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.BatchResponse{Success: false, Error: err.Error()})
		return
	}
	serials := make([]string, req.Count)
	for i := 0; i < req.Count; i++ {
		serials[i] = fmt.Sprintf("%s-%06d", date, start+int64(i))
	}
	startStr := fmt.Sprintf("%s-%06d", date, start)
	endStr := fmt.Sprintf("%s-%06d", date, end)
	writeJSON(w, http.StatusOK, api.BatchResponse{Success: true, Start: startStr, End: endStr, Serials: serials})
}

func (h *handler) status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.StatusResponse{Success: false, Error: "method not allowed"})
		return
	}
	var req api.StatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.StatusResponse{Success: false, Error: "invalid request body"})
		return
	}
	if req.BizType == "" {
		writeJSON(w, http.StatusBadRequest, api.StatusResponse{Success: false, Error: "biz_type is required"})
		return
	}
	state := h.gen.Status(req.BizType)
	writeJSON(w, http.StatusOK, api.StatusResponse{
		Success:        true,
		BizType:        req.BizType,
		Date:           state.Date,
		CurrentMax:     state.CurrentMax,
		TotalAllocated: state.TotalAllocated,
	})
}

func (h *handler) check(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.CheckResponse{Success: false, Error: "method not allowed"})
		return
	}
	var req api.CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.CheckResponse{Success: false, Error: "invalid request body"})
		return
	}
	if req.BizType == "" {
		writeJSON(w, http.StatusBadRequest, api.CheckResponse{Success: false, Error: "biz_type is required"})
		return
	}
	hasGap, rawGaps, currentMax, total := h.gen.Check(req.BizType)
	gaps := make([]api.Gap, len(rawGaps))
	for i, g := range rawGaps {
		gaps[i] = api.Gap{Start: g.Start, End: g.End}
	}
	writeJSON(w, http.StatusOK, api.CheckResponse{
		Success:        true,
		HasGap:         hasGap,
		Gaps:           gaps,
		CurrentMax:     currentMax,
		TotalAllocated: total,
	})
}

func (h *handler) reset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ResetResponse{Success: false, Error: "method not allowed"})
		return
	}
	var req api.ResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ResetResponse{Success: false, Error: "invalid request body"})
		return
	}
	if req.BizType == "" {
		writeJSON(w, http.StatusBadRequest, api.ResetResponse{Success: false, Error: "biz_type is required"})
		return
	}
	date, err := h.gen.Reset(req.BizType, req.Confirm)
	if err != nil {
		writeJSON(w, http.StatusForbidden, api.ResetResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, api.ResetResponse{Success: true, Date: date})
}

func main() {
	port := getPort()
	gen := serial.NewGenerator()
	h := &handler{gen: gen}

	mux := http.NewServeMux()
	mux.HandleFunc("/next", h.next)
	mux.HandleFunc("/batch", h.batch)
	mux.HandleFunc("/status", h.status)
	mux.HandleFunc("/check", h.check)
	mux.HandleFunc("/reset", h.reset)

	log.Printf("serial server listening on %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

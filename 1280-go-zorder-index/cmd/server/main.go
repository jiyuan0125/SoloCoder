package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/zorder-index/internal/api"
	"github.com/zorder-index/pkg/zorder"
)

func main() {
	port := getPort()

	mux := http.NewServeMux()

	mux.HandleFunc("/encode2d", handleEncode2D)
	mux.HandleFunc("/decode2d", handleDecode2D)
	mux.HandleFunc("/encode3d", handleEncode3D)
	mux.HandleFunc("/decode3d", handleDecode3D)
	mux.HandleFunc("/query", handleQuery)
	mux.HandleFunc("/health", handleHealth)

	addr := ":" + port
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getPort() string {
	portFlag := flag.String("port", "", "Port to listen on")
	flag.Parse()

	if *portFlag != "" {
		return *portFlag
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}

	return "8507"
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleEncode2D(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.Encode2DRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	method := r.URL.Query().Get("method")
	var code zorder.Code
	if method == "lookup" {
		code = zorder.Encode2DLookup(zorder.Point2D{X: req.X, Y: req.Y})
	} else {
		code = zorder.Encode2DBit(zorder.Point2D{X: req.X, Y: req.Y})
		method = "bit"
	}

	resp := api.Encode2DResponse{
		Code:       code,
		Method:     method,
		Coordinate: zorder.Point2D{X: req.X, Y: req.Y},
	}

	sendJSON(w, resp, http.StatusOK)
}

func handleDecode2D(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.Decode2DRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	method := r.URL.Query().Get("method")
	var point zorder.Point2D
	if method == "lookup" {
		point = zorder.Decode2DLookup(req.Code)
	} else {
		point = zorder.Decode2DBit(req.Code)
	}

	resp := api.Decode2DResponse{
		Code:       req.Code,
		Coordinate: point,
	}

	sendJSON(w, resp, http.StatusOK)
}

func handleEncode3D(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.Encode3DRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	code, err := zorder.Encode3D(zorder.Point3D{X: req.X, Y: req.Y, Z: req.Z})
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := api.Encode3DResponse{
		Code:       code,
		Coordinate: zorder.Point3D{X: req.X, Y: req.Y, Z: req.Z},
	}

	sendJSON(w, resp, http.StatusOK)
}

func handleDecode3D(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.Decode3DRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	point := zorder.Decode3D(req.Code)

	resp := api.Decode3DResponse{
		Code:       req.Code,
		Coordinate: point,
	}

	sendJSON(w, resp, http.StatusOK)
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.QueryRangesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rect := zorder.Rect2D{
		Min: zorder.Point2D{X: req.MinX, Y: req.MinY},
		Max: zorder.Point2D{X: req.MaxX, Y: req.MaxY},
	}

	ranges := zorder.QueryRangesOptimized(rect)

	resp := api.QueryRangesResponse{
		Rect:   rect,
		Ranges: ranges,
		Count:  len(ranges),
	}

	sendJSON(w, resp, http.StatusOK)
}

func sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, status int) {
	sendJSON(w, api.ErrorResponse{Error: message}, status)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

var _ = strconv.Itoa
var _ = fmt.Sprintf

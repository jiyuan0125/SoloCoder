package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"geohash/pkg/api"
	"geohash/pkg/geohash"
)

type handler struct{}

func newHandler() *handler {
	return &handler{}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, api.ErrorResponse{Error: err.Error()})
}

func (h *handler) encode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("只支持 POST 方法"))
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	geohashStr, err := geohash.Encode(req.Latitude, req.Longitude, req.Precision)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, api.EncodeResponse{Geohash: geohashStr})
}

func (h *handler) decode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("只支持 POST 方法"))
		return
	}

	var req api.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	lat, lng, err := geohash.Decode(req.Geohash)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, api.DecodeResponse{Latitude: lat, Longitude: lng})
}

func (h *handler) boundingBox(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("只支持 POST 方法"))
		return
	}

	var req api.BoundingBoxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	bbox, err := geohash.BoundingBoxOf(req.Geohash)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, api.BoundingBoxResponse{
		MinLat: bbox.MinLat,
		MaxLat: bbox.MaxLat,
		MinLng: bbox.MinLng,
		MaxLng: bbox.MaxLng,
	})
}

func (h *handler) neighbors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("只支持 POST 方法"))
		return
	}

	var req api.NeighborsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	neighbors, err := geohash.Neighbors(req.Geohash)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, api.NeighborsResponse{
		North:     neighbors[geohash.North],
		Northeast: neighbors[geohash.Northeast],
		East:      neighbors[geohash.East],
		Southeast: neighbors[geohash.Southeast],
		South:     neighbors[geohash.South],
		Southwest: neighbors[geohash.Southwest],
		West:      neighbors[geohash.West],
		Northwest: neighbors[geohash.Northwest],
	})
}

func (h *handler) proximitySearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("只支持 POST 方法"))
		return
	}

	var req api.ProximitySearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	geohashes, err := geohash.ProximitySearch(req.Latitude, req.Longitude, req.RadiusKm)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, api.ProximitySearchResponse{
		Geohashes: geohashes,
		Count:     len(geohashes),
	})
}

func getPort() string {
	port := "8080"

	if envPort := os.Getenv("GEOHASH_PORT"); envPort != "" {
		port = envPort
	}

	var flagPort string
	flag.StringVar(&flagPort, "port", "", "服务端监听端口")
	flag.Parse()

	if flagPort != "" {
		port = flagPort
	}

	return port
}

func main() {
	port := getPort()

	h := newHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/encode", h.encode)
	mux.HandleFunc("/decode", h.decode)
	mux.HandleFunc("/bbox", h.boundingBox)
	mux.HandleFunc("/neighbors", h.neighbors)
	mux.HandleFunc("/search", h.proximitySearch)

	addr := ":" + strings.TrimPrefix(port, ":")

	fmt.Printf("Geohash 服务端启动中，监听端口 %s...\n", port)
	fmt.Println("可用接口:")
	fmt.Println("  POST /encode - 经纬度编码为 Geohash")
	fmt.Println("  POST /decode - Geohash 解码为经纬度")
	fmt.Println("  POST /bbox   - 获取 Geohash 边界矩形")
	fmt.Println("  POST /neighbors - 获取 Geohash 邻居")
	fmt.Println("  POST /search - 邻近搜索")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "服务端启动失败: %v\n", err)
		os.Exit(1)
	}
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"httpcacheproxy/pkg/cache"
	"httpcacheproxy/pkg/common"
)

var (
	listenAddr  = flag.String("addr", ":8080", "Listen address")
	defaultTTL  = flag.Int("ttl", 60, "Default cache TTL in seconds")
	maxCacheSize = flag.Int64("size", 100*1024*1024, "Max cache size in bytes")
)

type Server struct {
	proxy *cache.Proxy
}

func NewServer(config *common.ServerConfig) *Server {
	store := cache.NewCacheStore(config.MaxCacheSize, config.DefaultTTL)
	return &Server{
		proxy: cache.NewProxy(store),
	}
}

func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse request: %v", err), http.StatusBadRequest)
		return
	}

	result, err := s.proxy.Handle(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Proxy error: %v", err), http.StatusInternalServerError)
		return
	}

	resp := common.ProxyResponse{
		StatusCode: result.StatusCode,
		Header:     result.Header,
		Body:       result.Body,
		FromCache:  result.FromCache,
		CacheAge:   result.CacheAge,
	}

	for k, values := range result.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}

	if result.FromCache {
		w.Header().Set("X-Cache", "HIT")
	} else {
		w.Header().Set("X-Cache", "MISS")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	flag.Parse()

	config := &common.ServerConfig{
		ListenAddr:  *listenAddr,
		DefaultTTL:  *defaultTTL,
		MaxCacheSize: *maxCacheSize,
	}

	server := NewServer(config)

	http.HandleFunc("/proxy", server.handleProxy)

	log.Printf("HTTP Cache Proxy Server listening on %s", config.ListenAddr)
	log.Printf("Default TTL: %d seconds", config.DefaultTTL)
	log.Printf("Max cache size: %d bytes", config.MaxCacheSize)

	if err := http.ListenAndServe(config.ListenAddr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

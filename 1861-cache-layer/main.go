package main

import (
	"cache-layer/api"
	"cache-layer/cache"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	defaultPort     = "8308"
	defaultCapacity = 1000
	defaultTTL      = 3600
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	capacity := defaultCapacity
	if capStr := os.Getenv("CAPACITY"); capStr != "" {
		if c, err := strconv.Atoi(capStr); err == nil && c >= 0 {
			capacity = c
		}
	}

	ttlDur := time.Duration(defaultTTL) * time.Second
	if ttlStr := os.Getenv("DEFAULT_TTL"); ttlStr != "" {
		if t, err := strconv.ParseInt(ttlStr, 10, 64); err == nil && t >= 0 {
			ttlDur = time.Duration(t) * time.Second
		}
	}

	c := cache.NewLRUCache(capacity, ttlDur)
	handler := api.NewHandler(c)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	log.Printf("cache service starting on port %s (capacity=%d, default_ttl=%s)", port, capacity, ttlDur)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

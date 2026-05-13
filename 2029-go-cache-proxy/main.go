package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"cache-proxy/proxy"
	"cache-proxy/storage"
)

type Config struct {
	Port          int
	CacheDir      string
	DefaultTTL    time.Duration
	MaxCacheSize  int64
	TTLRules      map[string]time.Duration
}

func main() {
	config := parseFlags()

	store, err := storage.NewStorage(config.CacheDir, config.DefaultTTL, config.MaxCacheSize)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	cacheProxy := proxy.NewCacheProxy(store)
	cacheProxy.SetTTLRules(config.TTLRules)

	mux := http.NewServeMux()

	mux.HandleFunc("/", cacheProxy.HandleCache)
	mux.HandleFunc("/cache/delete", cacheProxy.HandleDeleteCache)
	mux.HandleFunc("/stats", cacheProxy.HandleStats)
	mux.HandleFunc("/health", cacheProxy.HandleHealth)

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(config.Port),
		Handler:      mux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Starting cache proxy server on port %d...", config.Port)
		log.Printf("Cache directory: %s", config.CacheDir)
		log.Printf("Default TTL: %v", config.DefaultTTL)
		log.Printf("Max cache size: %d bytes", config.MaxCacheSize)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")
}

func parseFlags() *Config {
	config := &Config{
		TTLRules: make(map[string]time.Duration),
	}

	port := flag.Int("port", 9090, "Server port")
	cacheDir := flag.String("cache-dir", "", "Cache directory (default: ./cache)")
	defaultTTL := flag.Duration("default-ttl", 5*time.Minute, "Default cache TTL")
	maxCacheSize := flag.String("max-cache-size", "1GB", "Max cache size (e.g., 1GB, 512MB)")
	ttlRules := flag.String("ttl-rules", "", "TTL rules (format: pattern1=ttl1,pattern2=ttl2)")

	flag.Parse()

	config.Port = *port
	config.DefaultTTL = *defaultTTL

	if *cacheDir == "" {
		cwd, _ := os.Getwd()
		config.CacheDir = filepath.Join(cwd, "cache")
	} else {
		config.CacheDir = *cacheDir
	}

	size, err := parseSize(*maxCacheSize)
	if err != nil {
		log.Fatalf("Invalid max-cache-size: %v", err)
	}
	config.MaxCacheSize = size

	if *ttlRules != "" {
		rules, err := parseTTLRules(*ttlRules)
		if err != nil {
			log.Fatalf("Invalid ttl-rules: %v", err)
		}
		config.TTLRules = rules
	}

	return config
}

func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	multiplier := int64(1)

	switch {
	case strings.HasSuffix(sizeStr, "GB"):
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	case strings.HasSuffix(sizeStr, "MB"):
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	case strings.HasSuffix(sizeStr, "KB"):
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	case strings.HasSuffix(sizeStr, "B"):
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	}

	value, err := strconv.ParseInt(strings.TrimSpace(sizeStr), 10, 64)
	if err != nil {
		return 0, err
	}

	return value * multiplier, nil
}

func parseTTLRules(rulesStr string) (map[string]time.Duration, error) {
	rules := make(map[string]time.Duration)
	pairs := strings.Split(rulesStr, ",")

	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid TTL rule format: %s", pair)
		}
		pattern := strings.TrimSpace(parts[0])
		ttlStr := strings.TrimSpace(parts[1])

		ttl, err := time.ParseDuration(ttlStr)
		if err != nil {
			return nil, err
		}

		rules[pattern] = ttl
	}

	return rules, nil
}

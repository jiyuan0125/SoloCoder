package main

import (
	"dns-cache-service/common"
	"dns-cache-service/dnscache"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	port := common.DefaultPort
	if envPort := os.Getenv("DNS_CACHE_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	config := dnscache.DefaultConfig()
	if multiplier := os.Getenv("NXDOMAIN_TTL_MULTIPLIER"); multiplier != "" {
		if m, err := strconv.ParseFloat(multiplier, 64); err == nil {
			config.NXDomainTTLMultiplier = m
		}
	}
	if ttl := os.Getenv("DEFAULT_TTL"); ttl != "" {
		if t, err := strconv.Atoi(ttl); err == nil {
			config.DefaultTTL = t
		}
	}

	service := dnscache.NewService(config)
	service.StartPreloadAsync(common.DefaultPreloadDomains, []common.RecordType{common.TypeA, common.TypeAAAA})

	handler := NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("/query", handler.Query)
	mux.HandleFunc("/refresh", handler.Refresh)
	mux.HandleFunc("/stats", handler.Stats)
	mux.HandleFunc("/preload", handler.Preload)
	mux.HandleFunc("/entries", handler.Entries)

	log.Printf("DNS Cache Server starting on port %d...", port)
	log.Printf("NXDomain TTL multiplier: %.2f", config.NXDomainTTLMultiplier)
	log.Printf("Default TTL: %d seconds", config.DefaultTTL)

	if err := http.ListenAndServe(":"+strconv.Itoa(port), mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

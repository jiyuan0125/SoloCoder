package proxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cache-proxy/storage"
)

type CacheProxy struct {
	storage    *storage.Storage
	httpClient *http.Client
	ttlRules   map[string]time.Duration
}

func NewCacheProxy(s *storage.Storage) *CacheProxy {
	return &CacheProxy{
		storage: s,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		ttlRules: make(map[string]time.Duration),
	}
}

func (p *CacheProxy) SetTTLRules(rules map[string]time.Duration) {
	if rules == nil {
		rules = make(map[string]time.Duration)
	}
	p.ttlRules = rules
}

func (p *CacheProxy) getTTLForURL(targetURL string) time.Duration {
	for pattern, ttl := range p.ttlRules {
		if strings.Contains(targetURL, pattern) {
			return ttl
		}
	}
	return 0
}

func (p *CacheProxy) HandleCache(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}

	if _, err := url.Parse(targetURL); err != nil {
		http.Error(w, "Invalid 'url' parameter", http.StatusBadRequest)
		return
	}

	entry, isExpired, cacheErr := p.storage.GetWithStaleCheck(targetURL)

	if cacheErr == nil && !isExpired {
		p.storage.RecordHit(targetURL, true)
		p.storage.UpdateLastAccess(targetURL)
		p.serveFromCache(w, entry, false)
		return
	}

	resp, err := p.fetchFromUpstream(targetURL, r)
	if err != nil {
		if cacheErr == nil && isExpired {
			p.storage.RecordHit(targetURL, true)
			p.storage.UpdateLastAccess(targetURL)
			p.serveFromCache(w, entry, true)
			return
		}
		p.storage.RecordHit(targetURL, false)
		http.Error(w, fmt.Sprintf("Upstream error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if !isSuccessStatus(resp.StatusCode) && cacheErr == nil && isExpired {
		p.storage.RecordHit(targetURL, true)
		p.storage.UpdateLastAccess(targetURL)
		p.serveFromCache(w, entry, true)
		return
	}

	p.storage.RecordHit(targetURL, false)
	p.cacheResponse(targetURL, resp, w)
}

func (p *CacheProxy) cacheResponse(targetURL string, resp *http.Response, w http.ResponseWriter) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	ttl := p.getTTLForURL(targetURL)

	_, err = p.storage.Put(targetURL, bytes.NewReader(bodyBytes), ttl, resp.StatusCode, contentType)
	if err != nil {
		http.Error(w, "Failed to cache response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(resp.StatusCode)
	w.Write(bodyBytes)
}

func (p *CacheProxy) serveFromCache(w http.ResponseWriter, entry *storage.CacheEntry, isStale bool) {
	w.Header().Set("Content-Type", entry.ContentType)
	w.Header().Set("X-Cache", "HIT")
	if isStale {
		w.Header().Set("X-Stale", "true")
	}

	body, err := p.storage.ReadBody(entry)
	if err != nil {
		http.Error(w, "Failed to read cache", http.StatusInternalServerError)
		return
	}
	defer body.Close()

	w.WriteHeader(entry.StatusCode)
	io.Copy(w, body)
}

func (p *CacheProxy) fetchFromUpstream(targetURL string, originalReq *http.Request) (*http.Response, error) {
	req, err := http.NewRequest(originalReq.Method, targetURL, originalReq.Body)
	if err != nil {
		return nil, err
	}

	for key, values := range originalReq.Header {
		if key == "Host" {
			continue
		}
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}

	return p.httpClient.Do(req)
}

func (p *CacheProxy) HandleDeleteCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}

	err := p.storage.Delete(targetURL)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "Cache entry not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to delete cache: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Cache entry deleted",
	})
}

func (p *CacheProxy) HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 7
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 30 {
			days = d
		}
	}

	stats := p.storage.GetStats(days)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (p *CacheProxy) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "cache-proxy",
	})
}

func isSuccessStatus(code int) bool {
	return code >= 200 && code < 300
}

package common

import "net/http"

type ProxyRequest struct {
	Method string              `json:"method"`
	URL    string              `json:"url"`
	Header http.Header         `json:"header"`
	Body   []byte              `json:"body,omitempty"`
}

type ProxyResponse struct {
	StatusCode int               `json:"status_code"`
	Header     http.Header       `json:"header"`
	Body       []byte            `json:"body,omitempty"`
	FromCache  bool              `json:"from_cache"`
	CacheAge   int               `json:"cache_age,omitempty"`
}

type ServerConfig struct {
	ListenAddr string
	DefaultTTL int
	MaxCacheSize int64
}

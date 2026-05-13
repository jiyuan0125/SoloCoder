package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"api-gateway/backend"
	"api-gateway/logger"
	"api-gateway/router"
)

const DefaultTimeout = 30 * time.Second

type Proxy struct {
	router        *router.Router
	backendMgr    *backend.BackendManager
	accessLogger  *logger.AccessLogger
}

func NewProxy(r *router.Router, bm *backend.BackendManager, al *logger.AccessLogger) *Proxy {
	return &Proxy{
		router:       r,
		backendMgr:   bm,
		accessLogger: al,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	matchResult := p.router.Match(r.URL.Path)
	if !matchResult.Matched {
		http.Error(w, "Not Found", http.StatusNotFound)
		p.logAccess(r, "", 0, http.StatusNotFound, startTime)
		return
	}

	route := matchResult.Route
	backendKey := fmt.Sprintf("%s:%d", route.Backend.Host, route.Backend.Port)

	backendInst := p.backendMgr.GetBackend(route.Backend.Host, route.Backend.Port)
	if backendInst == nil {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		p.logAccess(r, backendKey, route.Backend.Port, http.StatusServiceUnavailable, startTime)
		return
	}

	if !backendInst.CanReceive() {
		if !backendInst.TryRecover() {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			p.logAccess(r, backendKey, route.Backend.Port, http.StatusServiceUnavailable, startTime)
			return
		}
	}

	statusCode := p.forwardRequest(w, r, route, backendInst)

	p.logAccess(r, route.Backend.Host, route.Backend.Port, statusCode, startTime)
}

func (p *Proxy) forwardRequest(w http.ResponseWriter, r *http.Request, route *router.Route, b *backend.Backend) int {
	timeout := DefaultTimeout
	if route.Timeout > 0 {
		timeout = time.Duration(route.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		b.RecordFailure(true)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return http.StatusInternalServerError
	}
	r.Body.Close()

	backendURL := fmt.Sprintf("http://%s:%d%s", b.Host, b.Port, r.URL.Path)

	if r.URL.RawQuery != "" {
		backendURL = fmt.Sprintf("%s?%s", backendURL, r.URL.RawQuery)
	}

	proxyReq, err := http.NewRequestWithContext(ctx, r.Method, backendURL, io.NopCloser(strings.NewReader(string(reqBody))))
	if err != nil {
		b.RecordFailure(true)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return http.StatusInternalServerError
	}

	proxyReq.Header = FilterHeaders(r.Header)

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		b.RecordFailure(true)
		http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
		return http.StatusGatewayTimeout
	}
	defer resp.Body.Close()

	is5xx := resp.StatusCode >= 500 && resp.StatusCode < 600
	if is5xx {
		b.RecordFailure(true)
	} else {
		b.RecordSuccess()
	}

	for key, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)

	return resp.StatusCode
}

func (p *Proxy) logAccess(r *http.Request, backendHost string, backendPort int, statusCode int, startTime time.Time) {
	duration := time.Since(startTime)
	clientIP := getClientIP(r)

	p.accessLogger.Record(clientIP, r.URL.Path, backendHost, backendPort, statusCode, duration)
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

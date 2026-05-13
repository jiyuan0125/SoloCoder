package proxy

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/example/apigateway/internal/model"
)

type FallbackResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func writeFallbackResponse(w http.ResponseWriter, code int, errMsg, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(FallbackResponse{
		Error:   errMsg,
		Message: message,
	})
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func hopHeaders() []string {
	return []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"TE",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}
}

func removeHopHeaders(h http.Header) {
	for _, hh := range hopHeaders() {
		h.Del(hh)
	}
	if c := h.Get("Connection"); c != "" {
		for _, part := range strings.Split(c, ",") {
			if part = strings.TrimSpace(part); part != "" {
				h.Del(part)
			}
		}
	}
}

func Forward(w http.ResponseWriter, r *http.Request, route *model.RouteRule) {
	targetURL, err := url.Parse(route.TargetURL)
	if err != nil {
		writeFallbackResponse(w, http.StatusBadGateway, "bad_gateway", "invalid target URL")
		return
	}

	targetPath := targetURL.Path
	if strings.HasSuffix(targetPath, "/") {
		targetPath = targetPath[:len(targetPath)-1]
	}

	reqPath := r.URL.Path
	if route.MatchType == model.MatchTypePrefix {
		reqPath = strings.TrimPrefix(reqPath, route.Path)
	}
	if reqPath == "" {
		reqPath = "/"
	}
	if !strings.HasPrefix(reqPath, "/") {
		reqPath = "/" + reqPath
	}

	newURL := *r.URL
	newURL.Scheme = targetURL.Scheme
	newURL.Host = targetURL.Host
	newURL.Path = targetPath + reqPath
	if newURL.RawQuery != "" {
		newURL.RawQuery = r.URL.RawQuery
	}

	timeout := route.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	outReq, err := http.NewRequestWithContext(ctx, r.Method, newURL.String(), r.Body)
	if err != nil {
		writeFallbackResponse(w, http.StatusBadGateway, "bad_gateway", "failed to create request")
		return
	}

	copyHeader(outReq.Header, r.Header)
	removeHopHeaders(outReq.Header)
	outReq.Host = targetURL.Host

	if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if prior, ok := outReq.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		outReq.Header.Set("X-Forwarded-For", clientIP)
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	resp, err := client.Do(outReq)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			writeFallbackResponse(w, http.StatusGatewayTimeout, "gateway_timeout", "backend service timeout")
			return
		}
		writeFallbackResponse(w, http.StatusBadGateway, "bad_gateway", "backend service unavailable")
		return
	}
	defer resp.Body.Close()

	removeHopHeaders(resp.Header)
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Proxy struct {
	config   *Config
	router   *Router
	balancer *Balancer
}

func NewProxy(config *Config) (*Proxy, error) {
	router := NewRouter(config.Routes)
	balancer := NewBalancer()

	return &Proxy{
		config:   config,
		router:   router,
		balancer: balancer,
	}, nil
}

func (p *Proxy) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handler)

	server := &http.Server{
		Addr:    p.config.ListenAddr,
		Handler: mux,
	}

	return server.ListenAndServe()
}

func (p *Proxy) handler(w http.ResponseWriter, r *http.Request) {
	backends, _ := p.router.Match(r)
	if backends == nil {
		http.NotFound(w, r)
		return
	}

	backend := p.balancer.Select(backends)
	if backend == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(w, "503 Service Unavailable: No healthy backend available")
		return
	}

	if !backend.Breaker.Acquire() {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(w, "503 Service Unavailable: Backend is in half-open state")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	r = r.WithContext(ctx)

	resp, err := p.forwardRequest(r, backend)
	if err != nil {
		backend.Breaker.RecordFailure()
		if ctx.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusGatewayTimeout)
			fmt.Fprintln(w, "504 Gateway Timeout")
		} else {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintln(w, "502 Bad Gateway")
		}
		log.Printf("Error forwarding to %s: %v", backend.Addr, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		backend.Breaker.RecordFailure()
	} else {
		backend.Breaker.RecordSuccess()
	}

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (p *Proxy) forwardRequest(r *http.Request, backend *Backend) (*http.Response, error) {
	targetURL := "http://" + backend.Addr + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		return nil, err
	}

	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	return client.Do(req)
}

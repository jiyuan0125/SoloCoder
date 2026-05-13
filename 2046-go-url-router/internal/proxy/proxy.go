package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"router/internal/router"
)

type Proxy struct {
	matcher *router.Matcher
	client  *http.Client
}

func New(matcher *router.Matcher) *Proxy {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &Proxy{
		matcher: matcher,
		client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route, pathMatch, methodMatch := p.matcher.Match(r.Method, r.URL.Path)

	if !pathMatch {
		http.NotFound(w, r)
		return
	}

	if !methodMatch {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	targetURL, err := buildTargetURL(route.Target, r.URL)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	outReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	copyHeaders(outReq.Header, r.Header)
	outReq.Host = r.Host
	outReq.ContentLength = r.ContentLength
	outReq.TransferEncoding = r.TransferEncoding
	if r.Close {
		outReq.Close = true
	}

	resp, err := p.client.Do(outReq)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func buildTargetURL(target string, original *url.URL) (string, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return "", err
	}

	if targetURL.Scheme == "" {
		targetURL.Scheme = "http"
	}

	if targetURL.Path == "" {
		targetURL.Path = "/"
	}

	remaining := original.Path
	if !strings.HasSuffix(targetURL.Path, "/") {
		targetURL.Path += "/"
	}
	if strings.HasPrefix(remaining, "/") {
		remaining = remaining[1:]
	}
	targetURL.Path += remaining

	if original.RawQuery != "" {
		targetURL.RawQuery = original.RawQuery
	}

	return targetURL.String(), nil
}

func copyHeaders(dst, src http.Header) {
	for k, vs := range src {
		if isHopByHopHeader(k) {
			continue
		}
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

func isHopByHopHeader(name string) bool {
	name = strings.ToLower(name)
	hop := []string{
		"connection",
		"keep-alive",
		"proxy-authenticate",
		"proxy-authorization",
		"te",
		"trailers",
		"transfer-encoding",
		"upgrade",
	}
	for _, h := range hop {
		if name == h {
			return true
		}
	}
	return false
}

func (p *Proxy) SetClient(client *http.Client) {
	if client != nil {
		p.client = client
	}
}

func (p *Proxy) Client() *http.Client {
	return p.client
}

func (p *Proxy) Matcher() *router.Matcher {
	return p.matcher
}

func (p *Proxy) BuildTargetURL(target string, original *url.URL) (string, error) {
	return buildTargetURL(target, original)
}

func (p *Proxy) IsHopByHopHeader(name string) bool {
	return isHopByHopHeader(name)
}

func (p *Proxy) CopyHeaders(dst, src http.Header) {
	copyHeaders(dst, src)
}

func (p *Proxy) NewOutRequest(ctx context.Context, method, targetURL string, body io.Reader, r *http.Request) (*http.Request, error) {
	outReq, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	copyHeaders(outReq.Header, r.Header)
	outReq.Host = r.Host
	outReq.ContentLength = r.ContentLength
	outReq.TransferEncoding = r.TransferEncoding
	if r.Close {
		outReq.Close = true
	}
	return outReq, nil
}

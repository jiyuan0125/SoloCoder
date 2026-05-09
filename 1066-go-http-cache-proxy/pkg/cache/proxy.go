package cache

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"httpcacheproxy/pkg/common"
)

type Proxy struct {
	store  *CacheStore
	client *http.Client
}

type ProxyResult struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	FromCache  bool
	CacheAge   int
}

func NewProxy(store *CacheStore) *Proxy {
	return &Proxy{
		store:  store,
		client: &http.Client{},
	}
}

func (p *Proxy) Handle(req *common.ProxyRequest) (*ProxyResult, error) {
	httpReq, err := buildHTTPRequest(req)
	if err != nil {
		return nil, err
	}

	cacheKey := BuildCacheKey(req.Method, req.URL)

	if req.Method == http.MethodGet || req.Method == http.MethodHead {
		cachedEntry, _ := p.findMatchingEntry(cacheKey, httpReq.Header)
		if cachedEntry != nil {
			return p.handleCachedEntry(cachedEntry, httpReq)
		}
	}

	return p.forwardRequest(httpReq, req.Method, req.URL)
}

func (p *Proxy) findMatchingEntry(cacheKey string, reqHeader http.Header) (*CacheEntry, string) {
	entries := p.store.GetAllVaryEntries(cacheKey)
	if entries == nil {
		return nil, ""
	}

	for varyKey, entry := range entries {
		if MatchesVary(entry.Header, reqHeader) {
			return entry, varyKey
		}
	}

	return nil, ""
}

func (p *Proxy) handleCachedEntry(entry *CacheEntry, req *http.Request) (*ProxyResult, error) {
	ccResp := ParseCacheControl(entry.Header)

	if ccResp.MustRevalidateOnUse() {
		validationReq := cloneRequest(req)
		if entry.ETag != nil {
			validationReq.Header.Set("If-None-Match", entry.ETag.String())
		}
		if !entry.LastModified.IsZero() {
			validationReq.Header.Set("If-Modified-Since", entry.LastModified.UTC().Format(http.TimeFormat))
		}

		validationResp, err := p.client.Do(validationReq)
		if err == nil && validationResp.StatusCode == http.StatusNotModified {
			validationResp.Body.Close()

			p.store.Update(entry.Key, entry.VaryKey, validationResp, time.Now(), time.Now())

			age := entry.CalculateAge()
			return &ProxyResult{
				StatusCode: entry.StatusCode,
				Header:     p.prepareResponseHeader(entry.Header, age),
				Body:       entry.Body,
				FromCache:  true,
				CacheAge:   age,
			}, nil
		}

		if err == nil {
			return p.processResponse(validationResp, req.Method, req.URL.String())
		}
	}

	age := entry.CalculateAge()
	return &ProxyResult{
		StatusCode: entry.StatusCode,
		Header:     p.prepareResponseHeader(entry.Header, age),
		Body:       entry.Body,
		FromCache:  true,
		CacheAge:   age,
	}, nil
}

func (p *Proxy) forwardRequest(req *http.Request, method, url string) (*ProxyResult, error) {
	requestTime := time.Now()

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return p.processResponse(resp, method, url, requestTime)
}

func (p *Proxy) processResponse(resp *http.Response, method, url string, requestTime ...time.Time) (*ProxyResult, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseTime := time.Now()
	reqTime := responseTime
	if len(requestTime) > 0 {
		reqTime = requestTime[0]
	}

	if (method == http.MethodGet || method == http.MethodHead) && resp.StatusCode == http.StatusOK {
		cacheKey := BuildCacheKey(method, url)
		varyFields := ParseVary(resp.Header)

		dummyHeader := make(http.Header)
		for _, field := range varyFields {
			dummyHeader[field] = resp.Header.Values(field)
		}

		varyKey := GenerateVaryKey(dummyHeader, varyFields)

		p.store.Set(cacheKey, varyKey, resp, body, reqTime, responseTime)
	}

	return &ProxyResult{
		StatusCode: resp.StatusCode,
		Header:     cloneHeader(resp.Header),
		Body:       body,
		FromCache:  false,
	}, nil
}

func (p *Proxy) prepareResponseHeader(header http.Header, age int) http.Header {
	result := cloneHeader(header)
	result.Set("Age", strconv.Itoa(age))
	return result
}

func buildHTTPRequest(req *common.ProxyRequest) (*http.Request, error) {
	var body io.Reader
	if len(req.Body) > 0 {
		body = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, values := range req.Header {
		for _, v := range values {
			httpReq.Header.Add(k, v)
		}
	}

	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"TE",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	for _, h := range hopByHopHeaders {
		httpReq.Header.Del(h)
	}

	return httpReq, nil
}

func cloneRequest(req *http.Request) *http.Request {
	cloned := req.Clone(req.Context())
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		cloned.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}
	return cloned
}

func cloneHeader(h http.Header) http.Header {
	result := make(http.Header)
	for k, v := range h {
		result[k] = append([]string(nil), v...)
	}
	return result
}

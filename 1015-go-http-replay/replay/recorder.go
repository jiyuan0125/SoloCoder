package replay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"http-replay/api"
)

type Recorder struct {
	sessionID     string
	targetURL     *url.URL
	proxy         *httputil.ReverseProxy
	server        *http.Server
	requests      []api.RecordedReq
	variableRules []api.VariableRule
	mu            sync.Mutex
	started       time.Time
	lastReqTime   time.Time
	stopped       bool
	listenAddr    string
	onStop        func(*api.RecordedSessionData)
}

func NewRecorder(targetURL string, variableRules []api.VariableRule, listenAddr string, onStop func(*api.RecordedSessionData)) (*Recorder, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL: %w", err)
	}

	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	rec := &Recorder{
		sessionID:     sessionID,
		targetURL:     parsedURL,
		variableRules: variableRules,
		listenAddr:    listenAddr,
		started:       time.Now(),
		requests:      make([]api.RecordedReq, 0),
		onStop:        onStop,
	}

	rec.proxy = httputil.NewSingleHostReverseProxy(parsedURL)
	rec.proxy.ModifyResponse = rec.captureResponse

	return rec, nil
}

func (r *Recorder) SessionID() string {
	return r.sessionID
}

func (r *Recorder) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.handleRequest)

	r.server = &http.Server{
		Addr:    r.listenAddr,
		Handler: mux,
	}

	go func() {
		_ = r.server.ListenAndServe()
	}()

	time.Sleep(100 * time.Millisecond)
	return nil
}

func (r *Recorder) Stop() (*api.RecordedSessionData, error) {
	r.mu.Lock()
	if r.stopped {
		r.mu.Unlock()
		return nil, nil
	}
	r.stopped = true
	r.mu.Unlock()

	if r.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = r.server.Shutdown(ctx)
	}

	endTime := time.Now()
	for i := 1; i < len(r.requests); i++ {
		r.requests[i].DelayFromPrevMs = r.requests[i].Timestamp.Sub(r.requests[i-1].Timestamp).Milliseconds()
	}

	data := &api.RecordedSessionData{
		Meta: api.SessionMeta{
			SessionID:    r.sessionID,
			TargetURL:    r.targetURL.String(),
			StartTime:    r.started,
			EndTime:      endTime,
			RequestCount: len(r.requests),
		},
		Requests: r.requests,
	}

	if r.onStop != nil {
		r.onStop(data)
	}

	return data, nil
}

func (r *Recorder) handleRequest(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now()

	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewReader(reqBody))

	headers := make(map[string]string)
	for k, v := range req.Header {
		if strings.ToLower(k) != "content-length" {
			headers[k] = strings.Join(v, ", ")
		}
	}

	fullURL := r.targetURL.String() + req.URL.RequestURI()

	recorded := api.RecordedReq{
		Index:     len(r.requests),
		Timestamp: startTime,
		Method:    req.Method,
		URL:       fullURL,
		Path:      req.URL.Path,
		Query:     req.URL.RawQuery,
		Headers:   headers,
		Body:      string(reqBody),
	}

	extractedVars := r.extractVariables(req, reqBody)
	if len(extractedVars) > 0 {
		recorded.ExtractedVars = extractedVars
	}

	wrapped := &responseCaptureWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	r.proxy.ServeHTTP(wrapped, req)

	recorded.ResponseStatus = wrapped.statusCode
	recorded.ResponseHeaders = make(map[string]string)
	for k, v := range wrapped.Header() {
		recorded.ResponseHeaders[k] = strings.Join(v, ", ")
	}
	recorded.ResponseBody = string(wrapped.body)
	recorded.RequestDuration = time.Since(startTime).Milliseconds()

	r.mu.Lock()
	if !r.stopped {
		r.requests = append(r.requests, recorded)
		r.lastReqTime = startTime
	}
	r.mu.Unlock()
}

func (r *Recorder) captureResponse(resp *http.Response) error {
	if resp == nil || resp.Body == nil {
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	resp.Body = ioutil.NopCloser(bytes.NewReader(body))
	return nil
}

func (r *Recorder) extractVariables(req *http.Request, body []byte) map[string]string {
	if len(r.variableRules) == 0 {
		return nil
	}

	vars := make(map[string]string)

	for _, rule := range r.variableRules {
		switch rule.Type {
		case "header":
			if rule.HeaderName != "" {
				if val := req.Header.Get(rule.HeaderName); val != "" {
					vars[rule.Name] = val
				}
			}
		case "jsonpath":
			if rule.JSONPath != "" && len(body) > 0 {
				if val, ok := extractJSONPath(body, rule.JSONPath); ok {
					vars[rule.Name] = val
				}
			}
		}
	}

	if len(vars) == 0 {
		return nil
	}
	return vars
}

type responseCaptureWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (w *responseCaptureWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseCaptureWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func extractJSONPath(data []byte, path string) (string, bool) {
	parts := strings.Split(path, ".")
	if len(parts) == 0 || parts[0] != "$" {
		return "", false
	}

	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return "", false
	}

	current := obj
	for i := 1; i < len(parts); i++ {
		key := parts[i]
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[key]; exists {
				current = val
			} else {
				return "", false
			}
		} else {
			return "", false
		}
	}

	switch v := current.(type) {
	case string:
		return v, true
	case float64:
		return fmt.Sprintf("%v", v), true
	case bool:
		return fmt.Sprintf("%v", v), true
	case nil:
		return "", true
	default:
		return "", false
	}
}

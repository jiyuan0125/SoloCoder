package replay

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"http-replay/api"
)

const (
	ModeStrict = "strict"
	ModeFast   = "fast"
)

type Player struct {
	storage *Storage
}

func NewPlayer(storage *Storage) *Player {
	return &Player{storage: storage}
}

type PlayOptions struct {
	TargetURL      string
	Mode           string
	Concurrency    int
	ReplaceRules   []api.ReplaceRule
	VariableValues []api.VariableValue
}

func (p *Player) Play(sessionID string, opts PlayOptions) (*api.PlayResponse, error) {
	session, err := p.storage.LoadSession(sessionID)
	if err != nil {
		return nil, err
	}

	if len(session.Requests) == 0 {
		return &api.PlayResponse{
			SessionID: sessionID,
			Total:     0,
			Success:   0,
			Failed:    0,
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Results:   []api.PlayResult{},
		}, nil
	}

	if opts.Mode == ModeStrict || opts.Mode == "" {
		return p.playStrict(session, opts)
	} else if opts.Mode == ModeFast {
		return p.playFast(session, opts)
	}

	return nil, fmt.Errorf("unknown mode: %s", opts.Mode)
}

func (p *Player) playStrict(session *api.RecordedSessionData, opts PlayOptions) (*api.PlayResponse, error) {
	startTime := time.Now()
	results := make([]api.PlayResult, len(session.Requests))
	successCount := 0
	failedCount := 0

	for i, req := range session.Requests {
		if i > 0 {
			delay := req.DelayFromPrevMs
			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
		}

		if req.RequestDuration > 0 {
			time.Sleep(time.Duration(req.RequestDuration) * time.Millisecond)
		}

		result := p.executeRequest(req, opts)
		results[i] = result
		if result.Success {
			successCount++
		} else {
			failedCount++
		}
	}

	endTime := time.Now()

	return &api.PlayResponse{
		SessionID:  session.Meta.SessionID,
		Total:      len(session.Requests),
		Success:    successCount,
		Failed:     failedCount,
		StartTime:  startTime,
		EndTime:    endTime,
		DurationMs: endTime.Sub(startTime).Milliseconds(),
		Results:    results,
	}, nil
}

func (p *Player) playFast(session *api.RecordedSessionData, opts PlayOptions) (*api.PlayResponse, error) {
	startTime := time.Now()
	results := make([]api.PlayResult, len(session.Requests))

	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	failedCount := 0

	semaphore := make(chan struct{}, concurrency)

	for i, req := range session.Requests {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(idx int, recorded api.RecordedReq) {
			defer wg.Done()
			defer func() { <-semaphore }()

			result := p.executeRequest(recorded, opts)
			mu.Lock()
			results[idx] = result
			if result.Success {
				successCount++
			} else {
				failedCount++
			}
			mu.Unlock()
		}(i, req)
	}

	wg.Wait()

	endTime := time.Now()

	return &api.PlayResponse{
		SessionID:  session.Meta.SessionID,
		Total:      len(session.Requests),
		Success:    successCount,
		Failed:     failedCount,
		StartTime:  startTime,
		EndTime:    endTime,
		DurationMs: endTime.Sub(startTime).Milliseconds(),
		Results:    results,
	}, nil
}

func (p *Player) executeRequest(req api.RecordedReq, opts PlayOptions) api.PlayResult {
	startTime := time.Now()
	result := api.PlayResult{
		Index:  req.Index,
		Method: req.Method,
	}

	targetURL := req.URL
	if opts.TargetURL != "" {
		targetURL = buildTargetURL(req, opts.TargetURL)
	}
	targetURL = applyReplaceRules(targetURL, opts.ReplaceRules)
	result.URL = targetURL

	var bodyBytes []byte
	if req.Body != "" {
		bodyBytes = []byte(req.Body)
		bodyBytes = []byte(applyVariableReplacements(string(bodyBytes), opts.VariableValues, req.ExtractedVars))
	}

	httpReq, err := http.NewRequest(req.Method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.DurationMs = time.Since(startTime).Milliseconds()
		return result
	}

	for k, v := range req.Headers {
		if strings.ToLower(k) != "content-length" {
			httpReq.Header.Set(k, v)
		}
	}
	applyVariableReplacementsToHeaders(httpReq.Header, opts.VariableValues, req.ExtractedVars)

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.DurationMs = time.Since(startTime).Milliseconds()
		return result
	}
	defer resp.Body.Close()

	_, _ = ioutil.ReadAll(resp.Body)

	result.StatusCode = resp.StatusCode
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
	result.DurationMs = time.Since(startTime).Milliseconds()

	return result
}

func buildTargetURL(req api.RecordedReq, baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	path := req.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if req.Query != "" {
		return baseURL + path + "?" + req.Query
	}
	return baseURL + path
}

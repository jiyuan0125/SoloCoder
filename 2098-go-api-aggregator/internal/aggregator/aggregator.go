package aggregator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

	"api-aggregator/internal/cache"
	"api-aggregator/internal/config"
	"api-aggregator/internal/db"
	"api-aggregator/internal/formatter"
	"api-aggregator/internal/httpclient"
)

type AggregateResult struct {
	StatusCode   int
	Body         []byte
	AllSuccess   bool
	AllFailed    bool
	FailedEndpoints []string
	Results      []*formatter.EndpointResult
}

func Aggregate(spec *config.AggregationSpec, query url.Values) (*AggregateResult, error) {
	cacheKey, err := generateCacheKey(spec, query)
	if err != nil {
		return nil, err
	}

	if spec.CacheTTL > 0 {
		if cached, err := cache.Get(cacheKey); err == nil && cached != nil {
			return &AggregateResult{
				StatusCode: 200,
				Body:       cached,
				AllSuccess: true,
			}, nil
		}
	}

	results := callEndpointsConcurrently(spec, query)

	allSuccess := true
	allFailed := true
	var failedEndpoints []string

	for _, r := range results {
		if r.Success {
			allFailed = false
		} else {
			allSuccess = false
			failedEndpoints = append(failedEndpoints, r.Endpoint)
		}
	}

	var body []byte
	var formatErr error

	if !allSuccess {
		body, formatErr = formatter.FormatResults(results, formatter.FormatResponse, nil)
	} else {
		body, formatErr = formatter.FormatResults(results, spec.FormatType, spec.FormatConfig)
	}

	if formatErr != nil {
		return &AggregateResult{
			StatusCode:   400,
			Body:         []byte(`{"error":"format error: ` + formatErr.Error() + `"}`),
			AllSuccess:   false,
			AllFailed:    false,
			FailedEndpoints: nil,
			Results:      results,
		}, nil
	}

	statusCode := 200
	if allFailed {
		statusCode = 502
	} else if !allSuccess {
		statusCode = 207
	}

	if allSuccess && spec.CacheTTL > 0 {
		cache.Set(cacheKey, body, spec.CacheTTL)
	}

	return &AggregateResult{
		StatusCode:      statusCode,
		Body:            body,
		AllSuccess:      allSuccess,
		AllFailed:       allFailed,
		FailedEndpoints: failedEndpoints,
		Results:         results,
	}, nil
}

func callEndpointsConcurrently(spec *config.AggregationSpec, query url.Values) []*formatter.EndpointResult {
	results := make([]*formatter.EndpointResult, len(spec.Endpoints))
	var wg sync.WaitGroup

	for i, ep := range spec.Endpoints {
		wg.Add(1)
		go func(idx int, endpoint *config.EndpointSpec) {
			defer wg.Done()

			fullURL := buildURLWithQuery(endpoint.URL, query)

			resp, err := httpclient.CallEndpoint(
				endpoint.Method,
				fullURL,
				endpoint.Headers,
				nil,
				endpoint.Timeout,
				endpoint.Retries,
			)

			result := &formatter.EndpointResult{
				Index:    idx,
				Endpoint: endpoint.URL,
			}

			if err != nil {
				result.Success = false
				result.Error = err.Error()
				result.StatusCode = 0
			} else {
				result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
				result.StatusCode = resp.StatusCode
				result.Body = resp.Body
				if !result.Success {
					result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
				}
			}

			results[idx] = result
		}(i, ep)
	}

	wg.Wait()
	return results
}

func buildURLWithQuery(baseURL string, extraQuery url.Values) string {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	q := parsed.Query()
	for k, vs := range extraQuery {
		if k == "url" || k == "method" || k == "timeout" || k == "retry" ||
			k == "format" || k == "format_type" || k == "cache_ttl" ||
			k == "config_id" {
			continue
		}
		for _, v := range vs {
			q.Add(k, v)
		}
	}

	parsed.RawQuery = q.Encode()
	return parsed.String()
}

func generateCacheKey(spec *config.AggregationSpec, query url.Values) (string, error) {
	data := struct {
		Spec  *config.AggregationSpec `json:"spec"`
		Query string                  `json:"query"`
	}{
		Spec:  spec,
		Query: config.GenerateQueryString(query),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

func AggregateWithStoredConfig(configID int64, query url.Values) (*AggregateResult, error) {
	cfg, err := db.GetAggregationConfig(configID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("config not found")
	}

	endpoints := make([]*config.EndpointSpec, 0, len(cfg.EndpointIDs))
	for _, epID := range cfg.EndpointIDs {
		ep, err := db.GetEndpoint(epID)
		if err != nil {
			return nil, err
		}
		if ep == nil {
			continue
		}
		endpoints = append(endpoints, config.EndpointToSpec(ep))
	}

	spec := &config.AggregationSpec{
		Endpoints:    endpoints,
		FormatType:   formatter.FormatType(cfg.FormatType),
		FormatConfig: cfg.FormatConfig,
		CacheTTL:     cfg.CacheTTL,
	}

	if err := config.ValidateSpec(spec); err != nil {
		return nil, err
	}

	return Aggregate(spec, query)
}

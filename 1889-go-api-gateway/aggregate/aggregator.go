package aggregate

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const DefaultTimeout = 10 * time.Second

type Request struct {
	URLs []string `json:"urls"`
}

type Result struct {
	URL     string          `json:"url"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error,omitempty"`
}

func Run(ctx context.Context, urls []string) []*Result {
	results := make([]*Result, len(urls))
	var wg sync.WaitGroup
	client := &http.Client{Timeout: DefaultTimeout}
	for i, u := range urls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			reqCtx, cancel := context.WithTimeout(ctx, DefaultTimeout)
			defer cancel()
			req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
			if err != nil {
				results[i] = &Result{URL: url, Error: err.Error()}
				return
			}
			resp, err := client.Do(req)
			if err != nil {
				results[i] = &Result{URL: url, Error: err.Error()}
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				results[i] = &Result{URL: url, Error: err.Error()}
				return
			}
			results[i] = &Result{URL: url, Data: body}
		}(i, u)
	}
	wg.Wait()
	return results
}

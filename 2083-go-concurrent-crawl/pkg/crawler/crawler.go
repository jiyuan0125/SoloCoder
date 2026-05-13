package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Crawler struct {
	config       *Config
	progress     *ProgressManager
	robots       *RobotsManager
	rateLimiter  *RateLimiter
	client       *http.Client
	queue        chan URLInfo
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	statsChan    chan struct{}
	warnedHighConcurrency bool
	warnedHighFailureRate bool
}

func NewCrawler(config *Config) *Crawler {
	config.Validate()

	if config.Concurrency > 100 {
		fmt.Fprintf(os.Stderr, "Warning: Concurrency value %d exceeds limit of 100, using 100 instead\n", config.Concurrency)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Crawler{
		config:       config,
		progress:     NewProgressManager(config.OutputDir),
		robots:       NewRobotsManager(config.UserAgent),
		rateLimiter:  NewRateLimiter(config.RateLimit),
		client:       client,
		queue:        make(chan URLInfo, 10000),
		statsChan:    make(chan struct{}, 1),
	}
}

func (c *Crawler) Start() error {
	if err := ValidateURL(c.config.StartURL); err != nil {
		return fmt.Errorf("%v", err)
	}

	normalizedStartURL, err := NormalizeURL(c.config.StartURL)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(c.config.OutputDir, 0755); err != nil {
		return err
	}

	loaded := c.progress.Load()
	if loaded {
		fmt.Println("Resuming from previous progress...")
		pending := c.progress.GetPendingURLs()
		if len(pending) > 0 {
			go func() {
				for _, info := range pending {
					c.queue <- info
				}
			}()
		}
	}

	if !c.progress.IsVisited(normalizedStartURL) {
		c.progress.AddURL(normalizedStartURL, 0)
		go func() {
			c.queue <- URLInfo{URL: normalizedStartURL, Depth: 0, State: StatePending}
		}()
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())

	if c.config.Timeout > 0 {
		go func() {
			time.Sleep(time.Duration(c.config.Timeout) * time.Second)
			fmt.Fprintf(os.Stderr, "\nWarning: Timeout reached after %d seconds. Stopping gracefully...\n", c.config.Timeout)
			c.cancel()
		}()
	}

	go c.displayProgress()

	for i := 0; i < c.config.Concurrency; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}

	go c.monitorShutdown()

	c.wg.Wait()
	c.cancel()

	fmt.Println("\nCrawling completed!")
	completed, failed, _ := c.progress.GetStats()
	fmt.Printf("Summary: Completed %d, Failed %d\n", completed, failed)

	return c.progress.Save()
}

func (c *Crawler) worker(id int) {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case info, ok := <-c.queue:
			if !ok {
				return
			}
			c.crawlURL(info)
		}
	}
}

func (c *Crawler) crawlURL(info URLInfo) {
	select {
	case <-c.ctx.Done():
		return
	default:
	}

	if info.Depth > c.config.MaxDepth {
		return
	}

	normalizedURL, err := NormalizeURL(info.URL)
	if err != nil {
		c.progress.MarkFailed(info.URL, err)
		fmt.Fprintf(os.Stderr, "Error normalizing URL %s: %v\n", info.URL, err)
		return
	}

	if c.progress.IsVisited(normalizedURL) {
		c.progress.MarkProcessing(normalizedURL)
	} else {
		c.progress.AddURL(normalizedURL, info.Depth)
		c.progress.MarkProcessing(normalizedURL)
	}

	if !c.robots.IsAllowed(normalizedURL) {
		c.progress.MarkCompleted(normalizedURL)
		return
	}

	domain, err := GetDomain(normalizedURL)
	if err == nil {
		c.rateLimiter.Wait(domain)
	}

	req, err := http.NewRequestWithContext(c.ctx, "GET", normalizedURL, nil)
	if err != nil {
		c.progress.MarkFailed(normalizedURL, err)
		fmt.Fprintf(os.Stderr, "Error creating request for %s: %v\n", normalizedURL, err)
		c.progress.Save()
		return
	}
	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		c.progress.MarkFailed(normalizedURL, err)
		fmt.Fprintf(os.Stderr, "Error fetching %s: %v\n", normalizedURL, err)
		c.progress.Save()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := fmt.Sprintf("HTTP status %d", resp.StatusCode)
		c.progress.MarkFailed(normalizedURL, fmt.Errorf(errMsg))
		fmt.Fprintf(os.Stderr, "Error fetching %s: %s\n", normalizedURL, errMsg)
		c.progress.Save()
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.progress.MarkFailed(normalizedURL, err)
		fmt.Fprintf(os.Stderr, "Error reading response body for %s: %v\n", normalizedURL, err)
		c.progress.Save()
		return
	}

	if err := c.savePage(normalizedURL, body); err != nil {
		c.progress.MarkFailed(normalizedURL, err)
		fmt.Fprintf(os.Stderr, "Error saving page %s: %v\n", normalizedURL, err)
		c.progress.Save()
		return
	}

	c.progress.MarkCompleted(normalizedURL)
	c.progress.Save()

	if info.Depth < c.config.MaxDepth {
		c.extractLinks(normalizedURL, body, info.Depth+1)
	}
}

func (c *Crawler) extractLinks(baseURL string, body []byte, depth int) {
	linkRegex := regexp.MustCompile(`<a\s+(?:[^>]*?\s+)?href="([^"]*)"`)
	matches := linkRegex.FindAllSubmatch(body, -1)

	for _, match := range matches {
		link := string(match[1])
		link = strings.TrimSpace(link)
		
		if link == "" || strings.HasPrefix(link, "#") || strings.HasPrefix(link, "javascript:") {
			continue
		}

		parsedBase, err := url.Parse(baseURL)
		if err != nil {
			continue
		}

		parsedLink, err := url.Parse(link)
		if err != nil {
			continue
		}

		absoluteURL := parsedBase.ResolveReference(parsedLink)
		
		if absoluteURL.Scheme != "http" && absoluteURL.Scheme != "https" {
			continue
		}

		normalized, err := NormalizeURL(absoluteURL.String())
		if err != nil {
			continue
		}

		if !c.progress.IsVisited(normalized) {
			c.progress.AddURL(normalized, depth)
			c.progress.Save()
			select {
			case c.queue <- URLInfo{URL: normalized, Depth: depth, State: StatePending}:
			case <-c.ctx.Done():
				return
			}
		}
	}
}

func (c *Crawler) savePage(rawURL string, content []byte) error {
	hash := HashURL(rawURL)
	filename := filepath.Join(c.config.OutputDir, fmt.Sprintf("%s.html", hash))
	return os.WriteFile(filename, content, 0644)
}

func (c *Crawler) displayProgress() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			completed, failed, queueSize := c.progress.GetStats()
			total := completed + failed
			
			fmt.Printf("\rProgress: Completed=%d, Failed=%d, Queue=%d", completed, failed, queueSize)
			
			if total > 0 && !c.warnedHighFailureRate {
				failureRate := float64(failed) / float64(total)
				if failureRate > 0.5 {
					fmt.Fprintf(os.Stderr, "\nWarning: Failure rate exceeds 50%% (%.2f%%)\n", failureRate*100)
					c.warnedHighFailureRate = true
				}
			}
		}
	}
}

func (c *Crawler) monitorShutdown() {
	<-c.ctx.Done()
	
	select {
	case <-time.After(2 * time.Second):
		close(c.queue)
	case <-time.After(10 * time.Second):
		fmt.Fprintln(os.Stderr, "Force closing after timeout...")
		close(c.queue)
	}
}

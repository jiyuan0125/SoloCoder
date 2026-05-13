package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
)

type RobotsManager struct {
	cache     map[string]*robotstxt.RobotsData
	cacheMu   sync.RWMutex
	userAgent string
	client    *http.Client
}

func NewRobotsManager(userAgent string) *RobotsManager {
	return &RobotsManager{
		cache:     make(map[string]*robotstxt.RobotsData),
		userAgent: userAgent,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (r *RobotsManager) IsAllowed(rawURL string) bool {
	domain, err := GetDomain(rawURL)
	if err != nil {
		return true
	}

	robots := r.getRobotsData(domain)
	if robots == nil {
		return true
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return true
	}

	return robots.TestAgent(parsedURL.Path, r.userAgent)
}

func (r *RobotsManager) getRobotsData(domain string) *robotstxt.RobotsData {
	r.cacheMu.RLock()
	if robots, exists := r.cache[domain]; exists {
		r.cacheMu.RUnlock()
		return robots
	}
	r.cacheMu.RUnlock()

	r.cacheMu.Lock()
	defer r.cacheMu.Unlock()

	if robots, exists := r.cache[domain]; exists {
		return robots
	}

	robots := r.fetchRobotsData(domain)
	r.cache[domain] = robots
	return robots
}

func (r *RobotsManager) fetchRobotsData(domain string) *robotstxt.RobotsData {
	robotsURL := fmt.Sprintf("https://%s/robots.txt", domain)
	
	req, err := http.NewRequest("GET", robotsURL, nil)
	if err != nil {
		robotsURL = fmt.Sprintf("http://%s/robots.txt", domain)
		req, err = http.NewRequest("GET", robotsURL, nil)
		if err != nil {
			return nil
		}
	}

	req.Header.Set("User-Agent", r.userAgent)
	
	resp, err := r.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		allowAll, _ := robotstxt.FromBytes([]byte{})
		return allowAll
	}

	if resp.StatusCode >= 400 {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	robots, err := robotstxt.FromBytes(body)
	if err != nil {
		return nil
	}

	return robots
}

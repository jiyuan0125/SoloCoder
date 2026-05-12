package proxy

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

type ProxyManager struct {
	backends []*url.URL
	next     uint64
}

func NewProxyManager(backendURLs []string) *ProxyManager {
	backends := make([]*url.URL, 0, len(backendURLs))
	for _, raw := range backendURLs {
		if u, err := url.Parse(raw); err == nil {
			backends = append(backends, u)
		}
	}
	return &ProxyManager{backends: backends}
}

func (p *ProxyManager) nextBackend() *url.URL {
	if len(p.backends) == 0 {
		return nil
	}
	idx := atomic.AddUint64(&p.next, 1) % uint64(len(p.backends))
	return p.backends[idx]
}

func (p *ProxyManager) Proxy(c *gin.Context) {
	target := p.nextBackend()
	if target == nil {
		c.JSON(502, gin.H{"error": "No available backend"})
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ServeHTTP(c.Writer, c.Request)
}

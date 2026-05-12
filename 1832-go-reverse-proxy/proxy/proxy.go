package proxy

import (
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gin-gonic/gin"

	"reverse-proxy/store"
	"reverse-proxy/types"
)

type Proxy struct {
	store  *store.Store
	client *http.Client
}

func New(store *store.Store) *Proxy {
	return &Proxy{
		store: store,
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (p *Proxy) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host
		backend := p.store.GetByHost(host)

		if backend == nil || backend.GetStatus() != types.StatusHealthy {
			c.JSON(http.StatusNotFound, gin.H{"error": "backend not found or unhealthy"})
			return
		}

		rp := httputil.NewSingleHostReverseProxy(backend.TargetURL)

		originalDirector := rp.Director
		rp.Director = func(req *http.Request) {
			originalDirector(req)

			req.Header.Del("X-Forwarded-Proto")

			if req.Header.Get("X-Real-IP") == "" {
				req.Header.Set("X-Real-IP", c.ClientIP())
			}
		}

		rp.Transport = &http.Transport{
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  true,
			MaxIdleConnsPerHost: 100,
		}

		rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "gateway timeout"})
		}

		rp.ServeHTTP(c.Writer, c.Request)
	}
}

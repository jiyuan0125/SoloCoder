package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"gateway/config"
)

type Router struct {
	cfg *config.Manager
}

func NewRouter(cfg *config.Manager) *Router {
	return &Router{cfg: cfg}
}

func (r *Router) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg := r.cfg.Get()
		route := r.matchRoute(cfg.Routes, req.URL.Path)

		if route == nil {
			http.NotFound(w, req)
			return
		}

		backendURL, err := url.Parse(route.Backend)
		if err != nil {
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(backendURL)

		if route.StripPrefix {
			originalPath := req.URL.Path
			req.URL.Path = strings.TrimPrefix(originalPath, route.Path)
			if !strings.HasPrefix(req.URL.Path, "/") {
				req.URL.Path = "/" + req.URL.Path
			}
		}

		originalDirector := proxy.Director
		proxy.Director = func(r *http.Request) {
			originalDirector(r)
			r.Host = backendURL.Host
		}

		proxy.ServeHTTP(w, req)
	})
}

func (r *Router) matchRoute(routes []config.Route, path string) *config.Route {
	for i := range routes {
		if strings.HasPrefix(path, routes[i].Path) {
			return &routes[i]
		}
	}
	return nil
}

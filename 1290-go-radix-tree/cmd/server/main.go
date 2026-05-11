package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/example/radix-router/pkg/common"
	"github.com/example/radix-router/pkg/httprouter"
	"github.com/example/radix-router/pkg/iprouter"
)

type Server struct {
	httpRouter *httprouter.Router
	ipRouter   *iprouter.Router
	mu         sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		httpRouter: httprouter.NewRouter(),
		ipRouter:   iprouter.NewRouter(),
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, common.ErrorResponse{Error: message})
}

func (s *Server) AddHTTPRoute(w http.ResponseWriter, r *http.Request) {
	var req common.AddHTTPRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Route.Method == "" || req.Route.Path == "" {
		s.writeError(w, http.StatusBadRequest, "Method and path are required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.httpRouter.AddRoute(req.Route.Method, req.Route.Path, req.Route.Handler)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, common.SuccessResponse{Success: true})
}

func (s *Server) DeleteHTTPRoute(w http.ResponseWriter, r *http.Request) {
	var req common.DeleteHTTPRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Method == "" || req.Path == "" {
		s.writeError(w, http.StatusBadRequest, "Method and path are required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	success := s.httpRouter.DeleteRoute(req.Method, req.Path)
	s.writeJSON(w, http.StatusOK, common.SuccessResponse{Success: success})
}

func (s *Server) MatchHTTP(w http.ResponseWriter, r *http.Request) {
	var req common.MatchHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Method == "" || req.Path == "" {
		s.writeError(w, http.StatusBadRequest, "Method and path are required")
		return
	}

	s.mu.RLock()
	start := time.Now()
	result := s.httpRouter.Match(req.Method, req.Path)
	latency := time.Since(start).Nanoseconds()
	s.mu.RUnlock()

	resp := common.MatchHTTPResponse{
		Found:  result.Found,
		Latency: latency,
	}

	if result.Found {
		resp.Route = common.HTTPRoute{
			Method:  result.Route.Method,
			Path:    result.Route.Path,
			Handler: result.Route.Handler,
		}
		resp.Params = result.Params
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) ListHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	routes := s.httpRouter.List()
	s.mu.RUnlock()

	httpRoutes := make([]common.HTTPRoute, 0, len(routes))
	for _, rt := range routes {
		httpRoutes = append(httpRoutes, common.HTTPRoute{
			Method:  rt.Method,
			Path:    rt.Path,
			Handler: rt.Handler,
		})
	}

	s.writeJSON(w, http.StatusOK, common.ListHTTPResponse{Routes: httpRoutes})
}

func (s *Server) AddIPRoute(w http.ResponseWriter, r *http.Request) {
	var req common.AddIPRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Route.CIDR == "" {
		s.writeError(w, http.StatusBadRequest, "CIDR is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.ipRouter.AddRoute(req.Route.CIDR, req.Route.Nexthop)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid CIDR: "+err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, common.SuccessResponse{Success: true})
}

func (s *Server) DeleteIPRoute(w http.ResponseWriter, r *http.Request) {
	var req common.DeleteIPRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.CIDR == "" {
		s.writeError(w, http.StatusBadRequest, "CIDR is required")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	success := s.ipRouter.DeleteRoute(req.CIDR)
	s.writeJSON(w, http.StatusOK, common.SuccessResponse{Success: success})
}

func (s *Server) MatchIP(w http.ResponseWriter, r *http.Request) {
	var req common.MatchIPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.IP == "" {
		s.writeError(w, http.StatusBadRequest, "IP is required")
		return
	}

	s.mu.RLock()
	start := time.Now()
	result := s.ipRouter.Match(req.IP)
	latency := time.Since(start).Nanoseconds()
	s.mu.RUnlock()

	resp := common.MatchIPResponse{
		Found:   result.Found,
		Latency: latency,
	}

	if result.Found {
		resp.Route = common.IPRoute{
			CIDR:    result.Route.CIDR,
			Nexthop: result.Route.Nexthop,
		}
		resp.MatchedCIDR = result.MatchedCIDR
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) ListIP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	routes := s.ipRouter.List()
	s.mu.RUnlock()

	ipRoutes := make([]common.IPRoute, 0, len(routes))
	for _, rt := range routes {
		ipRoutes = append(ipRoutes, common.IPRoute{
			CIDR:    rt.CIDR,
			Nexthop: rt.Nexthop,
		})
	}

	s.writeJSON(w, http.StatusOK, common.ListIPResponse{Routes: ipRoutes})
}

func (s *Server) Stats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	httpStats := s.httpRouter.Stats()
	ipStats := s.ipRouter.Stats()
	s.mu.RUnlock()

	resp := common.StatsResponse{
		HTTP: make(map[string]common.TreeStats),
		IP:   make(map[string]common.TreeStats),
	}

	for method, stats := range httpStats {
		resp.HTTP[method] = common.TreeStats{
			NodeCount: stats.NodeCount,
			Height:    stats.Height,
			IsEmpty:   stats.IsEmpty,
		}
	}

	for key, stats := range ipStats {
		resp.IP[key] = common.TreeStats{
			NodeCount: stats.NodeCount,
			Height:    stats.Height,
			IsEmpty:   stats.IsEmpty,
		}
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) BulkImport(w http.ResponseWriter, r *http.Request) {
	var req common.BulkImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, route := range req.HTTPRoutes {
		if route.Method != "" && route.Path != "" {
			s.httpRouter.AddRoute(route.Method, route.Path, route.Handler)
		}
	}

	for _, route := range req.IPRoutes {
		if route.CIDR != "" {
			s.ipRouter.AddRoute(route.CIDR, route.Nexthop)
		}
	}

	s.writeJSON(w, http.StatusOK, common.SuccessResponse{Success: true})
}

func (s *Server) BulkMatch(w http.ResponseWriter, r *http.Request) {
	var req common.BulkMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	httpResults := make([]common.MatchHTTPResponse, 0, len(req.HTTPMatches))
	for _, match := range req.HTTPMatches {
		start := time.Now()
		result := s.httpRouter.Match(match.Method, match.Path)
		latency := time.Since(start).Nanoseconds()

		resp := common.MatchHTTPResponse{
			Found:   result.Found,
			Latency: latency,
		}
		if result.Found {
			resp.Route = common.HTTPRoute{
				Method:  result.Route.Method,
				Path:    result.Route.Path,
				Handler: result.Route.Handler,
			}
			resp.Params = result.Params
		}
		httpResults = append(httpResults, resp)
	}

	ipResults := make([]common.MatchIPResponse, 0, len(req.IPMatches))
	for _, match := range req.IPMatches {
		start := time.Now()
		result := s.ipRouter.Match(match.IP)
		latency := time.Since(start).Nanoseconds()

		resp := common.MatchIPResponse{
			Found:   result.Found,
			Latency: latency,
		}
		if result.Found {
			resp.Route = common.IPRoute{
				CIDR:    result.Route.CIDR,
				Nexthop: result.Route.Nexthop,
			}
			resp.MatchedCIDR = result.MatchedCIDR
		}
		ipResults = append(ipResults, resp)
	}

	s.writeJSON(w, http.StatusOK, common.BulkMatchResponse{
		HTTPResults: httpResults,
		IPResults:   ipResults,
	})
}

func getPort() string {
	port := flag.String("port", "", "Server port")
	flag.Parse()

	if *port != "" {
		return *port
	}

	portEnv := os.Getenv("SERVER_PORT")
	if portEnv != "" {
		return portEnv
	}

	return "8517"
}

func main() {
	server := NewServer()

	http.HandleFunc("/api/http/add", server.AddHTTPRoute)
	http.HandleFunc("/api/http/delete", server.DeleteHTTPRoute)
	http.HandleFunc("/api/http/match", server.MatchHTTP)
	http.HandleFunc("/api/http/list", server.ListHTTP)

	http.HandleFunc("/api/ip/add", server.AddIPRoute)
	http.HandleFunc("/api/ip/delete", server.DeleteIPRoute)
	http.HandleFunc("/api/ip/match", server.MatchIP)
	http.HandleFunc("/api/ip/list", server.ListIP)

	http.HandleFunc("/api/stats", server.Stats)
	http.HandleFunc("/api/bulk/import", server.BulkImport)
	http.HandleFunc("/api/bulk/match", server.BulkMatch)

	port := getPort()
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	log.Printf("Server starting on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

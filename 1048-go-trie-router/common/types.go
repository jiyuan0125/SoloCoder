package common

import "net/http"

type RouteInfo struct {
	Path        string            `json:"path"`
	Method      string            `json:"method"`
	HandlerName string            `json:"handler_name"`
	RegisteredAt int64            `json:"registered_at"`
}

type MatchResult struct {
	Found       bool              `json:"found"`
	Route       *RouteInfo        `json:"route,omitempty"`
	Params      map[string]string `json:"params,omitempty"`
	AllowedMethods []string       `json:"allowed_methods,omitempty"`
	Status      int               `json:"status"`
}

type RegisterRequest struct {
	Path        string `json:"path"`
	Method      string `json:"method"`
	HandlerName string `json:"handler_name"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type QueryRequest struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

type QueryResponse struct {
	Found          bool              `json:"found"`
	Route          *RouteInfo        `json:"route,omitempty"`
	Params         map[string]string `json:"params,omitempty"`
	AllowedMethods []string          `json:"allowed_methods,omitempty"`
	Status         int               `json:"status"`
	Error          string            `json:"error,omitempty"`
}

type ListRequest struct {
	SortBy string `json:"sort_by"`
}

type ListResponse struct {
	Routes  []*RouteInfo `json:"routes"`
	Count   int          `json:"count"`
	Error   string       `json:"error,omitempty"`
}

func (m *MatchResult) IsMethodNotAllowed() bool {
	return m.Status == http.StatusMethodNotAllowed
}

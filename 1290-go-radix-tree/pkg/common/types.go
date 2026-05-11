package common

type HTTPRoute struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Handler string `json:"handler"`
}

type IPRoute struct {
	CIDR    string `json:"cidr"`
	Nexthop string `json:"nexthop"`
}

type AddHTTPRouteRequest struct {
	Route HTTPRoute `json:"route"`
}

type AddIPRouteRequest struct {
	Route IPRoute `json:"route"`
}

type DeleteHTTPRouteRequest struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type DeleteIPRouteRequest struct {
	CIDR string `json:"cidr"`
}

type MatchHTTPRequest struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type MatchHTTPResponse struct {
	Found   bool            `json:"found"`
	Route   HTTPRoute       `json:"route"`
	Params  map[string]string `json:"params"`
	Latency int64           `json:"latency_ns"`
}

type MatchIPRequest struct {
	IP string `json:"ip"`
}

type MatchIPResponse struct {
	Found       bool   `json:"found"`
	Route       IPRoute `json:"route"`
	MatchedCIDR string `json:"matched_cidr"`
	Latency     int64  `json:"latency_ns"`
}

type ListHTTPResponse struct {
	Routes []HTTPRoute `json:"routes"`
}

type ListIPResponse struct {
	Routes []IPRoute `json:"routes"`
}

type StatsResponse struct {
	HTTP map[string]TreeStats `json:"http"`
	IP   map[string]TreeStats `json:"ip"`
}

type TreeStats struct {
	NodeCount int  `json:"node_count"`
	Height    int  `json:"height"`
	IsEmpty   bool `json:"is_empty"`
}

type BulkImportRequest struct {
	HTTPRoutes []HTTPRoute `json:"http_routes"`
	IPRoutes   []IPRoute   `json:"ip_routes"`
}

type BulkMatchRequest struct {
	HTTPMatches []MatchHTTPRequest `json:"http_matches"`
	IPMatches   []MatchIPRequest   `json:"ip_matches"`
}

type BulkMatchResponse struct {
	HTTPResults []MatchHTTPResponse `json:"http_results"`
	IPResults   []MatchIPResponse   `json:"ip_results"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

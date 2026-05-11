package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/example/radix-router/pkg/common"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	if strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL[:len(baseURL)-1]
	}
	return &Client{baseURL: baseURL}
}

func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		return json.Unmarshal(respBody, result)
	}
	return nil
}

func (c *Client) AddHTTPRoute(method, path, handler string) error {
	req := common.AddHTTPRouteRequest{
		Route: common.HTTPRoute{
			Method:  method,
			Path:    path,
			Handler: handler,
		},
	}
	var result common.SuccessResponse
	return c.doRequest("POST", "/api/http/add", req, &result)
}

func (c *Client) DeleteHTTPRoute(method, path string) error {
	req := common.DeleteHTTPRouteRequest{
		Method: method,
		Path:   path,
	}
	var result common.SuccessResponse
	return c.doRequest("POST", "/api/http/delete", req, &result)
}

func (c *Client) MatchHTTP(method, path string) (*common.MatchHTTPResponse, error) {
	req := common.MatchHTTPRequest{
		Method: method,
		Path:   path,
	}
	var result common.MatchHTTPResponse
	err := c.doRequest("POST", "/api/http/match", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListHTTP() ([]common.HTTPRoute, error) {
	var result common.ListHTTPResponse
	err := c.doRequest("GET", "/api/http/list", nil, &result)
	if err != nil {
		return nil, err
	}
	return result.Routes, nil
}

func (c *Client) AddIPRoute(cidr, nexthop string) error {
	req := common.AddIPRouteRequest{
		Route: common.IPRoute{
			CIDR:    cidr,
			Nexthop: nexthop,
		},
	}
	var result common.SuccessResponse
	return c.doRequest("POST", "/api/ip/add", req, &result)
}

func (c *Client) DeleteIPRoute(cidr string) error {
	req := common.DeleteIPRouteRequest{
		CIDR: cidr,
	}
	var result common.SuccessResponse
	return c.doRequest("POST", "/api/ip/delete", req, &result)
}

func (c *Client) MatchIP(ip string) (*common.MatchIPResponse, error) {
	req := common.MatchIPRequest{
		IP: ip,
	}
	var result common.MatchIPResponse
	err := c.doRequest("POST", "/api/ip/match", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListIP() ([]common.IPRoute, error) {
	var result common.ListIPResponse
	err := c.doRequest("GET", "/api/ip/list", nil, &result)
	if err != nil {
		return nil, err
	}
	return result.Routes, nil
}

func (c *Client) Stats() (*common.StatsResponse, error) {
	var result common.StatsResponse
	err := c.doRequest("GET", "/api/stats", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) BulkImport(httpRoutes []common.HTTPRoute, ipRoutes []common.IPRoute) error {
	req := common.BulkImportRequest{
		HTTPRoutes: httpRoutes,
		IPRoutes:   ipRoutes,
	}
	var result common.SuccessResponse
	return c.doRequest("POST", "/api/bulk/import", req, &result)
}

func (c *Client) BulkMatch(httpMatches []common.MatchHTTPRequest, ipMatches []common.MatchIPRequest) (*common.BulkMatchResponse, error) {
	req := common.BulkMatchRequest{
		HTTPMatches: httpMatches,
		IPMatches:   ipMatches,
	}
	var result common.BulkMatchResponse
	err := c.doRequest("POST", "/api/bulk/match", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func printUsage() {
	fmt.Println(`Radix Router Client

Usage:
  client [command] [flags]

Commands:
  http-add      Add an HTTP route
  http-delete   Delete an HTTP route
  http-match    Match an HTTP request
  http-list     List all HTTP routes
  
  ip-add        Add an IP route
  ip-delete     Delete an IP route
  ip-match      Match an IP address
  ip-list       List all IP routes
  
  stats         Show tree statistics
  bulk-import   Bulk import routes from JSON file
  bulk-match    Bulk match requests from JSON file

Flags:
  --server      Server address (default: localhost:8517)
  --help        Show help

Examples:
  client http-add --method GET --path /api/users --handler GetUsers
  client http-match --method GET --path /api/users/123
  client ip-add --cidr 192.168.1.0/24 --nexthop 10.0.0.1
  client ip-match --ip 192.168.1.5
  client stats
  client bulk-import --file routes.json
  client bulk-match --file matches.json`)
}

func parseKVPair(s string) (string, string, bool) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	fs := flag.NewFlagSet("global", flag.ContinueOnError)
	server := fs.String("server", "localhost:8517", "Server address")

	args := os.Args[1:]
	cmd := args[0]

	for i := 1; i < len(args); i++ {
		if args[i] == "--server" && i+1 < len(args) {
			*server = args[i+1]
			args = append(args[:i], args[i+2:]...)
			i--
		}
	}

	if cmd == "--help" || cmd == "-h" {
		printUsage()
		os.Exit(0)
	}

	client := NewClient(*server)

	switch cmd {
	case "http-add":
		handleHTTPAdd(client, args[1:])
	case "http-delete":
		handleHTTPDelete(client, args[1:])
	case "http-match":
		handleHTTPMatch(client, args[1:])
	case "http-list":
		handleHTTPList(client)
	case "ip-add":
		handleIPAdd(client, args[1:])
	case "ip-delete":
		handleIPDelete(client, args[1:])
	case "ip-match":
		handleIPMatch(client, args[1:])
	case "ip-list":
		handleIPList(client)
	case "stats":
		handleStats(client)
	case "bulk-import":
		handleBulkImport(client, args[1:])
	case "bulk-match":
		handleBulkMatch(client, args[1:])
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func handleHTTPAdd(client *Client, args []string) {
	var method, path, handler string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--method":
			if i+1 < len(args) {
				method = args[i+1]
				i++
			}
		case "--path":
			if i+1 < len(args) {
				path = args[i+1]
				i++
			}
		case "--handler":
			if i+1 < len(args) {
				handler = args[i+1]
				i++
			}
		}
	}

	if method == "" || path == "" {
		log.Fatal("Method and path are required")
	}

	err := client.AddHTTPRoute(method, path, handler)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("HTTP route added successfully")
}

func handleHTTPDelete(client *Client, args []string) {
	var method, path string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--method":
			if i+1 < len(args) {
				method = args[i+1]
				i++
			}
		case "--path":
			if i+1 < len(args) {
				path = args[i+1]
				i++
			}
		}
	}

	if method == "" || path == "" {
		log.Fatal("Method and path are required")
	}

	err := client.DeleteHTTPRoute(method, path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("HTTP route deleted successfully")
}

func handleHTTPMatch(client *Client, args []string) {
	var method, path string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--method":
			if i+1 < len(args) {
				method = args[i+1]
				i++
			}
		case "--path":
			if i+1 < len(args) {
				path = args[i+1]
				i++
			}
		}
	}

	if method == "" || path == "" {
		log.Fatal("Method and path are required")
	}

	result, err := client.MatchHTTP(method, path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Result: %v\n", result.Found)
	if result.Found {
		fmt.Printf("Route: %s %s -> %s\n", result.Route.Method, result.Route.Path, result.Route.Handler)
		if len(result.Params) > 0 {
			fmt.Printf("Params: %v\n", result.Params)
		}
	}
	fmt.Printf("Latency: %d ns\n", result.Latency)
}

func handleHTTPList(client *Client) {
	routes, err := client.ListHTTP()
	if err != nil {
		log.Fatal(err)
	}

	if len(routes) == 0 {
		fmt.Println("No HTTP routes registered")
		return
	}

	fmt.Printf("HTTP Routes (%d):\n", len(routes))
	for _, rt := range routes {
		fmt.Printf("  %s %s -> %s\n", rt.Method, rt.Path, rt.Handler)
	}
}

func handleIPAdd(client *Client, args []string) {
	var cidr, nexthop string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--cidr":
			if i+1 < len(args) {
				cidr = args[i+1]
				i++
			}
		case "--nexthop":
			if i+1 < len(args) {
				nexthop = args[i+1]
				i++
			}
		}
	}

	if cidr == "" {
		log.Fatal("CIDR is required")
	}

	err := client.AddIPRoute(cidr, nexthop)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("IP route added successfully")
}

func handleIPDelete(client *Client, args []string) {
	var cidr string
	for i := 0; i < len(args); i++ {
		if args[i] == "--cidr" && i+1 < len(args) {
			cidr = args[i+1]
			i++
		}
	}

	if cidr == "" {
		log.Fatal("CIDR is required")
	}

	err := client.DeleteIPRoute(cidr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("IP route deleted successfully")
}

func handleIPMatch(client *Client, args []string) {
	var ip string
	for i := 0; i < len(args); i++ {
		if args[i] == "--ip" && i+1 < len(args) {
			ip = args[i+1]
			i++
		}
	}

	if ip == "" {
		log.Fatal("IP address is required")
	}

	result, err := client.MatchIP(ip)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Result: %v\n", result.Found)
	if result.Found {
		fmt.Printf("Matched CIDR: %s -> %s\n", result.MatchedCIDR, result.Route.Nexthop)
	}
	fmt.Printf("Latency: %d ns\n", result.Latency)
}

func handleIPList(client *Client) {
	routes, err := client.ListIP()
	if err != nil {
		log.Fatal(err)
	}

	if len(routes) == 0 {
		fmt.Println("No IP routes registered")
		return
	}

	fmt.Printf("IP Routes (%d):\n", len(routes))
	for _, rt := range routes {
		fmt.Printf("  %s -> %s\n", rt.CIDR, rt.Nexthop)
	}
}

func handleStats(client *Client) {
	stats, err := client.Stats()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Statistics ===")
	fmt.Println("\nHTTP Router:")
	if len(stats.HTTP) == 0 {
		fmt.Println("  (no HTTP routes)")
	} else {
		for method, s := range stats.HTTP {
			fmt.Printf("  %s:\n", method)
			fmt.Printf("    Nodes: %d, Height: %d, Empty: %v\n", s.NodeCount, s.Height, s.IsEmpty)
		}
	}

	fmt.Println("\nIP Router:")
	if len(stats.IP) == 0 {
		fmt.Println("  (no IP routes)")
	} else {
		for key, s := range stats.IP {
			fmt.Printf("  %s:\n", key)
			fmt.Printf("    Nodes: %d, Height: %d, Empty: %v\n", s.NodeCount, s.Height, s.IsEmpty)
		}
	}
}

func handleBulkImport(client *Client, args []string) {
	var filename string
	for i := 0; i < len(args); i++ {
		if args[i] == "--file" && i+1 < len(args) {
			filename = args[i+1]
			i++
		}
	}

	if filename == "" {
		log.Fatal("File path is required")
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var req common.BulkImportRequest
	if err := json.Unmarshal(data, &req); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	fmt.Printf("Importing %d HTTP routes and %d IP routes...\n", len(req.HTTPRoutes), len(req.IPRoutes))

	err = client.BulkImport(req.HTTPRoutes, req.IPRoutes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Bulk import completed successfully")
}

func handleBulkMatch(client *Client, args []string) {
	var filename string
	for i := 0; i < len(args); i++ {
		if args[i] == "--file" && i+1 < len(args) {
			filename = args[i+1]
			i++
		}
	}

	if filename == "" {
		log.Fatal("File path is required")
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var req common.BulkMatchRequest
	if err := json.Unmarshal(data, &req); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	results, err := client.BulkMatch(req.HTTPMatches, req.IPMatches)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Bulk Match Results ===")
	if len(results.HTTPResults) > 0 {
		fmt.Printf("\nHTTP Matches (%d):\n", len(results.HTTPResults))
		for i, r := range results.HTTPResults {
			req := req.HTTPMatches[i]
			if r.Found {
				fmt.Printf("  %s %s -> %s (%v, %d ns)\n", req.Method, req.Path, r.Route.Handler, r.Params, r.Latency)
			} else {
				fmt.Printf("  %s %s -> NOT FOUND (%d ns)\n", req.Method, req.Path, r.Latency)
			}
		}
	}

	if len(results.IPResults) > 0 {
		fmt.Printf("\nIP Matches (%d):\n", len(results.IPResults))
		for i, r := range results.IPResults {
			req := req.IPMatches[i]
			if r.Found {
				fmt.Printf("  %s -> %s via %s (%d ns)\n", req.IP, r.MatchedCIDR, r.Route.Nexthop, r.Latency)
			} else {
				fmt.Printf("  %s -> NOT FOUND (%d ns)\n", req.IP, r.Latency)
			}
		}
	}
}

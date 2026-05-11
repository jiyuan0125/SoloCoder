package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/zorder-index/internal/api"
	"github.com/zorder-index/pkg/zorder"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverAddr := getServerAddr()

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "encode2d":
		handleEncode2D(serverAddr, args)
	case "decode2d":
		handleDecode2D(serverAddr, args)
	case "encode3d":
		handleEncode3D(serverAddr, args)
	case "decode3d":
		handleDecode3D(serverAddr, args)
	case "query":
		handleQuery(serverAddr, args)
	case "health":
		handleHealth(serverAddr)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func getServerAddr() string {
	addr := os.Getenv("ZORDER_SERVER")
	if addr == "" {
		addr = "http://localhost:8507"
	}
	return addr
}

func printUsage() {
	fmt.Println("Z-Order Index Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  encode2d <x> <y> [--method=bit|lookup]  Encode 2D coordinate")
	fmt.Println("  decode2d <code>                          Decode 2D Morton code")
	fmt.Println("  encode3d <x> <y> <z>                     Encode 3D coordinate")
	fmt.Println("  decode3d <code>                          Decode 3D Morton code")
	fmt.Println("  query <min_x> <min_y> <max_x> <max_y>    Query prefix ranges")
	fmt.Println("  health                                   Check server health")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  ZORDER_SERVER  Server address (default: http://localhost:8507)")
}

func handleEncode2D(serverAddr string, args []string) {
	fs := flag.NewFlagSet("encode2d", flag.ExitOnError)
	method := fs.String("method", "bit", "Encoding method: bit or lookup")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: client encode2d <x> <y> [--method=bit|lookup]")
		os.Exit(1)
	}

	x, err := strconv.ParseUint(remaining[0], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid x value: %v\n", err)
		os.Exit(1)
	}

	y, err := strconv.ParseUint(remaining[1], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid y value: %v\n", err)
		os.Exit(1)
	}

	req := api.Encode2DRequest{X: uint32(x), Y: uint32(y)}
	var resp api.Encode2DResponse

	url := fmt.Sprintf("%s/encode2d?method=%s", serverAddr, *method)
	if err := postJSON(url, req, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Coordinate: (%d, %d)\n", resp.Coordinate.X, resp.Coordinate.Y)
	fmt.Printf("Morton Code: %s\n", uint64ToString(uint64(resp.Code)))
	fmt.Printf("Method: %s\n", resp.Method)
}

func handleDecode2D(serverAddr string, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: client decode2d <code>")
		os.Exit(1)
	}

	codeStr := args[0]
	code, err := parseCode(codeStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid code: %v\n", err)
		os.Exit(1)
	}

	req := api.Decode2DRequest{Code: zorder.Code(code)}
	var resp api.Decode2DResponse

	url := fmt.Sprintf("%s/decode2d", serverAddr)
	if err := postJSON(url, req, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Morton Code: %s\n", uint64ToString(uint64(resp.Code)))
	fmt.Printf("Coordinate: (%d, %d)\n", resp.Coordinate.X, resp.Coordinate.Y)
}

func handleEncode3D(serverAddr string, args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: client encode3d <x> <y> <z>")
		os.Exit(1)
	}

	x, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid x value: %v\n", err)
		os.Exit(1)
	}

	y, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid y value: %v\n", err)
		os.Exit(1)
	}

	z, err := strconv.ParseUint(args[2], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid z value: %v\n", err)
		os.Exit(1)
	}

	req := api.Encode3DRequest{X: uint32(x), Y: uint32(y), Z: uint32(z)}
	var resp api.Encode3DResponse

	url := fmt.Sprintf("%s/encode3d", serverAddr)
	if err := postJSON(url, req, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Coordinate: (%d, %d, %d)\n", resp.Coordinate.X, resp.Coordinate.Y, resp.Coordinate.Z)
	fmt.Printf("Morton Code: %s\n", uint64ToString(uint64(resp.Code)))
}

func handleDecode3D(serverAddr string, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: client decode3d <code>")
		os.Exit(1)
	}

	codeStr := args[0]
	code, err := parseCode(codeStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid code: %v\n", err)
		os.Exit(1)
	}

	req := api.Decode3DRequest{Code: zorder.Code(code)}
	var resp api.Decode3DResponse

	url := fmt.Sprintf("%s/decode3d", serverAddr)
	if err := postJSON(url, req, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Morton Code: %s\n", uint64ToString(uint64(resp.Code)))
	fmt.Printf("Coordinate: (%d, %d, %d)\n", resp.Coordinate.X, resp.Coordinate.Y, resp.Coordinate.Z)
}

func handleQuery(serverAddr string, args []string) {
	if len(args) < 4 {
		fmt.Fprintln(os.Stderr, "Usage: client query <min_x> <min_y> <max_x> <max_y>")
		os.Exit(1)
	}

	minX, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid min_x value: %v\n", err)
		os.Exit(1)
	}

	minY, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid min_y value: %v\n", err)
		os.Exit(1)
	}

	maxX, err := strconv.ParseUint(args[2], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid max_x value: %v\n", err)
		os.Exit(1)
	}

	maxY, err := strconv.ParseUint(args[3], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid max_y value: %v\n", err)
		os.Exit(1)
	}

	req := api.QueryRangesRequest{
		MinX: uint32(minX),
		MinY: uint32(minY),
		MaxX: uint32(maxX),
		MaxY: uint32(maxY),
	}
	var resp api.QueryRangesResponse

	url := fmt.Sprintf("%s/query", serverAddr)
	if err := postJSON(url, req, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Query Rect: [(%d, %d), (%d, %d)]\n",
		resp.Rect.Min.X, resp.Rect.Min.Y,
		resp.Rect.Max.X, resp.Rect.Max.Y)
	fmt.Printf("Prefix Ranges Found: %d\n", resp.Count)
	fmt.Println()
	for i, r := range resp.Ranges {
		fmt.Printf("Range %d: [%s, %s]\n", i+1,
			uint64ToString(uint64(r.Start)),
			uint64ToString(uint64(r.End)))
	}
}

func handleHealth(serverAddr string) {
	url := fmt.Sprintf("%s/health", serverAddr)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Health check failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Server not healthy. Status: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func postJSON(url string, req, resp interface{}) error {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	if httpResp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", httpResp.StatusCode, string(respBody))
	}

	return json.Unmarshal(respBody, resp)
}

func parseCode(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"")
	return strconv.ParseUint(s, 10, 64)
}

func uint64ToString(n uint64) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append(buf, '0'+byte(n%10))
		n /= 10
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

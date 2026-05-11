package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"serial-system/internal/api"
)

func getServerURL() string {
	if envURL := os.Getenv("SERIAL_SERVER"); envURL != "" {
		return envURL
	}
	return "http://localhost:8080"
}

func postJSON(url string, req interface{}, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	httpResp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer httpResp.Body.Close()
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if err := json.Unmarshal(respBody, resp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s <command> [options]\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  next  - get next serial number")
	fmt.Fprintln(os.Stderr, "  batch - get N serial numbers")
	fmt.Fprintln(os.Stderr, "  check - check for gaps")
	fmt.Fprintln(os.Stderr, "  reset - reset serial (requires confirm=true)")
	os.Exit(2)
}

func cmdNext(args []string) {
	fs := flag.NewFlagSet("next", flag.ExitOnError)
	bizType := fs.String("biz", "order", "business type")
	fs.Parse(args)

	server := getServerURL()
	var resp api.NextResponse
	if err := postJSON(server+"/next", api.NextRequest{BizType: *bizType}, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if !resp.Success {
		fmt.Fprintf(os.Stderr, "server error: %s\n", resp.Error)
		os.Exit(1)
	}
	fmt.Println(resp.Serial)
}

func cmdBatch(args []string) {
	fs := flag.NewFlagSet("batch", flag.ExitOnError)
	bizType := fs.String("biz", "order", "business type")
	count := fs.Int("n", 1, "count of serial numbers")
	fs.Parse(args)

	if *count <= 0 {
		fmt.Fprintln(os.Stderr, "count must be positive")
		os.Exit(2)
	}

	server := getServerURL()
	var resp api.BatchResponse
	if err := postJSON(server+"/batch", api.BatchRequest{BizType: *bizType, Count: *count}, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if !resp.Success {
		fmt.Fprintf(os.Stderr, "server error: %s\n", resp.Error)
		os.Exit(1)
	}
	fmt.Printf("start: %s, end: %s\n", resp.Start, resp.End)
	for _, s := range resp.Serials {
		fmt.Println(s)
	}
}

func cmdCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	bizType := fs.String("biz", "order", "business type")
	fs.Parse(args)

	server := getServerURL()
	var resp api.CheckResponse
	if err := postJSON(server+"/check", api.CheckRequest{BizType: *bizType}, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if !resp.Success {
		fmt.Fprintf(os.Stderr, "server error: %s\n", resp.Error)
		os.Exit(1)
	}
	fmt.Printf("current_max: %d, total_allocated: %d\n", resp.CurrentMax, resp.TotalAllocated)
	if resp.HasGap {
		fmt.Printf("has gaps:\n")
		for _, g := range resp.Gaps {
			fmt.Printf("  gap [%d - %d]\n", g.Start, g.End)
		}
		os.Exit(1)
	} else {
		fmt.Println("no gaps detected")
	}
}

func cmdReset(args []string) {
	fs := flag.NewFlagSet("reset", flag.ExitOnError)
	bizType := fs.String("biz", "order", "business type")
	confirm := fs.Bool("confirm", false, "confirm reset")
	fs.Parse(args)

	server := getServerURL()
	var resp api.ResetResponse
	if err := postJSON(server+"/reset", api.ResetRequest{BizType: *bizType, Confirm: *confirm}, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if !resp.Success {
		fmt.Fprintf(os.Stderr, "server error: %s\n", resp.Error)
		os.Exit(1)
	}
	fmt.Printf("reset to %s\n", resp.Date)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "next":
		cmdNext(args)
	case "batch":
		cmdBatch(args)
	case "check":
		cmdCheck(args)
	case "reset":
		cmdReset(args)
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
	}
}

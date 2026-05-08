package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/example/timeseries-downsample/pkg/common"
)

const defaultServer = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "write":
		cmdWrite(os.Args[2:])
	case "query":
		cmdQuery(os.Args[2:])
	case "config":
		cmdConfig(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `usage: ts-client <command> [options]

commands:
  write   - write data points
  query   - query downsampled data
  config  - get server config
`)
}

func getServer() string {
	if s := os.Getenv("TS_SERVER"); s != "" {
		return s
	}
	return defaultServer
}

func cmdWrite(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: ts-client write <metric> <value> [value...]")
		fmt.Fprintln(os.Stderr, "       (values can be 'timestamp,value' pairs or just 'value' for current time)")
		os.Exit(1)
	}

	metric := args[0]
	values := args[1:]

	now := time.Now().UTC()
	points := make([]common.DataPoint, 0, len(values))
	for _, v := range values {
		if strings.Contains(v, ",") {
			parts := strings.SplitN(v, ",", 2)
			ts, err := time.Parse(time.RFC3339, parts[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid timestamp: %s\n", parts[0])
				os.Exit(1)
			}
			var f float64
			fmt.Sscanf(parts[1], "%f", &f)
			points = append(points, common.DataPoint{Timestamp: ts, Value: f})
		} else {
			var f float64
			fmt.Sscanf(v, "%f", &f)
			points = append(points, common.DataPoint{Timestamp: now, Value: f})
		}
	}

	req := common.WriteRequest{
		Metric: metric,
		Data:   points,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(getServer()+"/ts/data", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result common.WriteResponse
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("success=%v, count=%d, message=%s\n", result.Success, result.Count, result.Message)
}

func cmdQuery(args []string) {
	if len(args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: ts-client query <metric> <window> <from> <to>")
		fmt.Fprintln(os.Stderr, "       window: 1m, 5m, 1h, 1d")
		fmt.Fprintln(os.Stderr, "       from/to: RFC3339 format, e.g., 2024-01-01T00:00:00Z")
		os.Exit(1)
	}

	metric := args[0]
	window := args[1]
	from := args[2]
	to := args[3]

	url := fmt.Sprintf("%s/ts/downsample?metric=%s&window=%s&from=%s&to=%s",
		getServer(), metric, window, from, to)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func cmdConfig(args []string) {
	resp, err := http.Get(getServer() + "/ts/config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

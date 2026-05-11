package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ntp-client/pkg/api"
)

const defaultServerURL = "http://localhost:8202"

func main() {
	var (
		serverURL  string
		ntpServer  string
		ntpPort    int
		timeout    time.Duration
		syncFlag   bool
		historyFlag bool
	)

	flag.StringVar(&serverURL, "server-url", defaultServerURL, "NTP service HTTP server URL")
	flag.StringVar(&ntpServer, "ntp-server", "pool.ntp.org", "NTP server address")
	flag.IntVar(&ntpPort, "ntp-port", 123, "NTP server port")
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "query timeout")
	flag.BoolVar(&syncFlag, "sync", false, "sync local system time on server")
	flag.BoolVar(&historyFlag, "history", false, "show sync history")
	flag.Parse()

	if envURL := os.Getenv("NTP_SERVER_URL"); envURL != "" {
		serverURL = envURL
	}

	if historyFlag {
		if err := showHistory(serverURL); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if ntpServer == "" && flag.NArg() > 0 {
		ntpServer = flag.Arg(0)
		if idx := strings.Index(ntpServer, ":"); idx > 0 {
			ntpServer = flag.Arg(0)[:idx]
			fmt.Sscanf(flag.Arg(0)[idx+1:], "%d", &ntpPort)
		}
	}

	if ntpServer == "" {
		fmt.Fprintln(os.Stderr, "usage: ntp-client [flags] <ntp-server[:port]>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if syncFlag {
		if err := doSync(serverURL, ntpServer, ntpPort, timeout); err != nil {
			fmt.Fprintf(os.Stderr, "sync error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := doQuery(serverURL, ntpServer, ntpPort, timeout); err != nil {
		fmt.Fprintf(os.Stderr, "query error: %v\n", err)
		os.Exit(1)
	}
}

func doQuery(serverURL, ntpServer string, ntpPort int, timeout time.Duration) error {
	req := api.QueryRequest{
		Server:  ntpServer,
		Port:    ntpPort,
		Timeout: timeout,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL+"/query", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, errResp.Error)
	}

	var result api.QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	printQueryResult(&result)
	return nil
}

func doSync(serverURL, ntpServer string, ntpPort int, timeout time.Duration) error {
	req := api.SyncRequest{
		Server:  ntpServer,
		Port:    ntpPort,
		Timeout: timeout,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL+"/sync", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, errResp.Error)
	}

	var result api.SyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	printSyncResult(&result)
	return nil
}

func showHistory(serverURL string) error {
	resp, err := http.Get(serverURL + "/history")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, errResp.Error)
	}

	var result api.HistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	printHistory(&result)
	return nil
}

func printQueryResult(r *api.QueryResponse) {
	fmt.Printf("Server:         %s\n", r.Server)
	fmt.Printf("Server Time:    %s\n", r.ServerTime.Format("2006-01-02 15:04:05.000000"))
	fmt.Printf("Time Offset:    %s\n", formatDuration(r.Offset))
	fmt.Printf("Round-Trip:     %s\n", formatDuration(r.Delay))
	fmt.Printf("Stratum:        %d\n", r.Stratum)
	fmt.Printf("Poll Interval:  %.0f seconds\n", r.PollSec)
	fmt.Printf("Clock Precision: %.6f seconds\n", r.PrecisionSec)

	localTime := time.Now()
	actual := localTime.Add(r.Offset)
	diff := actual.Sub(localTime)
	fmt.Println()
	if diff.Abs() < 10*time.Millisecond {
		fmt.Printf("Local time is within 10ms of NTP time (%.0f µs off)\n", diff.Abs().Microseconds())
	} else if diff.Abs() < 100*time.Millisecond {
		fmt.Printf("Local time is slightly off: %s (consider adjustment)\n", formatDuration(diff))
	} else {
		fmt.Printf("Local time needs adjustment: %s\n", formatDuration(diff))
		fmt.Println("Recommendation: run with -sync flag to synchronize")
	}
}

func printSyncResult(r *api.SyncResponse) {
	fmt.Printf("Server:         %s\n", r.Server)
	fmt.Printf("Server Time:    %s\n", r.ServerTime.Format("2006-01-02 15:04:05.000000"))
	fmt.Printf("Local Time:     %s\n", r.LocalTime.Format("2006-01-02 15:04:05.000000"))
	fmt.Printf("Time Offset:    %s\n", formatDuration(r.Offset))
	fmt.Printf("Round-Trip:     %s\n", formatDuration(r.Delay))
	fmt.Println()
	if r.Adjusted {
		fmt.Println("✓ System time has been adjusted successfully")
	} else {
		fmt.Printf("! %s\n", r.Message)
	}
}

func printHistory(r *api.HistoryResponse) {
	if len(r.Entries) == 0 {
		fmt.Println("No sync history available")
		return
	}

	fmt.Printf("Sync History (%d entries, most recent last):\n\n", len(r.Entries))
	for i, e := range r.Entries {
		fmt.Printf("[%d] %s\n", i+1, e.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("    Server:  %s\n", e.Server)
		fmt.Printf("    Offset:  %s\n", formatDuration(e.Offset))
		fmt.Printf("    Delay:   %s\n", formatDuration(e.Delay))
		fmt.Printf("    Stratum: %d\n", e.Stratum)
		fmt.Println()
	}
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0 µs"
	}

	sign := ""
	if d < 0 {
		sign = "-"
		d = -d
	}

	switch {
	case d >= time.Second:
		return fmt.Sprintf("%s%.3f s", sign, d.Seconds())
	case d >= time.Millisecond:
		return fmt.Sprintf("%s%.3f ms", sign, float64(d)/float64(time.Millisecond))
	case d >= time.Microsecond:
		return fmt.Sprintf("%s%.0f µs", sign, float64(d)/float64(time.Microsecond))
	default:
		return fmt.Sprintf("%s%d ns", sign, d.Nanoseconds())
	}
}

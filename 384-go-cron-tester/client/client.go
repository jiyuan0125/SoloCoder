package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"

	"cron-tester/internal/protocol"
)

type Client struct {
	host string
	port int
}

func NewClient(host string, port int) *Client {
	return &Client{
		host: host,
		port: port,
	}
}

func (c *Client) connect() (net.Conn, error) {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	return net.Dial("tcp", addr)
}

func (c *Client) sendRequest(request interface{}, response interface{}) error {
	conn, err := c.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	reqBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	_, err = conn.Write(reqBytes)
	if err != nil {
		return err
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf[:n], response)
}

func (c *Client) Validate(expr string) (*protocol.ValidateResponse, error) {
	req := protocol.ValidateRequest{
		Type: protocol.RequestTypeValidate,
		Expr: expr,
	}

	var resp protocol.ValidateResponse
	if err := c.sendRequest(&req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) NextTime(expr string, verbose bool) (*protocol.NextTimeResponse, error) {
	req := protocol.NextTimeRequest{
		Type:    protocol.RequestTypeNextTime,
		Expr:    expr,
		Verbose: verbose,
	}

	var resp protocol.NextTimeResponse
	if err := c.sendRequest(&req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) NextNTimes(expr string, n int, verbose bool) (*protocol.NextNTimesResponse, error) {
	req := protocol.NextNTimesRequest{
		Type:    protocol.RequestTypeNextNTimes,
		Expr:    expr,
		N:       n,
		Verbose: verbose,
	}

	var resp protocol.NextNTimesResponse
	if err := c.sendRequest(&req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func formatDuration(duration interface{}) string {
	switch v := duration.(type) {
	case float64:
		return formatDurationFloat64(v)
	case int64:
		return formatDurationInt64(v)
	case int:
		return formatDurationInt64(int64(v))
	default:
		return ""
	}
}

func formatDurationFloat64(d float64) string {
	duration := int64(d)
	return formatDurationInt64(duration)
}

func formatDurationInt64(duration int64) string {
	if duration < 0 {
		duration = -duration
	}

	seconds := duration / 1e9
	nanos := duration % 1e9

	if seconds == 0 {
		if nanos < 1e6 {
			return fmt.Sprintf("%dµs", nanos/1e3)
		}
		return fmt.Sprintf("%dms", nanos/1e6)
	}

	if seconds < 60 {
		if nanos > 5e8 {
			seconds++
		}
		return fmt.Sprintf("%ds", seconds)
	}

	minutes := seconds / 60
	remainingSeconds := seconds % 60

	if minutes < 60 {
		if remainingSeconds > 30 {
			minutes++
		}
		return fmt.Sprintf("%dm", minutes)
	}

	hours := minutes / 60
	remainingMinutes := minutes % 60

	if hours < 24 {
		if remainingMinutes > 30 {
			hours++
		}
		return fmt.Sprintf("%dh", hours)
	}

	days := hours / 24
	remainingHours := hours % 24

	if remainingHours > 12 {
		days++
	}
	return fmt.Sprintf("%dd", days)
}

func RunClient() {
	validateFlag := flag.Bool("validate", false, "Validate cron expression syntax")
	verboseFlag := flag.Bool("verbose", false, "Show time interval from now")
	oneOffFlag := flag.Bool("one-off", false, "Show only the next execution time")
	nFlag := flag.Int("n", 10, "Number of next execution times to show (default 10)")
	portFlag := flag.Int("port", protocol.DefaultPort, "Server port")
	hostFlag := flag.String("host", "localhost", "Server host")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <cron-expression>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s \"0 0 2 * * *\"\n", os.Args[0])
		os.Exit(1)
	}

	expr := args[0]

	client := NewClient(*hostFlag, *portFlag)

	if *validateFlag {
		resp, err := client.Validate(expr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if resp.Valid {
			fmt.Println("Expression is valid.")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
			os.Exit(1)
		}
		return
	}

	if *oneOffFlag {
		resp, err := client.NextTime(expr, *verboseFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if !resp.Valid {
			fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
			os.Exit(1)
		}

		if resp.WillNeverFire {
			fmt.Println("This expression will never fire.")
			return
		}

		if resp.NextTime != nil {
			if *verboseFlag && resp.NextTime.Delay != 0 {
				fmt.Printf("%s (in %s)\n", resp.NextTime.Time, formatDuration(resp.NextTime.Delay))
			} else {
				fmt.Println(resp.NextTime.Time)
			}
		}
		return
	}

	resp, err := client.NextNTimes(expr, *nFlag, *verboseFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !resp.Valid {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}

	if resp.WillNeverFire && resp.Count == 0 {
		fmt.Println("This expression will never fire.")
		return
	}

	for i, t := range resp.NextTimes {
		if *verboseFlag && t.Delay != 0 {
			fmt.Printf("%d: %s (in %s)\n", i+1, t.Time, formatDuration(t.Delay))
		} else {
			fmt.Printf("%d: %s\n", i+1, t.Time)
		}
	}

	if resp.WillNeverFire && resp.Count < *nFlag {
		fmt.Printf("\nWarning: Only %d execution times found before the expression will no longer fire.\n", resp.Count)
	}
}

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"cms/api"
)

func cmdCreate(serverURL string, args []string) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	name := fs.String("name", "", "sketch name")
	width := fs.Int("width", 0, "width")
	depth := fs.Int("depth", 0, "depth")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("name is required")
	}

	req := api.CreateRequest{
		Name:  *name,
		Width: *width,
		Depth: *depth,
	}

	var resp api.CreateResponse
	if err := sendRequest(serverURL+"/create", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	fmt.Println(resp.Message)
	return nil
}

func cmdAdd(serverURL string, args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	name := fs.String("name", "", "sketch name")
	item := fs.String("item", "", "item to add")
	count := fs.Int64("count", 1, "count")
	filePath := fs.String("file", "", "file path for batch add")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("name is required")
	}

	var items []api.AddItem

	if *filePath != "" {
		f, err := os.Open(*filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) == 1 {
				items = append(items, api.AddItem{Item: parts[0], Count: 1})
			} else if len(parts) >= 2 {
				c, err := strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					c = 1
				}
				items = append(items, api.AddItem{Item: parts[0], Count: c})
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	}

	if *item != "" {
		items = append(items, api.AddItem{Item: *item, Count: *count})
	}

	if len(items) == 0 {
		return fmt.Errorf("no items to add")
	}

	req := api.AddRequest{
		Name:  *name,
		Items: items,
	}

	var resp api.AddResponse
	if err := sendRequest(serverURL+"/add", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	fmt.Printf("Added %d items\n", len(items))
	return nil
}

func cmdQuery(serverURL string, args []string) error {
	fs := flag.NewFlagSet("query", flag.ContinueOnError)
	name := fs.String("name", "", "sketch name")
	item := fs.String("item", "", "item to query")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("name is required")
	}

	items := []string{*item}
	if *item == "" {
		items = fs.Args()
	}

	if len(items) == 0 {
		return fmt.Errorf("no items to query")
	}

	req := api.QueryRequest{
		Name:  *name,
		Items: items,
	}

	var resp api.QueryResponse
	if err := sendRequest(serverURL+"/query", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	for _, r := range resp.Results {
		fmt.Printf("%s: %d (error_bound: %.2f)\n", r.Item, r.Frequency, r.ErrorBound)
	}
	return nil
}

func cmdMerge(serverURL string, args []string) error {
	fs := flag.NewFlagSet("merge", flag.ContinueOnError)
	source := fs.String("source", "", "source sketch")
	target := fs.String("target", "", "target sketch")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *source == "" || *target == "" {
		return fmt.Errorf("source and target are required")
	}

	req := api.MergeRequest{
		Source: *source,
		Target: *target,
	}

	var resp api.MergeResponse
	if err := sendRequest(serverURL+"/merge", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	fmt.Println(resp.Message)
	return nil
}

func cmdStats(serverURL string, args []string) error {
	fs := flag.NewFlagSet("stats", flag.ContinueOnError)
	name := fs.String("name", "", "sketch name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	req := api.StatsRequest{
		Name: *name,
	}

	var resp api.StatsResponse

	url := serverURL + "/stats"
	if *name == "" {
		body, err := httpGet(url)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return err
		}
	} else {
		if err := sendRequest(url, req, &resp); err != nil {
			return err
		}
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	for _, s := range resp.Sketches {
		fmt.Printf("%s:\n", s.Name)
		fmt.Printf("  Width: %d\n", s.Width)
		fmt.Printf("  Depth: %d\n", s.Depth)
		fmt.Printf("  Total Count: %d\n", s.TotalCount)
	}
	return nil
}

func sendRequest(url string, reqBody, respBody interface{}) error {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return fmt.Errorf(errResp.Error)
		}
		return fmt.Errorf("server error: %d", resp.StatusCode)
	}

	if err := json.Unmarshal(body, respBody); err != nil {
		return err
	}
	return nil
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

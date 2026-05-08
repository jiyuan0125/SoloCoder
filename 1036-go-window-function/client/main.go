package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"windowfunc/api"
)

func makeURL(base, path string) string {
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + base
	}
	return base + path
}

func doPOST(url string, body interface{}, out interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(rb, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(rb))
	}
	if out != nil {
		return json.Unmarshal(rb, out)
	}
	return nil
}

func doGET(url string, out interface{}) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(rb, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(rb))
	}
	return json.Unmarshal(rb, out)
}

func cmdList(server string) error {
	var resp api.ListResponse
	if err := doGET(makeURL(server, "/list"), &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("list failed")
	}
	if len(resp.Datasets) == 0 {
		fmt.Println("No datasets found.")
		return nil
	}
	rows := make([]api.Record, len(resp.Datasets))
	for i, d := range resp.Datasets {
		rows[i] = api.Record{
			"name": d.Name,
			"size": d.Size,
		}
	}
	printTable(rows)
	return nil
}

func cmdUpload(server string, dataset string, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	var data []api.Record
	if err := json.Unmarshal(b, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}
	req := api.UploadRequest{
		Dataset: dataset,
		Data:    data,
	}
	var resp api.UploadResponse
	if err := doPOST(makeURL(server, "/upload"), req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("upload failed")
	}
	fmt.Println(resp.Message)
	return nil
}

func parseSortOrder(s string) api.SortOrder {
	switch strings.ToUpper(s) {
	case "DESC", "D":
		return api.SortDesc
	default:
		return api.SortAsc
	}
}

func parseNullsOrder(s string) api.NullsOrder {
	switch strings.ToUpper(s) {
	case "FIRST", "F":
		return api.NullsFirst
	default:
		return api.NullsLast
	}
}

func parseFuncSpecs(specs []string) ([]api.WindowFunction, error) {
	funcs := make([]api.WindowFunction, 0, len(specs))
	for _, spec := range specs {
		parts := strings.Split(spec, ";")
		fn := api.WindowFunction{
			Order:      api.SortAsc,
			NullsOrder: api.NullsLast,
			Offset:     1,
		}
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			kv := strings.SplitN(p, "=", 2)
			if len(kv) != 2 {
				return nil, fmt.Errorf("invalid function spec part: %s (expected key=value)", p)
			}
			key := strings.ToLower(strings.TrimSpace(kv[0]))
			value := strings.TrimSpace(kv[1])
			switch key {
			case "name":
				fn.Name = api.WindowFuncType(strings.ToUpper(value))
			case "partition_by", "partition":
				fn.PartitionBy = value
			case "order_by", "order":
				fn.OrderBy = value
			case "sort_order", "dir":
				fn.Order = parseSortOrder(value)
			case "nulls":
				fn.NullsOrder = parseNullsOrder(value)
			case "offset", "n":
				var n int
				if _, err := fmt.Sscanf(value, "%d", &n); err != nil {
					return nil, fmt.Errorf("invalid offset: %s", value)
				}
				fn.Offset = n
			case "field", "col":
				fn.Field = value
			case "alias", "as":
				fn.Alias = value
			default:
				return nil, fmt.Errorf("unknown key in function spec: %s", key)
			}
		}
		if fn.Name == "" {
			return nil, fmt.Errorf("function name is required in spec: %s", spec)
		}
		funcs = append(funcs, fn)
	}
	return funcs, nil
}

func cmdQuery(server string, dataset string, specs []string) error {
	funcs, err := parseFuncSpecs(specs)
	if err != nil {
		return err
	}
	req := api.QueryRequest{
		Dataset:   dataset,
		Functions: funcs,
	}
	var resp api.QueryResponse
	if err := doPOST(makeURL(server, "/query"), req, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("query failed")
	}
	printTable(resp.Data)
	return nil
}

func usage() {
	fmt.Print(`windowfunc client - command line tool for window function calculations

Usage:
  client [options] list
  client [options] upload -dataset <name> -file <data.json>
  client [options] query -dataset <name> -func <spec> [-func <spec> ...]

Function spec format: name=...[;partition_by=...;order_by=...;sort_order=ASC|DESC;nulls=FIRST|LAST;offset=N;field=...;alias=...]

Examples:
  client upload -dataset sales -file data.json
  client query -dataset sales -func "name=ROW_NUMBER;order_by=amount;sort_order=DESC"
  client query -dataset sales -func "name=RANK;partition_by=region;order_by=amount"
  client query -dataset sales -func "name=LAG;order_by=date;field=amount;offset=1"

Options:
  -server <addr>    Server address (default: localhost:8080)
`)
}

func main() {
	fs := flag.NewFlagSet("client", flag.ExitOnError)
	server := fs.String("server", "localhost:8080", "server address")
	dataset := fs.String("dataset", "", "dataset name")
	file := fs.String("file", "", "data file path for upload")
	funcs := stringArray{}
	fs.Var(&funcs, "func", "window function spec")

	fs.Usage = usage

	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	cmd := args[0]
	fs.Parse(args[1:])

	switch cmd {
	case "list":
		if err := cmdList(*server); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "upload":
		if *dataset == "" {
			fmt.Fprintln(os.Stderr, "Error: -dataset is required")
			os.Exit(1)
		}
		if *file == "" {
			fmt.Fprintln(os.Stderr, "Error: -file is required")
			os.Exit(1)
		}
		if err := cmdUpload(*server, *dataset, *file); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "query":
		if *dataset == "" {
			fmt.Fprintln(os.Stderr, "Error: -dataset is required")
			os.Exit(1)
		}
		if len(funcs) == 0 {
			fmt.Fprintln(os.Stderr, "Error: at least one -func is required")
			os.Exit(1)
		}
		if err := cmdQuery(*server, *dataset, funcs); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

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
	"path/filepath"

	"naturalsort/pkg/api"
)

func main() {
	server := flag.String("server", "http://localhost:8080", "sort server URL")
	desc := flag.Bool("desc", false, "sort descending")
	caseSensitive := flag.Bool("case-sensitive", false, "case sensitive sort")
	keepLeadingZeros := flag.Bool("keep-leading-zeros", false, "preserve leading zeros in comparison")
	files := flag.Bool("files", false, "sort file list (reads positional args)")
	stdin := flag.Bool("stdin", false, "sort lines from stdin")
	flag.Parse()

	items := collectInput(*files, *stdin)

	req := api.SortRequest{
		Strings:            items,
		Ascending:          !*desc,
		CaseSensitive:      *caseSensitive,
		IgnoreLeadingZeros: !*keepLeadingZeros,
	}

	sorted, err := callServer(*server, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for _, s := range sorted {
		fmt.Println(s)
	}
}

func collectInput(filesMode, stdinMode bool) []string {
	if stdinMode {
		return readLines(os.Stdin)
	}

	if filesMode {
		args := flag.Args()
		if len(args) > 0 {
			return listFiles(args)
		}
		return listCurrentDir()
	}

	args := flag.Args()
	if len(args) > 0 {
		return args
	}

	return readLines(os.Stdin)
}

func readLines(r io.Reader) []string {
	var lines []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

func listCurrentDir() []string {
	files, err := os.ReadDir(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading current directory: %v\n", err)
		os.Exit(1)
	}
	var names []string
	for _, f := range files {
		names = append(names, f.Name())
	}
	return names
}

func listFiles(paths []string) []string {
	var names []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			names = append(names, p)
			continue
		}
		if info.IsDir() {
			entries, err := os.ReadDir(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading directory %s: %v\n", p, err)
				os.Exit(1)
			}
			for _, e := range entries {
				names = append(names, filepath.Join(p, e.Name()))
			}
		} else {
			names = append(names, p)
		}
	}
	return names
}

func callServer(server string, req api.SortRequest) ([]string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := http.Post(server+"/sort", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("connect to server: %w", err)
	}
	defer resp.Body.Close()

	var result api.SortResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("server error: %s", result.Error)
	}

	return result.Strings, nil
}

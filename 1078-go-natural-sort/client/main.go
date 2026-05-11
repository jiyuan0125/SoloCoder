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

	"naturalsort/common"
)

const defaultServer = "http://localhost:8710"

func main() {
	serverURL := flag.String("server", defaultServer, "Sort server URL")
	caseSensitive := flag.Bool("case-sensitive", false, "Case sensitive comparison")
	descending := flag.Bool("descending", false, "Sort in descending order")
	keepLeadingZeros := flag.Bool("keep-leading-zeros", false, "Keep leading zeros in comparison")
	files := flag.Bool("files", false, "Treat arguments as files to list and sort")
	flag.Parse()

	var input []string

	if *files {
		args := flag.Args()
		if len(args) == 0 {
			args = []string{"."}
		}
		for _, dir := range args {
			entries, err := os.ReadDir(dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", dir, err)
				os.Exit(1)
			}
			for _, entry := range entries {
				input = append(input, filepath.Join(dir, entry.Name()))
			}
		}
	} else if flag.NArg() > 0 {
		input = flag.Args()
	} else {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					if len(line) > 0 {
						input = append(input, line)
					}
					break
				}
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
				os.Exit(1)
			}
			if len(line) > 0 && line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			input = append(input, line)
		}
	}

	if len(input) == 0 {
		fmt.Fprintln(os.Stderr, "No input provided")
		os.Exit(1)
	}

	req := common.SortRequest{
		Strings:          input,
		CaseSensitive:    *caseSensitive,
		Descending:       *descending,
		KeepLeadingZeros: *keepLeadingZeros,
	}

	sorted, err := sendRequest(*serverURL, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, s := range sorted {
		fmt.Println(s)
	}
}

func sendRequest(serverURL string, req common.SortRequest) ([]string, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(serverURL+"/sort", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("server error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var sortResp common.SortResponse
	if err := json.NewDecoder(resp.Body).Decode(&sortResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return sortResp.Sorted, nil
}

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"vfs/common"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &Client{baseURL: baseURL}
}

func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func (c *Client) getJSON(path string, result interface{}) error {
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("error %d: %s", resp.StatusCode, errResp.Error)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) List(path string) (*common.ListDirResponse, error) {
	urlPath := "/files"
	if path != "." && path != "" {
		urlPath = "/files/" + strings.TrimPrefix(path, "/")
	}
	var result common.ListDirResponse
	err := c.getJSON(urlPath, &result)
	return &result, err
}

func (c *Client) Cat(path string) ([]byte, error) {
	urlPath := "/files/" + strings.TrimPrefix(path, "/")
	resp, err := c.doRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("error %d: %s", resp.StatusCode, errResp.Error)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) Tree(path string) (*common.TreeResponse, error) {
	urlPath := "/tree"
	if path != "." && path != "" {
		urlPath = "/tree?path=" + strings.TrimPrefix(path, "/")
	}
	var result common.TreeResponse
	err := c.getJSON(urlPath, &result)
	return &result, err
}

func (c *Client) Diff() (*common.DiffResponse, error) {
	var result common.DiffResponse
	err := c.getJSON("/diff", &result)
	return &result, err
}

func (c *Client) Reload() (*common.ReloadResponse, error) {
	resp, err := c.doRequest(http.MethodPost, "/reload", bytes.NewReader(nil))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("error %d: %s", resp.StatusCode, errResp.Error)
	}

	var result common.ReloadResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}

func printList(resp *common.ListDirResponse) {
	fmt.Printf("Directory: %s\n\n", resp.Path)
	for _, entry := range resp.Entries {
		marker := " "
		if entry.IsDir {
			marker = "/"
		}
		fmt.Printf("%s%s\n", entry.Name, marker)
	}
}

func printTree(node *common.TreeNodeResponse, prefix string, isLast bool) {
	if node == nil {
		return
	}

	currentPrefix := ""
	if prefix != "" {
		if isLast {
			currentPrefix = prefix[:len(prefix)-4] + "    └── "
		} else {
			currentPrefix = prefix[:len(prefix)-4] + "    ├── "
		}
	}

	marker := " "
	if node.IsDir {
		marker = "/"
	}

	fmt.Printf("%s%s%s\n", currentPrefix, node.Name, marker)

	for i, child := range node.Children {
		nextPrefix := prefix
		if isLast {
			nextPrefix += "    "
		} else {
			nextPrefix += "│   "
		}
		printTree(child, nextPrefix, i == len(node.Children)-1)
	}
}

func printDiff(resp *common.DiffResponse) {
	if len(resp.Diffs) == 0 {
		fmt.Println("No differences found.")
		return
	}

	for _, d := range resp.Diffs {
		statusColor := ""
		switch d.Status {
		case "added":
			statusColor = "+ "
		case "removed":
			statusColor = "- "
		case "modified":
			statusColor = "M "
		default:
			statusColor = "? "
		}
		fmt.Printf("%s%s - %s\n", statusColor, d.Path, d.Details)
	}
}

func getServerURL() string {
	url := os.Getenv("VFS_SERVER")
	if url == "" {
		flagURL := flag.String("server", "localhost:8420", "Server address (host:port)")
		flag.Parse()
		url = *flagURL
	}
	return url
}

func printUsage() {
	fmt.Println(`VFS Client - Command line tool for interacting with VFS Server

Usage:
  vfs-client [flags] <command> [arguments]

Commands:
  ls [path]       List directory contents
  cat <path>      Output file content
  tree [path]     Display directory tree
  diff            Show differences between disk and embedded versions
  reload          Reload disk overlay cache

Flags:
  -server string  Server address (default "localhost:8420")
                  Can also be set via VFS_SERVER environment variable

Examples:
  vfs-client ls
  vfs-client ls /subdir
  vfs-client cat /test1.txt
  vfs-client tree
  vfs-client diff
  vfs-client reload
`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := "localhost:8420"
	if envURL := os.Getenv("VFS_SERVER"); envURL != "" {
		serverURL = envURL
	}

	var args []string
	cmd := ""
	for i, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			if arg == "-server" && i+1 < len(os.Args[1:]) {
				serverURL = os.Args[i+2]
				os.Args = append(os.Args[:i+1], os.Args[i+3:]...)
				break
			}
		} else if cmd == "" {
			cmd = arg
		} else {
			args = append(args, arg)
		}
	}

	if cmd == "" {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(serverURL)

	switch cmd {
	case "ls":
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		resp, err := client.List(path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printList(resp)

	case "cat":
		if len(args) < 1 {
			fmt.Println("Error: cat requires a file path")
			os.Exit(1)
		}
		data, err := client.Cat(args[0])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(string(data))

	case "tree":
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		resp, err := client.Tree(path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printTree(resp.Root, "", true)

	case "diff":
		resp, err := client.Diff()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printDiff(resp)

	case "reload":
		resp, err := client.Reload()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Success: %s\n", resp.Message)

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

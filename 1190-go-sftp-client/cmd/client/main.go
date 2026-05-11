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
	"strconv"
	"strings"

	"github.com/example/sftp-client/pkg/api"
)

type Client struct {
	serverURL    string
	connectionID string
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: serverURL,
	}
}

func (c *Client) connect(cfg *api.ConnectionConfig) error {
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	resp, err := http.Post(c.serverURL+"/api/config", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	var result api.ConnectionConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("connection failed: %s", result.Message)
	}

	c.connectionID = result.ID
	fmt.Printf("Connected successfully. Connection ID: %s\n", result.ID)
	return nil
}

func (c *Client) listDir(path string) error {
	if c.connectionID == "" {
		return fmt.Errorf("not connected")
	}

	req := api.FileOperationRequest{
		ConnectionID: c.connectionID,
		Operation:    "list",
		RemotePath:   path,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.serverURL+"/api/file", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result api.FileOperationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("list failed: %s", result.Message)
	}

	dataBytes, err := json.Marshal(result.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	var entries []api.DirEntry
	if err := json.Unmarshal(dataBytes, &entries); err != nil {
		return fmt.Errorf("failed to unmarshal entries: %w", err)
	}

	for _, entry := range entries {
		typeChar := "-"
		if entry.IsDir {
			typeChar = "d"
		} else if entry.IsSymlink {
			typeChar = "l"
		}
		fmt.Printf("%s %10d %s\n", typeChar, entry.Size, entry.Filename)
	}

	return nil
}

func (c *Client) stat(path string) error {
	if c.connectionID == "" {
		return fmt.Errorf("not connected")
	}

	req := api.FileOperationRequest{
		ConnectionID: c.connectionID,
		Operation:    "stat",
		RemotePath:   path,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.serverURL+"/api/file", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result api.FileOperationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("stat failed: %s", result.Message)
	}

	dataBytes, err := json.Marshal(result.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	var attrs api.FileAttrs
	if err := json.Unmarshal(dataBytes, &attrs); err != nil {
		return fmt.Errorf("failed to unmarshal attrs: %w", err)
	}

	fmt.Printf("Path: %s\n", path)
	fmt.Printf("Size: %d bytes\n", attrs.Size)
	fmt.Printf("Permissions: %04o\n", attrs.Perms&07777)
	fmt.Printf("Type: ")
	if attrs.IsDir {
		fmt.Println("directory")
	} else if attrs.IsSymlink {
		fmt.Println("symlink")
	} else {
		fmt.Println("file")
	}

	return nil
}

func (c *Client) upload(localPath, remotePath string) error {
	if c.connectionID == "" {
		return fmt.Errorf("not connected")
	}

	absLocal, err := filepath.Abs(localPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if remotePath == "" {
		remotePath = filepath.Base(localPath)
	}

	req := api.FileOperationRequest{
		ConnectionID: c.connectionID,
		Operation:    "upload",
		LocalPath:    absLocal,
		RemotePath:   remotePath,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.serverURL+"/api/file", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result api.FileOperationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("upload failed: %s", result.Message)
	}

	fmt.Println("Upload completed successfully")
	return nil
}

func (c *Client) download(remotePath, localPath string) error {
	if c.connectionID == "" {
		return fmt.Errorf("not connected")
	}

	if localPath == "" {
		localPath = filepath.Base(remotePath)
	}

	req := api.FileOperationRequest{
		ConnectionID: c.connectionID,
		Operation:    "download",
		RemotePath:   remotePath,
		LocalPath:    localPath,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.serverURL+"/api/file", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result api.FileOperationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("download failed: %s", result.Message)
	}

	fmt.Println("Download completed successfully")
	return nil
}

func (c *Client) view(remotePath string) error {
	if c.connectionID == "" {
		return fmt.Errorf("not connected")
	}

	url := fmt.Sprintf("%s/api/view?connection_id=%s&path=%s", c.serverURL, c.connectionID, remotePath)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("view failed: %s", string(body))
	}

	io.Copy(os.Stdout, resp.Body)
	return nil
}

func (c *Client) interactive() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("SFTP CLI Client")
	fmt.Println("Type 'help' for available commands")

	for {
		fmt.Printf("sftp> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "help", "?":
			fmt.Println("Available commands:")
			fmt.Println("  connect <host> <port> <user> [password|key_path]  - Connect to SSH server")
			fmt.Println("  ls <path>                                         - List directory")
			fmt.Println("  stat <path>                                       - Get file/directory info")
			fmt.Println("  put <local_path> [remote_path]                    - Upload file")
			fmt.Println("  get <remote_path> [local_path]                    - Download file")
			fmt.Println("  view <remote_path>                                - View file content")
			fmt.Println("  help, ?                                           - Show this help")
			fmt.Println("  quit, exit                                        - Exit")

		case "connect":
			if len(args) < 3 {
				fmt.Println("Usage: connect <host> <port> <user> [password|key_path]")
				continue
			}

			host := args[0]
			portStr := args[1]
			user := args[2]

			port, err := strconv.Atoi(portStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid port: %v\n", err)
				continue
			}

			cfg := &api.ConnectionConfig{
				Host: host,
				Port: port,
				User: user,
			}

			if len(args) >= 4 {
				arg := args[3]
				if strings.HasPrefix(arg, "~") || strings.HasPrefix(arg, "/") || strings.HasPrefix(arg, ".") {
					cfg.KeyPath = arg
				} else {
					cfg.Password = arg
				}
			}

			if err := c.connect(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
			}

		case "ls":
			path := "."
			if len(args) >= 1 {
				path = args[0]
			}
			if err := c.listDir(path); err != nil {
				fmt.Fprintf(os.Stderr, "List error: %v\n", err)
			}

		case "stat":
			if len(args) < 1 {
				fmt.Println("Usage: stat <path>")
				continue
			}
			if err := c.stat(args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "Stat error: %v\n", err)
			}

		case "put":
			if len(args) < 1 {
				fmt.Println("Usage: put <local_path> [remote_path]")
				continue
			}
			local := args[0]
			remote := ""
			if len(args) >= 2 {
				remote = args[1]
			}
			if err := c.upload(local, remote); err != nil {
				fmt.Fprintf(os.Stderr, "Upload error: %v\n", err)
			}

		case "get":
			if len(args) < 1 {
				fmt.Println("Usage: get <remote_path> [local_path]")
				continue
			}
			remote := args[0]
			local := ""
			if len(args) >= 2 {
				local = args[1]
			}
			if err := c.download(remote, local); err != nil {
				fmt.Fprintf(os.Stderr, "Download error: %v\n", err)
			}

		case "view":
			if len(args) < 1 {
				fmt.Println("Usage: view <remote_path>")
				continue
			}
			if err := c.view(args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "View error: %v\n", err)
			}

		case "quit", "exit":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", cmd)
		}
	}
}

func main() {
	serverURL := flag.String("server", "http://localhost:8204", "SFTP server URL")
	flag.Parse()

	client := NewClient(*serverURL)
	client.interactive()
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"ftp-client/common"
)

type Session struct {
	config     common.FTPConfig
	apiClient  *APIClient
	currentDir string
	connected  bool
}

func main() {
	serverURL := flag.String("server", "http://localhost:8202", "FTP server API URL")
	flag.Parse()

	session := &Session{
		apiClient: NewAPIClient(*serverURL),
		config: common.FTPConfig{
			Mode: common.ModePassive,
		},
	}

	fmt.Println("FTP Client CLI")
	fmt.Println("Type 'help' for commands, 'exit' to quit")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		prompt := "ftp> "
		if session.connected {
			prompt = fmt.Sprintf("ftp [%s]> ", session.currentDir)
		}
		fmt.Print(prompt)

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "exit", "quit", "bye":
			fmt.Println("Goodbye!")
			return
		case "help", "?":
			showHelp()
		case "open", "connect":
			handleOpen(session, parts)
		case "close", "disconnect":
			handleClose(session)
		case "user", "login":
			handleUser(session, parts)
		case "ls", "dir":
			handleList(session, parts)
		case "cd":
			handleCd(session, parts)
		case "pwd":
			handlePwd(session)
		case "mkdir", "md":
			handleMkdir(session, parts)
		case "rmdir", "rd":
			handleRmdir(session, parts)
		case "get", "recv":
			handleGet(session, parts)
		case "put", "send":
			handlePut(session, parts)
		case "mget":
			handleMget(session, parts)
		case "mput":
			handleMput(session, parts)
		case "queue", "jobs":
			handleQueue(session)
		case "cancel":
			handleCancel(session, parts)
		case "passive":
			handlePassive(session, parts)
		case "binary", "bin":
			fmt.Println("Binary mode set (default)")
		case "status":
			handleStatus(session)
		default:
			fmt.Printf("Unknown command: %s\n", cmd)
			fmt.Println("Type 'help' for available commands")
		}
	}
}

func showHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  open <host>[:port]          - Connect to FTP server")
	fmt.Println("  user <username> [password]  - Login with credentials")
	fmt.Println("  close / disconnect          - Close connection")
	fmt.Println("  ls / dir [path]             - List directory contents")
	fmt.Println("  cd <path>                   - Change directory")
	fmt.Println("  pwd                         - Print working directory")
	fmt.Println("  mkdir / md <path>           - Create directory")
	fmt.Println("  rmdir / rd <path>           - Remove directory")
	fmt.Println("  get / recv <remote> [local] - Download file")
	fmt.Println("  put / send <local> [remote] - Upload file")
	fmt.Println("  mget <pattern>              - Download multiple files")
	fmt.Println("  mput <pattern>              - Upload multiple files")
	fmt.Println("  queue / jobs                - Show transfer queue")
	fmt.Println("  cancel <id>                 - Cancel transfer")
	fmt.Println("  passive [on|off]            - Set passive mode")
	fmt.Println("  status                      - Show current status")
	fmt.Println("  help / ?                    - Show this help")
	fmt.Println("  exit / quit / bye           - Exit program")
	fmt.Println()
}

func handleOpen(session *Session, parts []string) {
	if len(parts) < 2 {
		fmt.Println("Usage: open <host>[:port]")
		return
	}

	hostPort := parts[1]
	host := hostPort
	port := 21

	if idx := strings.LastIndex(hostPort, ":"); idx > 0 {
		host = hostPort[:idx]
		if p, err := strconv.Atoi(hostPort[idx+1:]); err == nil {
			port = p
		}
	}

	session.config.Host = host
	session.config.Port = port

	if session.config.Username == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Username: ")
		username, _ := reader.ReadString('\n')
		session.config.Username = strings.TrimSpace(username)

		fmt.Print("Password: ")
		password, _ := reader.ReadString('\n')
		session.config.Password = strings.TrimSpace(password)
	}

	fmt.Printf("Connecting to %s:%d...\n", host, port)
	err := session.apiClient.TestConnection(session.config)
	if err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		return
	}

	session.connected = true
	session.currentDir = "/"
	fmt.Println("Connected successfully")
}

func handleUser(session *Session, parts []string) {
	if len(parts) < 2 {
		fmt.Println("Usage: user <username> [password]")
		return
	}

	session.config.Username = parts[1]
	if len(parts) > 2 {
		session.config.Password = parts[2]
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Password: ")
		password, _ := reader.ReadString('\n')
		session.config.Password = strings.TrimSpace(password)
	}

	if session.config.Host == "" {
		fmt.Println("Not connected to any server")
		return
	}

	err := session.apiClient.TestConnection(session.config)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}

	session.connected = true
	session.currentDir = "/"
	fmt.Println("Login successful")
}

func handleClose(session *Session) {
	session.connected = false
	session.currentDir = "/"
	fmt.Println("Connection closed")
}

func handleList(session *Session, parts []string) {
	if !session.connected {
		fmt.Println("Not connected")
		return
	}

	path := session.currentDir
	if len(parts) > 1 {
		path = resolvePath(session.currentDir, parts[1])
	}

	resp, err := session.apiClient.ListDir(session.config, path, false)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		return
	}

	if resp.CurrentDir != "" {
		session.currentDir = resp.CurrentDir
	}

	entries := resp.Entries
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDirectory != entries[j].IsDirectory {
			return entries[i].IsDirectory
		}
		return entries[i].Name < entries[j].Name
	})

	for _, e := range entries {
		perm := e.Permissions
		if perm == "" {
			if e.IsDirectory {
				perm = "drwxr-xr-x"
			} else {
				perm = "-rw-r--r--"
			}
		}

		size := formatSize(e.Size)
		modified := ""
		if !e.Modified.IsZero() {
			modified = e.Modified.Format("2006-01-02 15:04")
		}

		dirMarker := " "
		if e.IsDirectory {
			dirMarker = "/"
		}

		fmt.Printf("%s  %10s  %s  %s%s\n", perm, size, modified, e.Name, dirMarker)
	}
}

func handleCd(session *Session, parts []string) {
	if !session.connected {
		fmt.Println("Not connected")
		return
	}
	if len(parts) < 2 {
		fmt.Println("Usage: cd <path>")
		return
	}

	path := resolvePath(session.currentDir, parts[1])
	err := session.apiClient.ChangeDir(session.config, path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	session.currentDir = path
	fmt.Printf("Directory changed to %s\n", path)
}

func handlePwd(session *Session) {
	if !session.connected {
		fmt.Println("Not connected")
		return
	}
	fmt.Println(session.currentDir)
}

func handleMkdir(session *Session, parts []string) {
	if !session.connected {
		fmt.Println("Not connected")
		return
	}
	if len(parts) < 2 {
		fmt.Println("Usage: mkdir <path>")
		return
	}

	path := resolvePath(session.currentDir, parts[1])
	err := session.apiClient.MakeDir(session.config, path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Directory created: %s\n", path)
}

func handleRmdir(session *Session, parts []string) {
	if !session.connected {
		fmt.Println("Not connected")
		return
	}
	if len(parts) < 2 {
		fmt.Println("Usage: rmdir <path>")
		return
	}

	path := resolvePath(session.currentDir, parts[1])
	err := session.apiClient.RemoveDir(session.config, path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Directory removed: %s\n", path)
}

func handleGet(session *Session, parts []string) {
	if !session.connected {
		fmt.Println("Not connected")
		return
	}
	if len(parts) < 2 {
		fmt.Println("Usage: get <remote_file> [local_file]")
		return
	}

	remotePath := resolvePath(session.currentDir, parts[1])
	localPath := filepath.Base(remotePath)
	if len(parts) > 2 {
		localPath = parts[2]
	}

	transferID, err := session.apiClient.Download(session.config, remotePath, localPath, false)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Download started, transfer ID: %s\n", transferID)
}

func handlePut(session *Session, parts []string) {
	if !session.connected {
		fmt.Println("Not connected")
		return
	}
	if len(parts) < 2 {
		fmt.Println("Usage: put <local_file> [remote_file]")
		return
	}

	localPath := parts[1]
	remotePath := filepath.Base(localPath)
	if len(parts) > 2 {
		remotePath = resolvePath(session.currentDir, parts[2])
	} else {
		remotePath = resolvePath(session.currentDir, remotePath)
	}

	transferID, err := session.apiClient.Upload(session.config, localPath, remotePath, false)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Upload started, transfer ID: %s\n", transferID)
}

func handleMget(session *Session, parts []string) {
	if !session.connected {
		fmt.Println("Not connected")
		return
	}
	if len(parts) < 2 {
		fmt.Println("Usage: mget <pattern>")
		return
	}

	pattern := parts[1]
	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, match := range matches {
		remotePath := resolvePath(session.currentDir, filepath.Base(match))
		transferID, err := session.apiClient.Download(session.config, remotePath, filepath.Base(match), false)
		if err != nil {
			fmt.Printf("Error downloading %s: %v\n", match, err)
			continue
		}
		fmt.Printf("Downloading %s, transfer ID: %s\n", match, transferID)
	}
}

func handleMput(session *Session, parts []string) {
	if !session.connected {
		fmt.Println("Not connected")
		return
	}
	if len(parts) < 2 {
		fmt.Println("Usage: mput <pattern>")
		return
	}

	pattern := parts[1]
	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}
		remotePath := resolvePath(session.currentDir, filepath.Base(match))
		transferID, err := session.apiClient.Upload(session.config, match, remotePath, false)
		if err != nil {
			fmt.Printf("Error uploading %s: %v\n", match, err)
			continue
		}
		fmt.Printf("Uploading %s, transfer ID: %s\n", match, transferID)
	}
}

func handleQueue(session *Session) {
	resp, err := session.apiClient.GetQueue()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if !resp.Success {
		fmt.Printf("Error: %s\n", resp.Error)
		return
	}

	if len(resp.Transfers) == 0 {
		fmt.Println("No active transfers")
		return
	}

	fmt.Println("\nTransfer Queue:")
	fmt.Println(strings.Repeat("-", 100))
	for _, t := range resp.Transfers {
		percent := 0
		if t.TotalSize > 0 {
			percent = int(t.Transferred * 100 / t.TotalSize)
		}
		fmt.Printf("ID: %s\n", t.ID)
		fmt.Printf("  Type: %s  Status: %s\n", t.Type, t.Status)
		fmt.Printf("  Local: %s\n", t.LocalPath)
		fmt.Printf("  Remote: %s\n", t.RemotePath)
		fmt.Printf("  Progress: %s / %s (%d%%)\n",
			formatSize(t.Transferred), formatSize(t.TotalSize), percent)
		if t.Error != "" {
			fmt.Printf("  Error: %s\n", t.Error)
		}
		fmt.Println()
	}
}

func handleCancel(session *Session, parts []string) {
	if len(parts) < 2 {
		fmt.Println("Usage: cancel <transfer_id>")
		return
	}

	err := session.apiClient.CancelTransfer(parts[1])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Transfer cancelled")
}

func handlePassive(session *Session, parts []string) {
	if len(parts) < 2 {
		if session.config.Mode == common.ModePassive {
			fmt.Println("Passive mode is on")
		} else {
			fmt.Println("Passive mode is off")
		}
		return
	}

	switch strings.ToLower(parts[1]) {
	case "on", "1", "true":
		session.config.Mode = common.ModePassive
		fmt.Println("Passive mode enabled")
	case "off", "0", "false":
		session.config.Mode = common.ModeActive
		fmt.Println("Passive mode disabled (active mode)")
	default:
		fmt.Println("Usage: passive [on|off]")
	}
}

func handleStatus(session *Session) {
	if !session.connected {
		fmt.Println("Not connected")
		return
	}
	fmt.Printf("Connected to: %s:%d\n", session.config.Host, session.config.Port)
	fmt.Printf("Username: %s\n", session.config.Username)
	fmt.Printf("Current directory: %s\n", session.currentDir)
	fmt.Printf("Mode: %s\n", session.config.Mode)
}

func resolvePath(current, path string) string {
	if strings.HasPrefix(path, "/") {
		return cleanPath(path)
	}
	if current == "/" {
		return cleanPath("/" + path)
	}
	return cleanPath(current + "/" + path)
}

func cleanPath(path string) string {
	parts := strings.Split(path, "/")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		switch part {
		case "", ".":
		case "..":
			if len(result) > 0 {
				result = result[:len(result)-1]
			}
		default:
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return "/"
	}
	return "/" + strings.Join(result, "/")
}

func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2fGB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2fMB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2fKB", float64(size)/KB)
	default:
		return fmt.Sprintf("%dB", size)
	}
}

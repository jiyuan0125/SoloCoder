package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"socks5-proxy/common"
	"socks5-proxy/socks5"
)

const defaultServerURL = "http://localhost:8210"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	serverURL := getEnvOrDefault("SERVER_URL", defaultServerURL)

	cmd := os.Args[1]
	switch cmd {
	case "start":
		cmdStart(serverURL)
	case "status":
		cmdStatus(serverURL)
	case "connections":
		cmdConnections(serverURL)
	case "add-user":
		cmdAddUser(serverURL)
	case "remove-user":
		cmdRemoveUser(serverURL)
	case "list-users":
		cmdListUsers(serverURL)
	case "test":
		cmdTest(serverURL)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
	}
}

func cmdStart(serverURL string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("SOCKS5 listen port (e.g. 1080): ")
	portStr, _ := reader.ReadString('\n')
	portStr = strings.TrimSpace(portStr)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Printf("Invalid port: %v\n", err)
		return
	}

	fmt.Print("Require authentication? (y/N): ")
	authStr, _ := reader.ReadString('\n')
	authStr = strings.TrimSpace(authStr)
	requireAuth := strings.ToLower(authStr) == "y" || strings.ToLower(authStr) == "yes"

	users := []common.User{}
	if requireAuth {
		for {
			fmt.Print("Add user? (enter username, or leave empty to skip): ")
			username, _ := reader.ReadString('\n')
			username = strings.TrimSpace(username)
			if username == "" {
				break
			}

			fmt.Print("Password: ")
			password, _ := reader.ReadString('\n')
			password = strings.TrimSpace(password)

			users = append(users, common.User{Username: username, Password: password})
		}
	}

	req := common.StartProxyRequest{
		ListenPort:  port,
		RequireAuth: requireAuth,
		Users:       users,
	}

	var resp common.StartProxyResponse
	if err := httpPostJSON(fmt.Sprintf("%s/api/proxy/start", serverURL), req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("Success! Proxy running at %s\n", resp.Address)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
	}
}

func cmdStatus(serverURL string) {
	var resp common.ProxyStatusResponse
	if err := httpGetJSON(fmt.Sprintf("%s/api/proxy/status", serverURL), &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if resp.Running {
		fmt.Printf("Proxy is running at %s\n", resp.Address)
	} else {
		fmt.Println("Proxy is not running")
	}
}

func cmdConnections(serverURL string) {
	var resp common.ActiveConnectionsResponse
	if err := httpGetJSON(fmt.Sprintf("%s/api/proxy/connections", serverURL), &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Active TCP Connections: %d\n", len(resp.TCPConnections))
	for i, c := range resp.TCPConnections {
		fmt.Printf("  [%d] %s -> %s (started: %s)\n", i+1, c.ClientAddr, c.TargetAddr, c.StartTime)
	}

	fmt.Printf("\nActive UDP Associations: %d\n", len(resp.UDPAssociations))
	for i, a := range resp.UDPAssociations {
		fmt.Printf("  [%d] %s (relay: %s, started: %s)\n", i+1, a.ClientAddr, a.RelayAddr, a.StartTime)
	}
}

func cmdAddUser(serverURL string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	req := common.AddUserRequest{
		Username: username,
		Password: password,
	}

	var resp common.UserResponse
	if err := httpPostJSON(fmt.Sprintf("%s/api/users/add", serverURL), req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("Success! %s\n", resp.Message)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
	}
}

func cmdRemoveUser(serverURL string) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: remove-user <username>")
		return
	}

	req := common.RemoveUserRequest{
		Username: os.Args[2],
	}

	var resp common.UserResponse
	if err := httpDeleteJSON(fmt.Sprintf("%s/api/users/remove", serverURL), req, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("Success! %s\n", resp.Message)
	} else {
		fmt.Printf("Failed: %s\n", resp.Message)
	}
}

func cmdListUsers(serverURL string) {
	var resp common.ListUsersResponse
	if err := httpGetJSON(fmt.Sprintf("%s/api/users/list", serverURL), &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Users: %d\n", len(resp.Users))
	for _, u := range resp.Users {
		fmt.Printf("  - %s\n", u)
	}
}

func cmdTest(serverURL string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("SOCKS5 proxy host (default: 127.0.0.1): ")
	proxyHost, _ := reader.ReadString('\n')
	proxyHost = strings.TrimSpace(proxyHost)
	if proxyHost == "" {
		proxyHost = "127.0.0.1"
	}

	fmt.Print("SOCKS5 proxy port: ")
	portStr, _ := reader.ReadString('\n')
	portStr = strings.TrimSpace(portStr)
	proxyPort, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Printf("Invalid port: %v\n", err)
		return
	}

	fmt.Print("Username (leave empty for no auth): ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	password := ""
	if username != "" {
		fmt.Print("Password: ")
		password, _ = reader.ReadString('\n')
		password = strings.TrimSpace(password)
	}

	fmt.Print("Test URL (default: http://httpbin.org/get): ")
	testURL, _ := reader.ReadString('\n')
	testURL = strings.TrimSpace(testURL)
	if testURL == "" {
		testURL = "http://httpbin.org/get"
	}

	fmt.Printf("\nTesting connection to %s via %s:%d...\n", testURL, proxyHost, proxyPort)
	startTime := time.Now()

	status, body, err := testSOCKS5Proxy(proxyHost, proxyPort, username, password, testURL)
	elapsed := time.Since(startTime)

	if err != nil {
		fmt.Printf("Test failed: %v\n", err)
		return
	}

	fmt.Printf("Status: %d (took %v)\n", status, elapsed)
	fmt.Printf("Response preview: %.200s\n", body)
	fmt.Println("Test passed!")
}

func testSOCKS5Proxy(proxyHost string, proxyPort int, username, password, targetURL string) (int, string, error) {
	proxyAddr := fmt.Sprintf("%s:%d", proxyHost, proxyPort)

	conn, err := net.DialTimeout("tcp", proxyAddr, 10*time.Second)
	if err != nil {
		return 0, "", fmt.Errorf("failed to connect to proxy: %v", err)
	}
	defer conn.Close()

	methods := []byte{socks5.MethodNoAuth}
	if username != "" {
		methods = []byte{socks5.MethodUserPassAuth}
	}
	greeting := []byte{socks5.Version, byte(len(methods))}
	greeting = append(greeting, methods...)

	if _, err := conn.Write(greeting); err != nil {
		return 0, "", fmt.Errorf("failed to send greeting: %v", err)
	}

	resp := make([]byte, 2)
	if _, err := io.ReadFull(conn, resp); err != nil {
		return 0, "", fmt.Errorf("failed to read greeting response: %v", err)
	}

	if resp[0] != socks5.Version {
		return 0, "", fmt.Errorf("unsupported SOCKS version: %d", resp[0])
	}

	if resp[1] == socks5.MethodUserPassAuth {
		auth := []byte{socks5.AuthUserPassVersion, byte(len(username))}
		auth = append(auth, []byte(username)...)
		auth = append(auth, byte(len(password)))
		auth = append(auth, []byte(password)...)

		if _, err := conn.Write(auth); err != nil {
			return 0, "", fmt.Errorf("failed to send auth: %v", err)
		}

		authResp := make([]byte, 2)
		if _, err := io.ReadFull(conn, authResp); err != nil {
			return 0, "", fmt.Errorf("failed to read auth response: %v", err)
		}

		if authResp[1] != socks5.AuthSuccess {
			return 0, "", fmt.Errorf("authentication failed")
		}
	} else if resp[1] != socks5.MethodNoAuth {
		return 0, "", fmt.Errorf("unsupported auth method: %d", resp[1])
	}

	targetHost := ""
	targetPort := 80
	if strings.HasPrefix(targetURL, "http://") {
		rest := targetURL[7:]
		slashIdx := strings.Index(rest, "/")
		if slashIdx == -1 {
			slashIdx = len(rest)
		}
		hostPart := rest[:slashIdx]
		colonIdx := strings.Index(hostPart, ":")
		if colonIdx != -1 {
			targetHost = hostPart[:colonIdx]
			targetPort, _ = strconv.Atoi(hostPart[colonIdx+1:])
		} else {
			targetHost = hostPart
			targetPort = 80
		}
	}

	req := []byte{socks5.Version, socks5.CmdConnect, 0x00}

	ip := net.ParseIP(targetHost)
	if ip != nil {
		if ip.To4() != nil {
			req = append(req, socks5.AddrTypeIPv4)
			req = append(req, ip.To4()...)
		} else {
			req = append(req, socks5.AddrTypeIPv6)
			req = append(req, ip.To16()...)
		}
	} else {
		req = append(req, socks5.AddrTypeDomain)
		req = append(req, byte(len(targetHost)))
		req = append(req, []byte(targetHost)...)
	}

	req = append(req, byte(targetPort>>8), byte(targetPort))

	if _, err := conn.Write(req); err != nil {
		return 0, "", fmt.Errorf("failed to send connect request: %v", err)
	}

	respHeader := make([]byte, 4)
	if _, err := io.ReadFull(conn, respHeader); err != nil {
		return 0, "", fmt.Errorf("failed to read connect response: %v", err)
	}

	if respHeader[1] != socks5.ReplySuccess {
		return 0, "", fmt.Errorf("connect failed: reply code %d", respHeader[1])
	}

	bndAddr := &socks5.Address{Type: respHeader[3]}
	switch bndAddr.Type {
	case socks5.AddrTypeIPv4:
		ipBuf := make([]byte, 4)
		io.ReadFull(conn, ipBuf)
		bndAddr.IP = net.IP(ipBuf)
	case socks5.AddrTypeIPv6:
		ipBuf := make([]byte, 16)
		io.ReadFull(conn, ipBuf)
		bndAddr.IP = net.IP(ipBuf)
	case socks5.AddrTypeDomain:
		lenBuf := make([]byte, 1)
		io.ReadFull(conn, lenBuf)
		nameBuf := make([]byte, int(lenBuf[0]))
		io.ReadFull(conn, nameBuf)
		bndAddr.Name = string(nameBuf)
	}
	portBuf := make([]byte, 2)
	io.ReadFull(conn, portBuf)

	httpReq := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", targetURL, targetHost)
	if _, err := conn.Write([]byte(httpReq)); err != nil {
		return 0, "", fmt.Errorf("failed to send HTTP request: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	responseBuf := new(bytes.Buffer)
	_, err = io.Copy(responseBuf, conn)
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		response := responseBuf.String()
		if response != "" {
			status := parseHTTPStatus(response)
			return status, response, nil
		}
		return 0, "", fmt.Errorf("failed to read response: %v", err)
	}

	response := responseBuf.String()
	status := parseHTTPStatus(response)

	return status, response, nil
}

func parseHTTPStatus(response string) int {
	lines := strings.SplitN(response, "\r\n", 2)
	if len(lines) == 0 {
		return 0
	}

	parts := strings.Split(lines[0], " ")
	if len(parts) < 2 {
		return 0
	}

	status, _ := strconv.Atoi(parts[1])
	return status
}

func printUsage() {
	fmt.Println(`SOCKS5 Proxy CLI

Usage:
  client <command> [args]

Commands:
  start           Start a new SOCKS5 proxy server
  status          Check proxy status
  connections     List active connections
  add-user        Add a new user
  remove-user <u> Remove a user
  list-users      List all users
  test            Test SOCKS5 proxy connection
  help            Show this help

Environment Variables:
  SERVER_URL      Server HTTP endpoint (default: http://localhost:8210)`)
}

func httpGetJSON(url string, v interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(v)
}

func httpPostJSON(url string, req, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	return json.NewDecoder(httpResp.Body).Decode(resp)
}

func httpDeleteJSON(url string, req, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	return json.NewDecoder(httpResp.Body).Decode(resp)
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

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

	"ssh-tunnel/pkg/common"
)

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: strings.TrimSuffix(serverURL, "/"),
	}
}

func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, c.serverURL+path, bodyReader)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		return json.Unmarshal(respBody, result)
	}
	return nil
}

func (c *Client) CreateTunnel(cfg *common.TunnelConfig) (*common.TunnelState, error) {
	req := common.CreateTunnelRequest{TunnelConfig: *cfg}
	var resp common.CreateTunnelResponse
	if err := c.doRequest("POST", "/tunnels/create", &req, &resp); err != nil {
		return nil, err
	}
	return resp.Tunnel, nil
}

func (c *Client) StopTunnel(id string) error {
	req := common.StopTunnelRequest{ID: id}
	var resp common.StopTunnelResponse
	return c.doRequest("POST", "/tunnels/stop", &req, &resp)
}

func (c *Client) ListTunnels() ([]*common.TunnelState, error) {
	var resp common.ListTunnelsResponse
	if err := c.doRequest("GET", "/tunnels/list", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Tunnels, nil
}

func (c *Client) GetTunnel(id string) (*common.TunnelState, error) {
	var resp common.GetTunnelResponse
	if err := c.doRequest("GET", "/tunnels/get?id="+id, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Tunnel, nil
}

func printUsage() {
	fmt.Println("SSH隧道客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client <命令> [选项]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  create    创建新隧道")
	fmt.Println("  list      列出所有隧道")
	fmt.Println("  stop      停止指定隧道")
	fmt.Println("  get       查看指定隧道状态")
	fmt.Println("  help      显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  client create --id tunnel1 --mode local --ssh-server example.com --ssh-user user --auth-method key --key-file ~/.ssh/id_rsa --local-port 8080 --remote-host localhost --remote-port 80")
	fmt.Println("  client list")
	fmt.Println("  client stop --id tunnel1")
	fmt.Println("  client get --id tunnel1")
}

func printCreateUsage() {
	fmt.Println("创建隧道用法:")
	fmt.Println("  client create [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  --id              隧道唯一ID (必填)")
	fmt.Println("  --name            隧道名称")
	fmt.Println("  --mode            模式: local 或 remote (必填)")
	fmt.Println("  --ssh-server      SSH服务器地址 (必填)")
	fmt.Println("  --ssh-port        SSH服务器端口 (默认 22)")
	fmt.Println("  --ssh-user        SSH用户名 (必填)")
	fmt.Println("  --auth-method     认证方式: password 或 key (必填)")
	fmt.Println("  --password        密码 (password方式时必填)")
	fmt.Println("  --key-file        密钥文件路径 (key方式时必填)")
	fmt.Println("  --local-address   本地监听地址 (默认 127.0.0.1)")
	fmt.Println("  --local-port      本地监听端口 (必填)")
	fmt.Println("  --remote-host     远程目标主机 (必填)")
	fmt.Println("  --remote-port     远程目标端口 (必填)")
	fmt.Println("  --keepalive       Keepalive间隔秒数 (默认 30)")
	fmt.Println("  --max-retry       最大重试间隔秒数 (默认 60)")
	fmt.Println("  --server          服务端地址 (默认 http://localhost:8080)")
}

func cmdCreate(args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	
	id := fs.String("id", "", "隧道唯一ID")
	name := fs.String("name", "", "隧道名称")
	mode := fs.String("mode", "", "模式: local 或 remote")
	sshServer := fs.String("ssh-server", "", "SSH服务器地址")
	sshPort := fs.Int("ssh-port", 22, "SSH服务器端口")
	sshUser := fs.String("ssh-user", "", "SSH用户名")
	authMethod := fs.String("auth-method", "", "认证方式: password 或 key")
	password := fs.String("password", "", "密码")
	keyFile := fs.String("key-file", "", "密钥文件路径")
	localAddress := fs.String("local-address", "127.0.0.1", "本地监听地址")
	localPort := fs.Int("local-port", 0, "本地监听端口")
	remoteHost := fs.String("remote-host", "", "远程目标主机")
	remotePort := fs.Int("remote-port", 0, "远程目标端口")
	keepalive := fs.Int("keepalive", 30, "Keepalive间隔秒数")
	maxRetry := fs.Int("max-retry", 60, "最大重试间隔秒数")
	serverURL := fs.String("server", "http://localhost:8080", "服务端地址")
	
	_ = fs.Parse(args)

	if *id == "" || *mode == "" || *sshServer == "" || *sshUser == "" || 
	   *authMethod == "" || *localPort == 0 || *remoteHost == "" || *remotePort == 0 {
		printCreateUsage()
		os.Exit(1)
	}

	cfg := &common.TunnelConfig{
		ID:               *id,
		Name:             *name,
		Mode:             common.TunnelMode(*mode),
		SSHServer:        *sshServer,
		SSHPort:          *sshPort,
		SSHUser:          *sshUser,
		AuthMethod:       common.AuthMethod(*authMethod),
		Password:         *password,
		KeyFilePath:      *keyFile,
		LocalAddress:     *localAddress,
		LocalPort:        *localPort,
		RemoteAddress:    *remoteHost,
		RemotePort:       *remotePort,
		KeepaliveInterval: *keepalive,
		MaxRetryInterval:  *maxRetry,
	}

	if cfg.AuthMethod == common.AuthMethodPassword && cfg.Password == "" {
		fmt.Println("错误: password认证方式需要--password参数")
		os.Exit(1)
	}
	if cfg.AuthMethod == common.AuthMethodKey && cfg.KeyFilePath == "" {
		fmt.Println("错误: key认证方式需要--key-file参数")
		os.Exit(1)
	}

	client := NewClient(*serverURL)
	tunnel, err := client.CreateTunnel(cfg)
	if err != nil {
		fmt.Printf("创建隧道失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("隧道创建成功!\n")
	fmt.Printf("ID: %s\n", tunnel.ID)
	fmt.Printf("名称: %s\n", tunnel.Name)
	fmt.Printf("状态: %s\n", tunnel.Status)
}

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	serverURL := fs.String("server", "http://localhost:8080", "服务端地址")
	_ = fs.Parse(args)

	client := NewClient(*serverURL)
	tunnels, err := client.ListTunnels()
	if err != nil {
		fmt.Printf("获取隧道列表失败: %v\n", err)
		os.Exit(1)
	}

	if len(tunnels) == 0 {
		fmt.Println("没有运行中的隧道")
		return
	}

	fmt.Printf("%-20s %-30s %-15s\n", "ID", "名称", "状态")
	fmt.Println(strings.Repeat("-", 65))
	for _, t := range tunnels {
		fmt.Printf("%-20s %-30s %-15s\n", t.ID, t.Name, t.Status)
	}
}

func cmdStop(args []string) {
	fs := flag.NewFlagSet("stop", flag.ExitOnError)
	id := fs.String("id", "", "隧道ID")
	serverURL := fs.String("server", "http://localhost:8080", "服务端地址")
	_ = fs.Parse(args)

	if *id == "" {
		fmt.Println("用法: client stop --id <隧道ID>")
		os.Exit(1)
	}

	client := NewClient(*serverURL)
	if err := client.StopTunnel(*id); err != nil {
		fmt.Printf("停止隧道失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("隧道 %s 已停止\n", *id)
}

func cmdGet(args []string) {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	id := fs.String("id", "", "隧道ID")
	serverURL := fs.String("server", "http://localhost:8080", "服务端地址")
	_ = fs.Parse(args)

	if *id == "" {
		fmt.Println("用法: client get --id <隧道ID>")
		os.Exit(1)
	}

	client := NewClient(*serverURL)
	tunnel, err := client.GetTunnel(*id)
	if err != nil {
		fmt.Printf("获取隧道状态失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("ID: %s\n", tunnel.ID)
	fmt.Printf("名称: %s\n", tunnel.Name)
	fmt.Printf("状态: %s\n", tunnel.Status)
	if tunnel.Error != "" {
		fmt.Printf("错误: %s\n", tunnel.Error)
	}
	fmt.Printf("启动时间: %s\n", tunnel.StartedAt.Format("2006-01-02 15:04:05"))
	if !tunnel.LastActive.IsZero() {
		fmt.Printf("最后活动: %s\n", tunnel.LastActive.Format("2006-01-02 15:04:05"))
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "create":
		cmdCreate(args)
	case "list":
		cmdList(args)
	case "stop":
		cmdStop(args)
	case "get":
		cmdGet(args)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

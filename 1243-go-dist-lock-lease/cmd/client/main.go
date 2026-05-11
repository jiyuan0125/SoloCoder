package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"dist-lock/pkg/api"
)

const defaultServer = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	server := os.Getenv("SERVER")
	if server == "" {
		server = defaultServer
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "acquire":
		handleAcquire(server, args)
	case "release":
		handleRelease(server, args)
	case "renew":
		handleRenew(server, args)
	case "status":
		handleStatus(server, args)
	case "list":
		handleList(server, args)
	case "-h", "--help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("分布式锁客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client acquire <锁名称> <持有者ID> [过期时间(毫秒)]")
	fmt.Println("  client release <锁名称> <持有者ID>")
	fmt.Println("  client renew <锁名称> <持有者ID>")
	fmt.Println("  client status <锁名称>")
	fmt.Println("  client list")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  SERVER - 服务端地址 (默认: http://localhost:8080)")
}

func handleAcquire(server string, args []string) {
	if len(args) < 2 {
		fmt.Println("用法: client acquire <锁名称> <持有者ID> [过期时间(毫秒)]")
		os.Exit(1)
	}

	name := args[0]
	holderID := args[1]
	expireMs := int64(10000)

	if len(args) >= 3 {
		var err error
		expireMs, err = strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			fmt.Printf("过期时间格式错误: %v\n", err)
			os.Exit(1)
		}
	}

	req := api.AcquireRequest{
		Name:     name,
		HolderID: holderID,
		ExpireMs: expireMs,
	}

	var resp api.Response
	if err := sendJSONRequest(http.MethodPost, server+"/locks/acquire", req, &resp); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("成功获取锁: %s (持有者: %s, 过期时间: %dms)\n", name, holderID, expireMs)
	} else {
		fmt.Printf("获取锁失败: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleRelease(server string, args []string) {
	if len(args) < 2 {
		fmt.Println("用法: client release <锁名称> <持有者ID>")
		os.Exit(1)
	}

	name := args[0]
	holderID := args[1]

	req := api.ReleaseRequest{
		Name:     name,
		HolderID: holderID,
	}

	var resp api.Response
	if err := sendJSONRequest(http.MethodPost, server+"/locks/release", req, &resp); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("成功释放锁: %s\n", name)
	} else {
		fmt.Printf("释放锁失败: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleRenew(server string, args []string) {
	if len(args) < 2 {
		fmt.Println("用法: client renew <锁名称> <持有者ID>")
		os.Exit(1)
	}

	name := args[0]
	holderID := args[1]

	req := api.RenewRequest{
		Name:     name,
		HolderID: holderID,
	}

	var resp api.Response
	if err := sendJSONRequest(http.MethodPost, server+"/locks/renew", req, &resp); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Printf("成功续期锁: %s\n", name)
	} else {
		fmt.Printf("续期锁失败: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleStatus(server string, args []string) {
	if len(args) < 1 {
		fmt.Println("用法: client status <锁名称>")
		os.Exit(1)
	}

	name := args[0]

	var resp api.Response
	if err := sendGetRequest(server+"/locks/"+name, &resp); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		dataMap, ok := resp.Data.(map[string]interface{})
		if !ok {
			fmt.Println("响应数据格式错误")
			os.Exit(1)
		}

		fmt.Printf("锁名称: %s\n", dataMap["name"])
		fmt.Printf("持有者ID: %s\n", dataMap["holder_id"])
		fmt.Printf("过期时间: %.0fms\n", dataMap["expire_ms"])
		fmt.Printf("剩余时间: %.0fms\n", dataMap["remaining_ms"])
		fmt.Printf("创建时间: %s\n", dataMap["created_at"])
		fmt.Printf("最后续期时间: %s\n", dataMap["last_renewed_at"])
	} else {
		fmt.Printf("查询锁状态失败: %s\n", resp.Message)
		os.Exit(1)
	}
}

func handleList(server string, args []string) {
	var resp api.Response
	if err := sendGetRequest(server+"/locks", &resp); err != nil {
		fmt.Printf("请求失败: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		dataSlice, ok := resp.Data.([]interface{})
		if !ok {
			fmt.Println("响应数据格式错误")
			os.Exit(1)
		}

		if len(dataSlice) == 0 {
			fmt.Println("当前没有活跃的锁")
			return
		}

		fmt.Printf("当前有 %d 个活跃的锁:\n", len(dataSlice))
		fmt.Println("---------------------------")
		for i, item := range dataSlice {
			dataMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			fmt.Printf("锁 #%d:\n", i+1)
			fmt.Printf("  锁名称: %s\n", dataMap["name"])
			fmt.Printf("  持有者ID: %s\n", dataMap["holder_id"])
			fmt.Printf("  过期时间: %.0fms\n", dataMap["expire_ms"])
			fmt.Printf("  剩余时间: %.0fms\n", dataMap["remaining_ms"])
			fmt.Println()
		}
	} else {
		fmt.Printf("获取锁列表失败: %s\n", resp.Message)
		os.Exit(1)
	}
}

func sendJSONRequest(method, url string, reqBody interface{}, respBody interface{}) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if err := json.Unmarshal(respBodyBytes, respBody); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	return nil
}

func sendGetRequest(url string, respBody interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if err := json.Unmarshal(respBodyBytes, respBody); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	return nil
}

func init() {
	flag.Usage = printUsage
}

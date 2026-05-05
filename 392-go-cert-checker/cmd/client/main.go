package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"cert-checker/internal/certchecker"
	"cert-checker/internal/client"
	"cert-checker/internal/protocol"
)

const defaultServerAddr = "localhost:8080"
const defaultWarnDays = 30

func main() {
	serverMode := false
	serverAddr := defaultServerAddr
	warnDays := defaultWarnDays
	jsonOutput := false
	inputFile := ""
	var domains []string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") {
			flagName := strings.TrimLeft(arg, "-")

			switch flagName {
			case "server":
				serverMode = true
			case "json":
				jsonOutput = true
			case "addr":
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					serverAddr = args[i+1]
					i++
				} else {
					fmt.Fprintf(os.Stderr, "错误: --addr 参数需要一个值\n")
					os.Exit(1)
				}
			case "warn-days":
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					var days int
					_, err := fmt.Sscanf(args[i+1], "%d", &days)
					if err != nil || days < 0 {
						fmt.Fprintf(os.Stderr, "错误: --warn-days 参数必须是一个非负整数\n")
						os.Exit(1)
					}
					warnDays = days
					i++
				} else {
					fmt.Fprintf(os.Stderr, "错误: --warn-days 参数需要一个值\n")
					os.Exit(1)
				}
			case "file":
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					inputFile = args[i+1]
					i++
				} else {
					fmt.Fprintf(os.Stderr, "错误: --file 参数需要一个值\n")
					os.Exit(1)
				}
			case "help", "h":
				printUsage()
				os.Exit(0)
			default:
				if strings.Contains(flagName, "=") {
					parts := strings.SplitN(flagName, "=", 2)
					flagName = parts[0]
					flagValue := parts[1]

					switch flagName {
					case "addr":
						serverAddr = flagValue
					case "warn-days":
						var days int
						_, err := fmt.Sscanf(flagValue, "%d", &days)
						if err != nil || days < 0 {
							fmt.Fprintf(os.Stderr, "错误: --warn-days 参数必须是一个非负整数\n")
							os.Exit(1)
						}
						warnDays = days
					case "file":
						inputFile = flagValue
					case "server":
						serverMode = flagValue == "true"
					case "json":
						jsonOutput = flagValue == "true"
					default:
						fmt.Fprintf(os.Stderr, "错误: 未知参数 --%s\n", flagName)
						printUsage()
						os.Exit(1)
					}
				} else {
					fmt.Fprintf(os.Stderr, "错误: 未知参数 --%s\n", flagName)
					printUsage()
					os.Exit(1)
				}
			}
		} else {
			domains = append(domains, arg)
		}
	}

	if inputFile != "" {
		fileDomains, err := readDomainsFromFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取文件失败: %v\n", err)
			os.Exit(1)
		}
		domains = append(domains, fileDomains...)
	}

	if len(domains) == 0 && !serverMode {
		printUsage()
		os.Exit(1)
	}

	if serverMode {
		handleServerMode(serverAddr, domains, warnDays, jsonOutput)
	} else {
		handleDirectMode(domains, warnDays, jsonOutput)
	}
}

func readDomainsFromFile(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var domains []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			domains = append(domains, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return domains, nil
}

func handleDirectMode(domains []string, warnDays int, jsonOutput bool) {
	results := make([]protocol.CertInfo, 0, len(domains))

	for _, domain := range domains {
		result := certchecker.CheckCertificate(domain)
		results = append(results, result)
	}

	if jsonOutput {
		outputJSON(results)
	} else {
		outputHumanReadable(results, warnDays)
	}
}

func handleServerMode(serverAddr string, domains []string, warnDays int, jsonOutput bool) {
	c := client.NewClient(serverAddr)

	if len(domains) == 0 {
		domainsResp, err := c.ListDomains()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取域名列表失败: %v\n", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(domainsResp)
		} else {
			if len(domainsResp.Domains) == 0 {
				fmt.Println("当前没有管理的域名")
			} else {
				fmt.Println("当前管理的域名:")
				for _, domain := range domainsResp.Domains {
					fmt.Printf("  - %s\n", domain)
				}
			}
		}
		return
	}

	var results []protocol.CertInfo
	for _, domain := range domains {
		result, err := c.CheckDomain(domain, warnDays)
		if err != nil {
			fmt.Fprintf(os.Stderr, "检查域名 %s 失败: %v\n", domain, err)
			continue
		}
		results = append(results, result)
	}

	if jsonOutput {
		outputJSON(results)
	} else {
		outputHumanReadable(results, warnDays)
	}
}

func outputJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "序列化JSON失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}

func outputHumanReadable(results []protocol.CertInfo, warnDays int) {
	for i, result := range results {
		if i > 0 {
			fmt.Println(strings.Repeat("-", 60))
		}

		if result.Error != "" {
			fmt.Printf("域名: %s\n", result.Domain)
			fmt.Printf("错误: %s\n", result.Error)
			continue
		}

		fmt.Printf("域名: %s\n", result.Domain)
		fmt.Printf("通用名称 (CN): %s\n", result.CommonName)
		fmt.Printf("颁发者: %s\n", result.Issuer)
		fmt.Printf("生效日期: %s\n", result.ValidFrom.Format("2006-01-02 15:04:05"))
		fmt.Printf("过期日期: %s\n", result.ValidTo.Format("2006-01-02 15:04:05"))

		if result.IsExpired {
			fmt.Printf("状态: \033[31m已过期\033[0m\n")
		} else if result.RemainingDays <= warnDays {
			fmt.Printf("剩余天数: \033[33m%d 天 (警告)\033[0m\n", result.RemainingDays)
		} else {
			fmt.Printf("剩余天数: %d 天\n", result.RemainingDays)
		}

		if result.IsWildcard {
			fmt.Println("类型: 通配符证书")
		}

		if len(result.SANs) > 0 {
			fmt.Printf("主题备用名称 (SANs): %s\n", strings.Join(result.SANs, ", "))
		}

		if !result.RequestDomainOK {
			fmt.Printf("\033[31m警告: 请求域名不在证书的有效域名列表中\033[0m\n")
		}

		fmt.Printf("检查时间: %s\n", result.CheckedAt.Format("2006-01-02 15:04:05"))
	}
}

func printUsage() {
	fmt.Println("使用方法: cert-checker [选项] <域名1> [域名2] ...")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  --server          使用服务端模式，与后台服务通信")
	fmt.Println("  --addr <地址>     服务端地址 (默认: localhost:8080)")
	fmt.Println("  --warn-days <天数> 告警天数阈值 (默认: 30)")
	fmt.Println("  --json            输出JSON格式")
	fmt.Println("  --file <文件>     从文件读取域名列表")
	fmt.Println("  --help, -h        显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  cert-checker example.com:443")
	fmt.Println("  cert-checker baidu.com:443 --warn-days 30 --json")
	fmt.Println("  cert-checker example.com api.example.com:8443")
	fmt.Println("  cert-checker --file domains.txt")
	fmt.Println("  cert-checker --server example.com")
}

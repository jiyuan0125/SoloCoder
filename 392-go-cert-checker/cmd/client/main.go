package main

import (
	"bufio"
	"encoding/json"
	"flag"
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
	var serverMode bool
	var serverAddr string
	var warnDays int
	var jsonOutput bool
	var inputFile string

	flag.BoolVar(&serverMode, "server", false, "使用服务端模式")
	flag.StringVar(&serverAddr, "addr", defaultServerAddr, "服务端地址 (默认: localhost:8080)")
	flag.IntVar(&warnDays, "warn-days", defaultWarnDays, "告警天数阈值 (默认: 30)")
	flag.BoolVar(&jsonOutput, "json", false, "输出JSON格式")
	flag.StringVar(&inputFile, "file", "", "从文件读取域名列表")

	flag.Parse()

	domains := flag.Args()

	if inputFile != "" {
		fileDomains, err := readDomainsFromFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取文件失败: %v\n", err)
			os.Exit(1)
		}
		domains = append(domains, fileDomains...)
	}

	if len(domains) == 0 && !serverMode {
		fmt.Fprintln(os.Stderr, "使用方法: cert-checker [选项] <域名1> [域名2] ...")
		fmt.Fprintln(os.Stderr, "选项:")
		flag.PrintDefaults()
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

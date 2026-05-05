package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	
	"disk-usage/internal/client"
	"disk-usage/internal/common"
	"disk-usage/internal/protocol"
)

func main() {
	flag.Usage = func() {
		fmt.Println("磁盘使用分析工具 - 客户端")
		fmt.Println("")
		fmt.Println("用法: disk-usage-client [选项] <扫描路径>")
		fmt.Println("")
		fmt.Println("选项:")
		flag.PrintDefaults()
		fmt.Println("")
		fmt.Println("示例:")
		fmt.Println("  disk-usage-client /home/user --top 20")
		fmt.Println("  disk-usage-client /var/log --min-size 100M")
		fmt.Println("  disk-usage-client /usr --max-depth 5")
		fmt.Println("  disk-usage-client /tmp --output csv")
		fmt.Println("  disk-usage-client /data --compare")
		fmt.Println("  disk-usage-client / --summary")
		fmt.Println("  disk-usage-client /home --all")
	}
	
	top := flag.Int("top", 0, "显示前N个最大的目录 (0=全部)")
	minSize := flag.String("min-size", "", "过滤小于指定大小的目录 (如: 100M, 2G)")
	maxDepth := flag.Int("max-depth", 10, "最大显示深度 (默认10层)")
	output := flag.String("output", "text", "输出格式: text 或 csv")
	showAll := flag.Bool("all", false, "包含隐藏目录 (以.开头)")
	summary := flag.Bool("summary", false, "只显示总大小和文件总数")
	compare := flag.Bool("compare", false, "与上一次扫描对比，显示增长情况")
	serverAddr := flag.String("server", "localhost:8765", "服务端地址")
	
	flag.Parse()
	
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}
	
	scanPath := args[0]
	
	req := &protocol.ScanRequest{
		Path:     scanPath,
		Top:      *top,
		MaxDepth: *maxDepth,
		Output:   *output,
		All:      *showAll,
		Summary:  *summary,
		Compare:  *compare,
	}
	
	if *minSize != "" {
		size, err := common.ParseSize(*minSize)
		if err != nil {
			log.Fatalf("无效的 min-size 参数: %v", err)
		}
		req.MinSize = size
	}
	
	c := client.NewClient(*serverAddr)
	formatter := client.NewFormatter(*output)
	
	if *compare {
		result, err := c.SendCompareRequest(req)
		if err != nil {
			log.Fatalf("发送对比请求失败: %v", err)
		}
		
		err = formatter.FormatCompareResult(result)
		if err != nil {
			log.Fatalf("格式化对比结果失败: %v", err)
		}
	} else {
		result, err := c.SendScanRequest(req)
		if err != nil {
			log.Fatalf("发送扫描请求失败: %v", err)
		}
		
		err = formatter.FormatScanResult(result, *summary)
		if err != nil {
			log.Fatalf("格式化结果失败: %v", err)
		}
	}
}

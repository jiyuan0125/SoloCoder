package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	serverAddr = flag.String("server", "localhost:9876", "服务端地址")
	outputFile = flag.String("o", "", "输出文件路径")
	timeFormat = flag.String("format", "", "时间戳格式 (默认: 2006-01-02 15:04:05)")
	resume     = flag.Bool("resume", false, "从断点继续处理")
	status     = flag.String("status", "", "查看指定任务ID的状态")
	help       = flag.Bool("help", false, "显示帮助信息")
)

func showUsage() {
	fmt.Print(`日志合并工具 - 客户端

用法:
  log-merger [选项] <输入文件1> <输入文件2> ...
  log-merger --status <任务ID>
  log-merger --help

选项:
  --server <地址>    服务端地址 (默认: localhost:9876)
  -o <文件>          输出文件路径 (必需)
  --format <格式>    时间戳格式 (默认: 2006-01-02 15:04:05)
  --resume           从断点继续处理
  --status <任务ID>  查看指定任务的状态
  --help             显示帮助信息

示例:
  # 合并两个日志文件
  log-merger file1.log file2.log -o merged.log

  # 使用自定义时间戳格式
  log-merger --format "2006/01/02 15:04:05" file1.log file2.log -o merged.log

  # 从断点继续
  log-merger --resume file1.log file2.log -o merged.log

  # 查看任务状态
  log-merger --status 20260505103015_abc123
`)
}

func main() {
	flag.Parse()

	if *help {
		showUsage()
		os.Exit(0)
	}

	if *status != "" {
		err := showTaskStatus(*serverAddr, *status)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}
		return
	}

	inputFiles := flag.Args()

	if len(inputFiles) < 2 {
		fmt.Println("错误: 至少需要2个输入文件")
		fmt.Println()
		showUsage()
		os.Exit(1)
	}

	if *outputFile == "" {
		fmt.Println("错误: 必须指定输出文件 (-o 参数)")
		fmt.Println()
		showUsage()
		os.Exit(1)
	}

	for i, f := range inputFiles {
		absPath, err := filepath.Abs(f)
		if err != nil {
			fmt.Printf("警告: 无法获取文件绝对路径 %s: %v\n", f, err)
			continue
		}
		inputFiles[i] = absPath
	}

	absOutput, err := filepath.Abs(*outputFile)
	if err != nil {
		fmt.Printf("警告: 无法获取输出文件绝对路径: %v\n", err)
	} else {
		*outputFile = absOutput
	}

	fmt.Println("=== 提交日志合并任务 ===")
	fmt.Printf("输入文件: %v\n", inputFiles)
	fmt.Printf("输出文件: %s\n", *outputFile)
	if *timeFormat != "" {
		fmt.Printf("时间格式: %s\n", *timeFormat)
	}
	fmt.Printf("断点续传: %v\n", *resume)
	fmt.Printf("服务端: %s\n", *serverAddr)
	fmt.Println()

	taskID, err := submitTask(*serverAddr, inputFiles, *outputFile, *timeFormat, *resume)
	if err != nil {
		fmt.Printf("提交任务失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("任务已提交，任务ID: %s\n", taskID)
	fmt.Println()
	fmt.Println("使用以下命令查看任务状态:")
	fmt.Printf("  log-merger --status %s\n", taskID)
}

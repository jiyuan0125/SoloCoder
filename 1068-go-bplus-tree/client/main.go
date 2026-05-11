package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const defaultServerURL = "http://localhost:8300"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(defaultServerURL)
	command := os.Args[1]

	switch strings.ToLower(command) {
	case "insert":
		if len(os.Args) < 4 {
			fmt.Println("用法: bplus-tree-client insert <key> <value>")
			os.Exit(1)
		}
		client.Insert(os.Args[2], os.Args[3])

	case "search":
		if len(os.Args) < 3 {
			fmt.Println("用法: bplus-tree-client search <key>")
			os.Exit(1)
		}
		client.Search(os.Args[2])

	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("用法: bplus-tree-client delete <key>")
			os.Exit(1)
		}
		client.Delete(os.Args[2])

	case "range":
		if len(os.Args) < 4 {
			fmt.Println("用法: bplus-tree-client range <start> <end>")
			os.Exit(1)
		}
		client.Range(os.Args[2], os.Args[3])

	case "scan":
		direction := "forward"
		batchSize := 10
		startOffset := 0

		if len(os.Args) >= 3 {
			direction = os.Args[2]
		}
		if len(os.Args) >= 4 {
			if bs, err := strconv.Atoi(os.Args[3]); err == nil {
				batchSize = bs
			}
		}
		if len(os.Args) >= 5 {
			if so, err := strconv.Atoi(os.Args[4]); err == nil {
				startOffset = so
			}
		}

		client.Scan(direction, batchSize, startOffset)

	case "help":
		printUsage()

	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("B+树索引引擎命令行客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  bplus-tree-client <command> [arguments]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  insert <key> <value>    插入键值对")
	fmt.Println("  search <key>            搜索键值对")
	fmt.Println("  delete <key>            删除键值对")
	fmt.Println("  range <start> <end>     范围查询")
	fmt.Println("  scan [direction] [batchSize] [offset]  遍历扫描")
	fmt.Println("  help                    显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  bplus-tree-client insert name Alice")
	fmt.Println("  bplus-tree-client search name")
	fmt.Println("  bplus-tree-client range a z")
	fmt.Println("  bplus-tree-client scan forward 10 0")
	fmt.Println("  bplus-tree-client scan backward 5 0")
}

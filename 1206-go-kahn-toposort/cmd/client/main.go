package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"kahn-toposort/pkg/common"
	"net/http"
	"os"
)

func main() {
	var serverURL string
	var tasks []string
	var depsBefore []string
	var depsAfter []string

	flag.StringVar(&serverURL, "server", "http://localhost:8080", "服务端地址")
	flag.Var((*stringSlice)(&tasks), "task", "任务ID（可重复指定多个任务）")
	flag.Var((*stringSlice)(&depsBefore), "dep-before", "依赖的前置任务（必须在dep-after之前）")
	flag.Var((*stringSlice)(&depsAfter), "dep-after", "依赖的后置任务（必须在dep-before之后）")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "用法: %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "示例:\n")
		fmt.Fprintf(os.Stderr, "  %s -task A -task B -task C -dep-before A -dep-after B -dep-before B -dep-after C\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(tasks) == 0 {
		fmt.Fprintln(os.Stderr, "错误: 必须指定至少一个任务")
		flag.Usage()
		os.Exit(1)
	}

	if len(depsBefore) != len(depsAfter) {
		fmt.Fprintln(os.Stderr, "错误: dep-before和dep-after的数量必须一致")
		os.Exit(1)
	}

	dependencies := []common.Dependency{}
	for i := range depsBefore {
		dependencies = append(dependencies, common.Dependency{
			Before: depsBefore[i],
			After:  depsAfter[i],
		})
	}

	req := common.SortRequest{
		Tasks:        tasks,
		Dependencies: dependencies,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 序列化请求失败: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/sort", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 发送请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 读取响应失败: %v\n", err)
		os.Exit(1)
	}

	var result common.SortResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 解析响应失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "响应内容: %s\n", respBody)
		os.Exit(1)
	}

	fmt.Println("拓扑排序结果:")
	fmt.Printf("是否存在循环依赖: %v\n", result.HasCycle)

	if result.HasCycle {
		fmt.Printf("错误信息: %s\n", result.Error)
		os.Exit(2)
	}

	fmt.Println("执行层（同一层可并行）:")
	for i, level := range result.Levels {
		fmt.Printf("  第 %d 层: %v\n", i+1, level)
	}
}

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

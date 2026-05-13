package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"credits/pkg/storage"
)

var (
	dataDir string
	store   *storage.Storage
)

var rootCmd = &cobra.Command{
	Use:   "training-credits",
	Short: "培训学分管理系统",
	Long: `培训学分管理系统 - 一个用于管理员工培训学分的命令行工具。

功能包括：
- 课程管理（添加、查询、更新、删除）
- 员工管理（添加、查询、更新、删除）
- 学分记录
- 学分进度查询
- 部门排名查询
- 年度统计`,
}

func Execute(dir string) {
	dataDir = dir
	var err error
	store, err = storage.New(dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化存储失败: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(courseCmd)
	rootCmd.AddCommand(employeeCmd)
	rootCmd.AddCommand(creditCmd)
	rootCmd.AddCommand(progressCmd)
	rootCmd.AddCommand(rankCmd)
	rootCmd.AddCommand(yearlyCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "序列化输出失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

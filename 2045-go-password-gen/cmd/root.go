package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "passwordgen",
	Short: "一个强大的密码生成命令行工具",
	Long: `passwordgen 是一个基于 Go 和 Cobra 构建的密码生成工具，
支持多种密码类型、密码强度评估、批量生成等功能。

示例:
  passwordgen generate                    # 生成默认的 16 位密码
  passwordgen generate -l 20 -t numeric   # 生成 20 位纯数字密码
  passwordgen strength "MyPass123!"       # 评估密码强度
  passwordgen batch -n 100 -o output.txt  # 批量生成 100 个密码到文件`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var templatesDir string

var rootCmd = &cobra.Command{
	Use:   "msg-template",
	Short: "消息模板命令行工具",
	Long: `一个用于管理和渲染消息模板的命令行工具。
支持变量替换、默认值、条件逻辑和子模板导入。

模板语法：
  {{name}}            - 变量替换
  {{name|默认值}}      - 带默认值的变量
  {{#if condition}}   - 条件开始
  {{/if}}             - 条件结束
  {{> header}}        - 导入子模板（以_开头的模板为子模板）`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	wd, _ := os.Getwd()
	defaultDir := filepath.Join(wd, "templates")
	rootCmd.PersistentFlags().StringVarP(&templatesDir, "dir", "d", defaultDir, "模板目录路径")
}

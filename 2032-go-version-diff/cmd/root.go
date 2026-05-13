package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "config-diff",
	Short: "配置版本对比工具",
	Long:  `基于 Git 仓库的配置文件版本对比工具，支持 JSON 和 YAML 格式，可对比分支或 commit。`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

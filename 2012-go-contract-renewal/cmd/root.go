package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "contract-manager",
	Short: "合同管理命令行工具",
	Long: `合同管理命令行工具，支持：
- 添加、查看合同
- 手动/自动续约
- 到期提醒
- 自动标记已终止合同`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(renewCmd)
	rootCmd.AddCommand(checkCmd)
}

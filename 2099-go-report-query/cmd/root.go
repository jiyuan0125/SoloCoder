package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "report-query",
	Short: "报表查询命令行工具",
	Long: `一个基于 Go + Cobra 的报表查询系统，支持：
- 通过命令行参数指定表名、字段和筛选条件
- 支持聚合查询和 GROUP BY
- 输出为表格或 CSV
- 保存查询历史和快捷方式`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func exitWithError(msg string, code int) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}

func init() {
}

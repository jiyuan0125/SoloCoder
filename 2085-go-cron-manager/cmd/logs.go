package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gocron/internal/logger"
)

var logLimit int

var logsCmd = &cobra.Command{
	Use:   "logs [任务名称]",
	Short: "查看任务执行日志",
	Long:  `查看指定任务的最近执行日志，记录了开始时间、结束时间、退出码和输出内容。`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		l, err := logger.New(logDir)
		if err != nil {
			fmt.Printf("初始化日志失败: %v\n", err)
			os.Exit(1)
		}

		logs, err := l.GetLogs(name, logLimit)
		if err != nil {
			fmt.Printf("读取日志失败: %v\n", err)
			os.Exit(1)
		}

		if len(logs) == 0 {
			fmt.Printf("没有找到任务 %s 的执行日志\n", name)
			return
		}

		fmt.Printf("任务 %s 的最近 %d 条执行日志:\n", name, len(logs))
		fmt.Println("--------------------------------------------------")
		for _, log := range logs {
			fmt.Println(log)
		}
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().IntVarP(&logLimit, "limit", "n", 20, "显示最近的日志条数")
}

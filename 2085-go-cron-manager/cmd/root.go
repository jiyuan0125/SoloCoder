package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile string
	logDir     string
)

var rootCmd = &cobra.Command{
	Use:   "gocron",
	Short: "Go Cron Manager - 基于 Go + Cobra 实现的定时任务管理系统",
	Long: `gocron 是一个使用 Go 语言实现的定时任务管理系统，支持 cron 表达式定义定时任务，每个任务关联执行命令。支持查看、添加、删除、暂停任务，任务执行日志记录，自动重试，手动触发等功能。`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "crontab.yaml", "配置文件路径")
	rootCmd.PersistentFlags().StringVarP(&logDir, "log-dir", "l", "logs", "日志目录")
}

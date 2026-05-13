package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gocron/internal/config"
	"gocron/internal/logger"
	"gocron/internal/task"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [任务名称]",
	Short: "删除定时任务",
	Long:  `从配置文件中删除指定的定时任务。如果任务正在执行，将返回错误。`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := config.Load(configFile)
		if err != nil {
			fmt.Printf("加载配置文件失败: %v\n", err)
			os.Exit(1)
		}

		l, err := logger.New(logDir)
		if err != nil {
			fmt.Printf("初始化日志失败: %v\n", err)
			os.Exit(1)
		}

		mgr := task.NewManager(l)
		mgr.LoadFromConfig(cfg)

		if err := mgr.Delete(name); err != nil {
			fmt.Printf("删除任务失败: %v\n", err)
			os.Exit(1)
		}

		if err := mgr.SaveConfig(configFile); err != nil {
			fmt.Printf("保存配置文件失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("任务已删除: %s\n", name)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

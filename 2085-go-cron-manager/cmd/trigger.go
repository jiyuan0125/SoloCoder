package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gocron/internal/config"
	"gocron/internal/logger"
	"gocron/internal/task"
)

var triggerCmd = &cobra.Command{
	Use:   "trigger [任务名称]",
	Short: "手动触发定时任务",
	Long:  `手动触发指定的定时任务立即执行，不影响定时调度。如果任务正在执行，将返回提示信息。`,
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

		if _, exists := mgr.Get(name); !exists {
			fmt.Printf("任务不存在: %s\n", name)
			os.Exit(1)
		}

		fmt.Printf("正在手动触发任务: %s\n", name)
		err = mgr.Execute(cmd.Context(), name, true)
		if err != nil {
			fmt.Printf("任务执行失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("任务执行完成: %s\n", name)
	},
}

func init() {
	rootCmd.AddCommand(triggerCmd)
}

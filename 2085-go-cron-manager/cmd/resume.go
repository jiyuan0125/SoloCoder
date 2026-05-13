package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gocron/internal/config"
	"gocron/internal/logger"
	"gocron/internal/task"
)

var resumeCmd = &cobra.Command{
	Use:   "resume [任务名称]",
	Short: "恢复已暂停的定时任务",
	Long:  `恢复已暂停的定时任务，使其继续按 cron 表达式执行。`,
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

		if err := mgr.Resume(name); err != nil {
			fmt.Printf("恢复任务失败: %v\n", err)
			os.Exit(1)
		}

		if err := mgr.SaveConfig(configFile); err != nil {
			fmt.Printf("保存配置文件失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("任务已恢复: %s\n", name)
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}

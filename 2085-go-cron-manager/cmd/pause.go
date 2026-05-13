package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gocron/internal/config"
	"gocron/internal/logger"
	"gocron/internal/task"
)

var pauseCmd = &cobra.Command{
	Use:   "pause [任务名称]",
	Short: "暂停定时任务",
	Long:  `暂停指定的定时任务。暂停后任务不会被定时触发，但配置会被保留。`,
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

		if err := mgr.Pause(name); err != nil {
			fmt.Printf("暂停任务失败: %v\n", err)
			os.Exit(1)
		}

		if err := mgr.SaveConfig(configFile); err != nil {
			fmt.Printf("保存配置文件失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("任务已暂停: %s\n", name)
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)
}

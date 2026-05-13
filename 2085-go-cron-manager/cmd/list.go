package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gocron/internal/config"
	"gocron/internal/logger"
	"gocron/internal/task"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有定时任务",
	Long:  `列出所有已配置的定时任务，包括名称、cron 表达式、状态等信息。`,
	Run: func(cmd *cobra.Command, args []string) {
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

		tasks := mgr.List()
		if len(tasks) == 0 {
			fmt.Println("没有配置任何任务")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "名称\tCron\t命令\t状态\t重试\t重试间隔(秒)\t描述")
		fmt.Fprintln(w, "----\t----\t----\t----\t----\t----------\t----")

		for _, t := range tasks {
			status := "运行中"
			if t.Config.Paused {
				status = "已暂停"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%s\n",
				t.Name, t.Config.Cron, t.Config.Command, status,
				t.Config.Retries, t.Config.RetryDelay, t.Config.Description)
		}

		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

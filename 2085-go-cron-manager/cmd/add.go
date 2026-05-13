package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gocron/internal/config"
	"gocron/internal/logger"
	"gocron/internal/task"
)

var (
	addCron       string
	addCommand    string
	addDesc       string
	addRetries    int
	addRetryDelay int
)

var addCmd = &cobra.Command{
	Use:   "add [任务名称]",
	Short: "添加新的定时任务",
	Long:  `添加一个新的定时任务到配置文件中。`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if err := config.ValidateCron(addCron); err != nil {
			fmt.Printf("Cron 表达式语法错误: %v\n", err)
			os.Exit(1)
		}

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

		newTask := &config.TaskConfig{
			Name:        name,
			Cron:        addCron,
			Command:     addCommand,
			Description: addDesc,
			Retries:     addRetries,
			RetryDelay:  addRetryDelay,
			Paused:      false,
		}

		if err := mgr.Add(newTask); err != nil {
			fmt.Printf("添加任务失败: %v\n", err)
			os.Exit(1)
		}

		if err := mgr.SaveConfig(configFile); err != nil {
			fmt.Printf("保存配置文件失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("任务添加成功: %s\n", name)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&addCron, "cron", "e", "", "Cron 表达式 (必填)")
	addCmd.Flags().StringVarP(&addCommand, "command", "C", "", "要执行的命令 (必填)")
	addCmd.Flags().StringVarP(&addDesc, "description", "D", "", "任务描述")
	addCmd.Flags().IntVarP(&addRetries, "retries", "r", 0, "失败后重试次数")
	addCmd.Flags().IntVarP(&addRetryDelay, "retry-delay", "R", 1, "重试间隔 (秒)")
	addCmd.MarkFlagRequired("cron")
	addCmd.MarkFlagRequired("command")
}

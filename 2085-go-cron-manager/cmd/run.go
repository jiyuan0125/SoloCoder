package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"gocron/internal/config"
	"gocron/internal/daemon"
	"gocron/internal/logger"
	"gocron/internal/scheduler"
	"gocron/internal/task"
)

var daemonMode bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "启动定时任务调度器",
	Long:  `启动定时任务调度器，可选择前台或守护进程模式运行。`,
	Run: func(cmd *cobra.Command, args []string) {
		if daemonMode {
			if err := daemon.StartDaemon(); err != nil {
				fmt.Printf("启动守护进程失败: %v\n", err)
				os.Exit(1)
			}
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

		sched := scheduler.New(mgr)
		sched.Start()

		if !daemon.IsDaemonMode() {
			fmt.Println("定时任务调度器已启动")
			fmt.Println("按 Ctrl+C 停止调度器")
		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		<-sigCh
		if !daemon.IsDaemonMode() {
			fmt.Println("\n正在停止调度器...")
		}
		sched.Stop()
		if !daemon.IsDaemonMode() {
			fmt.Println("调度器已停止")
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "停止守护进程模式的定时任务调度器",
	Long:  `停止以守护进程模式运行的定时任务调度器。`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := daemon.StopDaemon(); err != nil {
			fmt.Printf("停止守护进程失败: %v\n", err)
			os.Exit(1)
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看守护进程状态",
	Long:  `查看以守护进程模式运行的定时任务调度器状态。`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := daemon.Status(); err != nil {
			fmt.Printf("获取守护进程状态失败: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	runCmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "以守护进程模式运行")
}

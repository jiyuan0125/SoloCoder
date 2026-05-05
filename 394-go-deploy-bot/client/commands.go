package main

import (
	"deploybot/common"
	"flag"
	"fmt"
	"os"
	"time"
)

func handleDeploy() {
	flagSet := flag.NewFlagSet("deploy", flag.ExitOnError)

	var (
		binaryPath       string
		backupDir        string
		healthCheckURL   string
		healthCheckTimeout int
		buildCommand     string
		pullCommand      string
		serverAddr       string
	)

	flagSet.StringVar(&binaryPath, "binary", "", "二进制文件路径")
	flagSet.StringVar(&backupDir, "backup", "backups", "备份目录")
	flagSet.StringVar(&healthCheckURL, "health-url", "", "健康检查URL")
	flagSet.IntVar(&healthCheckTimeout, "health-timeout", 5, "健康检查超时时间(秒)")
	flagSet.StringVar(&buildCommand, "build-cmd", "", "编译命令")
	flagSet.StringVar(&pullCommand, "pull-cmd", "", "拉取代码命令")
	flagSet.StringVar(&serverAddr, "server", "localhost:9876", "服务端地址")

	if err := flagSet.Parse(os.Args[2:]); err != nil {
		flagSet.Usage()
		os.Exit(1)
	}

	if binaryPath == "" {
		fmt.Println("错误: 必须指定 --binary 参数")
		flagSet.Usage()
		os.Exit(1)
	}

	config := &common.DeployConfig{
		BinaryPath:         binaryPath,
		BackupDir:          backupDir,
		HealthCheckURL:     healthCheckURL,
		HealthCheckTimeout: healthCheckTimeout,
		BuildCommand:       buildCommand,
		PullCommand:        pullCommand,
	}

	req := &common.DeployRequest{
		Command: common.CmdDeploy,
		Config:  config,
	}

	printDeployConfig(config)

	client, err := NewClient(serverAddr)
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("\n正在连接服务端...")
	resp, err := client.Send(req)
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
		os.Exit(1)
	}

	printResponse(resp)
}

func handleRollback() {
	flagSet := flag.NewFlagSet("rollback", flag.ExitOnError)

	var (
		backupPath     string
		binaryPath     string
		backupDir      string
		healthCheckURL string
		serverAddr     string
	)

	flagSet.StringVar(&backupPath, "backup-path", "", "备份文件路径")
	flagSet.StringVar(&binaryPath, "binary", "", "二进制文件路径")
	flagSet.StringVar(&backupDir, "backup", "", "备份目录")
	flagSet.StringVar(&healthCheckURL, "health-url", "", "健康检查URL")
	flagSet.StringVar(&serverAddr, "server", "localhost:9876", "服务端地址")

	if err := flagSet.Parse(os.Args[2:]); err != nil {
		flagSet.Usage()
		os.Exit(1)
	}

	if binaryPath == "" {
		fmt.Println("错误: 必须指定 --binary 参数")
		flagSet.Usage()
		os.Exit(1)
	}

	config := &common.DeployConfig{
		BinaryPath:     binaryPath,
		BackupDir:      backupDir,
		HealthCheckURL: healthCheckURL,
	}

	req := &common.DeployRequest{
		Command:    common.CmdRollback,
		BackupPath: backupPath,
		Config:     config,
	}

	fmt.Println("=== 回滚配置 ===")
	if backupPath != "" {
		fmt.Printf("备份路径: %s\n", backupPath)
	}
	fmt.Printf("二进制文件: %s\n", binaryPath)
	if backupDir != "" {
		fmt.Printf("备份目录: %s\n", backupDir)
	}
	if healthCheckURL != "" {
		fmt.Printf("健康检查URL: %s\n", healthCheckURL)
	}

	client, err := NewClient(serverAddr)
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("\n正在连接服务端...")
	resp, err := client.Send(req)
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
		os.Exit(1)
	}

	printResponse(resp)
}

func handleStatus() {
	flagSet := flag.NewFlagSet("status", flag.ExitOnError)

	var (
		serverAddr string
		binaryPath string
	)

	flagSet.StringVar(&serverAddr, "server", "localhost:9876", "服务端地址")
	flagSet.StringVar(&binaryPath, "binary", "", "二进制文件路径")

	if err := flagSet.Parse(os.Args[2:]); err != nil {
		flagSet.Usage()
		os.Exit(1)
	}

	var config *common.DeployConfig
	if binaryPath != "" {
		config = &common.DeployConfig{
			BinaryPath: binaryPath,
		}
	}

	req := &common.DeployRequest{
		Command: common.CmdStatus,
		Config:  config,
	}

	client, err := NewClient(serverAddr)
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	resp, err := client.Send(req)
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== 服务状态 ===")
	printResponse(resp)

	if resp.Deployments != nil && len(resp.Deployments) > 0 {
		printDeployments(resp.Deployments)
	}
}

func handleHistory() {
	flagSet := flag.NewFlagSet("history", flag.ExitOnError)

	var (
		serverAddr string
	)

	flagSet.StringVar(&serverAddr, "server", "localhost:9876", "服务端地址")

	if err := flagSet.Parse(os.Args[2:]); err != nil {
		flagSet.Usage()
		os.Exit(1)
	}

	req := &common.DeployRequest{
		Command: common.CmdHistory,
	}

	client, err := NewClient(serverAddr)
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	resp, err := client.Send(req)
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== 部署历史 ===")
	printResponse(resp)

	if resp.Deployments != nil && len(resp.Deployments) > 0 {
		printDeployments(resp.Deployments)
	}
}

func printDeployConfig(config *common.DeployConfig) {
	fmt.Println("=== 部署配置 ===")
	fmt.Printf("二进制文件: %s\n", config.BinaryPath)
	fmt.Printf("备份目录: %s\n", config.BackupDir)

	if config.PullCommand != "" {
		fmt.Printf("拉取命令: %s\n", config.PullCommand)
	}

	if config.BuildCommand != "" {
		fmt.Printf("编译命令: %s\n", config.BuildCommand)
	}

	if config.HealthCheckURL != "" {
		fmt.Printf("健康检查URL: %s\n", config.HealthCheckURL)
		fmt.Printf("健康检查超时: %d秒\n", config.HealthCheckTimeout)
	}
}

func printResponse(resp *common.DeployResponse) {
	if resp.Success {
		fmt.Printf("\n✓ 成功: %s\n", resp.Message)
	} else {
		fmt.Printf("\n✗ 失败: %s\n", resp.Message)
	}

	if resp.DeployID != "" {
		fmt.Printf("部署ID: %s\n", resp.DeployID)
	}

	if resp.Status != "" {
		fmt.Printf("状态: %s\n", resp.Status)
	}
}

func printDeployments(deployments []*common.Deployment) {
	fmt.Println("\n--- 详情 ---")
	for i, d := range deployments {
		fmt.Printf("\n[%d] ID: %s\n", i+1, d.ID)
		fmt.Printf("    类型: %s\n", map[bool]string{true: "回滚", false: "部署"}[d.IsRollback])
		fmt.Printf("    状态: %s\n", d.Status)
		fmt.Printf("    开始时间: %s\n", d.StartTime.Format(time.RFC3339))
		if !d.EndTime.IsZero() {
			fmt.Printf("    结束时间: %s\n", d.EndTime.Format(time.RFC3339))
		}
		if d.TotalDuration > 0 {
			fmt.Printf("    总耗时: %v\n", d.TotalDuration)
		}
		if d.BackupPath != "" {
			fmt.Printf("    备份路径: %s\n", d.BackupPath)
		}

		if d.Steps != nil && len(d.Steps) > 0 {
			fmt.Printf("    步骤:\n")
			for _, step := range d.Steps {
				statusSymbol := " "
				switch step.Status {
				case common.StepCompleted:
					statusSymbol = "✓"
				case common.StepFailed:
					statusSymbol = "✗"
				case common.StepRunning:
					statusSymbol = "●"
				}
				if step.Duration > 0 {
					fmt.Printf("      %s [%s] %-20s %v\n", statusSymbol, step.Status, step.Name, step.Duration)
				} else {
					fmt.Printf("      %s [%s] %s\n", statusSymbol, step.Status, step.Name)
				}
			}
		}
	}
}

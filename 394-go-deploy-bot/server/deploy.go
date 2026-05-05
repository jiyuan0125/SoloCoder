package main

import (
	"deploybot/common"
	"sync"
	"sync/atomic"
	"time"
)

type DeployManager struct {
	mu          sync.Mutex
	current     *common.Deployment
	stopChan    chan struct{}
	running     atomic.Bool
}

func NewDeployManager() *DeployManager {
	return &DeployManager{
		stopChan: make(chan struct{}),
	}
}

func (dm *DeployManager) Current() *common.Deployment {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.current
}

func (dm *DeployManager) Execute(deployment *common.Deployment) {
	dm.mu.Lock()
	if dm.current != nil && dm.current.Status == common.StatusRunning {
		dm.mu.Unlock()
		deployment.Status = common.StatusFailed
		Error("已有部署任务在运行中，拒绝新任务: %s", deployment.ID)
		return
	}
	dm.current = deployment
	dm.running.Store(true)
	dm.mu.Unlock()

	defer func() {
		dm.mu.Lock()
		dm.current = nil
		dm.running.Store(false)
		dm.mu.Unlock()
	}()

	dm.runDeployment(deployment)
}

func (dm *DeployManager) ExecuteRollback(deployment *common.Deployment) {
	dm.mu.Lock()
	if dm.current != nil && dm.current.Status == common.StatusRunning {
		dm.mu.Unlock()
		deployment.Status = common.StatusFailed
		Error("已有部署任务在运行中，拒绝回滚任务: %s", deployment.ID)
		return
	}
	dm.current = deployment
	dm.running.Store(true)
	dm.mu.Unlock()

	defer func() {
		dm.mu.Lock()
		dm.current = nil
		dm.running.Store(false)
		dm.mu.Unlock()
	}()

	dm.runRollback(deployment)
}

func (dm *DeployManager) Stop() {
	if dm.running.Swap(false) {
		close(dm.stopChan)
	}
}

func (dm *DeployManager) runDeployment(deployment *common.Deployment) {
	deployment.Status = common.StatusRunning
	deployment.StartTime = time.Now()
	deployment.Steps = []*common.StepRecord{}

	Info("开始部署任务: %s", deployment.ID)

	stepsFailed := false

	steps := []struct {
		name string
		fn   func() error
	}{
		{"拉取代码", func() error {
			return dm.stepPullCode(deployment)
		}},
		{"编译项目", func() error {
			return dm.stepBuild(deployment)
		}},
		{"备份当前版本", func() error {
			return dm.stepBackup(deployment)
		}},
		{"停止旧服务", func() error {
			return dm.stepStopService(deployment)
		}},
		{"替换二进制文件", func() error {
			return dm.stepReplaceBinary(deployment)
		}},
		{"启动新服务", func() error {
			return dm.stepStartService(deployment)
		}},
		{"健康检查", func() error {
			return dm.stepHealthCheck(deployment)
		}},
	}

	for _, step := range steps {
		stepErr := step.fn()
		if stepErr != nil {
			stepsFailed = true
			break
		}
	}

	if stepsFailed {
		Error("部署失败，开始回滚: %s", deployment.ID)

		rollbackStep := common.NewStepRecord("自动回滚")
		deployment.Steps = append(deployment.Steps, rollbackStep)
		rollbackStep.Start()

		rollbackErr := dm.doRollback(deployment)
		if rollbackErr != nil {
			rollbackStep.Fail("", rollbackErr.Error())
			Error("回滚也失败了: %v", rollbackErr)
			deployment.Status = common.StatusFailed
			LogSummary(time.Since(deployment.StartTime), deployment.Steps)
			deployment.EndTime = time.Now()
			deployment.TotalDuration = deployment.EndTime.Sub(deployment.StartTime)
			return
		}

		rollbackStep.Complete("回滚成功")
		deployment.Status = common.StatusRolledBack
	} else {
		deployment.Status = common.StatusSuccess
		Info("部署成功: %s", deployment.ID)
	}

	deployment.EndTime = time.Now()
	deployment.TotalDuration = deployment.EndTime.Sub(deployment.StartTime)
	LogSummary(deployment.TotalDuration, deployment.Steps)
}

func (dm *DeployManager) stepPullCode(deployment *common.Deployment) error {
	step := common.NewStepRecord("拉取代码")
	deployment.Steps = append(deployment.Steps, step)
	step.Start()

	if deployment.Config.PullCommand == "" {
		step.Complete("跳过拉取代码（未配置）")
		LogStep("拉取代码", "跳过", 0, "未配置拉取命令", nil)
		return nil
	}

	Info("执行拉取命令: %s", deployment.Config.PullCommand)
	output, err := executeShell(deployment.Config.PullCommand)
	if err != nil {
		step.Fail(output, err.Error())
		LogStep("拉取代码", "失败", step.Duration, output, err)
		Error("拉取代码失败: %v", err)
		return err
	}

	step.Complete(output)
	LogStep("拉取代码", "成功", step.Duration, output, nil)
	Info("拉取代码成功")
	return nil
}

func (dm *DeployManager) stepBuild(deployment *common.Deployment) error {
	step := common.NewStepRecord("编译项目")
	deployment.Steps = append(deployment.Steps, step)
	step.Start()

	if deployment.Config.BuildCommand == "" {
		step.Complete("跳过编译（未配置）")
		LogStep("编译项目", "跳过", 0, "未配置编译命令", nil)
		return nil
	}

	Info("执行编译命令: %s", deployment.Config.BuildCommand)
	output, err := executeShell(deployment.Config.BuildCommand)
	if err != nil {
		step.Fail(output, err.Error())
		LogStep("编译项目", "失败", step.Duration, output, err)
		Error("编译失败: %v", err)
		return err
	}

	step.Complete(output)
	LogStep("编译项目", "成功", step.Duration, output, nil)
	Info("编译成功")
	return nil
}

func (dm *DeployManager) stepBackup(deployment *common.Deployment) error {
	step := common.NewStepRecord("备份当前版本")
	deployment.Steps = append(deployment.Steps, step)
	step.Start()

	backupPath, err := createBackup(deployment.Config.BinaryPath, deployment.Config.BackupDir)
	if err != nil {
		step.Fail("", err.Error())
		LogStep("备份当前版本", "失败", step.Duration, "", err)
		Error("备份失败: %v", err)
		return err
	}

	deployment.BackupPath = backupPath
	step.Complete("备份到: " + backupPath)
	LogStep("备份当前版本", "成功", step.Duration, "备份路径: "+backupPath, nil)
	Info("备份成功: %s", backupPath)
	return nil
}

func (dm *DeployManager) stepStopService(deployment *common.Deployment) error {
	step := common.NewStepRecord("停止旧服务")
	deployment.Steps = append(deployment.Steps, step)
	step.Start()

	err := stopService(deployment.Config.BinaryPath)
	if err != nil {
		step.Fail("", err.Error())
		LogStep("停止旧服务", "失败", step.Duration, "", err)
		Error("停止服务失败: %v", err)
		return err
	}

	step.Complete("服务已停止")
	LogStep("停止旧服务", "成功", step.Duration, "", nil)
	Info("服务已停止")
	return nil
}

func (dm *DeployManager) stepReplaceBinary(deployment *common.Deployment) error {
	step := common.NewStepRecord("替换二进制文件")
	deployment.Steps = append(deployment.Steps, step)
	step.Start()

	err := replaceBinary(deployment.Config.BinaryPath)
	if err != nil {
		step.Fail("", err.Error())
		LogStep("替换二进制文件", "失败", step.Duration, "", err)
		Error("替换二进制文件失败: %v", err)
		return err
	}

	step.Complete("二进制文件已替换")
	LogStep("替换二进制文件", "成功", step.Duration, "", nil)
	Info("二进制文件已替换")
	return nil
}

func (dm *DeployManager) stepStartService(deployment *common.Deployment) error {
	step := common.NewStepRecord("启动新服务")
	deployment.Steps = append(deployment.Steps, step)
	step.Start()

	err := startService(deployment.Config.BinaryPath)
	if err != nil {
		step.Fail("", err.Error())
		LogStep("启动新服务", "失败", step.Duration, "", err)
		Error("启动服务失败: %v", err)
		return err
	}

	step.Complete("服务已启动")
	LogStep("启动新服务", "成功", step.Duration, "", nil)
	Info("服务已启动")
	return nil
}

func (dm *DeployManager) stepHealthCheck(deployment *common.Deployment) error {
	step := common.NewStepRecord("健康检查")
	deployment.Steps = append(deployment.Steps, step)
	step.Start()

	if deployment.Config.HealthCheckURL == "" {
		step.Complete("跳过健康检查（未配置）")
		LogStep("健康检查", "跳过", 0, "未配置健康检查URL", nil)
		return nil
	}

	timeout := 5
	if deployment.Config.HealthCheckTimeout > 0 {
		timeout = deployment.Config.HealthCheckTimeout
	}

	Info("执行健康检查: URL=%s, 超时=%ds", deployment.Config.HealthCheckURL, timeout)
	err := healthCheck(deployment.Config.HealthCheckURL, timeout)
	if err != nil {
		step.Fail("", err.Error())
		LogStep("健康检查", "失败", step.Duration, "", err)
		Error("健康检查失败: %v", err)
		return err
	}

	step.Complete("健康检查通过")
	LogStep("健康检查", "成功", step.Duration, "", nil)
	Info("健康检查通过")
	return nil
}

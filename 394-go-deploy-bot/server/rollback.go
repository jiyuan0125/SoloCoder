package main

import (
	"deploybot/common"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func (dm *DeployManager) runRollback(deployment *common.Deployment) {
	deployment.Status = common.StatusRunning
	deployment.StartTime = time.Now()
	deployment.Steps = []*common.StepRecord{}

	Info("开始回滚任务: %s", deployment.ID)

	err := dm.doRollback(deployment)

	deployment.EndTime = time.Now()
	deployment.TotalDuration = deployment.EndTime.Sub(deployment.StartTime)

	if err != nil {
		deployment.Status = common.StatusFailed
		Error("回滚失败: %v", err)
		LogSummary(deployment.TotalDuration, deployment.Steps)
		return
	}

	deployment.Status = common.StatusSuccess
	Info("回滚成功: %s", deployment.ID)
	LogSummary(deployment.TotalDuration, deployment.Steps)
}

func (dm *DeployManager) doRollback(deployment *common.Deployment) error {
	if deployment.BackupPath == "" {
		lastBackup := deployment.Config.BackupDir
		if lastBackup == "" {
			return fmt.Errorf("没有指定备份路径，也未找到最近的备份")
		}
	}

	if _, err := os.Stat(deployment.BackupPath); os.IsNotExist(err) {
		return fmt.Errorf("备份文件不存在: %s", deployment.BackupPath)
	}

	stepStop := common.NewStepRecord("停止当前服务")
	deployment.Steps = append(deployment.Steps, stepStop)
	stepStop.Start()

	err := stopService(deployment.Config.BinaryPath)
	if err != nil {
		stepStop.Fail("", err.Error())
		LogStep("停止当前服务", "失败", stepStop.Duration, "", err)
		return err
	}
	stepStop.Complete("服务已停止")
	LogStep("停止当前服务", "成功", stepStop.Duration, "", nil)

	stepRestore := common.NewStepRecord("恢复备份版本")
	deployment.Steps = append(deployment.Steps, stepRestore)
	stepRestore.Start()

	err = restoreFromBackup(deployment.BackupPath, deployment.Config.BinaryPath)
	if err != nil {
		stepRestore.Fail("", err.Error())
		LogStep("恢复备份版本", "失败", stepRestore.Duration, "", err)
		return err
	}
	stepRestore.Complete(fmt.Sprintf("从 %s 恢复", deployment.BackupPath))
	LogStep("恢复备份版本", "成功", stepRestore.Duration, "备份路径: "+deployment.BackupPath, nil)

	stepStart := common.NewStepRecord("启动恢复后的服务")
	deployment.Steps = append(deployment.Steps, stepStart)
	stepStart.Start()

	err = startService(deployment.Config.BinaryPath)
	if err != nil {
		stepStart.Fail("", err.Error())
		LogStep("启动恢复后的服务", "失败", stepStart.Duration, "", err)
		return err
	}
	stepStart.Complete("服务已启动")
	LogStep("启动恢复后的服务", "成功", stepStart.Duration, "", nil)

	if deployment.Config.HealthCheckURL != "" {
		stepHC := common.NewStepRecord("回滚后健康检查")
		deployment.Steps = append(deployment.Steps, stepHC)
		stepHC.Start()

		timeout := 5
		if deployment.Config.HealthCheckTimeout > 0 {
			timeout = deployment.Config.HealthCheckTimeout
		}

		err = healthCheck(deployment.Config.HealthCheckURL, timeout)
		if err != nil {
			stepHC.Fail("", err.Error())
			LogStep("回滚后健康检查", "失败", stepHC.Duration, "", err)
			return err
		}
		stepHC.Complete("健康检查通过")
		LogStep("回滚后健康检查", "成功", stepHC.Duration, "", nil)
	}

	return nil
}

func createBackup(binaryPath string, backupDir string) (string, error) {
	if backupDir == "" {
		backupDir = "backups"
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("创建备份目录失败: %v", err)
	}

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		Warn("二进制文件不存在，跳过备份: %s", binaryPath)
		return "", nil
	}

	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s-%s", filepath.Base(binaryPath), timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	input, err := os.ReadFile(binaryPath)
	if err != nil {
		return "", fmt.Errorf("读取源文件失败: %v", err)
	}

	info, err := os.Stat(binaryPath)
	if err != nil {
		return "", fmt.Errorf("获取源文件信息失败: %v", err)
	}

	if err := os.WriteFile(backupPath, input, info.Mode()); err != nil {
		return "", fmt.Errorf("写入备份文件失败: %v", err)
	}

	Debug("备份完成: %s -> %s", binaryPath, backupPath)
	return backupPath, nil
}

func restoreFromBackup(backupPath string, binaryPath string) error {
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("备份文件不存在: %s", backupPath)
	}

	input, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %v", err)
	}

	info, err := os.Stat(backupPath)
	if err != nil {
		return fmt.Errorf("获取备份文件信息失败: %v", err)
	}

	if err := os.WriteFile(binaryPath, input, info.Mode()); err != nil {
		return fmt.Errorf("写入目标文件失败: %v", err)
	}

	Debug("恢复备份完成: %s -> %s", backupPath, binaryPath)
	return nil
}

func replaceBinary(binaryPath string) error {
	newPath := binaryPath + ".new"

	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		Debug("未找到新二进制文件 %s，跳过替换", newPath)
		return nil
	}

	if err := os.Rename(newPath, binaryPath); err != nil {
		return fmt.Errorf("替换二进制文件失败: %v", err)
	}

	Debug("替换二进制文件完成: %s", binaryPath)
	return nil
}

func executeShell(command string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", command)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return outputStr, fmt.Errorf("命令执行失败: %v, 输出: %s", err, outputStr)
	}

	return outputStr, nil
}
